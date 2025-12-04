package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"formera/internal/database"
	"formera/internal/storage"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	storage     storage.Storage
	rateLimiter *rateLimiter
}

// rateLimiter implements a simple token bucket rate limiter per user
type rateLimiter struct {
	mu       sync.Mutex
	limits   map[uint]*userLimit
	maxUploads int           // Maximum uploads per window
	window     time.Duration // Time window
}

type userLimit struct {
	count     int
	resetTime time.Time
}

func newRateLimiter(maxUploads int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limits:     make(map[uint]*userLimit),
		maxUploads: maxUploads,
		window:     window,
	}
}

func (r *rateLimiter) allow(userID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	limit, exists := r.limits[userID]
	if !exists || now.After(limit.resetTime) {
		r.limits[userID] = &userLimit{
			count:     1,
			resetTime: now.Add(r.window),
		}
		return true
	}

	if limit.count >= r.maxUploads {
		return false
	}

	limit.count++
	return true
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(store storage.Storage) *UploadHandler {
	return &UploadHandler{
		storage: store,
		// Rate limit: 20 uploads per 5 minutes per user
		rateLimiter: newRateLimiter(20, 5*time.Minute),
	}
}

// UploadImage godoc
// @Summary      Upload image
// @Description  Upload an image file (for form design backgrounds)
// @Tags         Uploads
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Image file"
// @Success      200 {object} storage.UploadResult
// @Failure      400 {object} ErrorResponse "Invalid file"
// @Failure      401 {object} ErrorResponse
// @Failure      429 {object} ErrorResponse "Rate limit exceeded"
// @Security     BearerAuth
// @Router       /uploads/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// Get authenticated user
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Rate limiting (hash user ID string to uint for rate limiter)
	if !h.rateLimiter.allow(hashIP(userID)) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Too many uploads. Please wait a few minutes.",
		})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Get content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from file extension
		contentType = detectContentType(header.Filename)
	}

	// Validate image upload
	if err := storage.ValidateImageUpload(contentType, header.Size); err != nil {
		switch err {
		case storage.ErrInvalidFileType:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid file type. Allowed: JPEG, PNG, GIF, WebP, SVG",
			})
		case storage.ErrFileTooLarge:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File too large. Maximum size: %d MB", storage.MaxImageSize/(1024*1024)),
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// Additional security: Read and verify magic bytes
	if !verifyImageMagicBytes(file, contentType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File content does not match declared type",
		})
		return
	}

	// Reset file reader position after magic byte check
	_, _ = file.Seek(0, 0)

	// Upload to storage
	result, err := h.storage.Upload(header.Filename, contentType, header.Size, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	// Track file in database for cleanup
	fileRecord := storage.FileRecord{
		ID:        result.ID,
		UserID:    userID,
		Filename:  result.Filename,
		MimeType:  result.MimeType,
		Size:      result.Size,
		Path:      result.Path,
		URL:       result.URL, // Kept for backward compatibility
		CreatedAt: time.Now(),
	}
	database.DB.Create(&fileRecord)

	c.JSON(http.StatusOK, result)
}

// UploadFile godoc
// @Summary      Upload file
// @Description  Upload a file (for form submissions)
// @Tags         Uploads
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "File to upload"
// @Success      200 {object} storage.UploadResult
// @Failure      400 {object} ErrorResponse "Invalid file"
// @Failure      429 {object} ErrorResponse "Rate limit exceeded"
// @Router       /public/upload [post]
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// Get authenticated user (or allow anonymous for public form submissions)
	userID := c.GetString("user_id")

	// For public uploads, use IP-based rate limiting
	var rateLimitKey uint
	if userID != "" {
		rateLimitKey = hashIP(userID)
	} else {
		// Use hash of IP for anonymous uploads
		rateLimitKey = hashIP(c.ClientIP())
	}

	// Rate limiting
	if !h.rateLimiter.allow(rateLimitKey) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Too many uploads. Please wait a few minutes.",
		})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Get content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(header.Filename)
	}

	// Validate file upload
	if err := storage.ValidateFileUpload(contentType, header.Size); err != nil {
		switch err {
		case storage.ErrInvalidFileType:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid file type",
			})
		case storage.ErrFileTooLarge:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File too large. Maximum size: %d MB", storage.MaxFileSize/(1024*1024)),
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// Upload to storage
	result, err := h.storage.Upload(header.Filename, contentType, header.Size, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	// Track file in database for cleanup
	fileRecord := storage.FileRecord{
		ID:        result.ID,
		UserID:    userID,
		Filename:  result.Filename,
		MimeType:  result.MimeType,
		Size:      result.Size,
		Path:      result.Path,
		URL:       result.URL, // Kept for backward compatibility
		CreatedAt: time.Now(),
	}
	database.DB.Create(&fileRecord)

	c.JSON(http.StatusOK, result)
}

// GetFile godoc
// @Summary      Get file
// @Description  Serve a file by path (streams from storage)
// @Tags         Files
// @Produce      octet-stream
// @Param        path path string true "File path"
// @Success      200 {file} file "File content"
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /files/{path} [get]
func (h *UploadHandler) GetFile(c *gin.Context) {
	// Get the file path from URL parameter (e.g., "images/2025/12/abc123.png")
	filePath := c.Param("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File path required"})
		return
	}

	// Remove leading slash if present
	if len(filePath) > 0 && filePath[0] == '/' {
		filePath = filePath[1:]
	}

	// Security: Sanitize path to prevent path traversal attacks
	filePath = storage.SanitizePath(filePath)
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path"})
		return
	}

	// Get file content for streaming
	fileContent, err := h.storage.GetFileByPath(filePath)
	if err != nil {
		if err == storage.ErrFileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file"})
		return
	}
	defer fileContent.Reader.Close()

	// Set headers for caching (1 year for immutable content-addressed files)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")

	// Use Gin's DataFromReader for efficient streaming
	c.DataFromReader(http.StatusOK, fileContent.Size, fileContent.ContentType, fileContent.Reader, nil)
}

// DeleteFile godoc
// @Summary      Delete file
// @Description  Delete an uploaded file
// @Tags         Uploads
// @Produce      json
// @Param        id path string true "File ID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /uploads/{id} [delete]
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	// Only authenticated users can delete
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID required"})
		return
	}

	// Validate file ID format (should be 32 hex characters)
	if len(fileID) != 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	for _, ch := range fileID {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
			return
		}
	}

	// Check file ownership - users can only delete their own files (admins can delete any)
	isAdmin := c.GetString("role") == "admin"
	var fileRecord storage.FileRecord

	if isAdmin {
		// Admins can delete any file
		if err := database.DB.Where("id = ?", fileID).First(&fileRecord).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
	} else {
		// Regular users can only delete their own files
		if err := database.DB.Where("id = ? AND user_id = ?", fileID, userID).First(&fileRecord).Error; err != nil {
			// Check if file exists at all (for proper error message)
			var exists storage.FileRecord
			if database.DB.Where("id = ?", fileID).First(&exists).Error == nil {
				// File exists but belongs to another user
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
	}

	if err := h.storage.Delete(fileID); err != nil {
		if err == storage.ErrFileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	// Remove file record from database
	database.DB.Delete(&storage.FileRecord{}, "id = ?", fileID)

	c.JSON(http.StatusOK, gin.H{"message": "File deleted"})
}

// detectContentType guesses content type from filename extension
func detectContentType(filename string) string {
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

	for ext, mime := range types {
		if len(filename) > len(ext) && filename[len(filename)-len(ext):] == ext {
			return mime
		}
	}

	return "application/octet-stream"
}

// verifyImageMagicBytes checks if file content matches declared MIME type
func verifyImageMagicBytes(file interface{ Read([]byte) (int, error) }, contentType string) bool {
	// Read first 16 bytes for magic number verification
	header := make([]byte, 16)
	n, err := file.Read(header)
	if err != nil || n < 4 {
		return false
	}

	switch contentType {
	case "image/jpeg":
		// JPEG magic: FF D8 FF
		return header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF

	case "image/png":
		// PNG magic: 89 50 4E 47 0D 0A 1A 0A
		return header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47

	case "image/gif":
		// GIF magic: GIF87a or GIF89a
		return header[0] == 'G' && header[1] == 'I' && header[2] == 'F' && header[3] == '8'

	case "image/webp":
		// WebP magic: RIFF....WEBP
		return header[0] == 'R' && header[1] == 'I' && header[2] == 'F' && header[3] == 'F' &&
			n >= 12 && header[8] == 'W' && header[9] == 'E' && header[10] == 'B' && header[11] == 'P'

	case "image/svg+xml":
		// SVG: Check for XML declaration or <svg tag
		str := string(header[:n])
		return len(str) >= 5 && (str[:5] == "<?xml" || str[:4] == "<svg" || str[:5] == "<!DOC")

	default:
		return false
	}
}

// hashIP creates a simple hash of an IP address for rate limiting
func hashIP(ip string) uint {
	var hash uint = 5381
	for i := 0; i < len(ip); i++ {
		hash = ((hash << 5) + hash) + uint(ip[i])
	}
	return hash
}
