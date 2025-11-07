package handlers

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/ashmitsharp/cashlens-api/internal/database/db"
	"github.com/ashmitsharp/cashlens-api/internal/models"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	// PresignedURLExpiryMinutes is the expiry time for presigned URLs in minutes
	PresignedURLExpiryMinutes = 15
	// PresignedURLExpirySeconds is the expiry time for presigned URLs in seconds
	PresignedURLExpirySeconds = PresignedURLExpiryMinutes * 60
)

var (
	// AllowedContentTypes defines the content types that are allowed for upload
	AllowedContentTypes = map[string]bool{
		"text/csv":                 true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"application/pdf": true,
	}
)

// StorageService interface defines methods for S3 operations
type StorageService interface {
	GenerateUploadKey(userID, filename string) (string, error)
	GeneratePresignedURL(key, contentType string, expiryMinutes int) (string, error)
	DownloadFile(key string) (io.ReadCloser, error)
}

// Parser interface defines methods for parsing bank statement files
type Parser interface {
	ParseFile(file io.Reader, filename string) ([]models.ParsedTransaction, error)
}

// Categorizer interface defines methods for categorizing transactions
type Categorizer interface {
	Categorize(ctx context.Context, description string, userID uuid.UUID) (string, error)
	LoadGlobalRules(ctx context.Context) error
	InvalidateUserCache(userID uuid.UUID)
	GetStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
}

// UploadHandler handles file upload-related requests
type UploadHandler struct {
	storage     StorageService
	parser      Parser
	categorizer Categorizer
	db          *db.Queries
}

// NewUploadHandler creates a new upload handler instance (backward compatible)
func NewUploadHandler(storage StorageService) *UploadHandler {
	return &UploadHandler{
		storage:     storage,
		parser:      nil,
		categorizer: nil,
		db:          nil,
	}
}

// NewUploadHandlerWithParser creates a new upload handler with parser support
func NewUploadHandlerWithParser(storage StorageService, parser Parser) *UploadHandler {
	return &UploadHandler{
		storage:     storage,
		parser:      parser,
		categorizer: nil,
		db:          nil,
	}
}

// NewUploadHandlerFull creates a fully configured upload handler
func NewUploadHandlerFull(storage StorageService, parser Parser, categorizer Categorizer, database *db.Queries) *UploadHandler {
	return &UploadHandler{
		storage:     storage,
		parser:      parser,
		categorizer: categorizer,
		db:          database,
	}
}

// GetPresignedURL generates a presigned URL for file upload
// Query params: filename (required), content_type (required)
// Returns: upload_url, file_key, expires_in
func (h *UploadHandler) GetPresignedURL(c fiber.Ctx) error {
	// 1. Get query parameters
	filename := c.Query("filename")
	contentType := c.Query("content_type")

	// 2. Validate filename
	if filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "filename is required",
		})
	}

	// 3. Validate content_type
	if contentType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "content_type is required",
		})
	}

	// 4. Validate content type against allowed types
	if !AllowedContentTypes[contentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unsupported file type",
		})
	}

	// 5. Get user_id from context (set by auth middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user_id not found",
		})
	}

	// 6. Generate upload key
	key, err := h.storage.GenerateUploadKey(userID, filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to generate upload key",
			"details": err.Error(),
		})
	}

	// 7. Generate presigned URL
	url, err := h.storage.GeneratePresignedURL(key, contentType, PresignedURLExpiryMinutes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to generate presigned URL",
			"details": err.Error(),
		})
	}

	// 8. Return successful response
	return c.JSON(fiber.Map{
		"upload_url": url,
		"file_key":   key,
		"expires_in": PresignedURLExpirySeconds,
	})
}

// ProcessUploadRequest represents the request body for ProcessUpload
type ProcessUploadRequest struct {
	FileKey string `json:"file_key"`
}

// ProcessUpload processes an uploaded file from S3 and returns summary statistics
// POST /v1/upload/process
// Body: {"file_key": "uploads/user123/1699564800-uuid-statement.csv"}
func (h *UploadHandler) ProcessUpload(c fiber.Ctx) error {
	// 1. Parse request body
	var req ProcessUploadRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// 2. Validate file_key
	if req.FileKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file_key is required",
		})
	}

	// 3. Authenticate and authorize
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user_id not found",
		})
	}

	// 4. Security check: Verify file belongs to user
	if !isFileOwnedByUser(req.FileKey, userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden - cannot access this file",
		})
	}

	// 5. Download file from S3
	reader, err := h.storage.DownloadFile(req.FileKey)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "file not found in storage",
		})
	}
	defer reader.Close()

	// 6. Parse file and extract transactions
	filename := filepath.Base(req.FileKey)
	transactions, err := h.parser.ParseFile(reader, filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "failed to parse file",
			"details": err.Error(),
		})
	}

	// 7. Categorize and save transactions if categorizer and db are available
	var categorizedCount int
	var accuracyPercent float64

	if h.categorizer != nil && h.db != nil {
		// Parse user ID from string to UUID
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "invalid user ID format",
			})
		}

		// Ensure categorizer has loaded global rules
		if err := h.categorizer.LoadGlobalRules(c.Context()); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "failed to load categorization rules",
				"details": err.Error(),
			})
		}

		// Categorize and save each transaction
		for _, txn := range transactions {
			// Categorize transaction
			category, err := h.categorizer.Categorize(c.Context(), txn.Description, userUUID)
			if err != nil {
				// Log error but continue processing
				fmt.Printf("Failed to categorize transaction: %v\n", err)
				category = ""
			}

			if category != "" {
				categorizedCount++
			}

			// Convert uuid.UUID to pgtype.UUID
			var pgUserID pgtype.UUID
			pgUserID.Bytes = userUUID
			pgUserID.Valid = true

			// Convert float64 to pgtype.Numeric
			var pgAmount pgtype.Numeric
			pgAmount.Scan(txn.Amount)

			// Determine transaction type
			txnType := txn.TxnType
			if txnType == "" {
				if txn.Amount > 0 {
					txnType = "credit"
				} else {
					txnType = "debit"
				}
			}

			// Save transaction to database
			_, err = h.db.CreateTransaction(c.Context(), db.CreateTransactionParams{
				UserID:      pgUserID,
				TxnDate:     pgtype.Date{Time: txn.TxnDate, Valid: true},
				Description: txn.Description,
				Amount:      pgAmount,
				TxnType:     txnType,
				Category:    pgtype.Text{String: category, Valid: category != ""},
				IsReviewed:  false,
				RawData:     pgtype.Text{String: txn.RawData, Valid: txn.RawData != ""},
			})

			if err != nil {
				// Log error but continue processing other transactions
				fmt.Printf("Failed to save transaction: %v\n", err)
			}
		}

		// Calculate accuracy
		if len(transactions) > 0 {
			accuracyPercent = (float64(categorizedCount) / float64(len(transactions))) * 100
		}
	}

	// 8. Build and return summary response
	summary := buildProcessSummaryWithCategorization(req.FileKey, filename, transactions, categorizedCount, accuracyPercent)
	return c.JSON(summary)
}

// isFileOwnedByUser checks if a file key belongs to the specified user
func isFileOwnedByUser(fileKey, userID string) bool {
	expectedPrefix := fmt.Sprintf("uploads/%s/", userID)
	return strings.HasPrefix(fileKey, expectedPrefix)
}

// buildProcessSummary creates the summary response from parsed transactions (deprecated)
func buildProcessSummary(fileKey, filename string, transactions []models.ParsedTransaction) fiber.Map {
	return buildProcessSummaryWithCategorization(fileKey, filename, transactions, 0, 0.0)
}

// buildProcessSummaryWithCategorization creates the summary response with categorization stats
func buildProcessSummaryWithCategorization(fileKey, filename string, transactions []models.ParsedTransaction, categorizedCount int, accuracyPercent float64) fiber.Map {
	totalRows := len(transactions)

	// Calculate date range
	dateRange := calculateDateRange(transactions)

	// Detect bank from filename
	bank := detectBankFromFilename(filename)

	uncategorizedCount := totalRows - categorizedCount

	return fiber.Map{
		"file_key":            fileKey,
		"total_rows":          totalRows,
		"categorized":         categorizedCount,
		"uncategorized":       uncategorizedCount,
		"accuracy_percent":    accuracyPercent,
		"transactions_parsed": totalRows,
		"bank_detected":       bank,
		"date_range":          dateRange,
		"status":              "success",
	}
}

// calculateDateRange finds the earliest and latest transaction dates
func calculateDateRange(transactions []models.ParsedTransaction) fiber.Map {
	var minDate, maxDate time.Time

	if len(transactions) > 0 {
		minDate = transactions[0].TxnDate
		maxDate = transactions[0].TxnDate

		for _, txn := range transactions {
			if txn.TxnDate.Before(minDate) {
				minDate = txn.TxnDate
			}
			if txn.TxnDate.After(maxDate) {
				maxDate = txn.TxnDate
			}
		}
	}

	return fiber.Map{
		"from": minDate,
		"to":   maxDate,
	}
}

// detectBankFromFilename attempts to detect the bank from the filename
func detectBankFromFilename(filename string) string {
	lowerFilename := strings.ToLower(filename)

	bankKeywords := map[string]string{
		"hdfc":  "HDFC",
		"icici": "ICICI",
		"sbi":   "SBI",
		"axis":  "Axis",
		"kotak": "Kotak",
	}

	for keyword, bankName := range bankKeywords {
		if strings.Contains(lowerFilename, keyword) {
			return bankName
		}
	}

	// Default to UNKNOWN if no bank detected
	return "UNKNOWN"
}
