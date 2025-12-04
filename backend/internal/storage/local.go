package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	basePath   string
	baseURL    string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Create subdirectories for organization
	subdirs := []string{"images", "files"}
	for _, subdir := range subdirs {
		path := filepath.Join(basePath, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create subdirectory %s: %w", subdir, err)
		}
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// Upload stores a file on the local filesystem
func (s *LocalStorage) Upload(filename string, contentType string, size int64, reader io.Reader) (*UploadResult, error) {
	// Generate unique file ID
	fileID, err := generateFileID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate file ID: %w", err)
	}

	// Sanitize filename and get extension
	sanitizedName := SanitizeFilename(filename)
	ext := filepath.Ext(sanitizedName)
	if ext == "" {
		ext = GetExtensionFromMimeType(contentType)
	}

	// Determine subdirectory based on content type
	subdir := "files"
	if AllowedImageTypes[contentType] {
		subdir = "images"
	}

	// Create date-based subdirectory for better organization
	dateDir := time.Now().Format("2006/01")
	fullDir := filepath.Join(s.basePath, subdir, dateDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create date directory: %w", err)
	}

	// Full file path
	storedFilename := fileID + ext
	fullPath := filepath.Join(fullDir, storedFilename)

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content with size limit
	written, err := io.Copy(file, io.LimitReader(reader, size+1))
	if err != nil {
		os.Remove(fullPath) // Cleanup on error
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Verify size matches
	if written > size {
		os.Remove(fullPath) // Cleanup on error
		return nil, ErrFileTooLarge
	}

	// Build the relative path for URL
	relativePath := filepath.Join(subdir, dateDir, storedFilename)
	url := fmt.Sprintf("%s/%s", s.baseURL, relativePath)

	return &UploadResult{
		ID:       fileID,
		Path:     relativePath, // Store relative path for database
		URL:      url,          // Full URL for immediate use
		Filename: sanitizedName,
		Size:     written,
		MimeType: contentType,
	}, nil
}

// UploadToFiles stores a file always in the files/ directory (for form submissions)
func (s *LocalStorage) UploadToFiles(filename string, contentType string, size int64, reader io.Reader) (*UploadResult, error) {
	// Generate short unique prefix for collision avoidance
	prefix, err := generateShortPrefix()
	if err != nil {
		return nil, fmt.Errorf("failed to generate prefix: %w", err)
	}

	// Sanitize filename - keep original name but make it safe
	sanitizedName := SanitizeFilename(filename)
	if sanitizedName == "" {
		// Fallback if filename is empty after sanitization
		ext := GetExtensionFromMimeType(contentType)
		sanitizedName = "file" + ext
	}

	// Always use "files" subdirectory for submission uploads
	subdir := "files"

	// Create date-based subdirectory for better organization
	dateDir := time.Now().Format("2006/01")
	fullDir := filepath.Join(s.basePath, subdir, dateDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create date directory: %w", err)
	}

	// Full file path - prefix + original filename for uniqueness while keeping recognizable name
	storedFilename := prefix + "_" + sanitizedName
	fullPath := filepath.Join(fullDir, storedFilename)

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content with size limit
	written, err := io.Copy(file, io.LimitReader(reader, size+1))
	if err != nil {
		os.Remove(fullPath) // Cleanup on error
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Verify size matches
	if written > size {
		os.Remove(fullPath) // Cleanup on error
		return nil, ErrFileTooLarge
	}

	// Build the relative path for URL
	relativePath := filepath.Join(subdir, dateDir, storedFilename)
	url := fmt.Sprintf("%s/%s", s.baseURL, relativePath)

	return &UploadResult{
		ID:       prefix,
		Path:     relativePath, // Store relative path for database
		URL:      url,          // Full URL for immediate use
		Filename: sanitizedName,
		Size:     written,
		MimeType: contentType,
	}, nil
}

// GetURLByPath returns the URL for a file given its relative path
func (s *LocalStorage) GetURLByPath(path string) (string, error) {
	// Check if the file exists
	fullPath := filepath.Join(s.basePath, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", ErrFileNotFound
	}
	return fmt.Sprintf("%s/%s", s.baseURL, path), nil
}

// GetFileByPath retrieves a file's content from local storage for streaming
func (s *LocalStorage) GetFileByPath(path string) (*FileContent, error) {
	fullPath := filepath.Join(s.basePath, path)

	// Check if file exists and get info
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}
	if err != nil {
		return nil, err
	}

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	// Detect content type from extension
	contentType := detectContentTypeFromPath(path)

	return &FileContent{
		Reader:      file,
		ContentType: contentType,
		Size:        info.Size(),
	}, nil
}

// detectContentTypeFromPath returns MIME type based on file extension
func detectContentTypeFromPath(path string) string {
	ext := filepath.Ext(path)
	types := map[string]string{
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

	if ct, ok := types[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// GetURL returns the URL for accessing a file
func (s *LocalStorage) GetURL(fileID string) (string, error) {
	// For local storage, we need to find the file
	// This is a simple implementation - in production you might want to store metadata in DB

	// Search in both subdirectories
	for _, subdir := range []string{"images", "files"} {
		pattern := filepath.Join(s.basePath, subdir, "*", "*", fileID+"*")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		if len(matches) > 0 {
			// Extract relative path
			relPath, err := filepath.Rel(s.basePath, matches[0])
			if err != nil {
				continue
			}
			return fmt.Sprintf("%s/%s", s.baseURL, relPath), nil
		}
	}

	return "", ErrFileNotFound
}

// Delete removes a file from local storage
func (s *LocalStorage) Delete(fileID string) error {
	// Search for the file
	for _, subdir := range []string{"images", "files"} {
		pattern := filepath.Join(s.basePath, subdir, "*", "*", fileID+"*")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			if err := os.Remove(match); err != nil {
				return fmt.Errorf("failed to delete file: %w", err)
			}
		}
		if len(matches) > 0 {
			return nil
		}
	}

	return ErrFileNotFound
}

// Type returns the storage type
func (s *LocalStorage) Type() StorageType {
	return StorageTypeLocal
}

// generateFileID creates a random file ID
func generateFileID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateShortPrefix creates a short random prefix (8 chars) for filename uniqueness
func generateShortPrefix() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
