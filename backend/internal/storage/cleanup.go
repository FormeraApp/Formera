package storage

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// CleanupConfig contains configuration for the file cleanup scheduler
type CleanupConfig struct {
	// Enabled determines if cleanup is active
	Enabled bool
	// Interval between cleanup runs
	Interval time.Duration
	// MinAge is the minimum age of orphaned files before deletion
	MinAge time.Duration
	// DryRun if true, only logs what would be deleted without actually deleting
	DryRun bool
}

// DefaultCleanupConfig returns sensible defaults
func DefaultCleanupConfig() CleanupConfig {
	return CleanupConfig{
		Enabled:  true,
		Interval: 24 * time.Hour, // Run once per day
		MinAge:   7 * 24 * time.Hour, // Only delete files orphaned for 7+ days
		DryRun:   false,
	}
}

// CleanupScheduler manages periodic cleanup of orphaned files
type CleanupScheduler struct {
	storage Storage
	db      *gorm.DB
	config  CleanupConfig
	stopCh  chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	running bool
}

// CleanupResult contains the results of a cleanup run
type CleanupResult struct {
	ScannedFiles  int
	DeletedFiles  int
	DeletedBytes  int64
	Errors        []string
	Duration      time.Duration
}

// NewCleanupScheduler creates a new cleanup scheduler
func NewCleanupScheduler(storage Storage, db *gorm.DB, config CleanupConfig) *CleanupScheduler {
	return &CleanupScheduler{
		storage: storage,
		db:      db,
		config:  config,
		stopCh:  make(chan struct{}),
	}
}

// Start begins the cleanup scheduler
func (c *CleanupScheduler) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	if !c.config.Enabled {
		log.Println("File cleanup scheduler is disabled")
		return
	}

	log.Printf("Starting file cleanup scheduler (interval: %v, min age: %v, dry run: %v)",
		c.config.Interval, c.config.MinAge, c.config.DryRun)

	c.wg.Add(1)
	go c.run()
}

// Stop stops the cleanup scheduler
func (c *CleanupScheduler) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.running = false
	c.mu.Unlock()

	close(c.stopCh)
	c.wg.Wait()
	log.Println("File cleanup scheduler stopped")
}

// run is the main scheduler loop
func (c *CleanupScheduler) run() {
	defer c.wg.Done()

	// Run immediately on start
	result := c.RunCleanup()
	c.logResult(result)

	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result := c.RunCleanup()
			c.logResult(result)
		case <-c.stopCh:
			return
		}
	}
}

// RunCleanup performs a single cleanup run
func (c *CleanupScheduler) RunCleanup() *CleanupResult {
	start := time.Now()
	result := &CleanupResult{}

	// Get all file records
	var files []FileRecord
	if err := c.db.Find(&files).Error; err != nil {
		result.Errors = append(result.Errors, "Failed to query file records: "+err.Error())
		result.Duration = time.Since(start)
		return result
	}

	result.ScannedFiles = len(files)
	cutoffTime := time.Now().Add(-c.config.MinAge)

	for _, file := range files {
		// Skip files that are too new
		if file.CreatedAt.After(cutoffTime) {
			continue
		}

		// Check if file is orphaned
		orphaned, err := file.IsOrphaned(c.db)
		if err != nil {
			result.Errors = append(result.Errors, "Error checking file "+file.ID+": "+err.Error())
			continue
		}

		if orphaned {
			if c.config.DryRun {
				log.Printf("[DRY RUN] Would delete orphaned file: %s (%s, %d bytes)",
					file.ID, file.Filename, file.Size)
				result.DeletedFiles++
				result.DeletedBytes += file.Size
			} else {
				// Delete from storage
				if err := c.storage.Delete(file.ID); err != nil && err != ErrFileNotFound {
					result.Errors = append(result.Errors, "Failed to delete file "+file.ID+": "+err.Error())
					continue
				}

				// Delete record from database
				if err := c.db.Delete(&file).Error; err != nil {
					result.Errors = append(result.Errors, "Failed to delete record "+file.ID+": "+err.Error())
					continue
				}

				result.DeletedFiles++
				result.DeletedBytes += file.Size
			}
		}
	}

	result.Duration = time.Since(start)
	return result
}

func (c *CleanupScheduler) logResult(result *CleanupResult) {
	if result.DeletedFiles > 0 || len(result.Errors) > 0 {
		prefix := ""
		if c.config.DryRun {
			prefix = "[DRY RUN] "
		}
		log.Printf("%sCleanup completed: scanned %d files, deleted %d files (%.2f MB) in %v",
			prefix, result.ScannedFiles, result.DeletedFiles,
			float64(result.DeletedBytes)/(1024*1024), result.Duration)

		if len(result.Errors) > 0 {
			log.Printf("Cleanup errors (%d):", len(result.Errors))
			for _, err := range result.Errors {
				log.Printf("  - %s", err)
			}
		}
	}
}

// FileRecord represents a tracked file (mirrors the model)
type FileRecord struct {
	ID        string    `gorm:"primaryKey"`
	UserID    string    `gorm:"index"`
	Filename  string
	MimeType  string
	Size      int64
	Path      string // Relative path (e.g., "images/2025/12/abc123.png")
	URL       string // Deprecated: kept for backward compatibility
	CreatedAt time.Time
}

// IsOrphaned checks if this file is referenced anywhere in the database
func (f *FileRecord) IsOrphaned(db *gorm.DB) (bool, error) {
	var count int64

	// Check form settings (design background images)
	err := db.Table("forms").
		Where("settings LIKE ?", "%"+f.ID+"%").
		Or("settings LIKE ?", "%"+f.URL+"%").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	// Check form fields (image fields, file references)
	err = db.Table("forms").
		Where("fields LIKE ?", "%"+f.ID+"%").
		Or("fields LIKE ?", "%"+f.URL+"%").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	// Check submissions (file uploads in submissions)
	err = db.Table("submissions").
		Where("data LIKE ?", "%"+f.ID+"%").
		Or("data LIKE ?", "%"+f.URL+"%").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	return true, nil
}
