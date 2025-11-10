package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ashmitsharp/cashlens-api/internal/database/db"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// TransactionHandler handles transaction-related requests
type TransactionHandler struct {
	db          *db.Queries
	categorizer Categorizer
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(database *db.Queries, categorizer Categorizer) *TransactionHandler {
	return &TransactionHandler{
		db:          database,
		categorizer: categorizer,
	}
}

// getUserUUIDFromClerkID looks up the user's database UUID from their Clerk ID
func (h *TransactionHandler) getUserUUIDFromClerkID(ctx context.Context, clerkUserID string) (uuid.UUID, error) {
	user, err := h.db.GetUserByClerkID(ctx, clerkUserID)
	if err != nil {
		return uuid.Nil, err
	}

	var userUUID uuid.UUID
	copy(userUUID[:], user.ID.Bytes[:])
	return userUUID, nil
}

// GetTransactions returns transactions with optional filtering
// GET /v1/transactions?status=all|categorized|uncategorized&limit=50&offset=0
func (h *TransactionHandler) GetTransactions(c fiber.Ctx) error {
	// 1. Get clerk_user_id from context
	clerkUserID, ok := c.Locals("clerk_user_id").(string)
	if !ok || clerkUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user not authenticated",
		})
	}

	// 2. Look up user's UUID from clerk_user_id
	userUUID, err := h.getUserUUIDFromClerkID(c.Context(), clerkUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found in database",
		})
	}

	// 3. Parse query parameters
	status := c.Query("status", "all")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 4. Convert to pgtype.UUID
	var pgUserID pgtype.UUID
	pgUserID.Bytes = userUUID
	pgUserID.Valid = true

	// 4. Query transactions based on status filter
	var transactions interface{}
	var totalCount int64

	switch status {
	case "uncategorized":
		transactions, err = h.db.GetUncategorizedTransactions(c.Context(), db.GetUncategorizedTransactionsParams{
			UserID: pgUserID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to fetch uncategorized transactions",
			})
		}
		totalCount, _ = h.db.CountUncategorizedTransactions(c.Context(), pgUserID)

	case "categorized":
		transactions, err = h.db.GetCategorizedTransactions(c.Context(), db.GetCategorizedTransactionsParams{
			UserID: pgUserID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to fetch categorized transactions",
			})
		}
		totalCount, _ = h.db.CountCategorizedTransactions(c.Context(), pgUserID)

	default: // "all"
		transactions, err = h.db.GetUserTransactions(c.Context(), db.GetUserTransactionsParams{
			UserID: pgUserID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to fetch transactions",
			})
		}
		totalCount, _ = h.db.CountUserTransactions(c.Context(), pgUserID)
	}

	// 5. Return response
	return c.JSON(fiber.Map{
		"transactions": transactions,
		"total":        totalCount,
		"limit":        limit,
		"offset":       offset,
	})
}

// UpdateTransactionRequest represents the request body for updating a transaction
type UpdateTransactionRequest struct {
	Category string `json:"category"`
}

// UpdateTransaction updates a transaction's category and marks it as reviewed
// PUT /v1/transactions/:id
func (h *TransactionHandler) UpdateTransaction(c fiber.Ctx) error {
	// 1. Get clerk_user_id from context
	clerkUserID, ok := c.Locals("clerk_user_id").(string)
	if !ok || clerkUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user not authenticated",
		})
	}

	// 1.5. Look up user's UUID
	userUUID, err := h.getUserUUIDFromClerkID(c.Context(), clerkUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found in database",
		})
	}

	// 2. Get transaction ID from URL
	txnIDStr := c.Params("id")
	txnID, err := uuid.Parse(txnIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid transaction ID",
		})
	}

	// 3. Parse request body
	var req UpdateTransactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// 4. Validate category
	if req.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "category is required",
		})
	}

	// 5. Convert to pgtype.UUID
	var pgTxnID pgtype.UUID
	pgTxnID.Bytes = txnID
	pgTxnID.Valid = true

	// 6. Get transaction to verify ownership
	transaction, err := h.db.GetTransactionByID(c.Context(), pgTxnID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "transaction not found",
		})
	}

	// 7. Verify user owns this transaction
	var transactionUserID uuid.UUID
	copy(transactionUserID[:], transaction.UserID.Bytes[:])
	if transactionUserID != userUUID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden - cannot update this transaction",
		})
	}

	// 8. Update transaction
	updated, err := h.db.UpdateTransactionCategory(c.Context(), db.UpdateTransactionCategoryParams{
		ID:         pgTxnID,
		Category:   pgtype.Text{String: req.Category, Valid: true},
		IsReviewed: true,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update transaction",
		})
	}

	// 9. Invalidate user cache to force reload with new rule
	if h.categorizer != nil {
		h.categorizer.InvalidateUserCache(userUUID)
	}

	// 10. Return updated transaction
	return c.JSON(fiber.Map{
		"transaction": updated,
		"message":     "Transaction updated successfully",
	})
}

// BulkUpdateRequest represents the request body for bulk updating transactions
type BulkUpdateRequest struct {
	TransactionIDs []string `json:"transaction_ids"`
	Category       string   `json:"category"`
}

// BulkUpdateTransactions updates multiple transactions at once
// PUT /v1/transactions/bulk
func (h *TransactionHandler) BulkUpdateTransactions(c fiber.Ctx) error {
	// 1. Get clerk_user_id from context
	clerkUserID, ok := c.Locals("clerk_user_id").(string)
	if !ok || clerkUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user not authenticated",
		})
	}

	// 1.5. Look up user's UUID
	userUUID, err := h.getUserUUIDFromClerkID(c.Context(), clerkUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found in database",
		})
	}

	// 2. Parse request body
	var req BulkUpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// 3. Validate request
	if len(req.TransactionIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "transaction_ids cannot be empty",
		})
	}

	if req.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "category is required",
		})
	}

	// 4. Update each transaction
	updatedCount := 0
	failedIDs := []string{}

	for _, txnIDStr := range req.TransactionIDs {
		txnID, err := uuid.Parse(txnIDStr)
		if err != nil {
			failedIDs = append(failedIDs, txnIDStr)
			continue
		}

		// Convert to pgtype.UUID
		var pgTxnID pgtype.UUID
		pgTxnID.Bytes = txnID
		pgTxnID.Valid = true

		// Get transaction to verify ownership
		transaction, err := h.db.GetTransactionByID(c.Context(), pgTxnID)
		if err != nil {
			failedIDs = append(failedIDs, txnIDStr)
			continue
		}

		// Verify user owns this transaction
		var transactionUserID uuid.UUID
		copy(transactionUserID[:], transaction.UserID.Bytes[:])
		if transactionUserID != userUUID {
			failedIDs = append(failedIDs, txnIDStr)
			continue
		}

		// Update transaction
		_, err = h.db.UpdateTransactionCategory(c.Context(), db.UpdateTransactionCategoryParams{
			ID:         pgTxnID,
			Category:   pgtype.Text{String: req.Category, Valid: true},
			IsReviewed: true,
		})
		if err != nil {
			failedIDs = append(failedIDs, txnIDStr)
			continue
		}

		updatedCount++
	}

	// 5. Invalidate user cache
	if h.categorizer != nil {
		h.categorizer.InvalidateUserCache(userUUID)
	}

	// 6. Return response
	response := fiber.Map{
		"updated_count": updatedCount,
		"total_count":   len(req.TransactionIDs),
		"message":       fmt.Sprintf("Successfully updated %d transactions", updatedCount),
	}

	if len(failedIDs) > 0 {
		response["failed_ids"] = failedIDs
		response["failed_count"] = len(failedIDs)
	}

	return c.JSON(response)
}

// GetTransactionStats returns categorization statistics for the user
// GET /v1/transactions/stats
func (h *TransactionHandler) GetTransactionStats(c fiber.Ctx) error {
	// 1. Get clerk_user_id from context
	clerkUserID, ok := c.Locals("clerk_user_id").(string)
	if !ok || clerkUserID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user not authenticated",
		})
	}

	// 2. Look up user's UUID
	userUUID, err := h.getUserUUIDFromClerkID(c.Context(), clerkUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found in database",
		})
	}

	// Convert to pgtype.UUID
	var pgUserID pgtype.UUID
	pgUserID.Bytes = userUUID
	pgUserID.Valid = true

	// 3. Get stats
	stats, err := h.db.GetTransactionStats(c.Context(), pgUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch transaction statistics",
		})
	}

	// 4. Return stats
	return c.JSON(fiber.Map{
		"total_transactions":      stats.TotalCount,
		"categorized_count":       stats.CategorizedCount,
		"uncategorized_count":     stats.UncategorizedCount,
		"accuracy_percent":        stats.AccuracyPercent,
	})
}
