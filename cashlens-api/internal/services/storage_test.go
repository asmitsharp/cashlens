package services

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStorageService tests the constructor
func TestNewStorageService(t *testing.T) {
	tests := []struct {
		name     string
		bucket   string
		region   string
		endpoint string
		wantErr  bool
	}{
		{
			name:     "valid configuration",
			bucket:   "test-bucket",
			region:   "us-east-1",
			endpoint: "http://localhost:4566",
			wantErr:  false,
		},
		{
			name:     "valid configuration without endpoint",
			bucket:   "test-bucket",
			region:   "ap-south-1",
			endpoint: "",
			wantErr:  false,
		},
		{
			name:     "empty bucket",
			bucket:   "",
			region:   "us-east-1",
			endpoint: "http://localhost:4566",
			wantErr:  true,
		},
		{
			name:     "empty region",
			bucket:   "test-bucket",
			region:   "",
			endpoint: "http://localhost:4566",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewStorageService(tt.bucket, tt.region, tt.endpoint)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.Equal(t, tt.bucket, service.bucket)
				assert.Equal(t, tt.region, service.region)
			}
		})
	}
}

// TestGenerateUploadKey tests the upload key generation
func TestGenerateUploadKey(t *testing.T) {
	service := &StorageService{}

	tests := []struct {
		name        string
		userID      string
		filename    string
		wantContain []string
		wantErr     bool
	}{
		{
			name:        "valid input",
			userID:      "user123",
			filename:    "statement.csv",
			wantContain: []string{"uploads/", "user123/", "statement.csv"},
			wantErr:     false,
		},
		{
			name:        "filename with spaces",
			userID:      "user456",
			filename:    "my statement file.csv",
			wantContain: []string{"uploads/", "user456/", ".csv"},
			wantErr:     false,
		},
		{
			name:        "filename with special characters",
			userID:      "user789",
			filename:    "statement@2024.csv",
			wantContain: []string{"uploads/", "user789/", ".csv"},
			wantErr:     false,
		},
		{
			name:     "empty user ID",
			userID:   "",
			filename: "statement.csv",
			wantErr:  true,
		},
		{
			name:     "empty filename",
			userID:   "user123",
			filename: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := service.GenerateUploadKey(tt.userID, tt.filename)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, key)

				// Check key contains expected parts
				for _, contain := range tt.wantContain {
					assert.Contains(t, key, contain)
				}

				// Verify format: uploads/{userID}/{timestamp}-{filename}
				parts := strings.Split(key, "/")
				assert.Equal(t, 3, len(parts), "key should have 3 parts separated by /")
				assert.Equal(t, "uploads", parts[0])
				assert.Equal(t, tt.userID, parts[1])
			}
		})
	}
}

// TestGeneratePresignedURL tests presigned URL generation
func TestGeneratePresignedURL(t *testing.T) {
	// Skip if LocalStack is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service, err := NewStorageService("cashlens-uploads", "us-east-1", "http://localhost:4566")
	require.NoError(t, err, "Failed to create storage service")

	tests := []struct {
		name           string
		key            string
		contentType    string
		expiryMinutes  int
		wantErr        bool
		wantContainURL []string
	}{
		{
			name:           "valid presigned URL for CSV",
			key:            "uploads/user123/test.csv",
			contentType:    "text/csv",
			expiryMinutes:  15,
			wantErr:        false,
			wantContainURL: []string{"uploads", "user123", "test.csv", "X-Amz-Algorithm", "X-Amz-Credential"},
		},
		{
			name:           "valid presigned URL for XLSX",
			key:            "uploads/user456/test.xlsx",
			contentType:    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			expiryMinutes:  30,
			wantErr:        false,
			wantContainURL: []string{"uploads", "user456", "test.xlsx"},
		},
		{
			name:          "empty key",
			key:           "",
			contentType:   "text/csv",
			expiryMinutes: 15,
			wantErr:       true,
		},
		{
			name:          "invalid expiry (0 minutes)",
			key:           "uploads/user123/test.csv",
			contentType:   "text/csv",
			expiryMinutes: 0,
			wantErr:       true,
		},
		{
			name:          "invalid expiry (negative)",
			key:           "uploads/user123/test.csv",
			contentType:   "text/csv",
			expiryMinutes: -10,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := service.GeneratePresignedURL(tt.key, tt.contentType, tt.expiryMinutes)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, url)

				// Check URL contains expected parts
				for _, contain := range tt.wantContainURL {
					assert.Contains(t, url, contain, fmt.Sprintf("URL should contain '%s'", contain))
				}

				// Verify it's a valid HTTP URL
				assert.True(t, strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
			}
		})
	}
}

// TestDownloadFile tests file download functionality
func TestDownloadFile(t *testing.T) {
	// Skip if LocalStack is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service, err := NewStorageService("cashlens-uploads", "us-east-1", "http://localhost:4566")
	require.NoError(t, err, "Failed to create storage service")

	// Setup: Create bucket and upload test file
	ctx := context.Background()
	testContent := "test file content for download"
	testKey := "uploads/testuser/test-download.txt"

	// Create bucket first
	_, err = service.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(service.bucket),
	})
	// Ignore error if bucket already exists

	// Upload test file
	_, err = service.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(service.bucket),
		Key:    aws.String(testKey),
		Body:   strings.NewReader(testContent),
	})
	require.NoError(t, err, "Failed to upload test file")

	tests := []struct {
		name        string
		key         string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "download existing file",
			key:         testKey,
			wantContent: testContent,
			wantErr:     false,
		},
		{
			name:    "download non-existent file",
			key:     "uploads/testuser/nonexistent.txt",
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := service.DownloadFile(tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, reader)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reader)
				defer reader.Close()

				// Read and verify content
				content, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantContent, string(content))
			}
		})
	}

	// Cleanup: Delete test file
	_, err = service.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(service.bucket),
		Key:    aws.String(testKey),
	})
	assert.NoError(t, err, "Failed to cleanup test file")
}

// TestDeleteFile tests file deletion functionality
func TestDeleteFile(t *testing.T) {
	// Skip if LocalStack is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service, err := NewStorageService("cashlens-uploads", "us-east-1", "http://localhost:4566")
	require.NoError(t, err, "Failed to create storage service")

	ctx := context.Background()

	// Create bucket if not exists
	_, err = service.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(service.bucket),
	})
	// Ignore error if bucket already exists

	tests := []struct {
		name       string
		key        string
		setupFile  bool
		wantErr    bool
		errMessage string
	}{
		{
			name:      "delete existing file",
			key:       "uploads/testuser/test-delete.txt",
			setupFile: true,
			wantErr:   false,
		},
		{
			name:      "delete non-existent file (should succeed - S3 behavior)",
			key:       "uploads/testuser/nonexistent-delete.txt",
			setupFile: false,
			wantErr:   false, // S3 DeleteObject succeeds even if file doesn't exist
		},
		{
			name:       "empty key",
			key:        "",
			setupFile:  false,
			wantErr:    true,
			errMessage: "key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Upload file if needed
			if tt.setupFile {
				_, err := service.s3Client.PutObject(ctx, &s3.PutObjectInput{
					Bucket: aws.String(service.bucket),
					Key:    aws.String(tt.key),
					Body:   strings.NewReader("test content for deletion"),
				})
				require.NoError(t, err, "Failed to setup test file")
			}

			// Test deletion
			err := service.DeleteFile(tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)

				// Verify file is deleted by trying to download
				if tt.setupFile {
					_, err := service.DownloadFile(tt.key)
					assert.Error(t, err, "File should not exist after deletion")
				}
			}
		})
	}
}

// TestGeneratePresignedURL_ExpiryValidation tests expiry time validation
func TestGeneratePresignedURL_ExpiryValidation(t *testing.T) {
	service := &StorageService{
		bucket: "test-bucket",
		region: "us-east-1",
	}

	// Test that expiry is properly used
	key := "uploads/user123/test.csv"
	contentType := "text/csv"

	// This should fail because we haven't initialized the S3 client
	// But it will test the validation logic
	_, err := service.GeneratePresignedURL(key, contentType, 15)
	// We expect an error because s3Client is nil
	assert.Error(t, err)
}

// TestStorageService_Integration tests the full workflow
func TestStorageService_Integration(t *testing.T) {
	// Skip if LocalStack is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create service
	service, err := NewStorageService("cashlens-uploads", "us-east-1", "http://localhost:4566")
	require.NoError(t, err, "Failed to create storage service")

	ctx := context.Background()

	// Create bucket
	_, err = service.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(service.bucket),
	})
	// Ignore error if bucket already exists

	userID := "integration-test-user"
	filename := "integration-test.csv"

	// Step 1: Generate upload key
	uploadKey, err := service.GenerateUploadKey(userID, filename)
	require.NoError(t, err)
	assert.Contains(t, uploadKey, userID)
	assert.Contains(t, uploadKey, filename)

	// Step 2: Generate presigned URL
	presignedURL, err := service.GeneratePresignedURL(uploadKey, "text/csv", 15)
	require.NoError(t, err)
	assert.NotEmpty(t, presignedURL)

	// Step 3: Upload a file using the key (simulating what would happen after presigned URL upload)
	testContent := "integration test content"
	_, err = service.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(service.bucket),
		Key:    aws.String(uploadKey),
		Body:   strings.NewReader(testContent),
	})
	require.NoError(t, err, "Failed to upload file")

	// Wait a bit for S3 consistency (not needed with LocalStack, but good practice)
	time.Sleep(100 * time.Millisecond)

	// Step 4: Download the file
	reader, err := service.DownloadFile(uploadKey)
	require.NoError(t, err)
	defer reader.Close()

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))

	// Step 5: Delete the file
	err = service.DeleteFile(uploadKey)
	require.NoError(t, err)

	// Step 6: Verify deletion
	_, err = service.DownloadFile(uploadKey)
	assert.Error(t, err, "File should not exist after deletion")
}

// TestStorageService_LocalStackConfig tests LocalStack-specific configuration
func TestStorageService_LocalStackConfig(t *testing.T) {
	service, err := NewStorageService("cashlens-uploads", "us-east-1", "http://localhost:4566")
	require.NoError(t, err)

	assert.NotNil(t, service.s3Client)
	assert.Equal(t, "cashlens-uploads", service.bucket)
	assert.Equal(t, "us-east-1", service.region)
}

// TestStorageService_ProductionConfig tests production AWS configuration
func TestStorageService_ProductionConfig(t *testing.T) {
	// This should work even without actual AWS credentials
	// because we're just testing the constructor
	service, err := NewStorageService("cashlens-uploads", "ap-south-1", "")
	require.NoError(t, err)

	assert.NotNil(t, service.s3Client)
	assert.Equal(t, "cashlens-uploads", service.bucket)
	assert.Equal(t, "ap-south-1", service.region)
}
