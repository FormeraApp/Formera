package storage

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
)

// Common errors
var (
	ErrInvalidFileType = errors.New("invalid file type")
	ErrFileTooLarge    = errors.New("file too large")
	ErrUploadFailed    = errors.New("upload failed")
	ErrFileNotFound    = errors.New("file not found")
)

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
)

// UploadResult contains information about an uploaded file
type UploadResult struct {
	ID       string `json:"id"`
	Path     string `json:"path"`     // Relative path (e.g., "images/2025/12/abc123.png")
	URL      string `json:"url"`      // Full URL for immediate use
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

// FileContent represents the content of a file for streaming
type FileContent struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
}

// Storage defines the interface for file storage backends
type Storage interface {
	// Upload stores a file and returns the result
	Upload(filename string, contentType string, size int64, reader io.Reader) (*UploadResult, error)

	// GetURL returns the URL for accessing a file by ID (searches for file)
	GetURL(fileID string) (string, error)

	// GetURLByPath returns the URL for accessing a file by its relative path
	GetURLByPath(path string) (string, error)

	// GetFileByPath retrieves a file's content for streaming/proxying
	GetFileByPath(path string) (*FileContent, error)

	// Delete removes a file from storage
	Delete(fileID string) error

	// Type returns the storage type
	Type() StorageType
}

// AllowedImageTypes contains permitted MIME types for image uploads
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/svg+xml": true,
}

// AllowedFileTypes contains permitted MIME types for general file uploads
var AllowedFileTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"image/svg+xml":   true,
	"application/pdf": true,
	"text/plain":      true,
	"text/csv":        true,
	"application/json": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}

// MaxImageSize is the maximum allowed size for image uploads (5MB)
const MaxImageSize = 5 * 1024 * 1024

// MaxFileSize is the maximum allowed size for general file uploads (25MB)
const MaxFileSize = 25 * 1024 * 1024

// ValidateImageUpload checks if a file is a valid image upload
func ValidateImageUpload(contentType string, size int64) error {
	if !AllowedImageTypes[contentType] {
		return ErrInvalidFileType
	}
	if size > MaxImageSize {
		return ErrFileTooLarge
	}
	return nil
}

// ValidateFileUpload checks if a file is a valid general file upload
func ValidateFileUpload(contentType string, size int64) error {
	if !AllowedFileTypes[contentType] {
		return ErrInvalidFileType
	}
	if size > MaxFileSize {
		return ErrFileTooLarge
	}
	return nil
}

// SanitizePath sanitizes a file path to prevent path traversal attacks
// Returns empty string if the path is invalid or attempts traversal
func SanitizePath(path string) string {
	// Remove null bytes
	path = strings.ReplaceAll(path, "\x00", "")

	// Clean the path to resolve . and .. components
	cleaned := filepath.Clean(path)

	// Reject absolute paths
	if filepath.IsAbs(cleaned) {
		return ""
	}

	// Reject paths that try to traverse upward
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") || strings.HasSuffix(cleaned, "/..") {
		return ""
	}

	// Only allow paths starting with known subdirectories
	if !strings.HasPrefix(cleaned, "images/") && !strings.HasPrefix(cleaned, "files/") {
		return ""
	}

	return cleaned
}

// SanitizeFilename removes potentially dangerous characters from filenames
func SanitizeFilename(filename string) string {
	// Get base name only (no path traversal)
	filename = filepath.Base(filename)

	// Remove any null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Replace problematic characters
	replacer := strings.NewReplacer(
		"..", "",
		"/", "",
		"\\", "",
		"<", "",
		">", "",
		":", "",
		"\"", "",
		"|", "",
		"?", "",
		"*", "",
	)

	return replacer.Replace(filename)
}

// GetExtensionFromMimeType returns a file extension for a given MIME type
func GetExtensionFromMimeType(mimeType string) string {
	extensions := map[string]string{
		"image/jpeg":    ".jpg",
		"image/png":     ".png",
		"image/gif":     ".gif",
		"image/webp":    ".webp",
		"image/svg+xml": ".svg",
		"application/pdf": ".pdf",
		"text/plain":    ".txt",
		"text/csv":      ".csv",
		"application/json": ".json",
		"application/msword": ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.ms-excel": ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
	}

	if ext, ok := extensions[mimeType]; ok {
		return ext
	}
	return ""
}
