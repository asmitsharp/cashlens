package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// StorageService handles S3 file operations
type StorageService struct {
	s3Client *s3.Client
	bucket   string
	region   string
}

// NewStorageService creates a new storage service instance
// For LocalStack: endpoint should be "http://localhost:4566"
// For production AWS: endpoint should be ""
func NewStorageService(bucket, region, endpoint string) (*StorageService, error) {
	// Validate required parameters
	if bucket == "" {
		return nil, fmt.Errorf("bucket cannot be empty")
	}
	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}

	ctx := context.Background()

	// Configure AWS SDK
	var cfg aws.Config
	var err error

	if endpoint != "" {
		// LocalStack configuration with custom endpoint
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				"test",      // Access Key ID (LocalStack accepts any value)
				"test",      // Secret Access Key
				"",          // Session Token
			)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		// Create S3 client with custom endpoint for LocalStack
		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // Required for LocalStack
		})

		return &StorageService{
			s3Client: client,
			bucket:   bucket,
			region:   region,
		}, nil
	}

	// Production AWS configuration
	cfg, err = config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &StorageService{
		s3Client: client,
		bucket:   bucket,
		region:   region,
	}, nil
}

// GenerateUploadKey creates a unique S3 key for file uploads
// Format: uploads/{userID}/{timestamp}-{filename}
func (s *StorageService) GenerateUploadKey(userID, filename string) (string, error) {
	// Validate inputs
	if userID == "" {
		return "", fmt.Errorf("userID cannot be empty")
	}
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	// Sanitize filename (remove spaces and special chars, keep extension)
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	// Replace spaces and special characters with hyphens
	baseName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, baseName)

	// Generate timestamp
	timestamp := time.Now().UTC().Unix()

	// Generate unique identifier
	uniqueID := uuid.New().String()[:8]

	// Format: uploads/{userID}/{timestamp}-{uniqueID}-{filename}
	key := fmt.Sprintf("uploads/%s/%d-%s-%s%s", userID, timestamp, uniqueID, baseName, ext)

	return key, nil
}

// GeneratePresignedURL generates a presigned PUT URL for file uploads
func (s *StorageService) GeneratePresignedURL(key, contentType string, expiryMinutes int) (string, error) {
	// Validate inputs
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}
	if expiryMinutes <= 0 {
		return "", fmt.Errorf("expiryMinutes must be greater than 0")
	}
	if s.s3Client == nil {
		return "", fmt.Errorf("s3 client is not initialized")
	}

	// Create presign client
	presignClient := s3.NewPresignClient(s.s3Client)

	// Create PutObject input
	putObjectInput := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	// Add content type if provided
	if contentType != "" {
		putObjectInput.ContentType = aws.String(contentType)
	}

	// Generate presigned URL
	presignedReq, err := presignClient.PresignPutObject(
		context.Background(),
		putObjectInput,
		s3.WithPresignExpires(time.Duration(expiryMinutes)*time.Minute),
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}

// DownloadFile downloads a file from S3 and returns a reader
func (s *StorageService) DownloadFile(key string) (io.ReadCloser, error) {
	// Validate inputs
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}
	if s.s3Client == nil {
		return nil, fmt.Errorf("s3 client is not initialized")
	}

	// Get object from S3
	result, err := s.s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return result.Body, nil
}

// DeleteFile deletes a file from S3
func (s *StorageService) DeleteFile(key string) error {
	// Validate inputs
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if s.s3Client == nil {
		return fmt.Errorf("s3 client is not initialized")
	}

	// Delete object from S3
	_, err := s.s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
