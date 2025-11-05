package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ashmitsharp/cashlens-api/internal/models"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorageService is a mock implementation of StorageService for testing
type MockStorageService struct {
	GenerateUploadKeyFunc    func(userID, filename string) (string, error)
	GeneratePresignedURLFunc func(key, contentType string, expiryMinutes int) (string, error)
	DownloadFileFunc         func(key string) (io.ReadCloser, error)
}

func (m *MockStorageService) GenerateUploadKey(userID, filename string) (string, error) {
	if m.GenerateUploadKeyFunc != nil {
		return m.GenerateUploadKeyFunc(userID, filename)
	}
	return fmt.Sprintf("uploads/%s/mock-%s", userID, filename), nil
}

func (m *MockStorageService) GeneratePresignedURL(key, contentType string, expiryMinutes int) (string, error) {
	if m.GeneratePresignedURLFunc != nil {
		return m.GeneratePresignedURLFunc(key, contentType, expiryMinutes)
	}
	return fmt.Sprintf("https://s3.amazonaws.com/bucket/%s?signature=mock", key), nil
}

func (m *MockStorageService) DownloadFile(key string) (io.ReadCloser, error) {
	if m.DownloadFileFunc != nil {
		return m.DownloadFileFunc(key)
	}
	return nil, fmt.Errorf("file not found")
}

// MockParser is a mock implementation of Parser for testing
type MockParser struct {
	ParseFileFunc func(file io.Reader, filename string) ([]models.ParsedTransaction, error)
}

func (m *MockParser) ParseFile(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
	if m.ParseFileFunc != nil {
		return m.ParseFileFunc(file, filename)
	}
	return nil, fmt.Errorf("parse failed")
}

// TestGetPresignedURL_Success tests successful presigned URL generation
func TestGetPresignedURL_Success(t *testing.T) {
	// Setup mock storage service
	mockStorage := &MockStorageService{
		GenerateUploadKeyFunc: func(userID, filename string) (string, error) {
			return fmt.Sprintf("uploads/%s/1699564800-uuid-%s", userID, filename), nil
		},
		GeneratePresignedURLFunc: func(key, contentType string, expiryMinutes int) (string, error) {
			return fmt.Sprintf("https://s3.amazonaws.com/bucket/%s?X-Amz-Signature=abc123", key), nil
		},
	}

	// Create handler
	handler := NewUploadHandler(mockStorage)

	// Setup Fiber app
	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		// Simulate auth middleware setting user_id
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	// Create test request
	req := httptest.NewRequest("GET", "/presigned-url?filename=test.csv&content_type=text/csv", nil)

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response status
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Parse response body
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Assert response structure
	assert.Contains(t, result, "upload_url")
	assert.Contains(t, result, "file_key")
	assert.Contains(t, result, "expires_in")

	// Assert response values
	assert.Contains(t, result["upload_url"].(string), "https://s3.amazonaws.com")
	assert.Contains(t, result["file_key"].(string), "uploads/user123")
	assert.Equal(t, float64(900), result["expires_in"].(float64))
}

// TestGetPresignedURL_MissingFilename tests error when filename is missing
func TestGetPresignedURL_MissingFilename(t *testing.T) {
	mockStorage := &MockStorageService{}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	// Missing filename parameter
	req := httptest.NewRequest("GET", "/presigned-url?content_type=text/csv", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 400 Bad Request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "filename")
}

// TestGetPresignedURL_MissingContentType tests error when content_type is missing
func TestGetPresignedURL_MissingContentType(t *testing.T) {
	mockStorage := &MockStorageService{}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	// Missing content_type parameter
	req := httptest.NewRequest("GET", "/presigned-url?filename=test.csv", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 400 Bad Request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "content_type")
}

// TestGetPresignedURL_MissingUserID tests error when user_id is not set (auth failure)
func TestGetPresignedURL_MissingUserID(t *testing.T) {
	mockStorage := &MockStorageService{}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", handler.GetPresignedURL) // No user_id set

	req := httptest.NewRequest("GET", "/presigned-url?filename=test.csv&content_type=text/csv", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 401 Unauthorized
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "unauthorized")
}

// TestGetPresignedURL_InvalidContentType tests error for unsupported content types
func TestGetPresignedURL_InvalidContentType(t *testing.T) {
	mockStorage := &MockStorageService{}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	// Test invalid content types
	invalidTypes := []string{
		"image/jpeg",
		"application/json",
		"text/plain",
		"video/mp4",
		"application/zip",
	}

	for _, contentType := range invalidTypes {
		t.Run(contentType, func(t *testing.T) {
			req := httptest.NewRequest("GET",
				fmt.Sprintf("/presigned-url?filename=test.csv&content_type=%s", contentType),
				nil)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assert 400 Bad Request
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "error")
			assert.Contains(t, result["error"].(string), "unsupported")
		})
	}
}

// TestGetPresignedURL_ValidContentTypes tests all valid content types
func TestGetPresignedURL_ValidContentTypes(t *testing.T) {
	mockStorage := &MockStorageService{}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	// Test valid content types
	validTypes := []struct {
		contentType string
		filename    string
	}{
		{"text/csv", "test.csv"},
		{"application/vnd.ms-excel", "test.xls"},
		{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "test.xlsx"},
		{"application/pdf", "test.pdf"},
	}

	for _, tc := range validTypes {
		t.Run(tc.contentType, func(t *testing.T) {
			req := httptest.NewRequest("GET",
				fmt.Sprintf("/presigned-url?filename=%s&content_type=%s", tc.filename, tc.contentType),
				nil)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assert 200 OK
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "upload_url")
			assert.Contains(t, result, "file_key")
			assert.Contains(t, result, "expires_in")
		})
	}
}

// TestGetPresignedURL_GenerateKeyError tests error when GenerateUploadKey fails
func TestGetPresignedURL_GenerateKeyError(t *testing.T) {
	mockStorage := &MockStorageService{
		GenerateUploadKeyFunc: func(userID, filename string) (string, error) {
			return "", fmt.Errorf("failed to generate key")
		},
	}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	req := httptest.NewRequest("GET", "/presigned-url?filename=test.csv&content_type=text/csv", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 500 Internal Server Error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
}

// TestGetPresignedURL_GenerateURLError tests error when GeneratePresignedURL fails
func TestGetPresignedURL_GenerateURLError(t *testing.T) {
	mockStorage := &MockStorageService{
		GenerateUploadKeyFunc: func(userID, filename string) (string, error) {
			return "uploads/user123/test.csv", nil
		},
		GeneratePresignedURLFunc: func(key, contentType string, expiryMinutes int) (string, error) {
			return "", fmt.Errorf("S3 service unavailable")
		},
	}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	req := httptest.NewRequest("GET", "/presigned-url?filename=test.csv&content_type=text/csv", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 500 Internal Server Error
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
}

// TestGetPresignedURL_EmptyFilename tests error when filename is empty string
func TestGetPresignedURL_EmptyFilename(t *testing.T) {
	mockStorage := &MockStorageService{}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	// Empty filename parameter
	req := httptest.NewRequest("GET", "/presigned-url?filename=&content_type=text/csv", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 400 Bad Request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "filename")
}

// TestGetPresignedURL_EmptyContentType tests error when content_type is empty string
func TestGetPresignedURL_EmptyContentType(t *testing.T) {
	mockStorage := &MockStorageService{}
	handler := NewUploadHandler(mockStorage)

	app := fiber.New()
	app.Get("/presigned-url", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.GetPresignedURL(c)
	})

	// Empty content_type parameter
	req := httptest.NewRequest("GET", "/presigned-url?filename=test.csv&content_type=", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert 400 Bad Request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "content_type")
}

// TestNewUploadHandler tests the constructor
func TestNewUploadHandler(t *testing.T) {
	mockStorage := &MockStorageService{}
	handler := NewUploadHandler(mockStorage)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.storage)
}

// =============================================================================
// ProcessUpload Tests (RED Phase - TDD)
// =============================================================================

// TestProcessUpload_Success_CSV tests successful CSV file processing
func TestProcessUpload_Success_CSV(t *testing.T) {
	// Load test CSV file
	testFile, err := os.Open("../../testdata/hdfc_sample.csv")
	require.NoError(t, err)
	defer testFile.Close()

	fileContent, err := io.ReadAll(testFile)
	require.NoError(t, err)

	// Setup mocks
	mockStorage := &MockStorageService{
		DownloadFileFunc: func(key string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(fileContent)), nil
		},
	}

	mockParser := &MockParser{
		ParseFileFunc: func(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
			// Return realistic parsed data
			return []models.ParsedTransaction{
				{Description: "AWS Services", Amount: -5000.50},
				{Description: "Salary Credit", Amount: 50000.00},
				{Description: "Office Rent", Amount: -15000.00},
			}, nil
		},
	}

	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.ProcessUpload(c)
	})

	// Create request body
	reqBody := map[string]string{
		"file_key": "uploads/user123/1699564800-uuid-statement.csv",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "uploads/user123/1699564800-uuid-statement.csv", result["file_key"])
	assert.Equal(t, float64(3), result["total_rows"])
	assert.Equal(t, float64(3), result["transactions_parsed"])
	assert.Contains(t, result, "bank_detected")
	assert.Contains(t, result, "date_range")
}

// TestProcessUpload_Success_XLSX tests successful XLSX file processing
func TestProcessUpload_Success_XLSX(t *testing.T) {
	// Load test XLSX file
	testFile, err := os.Open("../../testdata/icici_sample.xlsx")
	require.NoError(t, err)
	defer testFile.Close()

	fileContent, err := io.ReadAll(testFile)
	require.NoError(t, err)

	// Setup mocks
	mockStorage := &MockStorageService{
		DownloadFileFunc: func(key string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(fileContent)), nil
		},
	}

	mockParser := &MockParser{
		ParseFileFunc: func(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
			return []models.ParsedTransaction{
				{Description: "UPI Payment", Amount: -1500.00},
				{Description: "Interest Credit", Amount: 250.00},
			}, nil
		},
	}

	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user456")
		return handler.ProcessUpload(c)
	})

	reqBody := map[string]string{
		"file_key": "uploads/user456/1699564800-uuid-statement.xlsx",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "uploads/user456/1699564800-uuid-statement.xlsx", result["file_key"])
	assert.Equal(t, float64(2), result["total_rows"])
	assert.Equal(t, float64(2), result["transactions_parsed"])
}

// TestProcessUpload_Success_PDF tests successful PDF file processing
func TestProcessUpload_Success_PDF(t *testing.T) {
	mockStorage := &MockStorageService{
		DownloadFileFunc: func(key string) (io.ReadCloser, error) {
			// Return mock PDF content
			return io.NopCloser(bytes.NewReader([]byte("%PDF-1.4 mock content"))), nil
		},
	}

	mockParser := &MockParser{
		ParseFileFunc: func(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
			return []models.ParsedTransaction{
				{Description: "Card Payment", Amount: -2500.00},
			}, nil
		},
	}

	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user789")
		return handler.ProcessUpload(c)
	})

	reqBody := map[string]string{
		"file_key": "uploads/user789/1699564800-uuid-statement.pdf",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, float64(1), result["transactions_parsed"])
}

// TestProcessUpload_MissingFileKey tests error when file_key is missing
func TestProcessUpload_MissingFileKey(t *testing.T) {
	mockStorage := &MockStorageService{}
	mockParser := &MockParser{}
	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.ProcessUpload(c)
	})

	// Empty request body
	reqBody := map[string]string{}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "file_key")
}

// TestProcessUpload_InvalidJSON tests error with invalid JSON body
func TestProcessUpload_InvalidJSON(t *testing.T) {
	mockStorage := &MockStorageService{}
	mockParser := &MockParser{}
	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.ProcessUpload(c)
	})

	req := httptest.NewRequest("POST", "/process", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
}

// TestProcessUpload_Unauthorized tests error when user_id is missing
func TestProcessUpload_Unauthorized(t *testing.T) {
	mockStorage := &MockStorageService{}
	mockParser := &MockParser{}
	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", handler.ProcessUpload) // No user_id set

	reqBody := map[string]string{
		"file_key": "uploads/user123/test.csv",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "unauthorized")
}

// TestProcessUpload_Forbidden tests error when user tries to access another user's file
func TestProcessUpload_Forbidden(t *testing.T) {
	mockStorage := &MockStorageService{}
	mockParser := &MockParser{}
	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.ProcessUpload(c)
	})

	// Trying to access user456's file
	reqBody := map[string]string{
		"file_key": "uploads/user456/test.csv",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "forbidden")
}

// TestProcessUpload_FileNotFound tests error when file doesn't exist in S3
func TestProcessUpload_FileNotFound(t *testing.T) {
	mockStorage := &MockStorageService{
		DownloadFileFunc: func(key string) (io.ReadCloser, error) {
			return nil, fmt.Errorf("NoSuchKey: The specified key does not exist")
		},
	}
	mockParser := &MockParser{}
	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.ProcessUpload(c)
	})

	reqBody := map[string]string{
		"file_key": "uploads/user123/nonexistent.csv",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "not found")
}

// TestProcessUpload_ParseError tests error when parser fails
func TestProcessUpload_ParseError(t *testing.T) {
	mockStorage := &MockStorageService{
		DownloadFileFunc: func(key string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("invalid,csv,content"))), nil
		},
	}
	mockParser := &MockParser{
		ParseFileFunc: func(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
			return nil, fmt.Errorf("unknown bank format")
		},
	}
	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.ProcessUpload(c)
	})

	reqBody := map[string]string{
		"file_key": "uploads/user123/invalid.csv",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "error")
	assert.Contains(t, result["error"].(string), "parse")
}

// TestProcessUpload_EmptyFile tests error when file has no transactions
func TestProcessUpload_EmptyFile(t *testing.T) {
	mockStorage := &MockStorageService{
		DownloadFileFunc: func(key string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("Date,Narration,Amount\n"))), nil
		},
	}
	mockParser := &MockParser{
		ParseFileFunc: func(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
			return []models.ParsedTransaction{}, nil // Empty result
		},
	}
	handler := NewUploadHandlerWithParser(mockStorage, mockParser)

	app := fiber.New()
	app.Post("/process", func(c fiber.Ctx) error {
		c.Locals("user_id", "user123")
		return handler.ProcessUpload(c)
	})

	reqBody := map[string]string{
		"file_key": "uploads/user123/empty.csv",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should succeed but with 0 transactions
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, float64(0), result["total_rows"])
	assert.Equal(t, float64(0), result["transactions_parsed"])
}

// TestProcessUpload_BankDetection tests bank detection from filename
func TestProcessUpload_BankDetection(t *testing.T) {
	testCases := []struct {
		name         string
		filename     string
		expectedBank string
	}{
		{"HDFC detection", "hdfc_statement.csv", "HDFC"},
		{"ICICI detection", "ICICI-Bank-Statement.csv", "ICICI"},
		{"SBI detection", "sbi_account_statement.csv", "SBI"},
		{"Axis detection", "axis-statement-2024.csv", "Axis"},
		{"Kotak detection", "kotak_mahindra.csv", "Kotak"},
		{"Unknown bank", "random_bank_statement.csv", "UNKNOWN"},
		{"Uppercase filename", "HDFC-STATEMENT.CSV", "HDFC"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage := &MockStorageService{
				DownloadFileFunc: func(key string) (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader([]byte("mock content"))), nil
				},
			}
			mockParser := &MockParser{
				ParseFileFunc: func(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
					return []models.ParsedTransaction{
						{Description: "Test", Amount: -100.00},
					}, nil
				},
			}
			handler := NewUploadHandlerWithParser(mockStorage, mockParser)

			app := fiber.New()
			app.Post("/process", func(c fiber.Ctx) error {
				c.Locals("user_id", "user123")
				return handler.ProcessUpload(c)
			})

			reqBody := map[string]string{
				"file_key": fmt.Sprintf("uploads/user123/%s", tc.filename),
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedBank, result["bank_detected"])
		})
	}
}

// TestProcessUpload_DateRangeCalculation tests date range calculation with various transaction sets
func TestProcessUpload_DateRangeCalculation(t *testing.T) {
	testCases := []struct {
		name         string
		transactions []models.ParsedTransaction
		expectDates  bool
	}{
		{
			name: "Multiple transactions with date range",
			transactions: []models.ParsedTransaction{
				{Description: "Txn1", Amount: -100.00, TxnDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
				{Description: "Txn2", Amount: -200.00, TxnDate: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
				{Description: "Txn3", Amount: 500.00, TxnDate: time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC)},
			},
			expectDates: true,
		},
		{
			name: "Single transaction",
			transactions: []models.ParsedTransaction{
				{Description: "Txn1", Amount: -100.00, TxnDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			},
			expectDates: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage := &MockStorageService{
				DownloadFileFunc: func(key string) (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader([]byte("mock content"))), nil
				},
			}
			mockParser := &MockParser{
				ParseFileFunc: func(file io.Reader, filename string) ([]models.ParsedTransaction, error) {
					return tc.transactions, nil
				},
			}
			handler := NewUploadHandlerWithParser(mockStorage, mockParser)

			app := fiber.New()
			app.Post("/process", func(c fiber.Ctx) error {
				c.Locals("user_id", "user123")
				return handler.ProcessUpload(c)
			})

			reqBody := map[string]string{
				"file_key": "uploads/user123/test.csv",
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/process", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			assert.Contains(t, result, "date_range")
			dateRange := result["date_range"].(map[string]interface{})

			if tc.expectDates {
				assert.NotNil(t, dateRange["from"])
				assert.NotNil(t, dateRange["to"])
			}
		})
	}
}
