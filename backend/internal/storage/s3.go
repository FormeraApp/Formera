package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage implements Storage interface for AWS S3
type S3Storage struct {
	client          *s3.Client
	presignClient   *s3.PresignClient
	bucket          string
	region          string
	prefix          string
	presignDuration time.Duration
}

// S3Config contains configuration for S3 storage
type S3Config struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string        // Optional: for S3-compatible services like MinIO
	Prefix          string        // Optional: prefix for all stored files
	PresignDuration time.Duration
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	// Build AWS config
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with optional custom endpoint
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // Required for MinIO and some S3-compatible services
		}
	})

	presignDuration := cfg.PresignDuration
	if presignDuration == 0 {
		presignDuration = 1 * time.Hour // Default 1 hour
	}

	return &S3Storage{
		client:          client,
		presignClient:   s3.NewPresignClient(client),
		bucket:          cfg.Bucket,
		region:          cfg.Region,
		prefix:          cfg.Prefix,
		presignDuration: presignDuration,
	}, nil
}

// Upload stores a file in S3
func (s *S3Storage) Upload(filename string, contentType string, size int64, reader io.Reader) (*UploadResult, error) {
	ctx := context.TODO()

	// Generate unique file ID
	fileID, err := s.generateFileID()
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

	// Create the relative path (without prefix - that's S3-specific)
	dateDir := time.Now().Format("2006/01")
	storedFilename := fileID + ext
	relativePath := fmt.Sprintf("%s/%s/%s", subdir, dateDir, storedFilename)

	// Create the full S3 key (with prefix)
	key := s.prefix + relativePath

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate presigned URL for immediate use
	url, err := s.getPresignedURL(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &UploadResult{
		ID:       fileID,
		Path:     relativePath, // Store relative path for database
		URL:      url,          // Presigned URL for immediate use
		Filename: sanitizedName,
		Size:     size,
		MimeType: contentType,
	}, nil
}

// UploadToFiles stores a file always in the files/ directory (for form submissions)
func (s *S3Storage) UploadToFiles(filename string, contentType string, size int64, reader io.Reader) (*UploadResult, error) {
	ctx := context.TODO()

	// Generate short unique prefix for collision avoidance
	prefix, err := s.generateShortPrefix()
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

	// Create the relative path (without prefix - that's S3-specific)
	dateDir := time.Now().Format("2006/01")
	storedFilename := prefix + "_" + sanitizedName
	relativePath := fmt.Sprintf("%s/%s/%s", subdir, dateDir, storedFilename)

	// Create the full S3 key (with prefix)
	key := s.prefix + relativePath

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate presigned URL for immediate use
	url, err := s.getPresignedURL(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &UploadResult{
		ID:       prefix,
		Path:     relativePath, // Store relative path for database
		URL:      url,          // Presigned URL for immediate use
		Filename: sanitizedName,
		Size:     size,
		MimeType: contentType,
	}, nil
}

// GetURLByPath returns a presigned URL for a file given its relative path
func (s *S3Storage) GetURLByPath(path string) (string, error) {
	// Build the full S3 key by adding our prefix
	key := s.prefix + path
	return s.getPresignedURL(key)
}

// GetFileByPath retrieves a file's content from S3 for streaming/proxying
func (s *S3Storage) GetFileByPath(path string) (*FileContent, error) {
	ctx := context.TODO()

	// Build the full S3 key by adding our prefix
	key := s.prefix + path

	// Get the object from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, ErrFileNotFound
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	var size int64
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return &FileContent{
		Reader:      result.Body,
		ContentType: contentType,
		Size:        size,
	}, nil
}

// GetURL returns a presigned URL for accessing a file
func (s *S3Storage) GetURL(fileID string) (string, error) {
	ctx := context.TODO()

	// Search for the file in S3
	// We need to list objects with the prefix to find the full key
	for _, subdir := range []string{"images", "files"} {
		prefix := fmt.Sprintf("%s%s/", s.prefix, subdir)

		paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
			Bucket: aws.String(s.bucket),
			Prefix: aws.String(prefix),
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, obj := range page.Contents {
				if obj.Key != nil && filepath.Base(*obj.Key)[:32] == fileID {
					return s.getPresignedURL(*obj.Key)
				}
			}
		}
	}

	return "", ErrFileNotFound
}

// Delete removes a file from S3
func (s *S3Storage) Delete(fileID string) error {
	ctx := context.TODO()

	// Find and delete the file
	for _, subdir := range []string{"images", "files"} {
		prefix := fmt.Sprintf("%s%s/", s.prefix, subdir)

		paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
			Bucket: aws.String(s.bucket),
			Prefix: aws.String(prefix),
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, obj := range page.Contents {
				if obj.Key != nil && filepath.Base(*obj.Key)[:32] == fileID {
					_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
						Bucket: aws.String(s.bucket),
						Key:    obj.Key,
					})
					if err != nil {
						return fmt.Errorf("failed to delete from S3: %w", err)
					}
					return nil
				}
			}
		}
	}

	return ErrFileNotFound
}

// Type returns the storage type
func (s *S3Storage) Type() StorageType {
	return StorageTypeS3
}

// getPresignedURL generates a presigned URL for an S3 object
func (s *S3Storage) getPresignedURL(key string) (string, error) {
	presignedReq, err := s.presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.presignDuration
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign request: %w", err)
	}

	return presignedReq.URL, nil
}

// generateFileID creates a random file ID
func (s *S3Storage) generateFileID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateShortPrefix creates a short random prefix (8 chars) for filename uniqueness
func (s *S3Storage) generateShortPrefix() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
