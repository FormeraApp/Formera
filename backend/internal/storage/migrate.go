package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MigrationResult contains statistics about the migration
type MigrationResult struct {
	TotalFiles     int
	MigratedFiles  int
	SkippedFiles   int
	FailedFiles    int
	TotalBytes     int64
	MigratedBytes  int64
	Errors         []string
	Duration       time.Duration
}

const migrationMarkerFile = ".migration_complete"

// IsMigrationComplete checks if migration has already been completed
func IsMigrationComplete(localPath string) bool {
	markerPath := filepath.Join(localPath, migrationMarkerFile)
	_, err := os.Stat(markerPath)
	return err == nil
}

// MarkMigrationComplete creates a marker file indicating migration is done
func MarkMigrationComplete(localPath string) error {
	markerPath := filepath.Join(localPath, migrationMarkerFile)
	content := fmt.Sprintf("Migration completed at %s\n", time.Now().Format(time.RFC3339))
	return os.WriteFile(markerPath, []byte(content), 0644)
}

// MigrateLocalToS3 migrates all files from local storage to S3
// It preserves the directory structure and skips files that already exist in S3
func MigrateLocalToS3(localPath string, s3Storage *S3Storage, deleteAfterMigrate bool) (*MigrationResult, error) {
	startTime := time.Now()
	result := &MigrationResult{
		Errors: []string{},
	}

	// Check if local path exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		log.Printf("Migration: Local storage path does not exist: %s", localPath)
		return result, nil
	}

	// Check if migration was already completed
	if IsMigrationComplete(localPath) {
		log.Println("Migration: Already completed (marker file found), skipping")
		return result, nil
	}

	log.Printf("Migration: Starting migration from %s to S3 bucket %s", localPath, s3Storage.bucket)

	// Walk through all files in local storage
	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error accessing %s: %v", path, err))
			return nil // Continue with other files
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip migration marker file
		if info.Name() == migrationMarkerFile {
			return nil
		}

		result.TotalFiles++
		result.TotalBytes += info.Size()

		// Get relative path for S3 key
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error getting relative path for %s: %v", path, err))
			result.FailedFiles++
			return nil
		}

		// Convert to forward slashes for S3
		s3Key := s3Storage.prefix + strings.ReplaceAll(relPath, "\\", "/")

		// Check if file already exists in S3
		exists, err := s3Storage.objectExists(s3Key)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error checking S3 for %s: %v", s3Key, err))
			result.FailedFiles++
			return nil
		}

		if exists {
			log.Printf("Migration: Skipping %s (already exists in S3)", relPath)
			result.SkippedFiles++
			return nil
		}

		// Upload file to S3
		if err := uploadFileToS3(path, s3Key, s3Storage); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error uploading %s: %v", relPath, err))
			result.FailedFiles++
			return nil
		}

		log.Printf("Migration: Uploaded %s (%d bytes)", relPath, info.Size())
		result.MigratedFiles++
		result.MigratedBytes += info.Size()

		// Optionally delete local file after successful migration
		if deleteAfterMigrate {
			if err := os.Remove(path); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error deleting local file %s: %v", path, err))
			}
		}

		return nil
	})

	result.Duration = time.Since(startTime)

	if err != nil {
		return result, fmt.Errorf("migration walk error: %w", err)
	}

	// Clean up empty directories if deleteAfterMigrate is true
	if deleteAfterMigrate && result.MigratedFiles > 0 {
		cleanupEmptyDirs(localPath)
	}

	log.Printf("Migration completed: %d files migrated, %d skipped, %d failed (%.2f MB in %v)",
		result.MigratedFiles, result.SkippedFiles, result.FailedFiles,
		float64(result.MigratedBytes)/(1024*1024), result.Duration)

	// Create marker file to prevent re-running migration on next startup
	if result.FailedFiles == 0 && (result.MigratedFiles > 0 || result.SkippedFiles > 0) {
		if err := MarkMigrationComplete(localPath); err != nil {
			log.Printf("Warning: Could not create migration marker file: %v", err)
		} else {
			log.Println("Migration marker file created - migration will be skipped on future startups")
		}
	}

	return result, nil
}

// objectExists checks if an object exists in S3
func (s *S3Storage) objectExists(key string) (bool, error) {
	_, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// uploadFileToS3 uploads a single file to S3
func uploadFileToS3(localPath, s3Key string, s3Storage *S3Storage) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info for size
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Detect content type
	contentType := detectMimeType(localPath)

	// Upload to S3
	_, err = s3Storage.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(s3Storage.bucket),
		Key:           aws.String(s3Key),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(info.Size()),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// detectMimeType detects MIME type from file extension
func detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".json": "application/json",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

// cleanupEmptyDirs removes empty directories recursively
func cleanupEmptyDirs(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == root {
			return nil
		}

		// Try to remove directory (will fail if not empty)
		os.Remove(path)
		return nil
	})
}

// MigrateS3ToLocal migrates files from S3 back to local storage
// Useful for switching back to local storage or backup
func MigrateS3ToLocal(s3Storage *S3Storage, localPath string, deleteAfterMigrate bool) (*MigrationResult, error) {
	startTime := time.Now()
	result := &MigrationResult{
		Errors: []string{},
	}

	ctx := context.TODO()

	log.Printf("Migration: Starting migration from S3 bucket %s to %s", s3Storage.bucket, localPath)

	// Ensure local path exists
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local directory: %w", err)
	}

	// List all objects in S3
	paginator := s3.NewListObjectsV2Paginator(s3Storage.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Storage.bucket),
		Prefix: aws.String(s3Storage.prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return result, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}

			result.TotalFiles++
			result.TotalBytes += *obj.Size

			// Get relative path (remove prefix)
			relPath := strings.TrimPrefix(*obj.Key, s3Storage.prefix)
			localFilePath := filepath.Join(localPath, relPath)

			// Check if file already exists locally
			if _, err := os.Stat(localFilePath); err == nil {
				log.Printf("Migration: Skipping %s (already exists locally)", relPath)
				result.SkippedFiles++
				continue
			}

			// Download from S3
			if err := downloadFromS3(*obj.Key, localFilePath, s3Storage); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error downloading %s: %v", relPath, err))
				result.FailedFiles++
				continue
			}

			log.Printf("Migration: Downloaded %s (%d bytes)", relPath, *obj.Size)
			result.MigratedFiles++
			result.MigratedBytes += *obj.Size

			// Optionally delete from S3 after successful migration
			if deleteAfterMigrate {
				_, err := s3Storage.client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket: aws.String(s3Storage.bucket),
					Key:    obj.Key,
				})
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Error deleting S3 object %s: %v", *obj.Key, err))
				}
			}
		}
	}

	result.Duration = time.Since(startTime)

	log.Printf("Migration completed: %d files migrated, %d skipped, %d failed (%.2f MB in %v)",
		result.MigratedFiles, result.SkippedFiles, result.FailedFiles,
		float64(result.MigratedBytes)/(1024*1024), result.Duration)

	return result, nil
}

// downloadFromS3 downloads a single file from S3 to local storage
func downloadFromS3(s3Key, localPath string, s3Storage *S3Storage) error {
	ctx := context.TODO()

	// Create directory if needed
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get object from S3
	resp, err := s3Storage.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3Storage.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer resp.Body.Close()

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(localPath) // Cleanup on error
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
