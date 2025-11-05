# Storage Service Usage Guide

## Overview

The `StorageService` provides a clean interface for S3 file operations including presigned URL generation, file download, and file deletion. It supports both AWS S3 (production) and LocalStack (local development).

## Setup

### Dependencies

```bash
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/credentials
```

### Configuration

#### LocalStack (Development)

```go
service, err := services.NewStorageService(
    "cashlens-uploads",           // bucket name
    "us-east-1",                   // region
    "http://localhost:4566",       // LocalStack endpoint
)
```

#### AWS S3 (Production)

```go
service, err := services.NewStorageService(
    "cashlens-uploads",           // bucket name
    "ap-south-1",                  // region
    "",                            // empty endpoint for AWS
)
```

## Usage Examples

### 1. Generate Upload Key

Creates a unique, timestamped key for file uploads:

```go
userID := "user123"
filename := "bank-statement.csv"

uploadKey, err := service.GenerateUploadKey(userID, filename)
if err != nil {
    log.Fatalf("Failed to generate upload key: %v", err)
}

// uploadKey format: uploads/user123/1699564800-a1b2c3d4-bank-statement.csv
fmt.Println("Upload key:", uploadKey)
```

### 2. Generate Presigned URL

Generate a presigned PUT URL for direct browser uploads:

```go
presignedURL, err := service.GeneratePresignedURL(
    uploadKey,
    "text/csv",  // content type
    15,          // expiry in minutes
)
if err != nil {
    log.Fatalf("Failed to generate presigned URL: %v", err)
}

// Send this URL to the frontend
fmt.Println("Presigned URL:", presignedURL)
```

### 3. Download File

Download a file from S3:

```go
reader, err := service.DownloadFile(uploadKey)
if err != nil {
    log.Fatalf("Failed to download file: %v", err)
}
defer reader.Close()

// Read file content
content, err := io.ReadAll(reader)
if err != nil {
    log.Fatalf("Failed to read file content: %v", err)
}

fmt.Printf("Downloaded %d bytes\n", len(content))
```

### 4. Delete File

Delete a file from S3:

```go
err := service.DeleteFile(uploadKey)
if err != nil {
    log.Fatalf("Failed to delete file: %v", err)
}

fmt.Println("File deleted successfully")
```

## Complete Workflow Example

```go
package main

import (
    "fmt"
    "io"
    "log"
    "strings"

    "github.com/ashmitsharp/cashlens-api/internal/services"
)

func main() {
    // 1. Initialize storage service
    service, err := services.NewStorageService(
        "cashlens-uploads",
        "us-east-1",
        "http://localhost:4566", // Use "" for production AWS
    )
    if err != nil {
        log.Fatalf("Failed to initialize storage: %v", err)
    }

    // 2. Generate upload key
    userID := "user123"
    filename := "statement.csv"
    uploadKey, err := service.GenerateUploadKey(userID, filename)
    if err != nil {
        log.Fatalf("Failed to generate key: %v", err)
    }
    fmt.Printf("Generated key: %s\n", uploadKey)

    // 3. Generate presigned URL for frontend
    presignedURL, err := service.GeneratePresignedURL(uploadKey, "text/csv", 15)
    if err != nil {
        log.Fatalf("Failed to generate presigned URL: %v", err)
    }
    fmt.Printf("Presigned URL: %s\n", presignedURL)

    // ... Frontend uploads file using presigned URL ...

    // 4. Download and process file (backend)
    reader, err := service.DownloadFile(uploadKey)
    if err != nil {
        log.Fatalf("Failed to download: %v", err)
    }
    defer reader.Close()

    content, err := io.ReadAll(reader)
    if err != nil {
        log.Fatalf("Failed to read: %v", err)
    }
    fmt.Printf("Downloaded %d bytes\n", len(content))

    // 5. Delete file after processing
    err = service.DeleteFile(uploadKey)
    if err != nil {
        log.Fatalf("Failed to delete: %v", err)
    }
    fmt.Println("File deleted successfully")
}
```

## HTTP Handler Example

```go
package handlers

import (
    "github.com/ashmitsharp/cashlens-api/internal/services"
    "github.com/gofiber/fiber/v3"
)

type UploadHandler struct {
    storage *services.StorageService
}

func NewUploadHandler(storage *services.StorageService) *UploadHandler {
    return &UploadHandler{storage: storage}
}

// GeneratePresignedURL returns a presigned URL for file upload
func (h *UploadHandler) GeneratePresignedURL(c fiber.Ctx) error {
    var req struct {
        Filename    string `json:"filename"`
        ContentType string `json:"content_type"`
    }

    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    // Get user ID from context (set by auth middleware)
    userID := c.Locals("user_id").(string)

    // Generate upload key
    uploadKey, err := h.storage.GenerateUploadKey(userID, req.Filename)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to generate upload key",
        })
    }

    // Generate presigned URL (15 minutes expiry)
    presignedURL, err := h.storage.GeneratePresignedURL(uploadKey, req.ContentType, 15)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to generate presigned URL",
        })
    }

    return c.JSON(fiber.Map{
        "upload_url": presignedURL,
        "file_key":   uploadKey,
        "expires_in": 900, // 15 minutes in seconds
    })
}
```

## Testing with LocalStack

### Start LocalStack

```bash
# From project root
cd /Users/asmitsingh/Desktop/side/cashlens
docker compose up -d localstack
```

### Run Tests

```bash
# Run all storage tests with integration tests
go test -v ./internal/services -run Storage

# Run tests in short mode (skip integration tests)
go test -v ./internal/services -run Storage -short

# Run tests with coverage
go test -v ./internal/services -run Storage -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

Current test coverage for storage.go:

- `NewStorageService`: 80.0%
- `GenerateUploadKey`: 100.0%
- `GeneratePresignedURL`: 92.9%
- `DownloadFile`: 87.5%
- `DeleteFile`: 75.0%

**Overall Average**: 87.08% (exceeds 80% requirement)

## Error Handling

All methods return descriptive errors:

```go
// Example error handling
uploadKey, err := service.GenerateUploadKey("", "file.csv")
if err != nil {
    // Error: "userID cannot be empty"
    log.Printf("Validation error: %v", err)
}

reader, err := service.DownloadFile("non-existent-key")
if err != nil {
    // Error: "failed to download file from S3: ..."
    log.Printf("S3 error: %v", err)
}
```

## Best Practices

1. **Always defer Close()** when downloading files:
   ```go
   reader, err := service.DownloadFile(key)
   if err != nil {
       return err
   }
   defer reader.Close() // Important!
   ```

2. **Use appropriate expiry times** for presigned URLs:
   - Short uploads (CSV): 15 minutes
   - Large files: 30-60 minutes

3. **Validate user input** before generating keys:
   ```go
   if userID == "" || filename == "" {
       return errors.New("invalid input")
   }
   ```

4. **Handle S3 errors gracefully**:
   ```go
   if err := service.DeleteFile(key); err != nil {
       log.Printf("Failed to delete %s: %v", key, err)
       // Don't fail the entire operation
   }
   ```

## Environment Variables

```bash
# LocalStack (development)
S3_BUCKET=cashlens-uploads
S3_REGION=us-east-1
AWS_ENDPOINT=http://localhost:4566

# AWS S3 (production)
S3_BUCKET=cashlens-uploads
S3_REGION=ap-south-1
AWS_ENDPOINT=  # Empty for production
```

## Troubleshooting

### LocalStack connection fails

```bash
# Check LocalStack is running
docker compose ps

# Check LocalStack logs
docker compose logs localstack

# Restart LocalStack
docker compose restart localstack
```

### Presigned URL returns 403

- Check bucket exists in LocalStack/AWS
- Verify credentials (LocalStack accepts any credentials)
- Ensure region matches bucket region

### File download fails

- Verify file exists with correct key
- Check bucket permissions
- Ensure LocalStack is accessible at specified endpoint
