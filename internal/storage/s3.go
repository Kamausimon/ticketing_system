package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// StorageService handles file storage with S3 primary and local fallback
type StorageService struct {
	s3Client  *s3.Client
	bucket    string
	region    string
	s3Enabled bool
	localPath string
	publicURL string
	ctx       context.Context
}

// UploadResult contains the result of an upload operation
type UploadResult struct {
	URL      string
	Key      string
	Size     int64
	MimeType string
	Backend  string // "s3" or "local"
}

// NewStorageService creates a new storage service
func NewStorageService(accessKey, secretKey, region, bucket, localPath, publicURL string) (*StorageService, error) {
	ctx := context.Background()
	service := &StorageService{
		bucket:    bucket,
		region:    region,
		localPath: localPath,
		publicURL: publicURL,
		s3Enabled: false,
		ctx:       ctx,
	}

	// Try to initialize S3 if credentials provided
	if accessKey != "" && secretKey != "" && region != "" && bucket != "" {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			)),
		)
		if err != nil {
			fmt.Printf("⚠️  Failed to load S3 config: %v (using local storage)\n", err)
		} else {
			service.s3Client = s3.NewFromConfig(cfg)

			// Test S3 connection
			_, err = service.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
				Bucket: aws.String(bucket),
			})
			if err != nil {
				fmt.Printf("⚠️  S3 bucket not accessible: %v (using local storage)\n", err)
			} else {
				service.s3Enabled = true
				fmt.Println("✅ S3 storage initialized successfully")
			}
		}
	}

	// Ensure local storage directory exists
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}

	if !service.s3Enabled {
		fmt.Println("✅ Local file storage initialized (S3 fallback)")
	}

	return service, nil
}

// UploadFile uploads a file to S3 or local storage
func (s *StorageService) UploadFile(file multipart.File, fileHeader *multipart.FileHeader, folder string) (*UploadResult, error) {
	// Read file content
	buffer := bytes.NewBuffer(nil)
	size, err := io.Copy(buffer, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	key := filepath.Join(folder, filename)

	// Try S3 first if enabled
	if s.s3Enabled {
		result, err := s.uploadToS3(buffer.Bytes(), key, fileHeader.Header.Get("Content-Type"))
		if err == nil {
			return result, nil
		}
		fmt.Printf("⚠️  S3 upload failed: %v (using local storage)\n", err)
	}

	// Fallback to local storage
	return s.uploadToLocal(buffer.Bytes(), key, fileHeader.Header.Get("Content-Type"), size)
}

// uploadToS3 uploads to Amazon S3
func (s *StorageService) uploadToS3(data []byte, key, contentType string) (*UploadResult, error) {
	_, err := s.s3Client.PutObject(s.ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)

	return &UploadResult{
		URL:      url,
		Key:      key,
		Size:     int64(len(data)),
		MimeType: contentType,
		Backend:  "s3",
	}, nil
}

// uploadToLocal uploads to local filesystem
func (s *StorageService) uploadToLocal(data []byte, key, contentType string, size int64) (*UploadResult, error) {
	fullPath := filepath.Join(s.localPath, key)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("%s/%s", strings.TrimRight(s.publicURL, "/"), key)

	return &UploadResult{
		URL:      url,
		Key:      key,
		Size:     size,
		MimeType: contentType,
		Backend:  "local",
	}, nil
}

// DeleteFile deletes a file from S3 or local storage
func (s *StorageService) DeleteFile(key string) error {
	// Try S3 first if enabled
	if s.s3Enabled {
		_, err := s.s3Client.DeleteObject(s.ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		})
		if err == nil {
			return nil
		}
		fmt.Printf("⚠️  S3 delete failed: %v (trying local storage)\n", err)
	}

	// Fallback to local storage
	fullPath := filepath.Join(s.localPath, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete local file: %w", err)
	}

	return nil
}

// GetFileURL returns the public URL for a file
func (s *StorageService) GetFileURL(key string) string {
	if s.s3Enabled {
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(s.publicURL, "/"), key)
}

// GeneratePresignedURL generates a presigned URL for temporary access
func (s *StorageService) GeneratePresignedURL(key string, expiration time.Duration) (string, error) {
	if !s.s3Enabled {
		// For local storage, just return the regular URL
		return s.GetFileURL(key), nil
	}

	presignClient := s3.NewPresignClient(s.s3Client)

	request, err := presignClient.PresignGetObject(s.ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// FileExists checks if a file exists
func (s *StorageService) FileExists(key string) bool {
	// Try S3 first if enabled
	if s.s3Enabled {
		_, err := s.s3Client.HeadObject(s.ctx, &s3.HeadObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		})
		return err == nil
	}

	// Check local storage
	fullPath := filepath.Join(s.localPath, key)
	_, err := os.Stat(fullPath)
	return err == nil
}

// IsS3Enabled returns true if S3 is available
func (s *StorageService) IsS3Enabled() bool {
	return s.s3Enabled
}

// GetBackendInfo returns information about the storage backend
func (s *StorageService) GetBackendInfo() map[string]interface{} {
	if s.s3Enabled {
		return map[string]interface{}{
			"backend":   "s3",
			"bucket":    s.bucket,
			"region":    s.region,
			"available": true,
		}
	}
	return map[string]interface{}{
		"backend":   "local",
		"path":      s.localPath,
		"available": true,
	}
}
