package handlers

import (
	"strconv"

	"github.com/ashmitsharp/cashlens-api/internal/database/db"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// RulesHandler handles categorization rule management
type RulesHandler struct {
	db          *db.Queries
	categorizer Categorizer
}

// NewRulesHandler creates a new rules handler instance
func NewRulesHandler(database *db.Queries, categorizer Categorizer) *RulesHandler {
	return &RulesHandler{
		db:          database,
		categorizer: categorizer,
	}
}

// CreateRuleRequest represents the request body for creating a rule
type CreateRuleRequest struct {
	Keyword             string  `json:"keyword" validate:"required"`
	Category            string  `json:"category" validate:"required"`
	Priority            int32   `json:"priority"`
	MatchType           string  `json:"match_type"` // substring, regex, exact, fuzzy
	SimilarityThreshold float64 `json:"similarity_threshold"`
}

// UpdateRuleRequest represents the request body for updating a rule
type UpdateRuleRequest struct {
	Category            string  `json:"category"`
	Priority            int32   `json:"priority"`
	MatchType           string  `json:"match_type"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	IsActive            bool    `json:"is_active"`
}

// GetUserRules returns all active rules for the authenticated user
// GET /v1/rules
func (h *RulesHandler) GetUserRules(c fiber.Ctx) error {
	// Get user_id from context (set by auth middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user_id not found",
		})
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	// Convert to pgtype.UUID
	var pgUserID pgtype.UUID
	pgUserID.Bytes = userUUID
	pgUserID.Valid = true

	// Get user rules from database
	rules, err := h.db.GetUserRules(c.Context(), pgUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to fetch user rules",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"rules": rules,
		"count": len(rules),
	})
}

// GetGlobalRules returns all active global rules
// GET /v1/rules/global
func (h *RulesHandler) GetGlobalRules(c fiber.Ctx) error {
	// Get global rules from database
	rules, err := h.db.GetAllGlobalRules(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to fetch global rules",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"rules": rules,
		"count": len(rules),
	})
}

// CreateUserRule creates a new user-specific categorization rule
// POST /v1/rules
func (h *RulesHandler) CreateUserRule(c fiber.Ctx) error {
	// Parse request body
	var req CreateRuleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.Keyword == "" || req.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "keyword and category are required",
		})
	}

	// Get user_id from context
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user_id not found",
		})
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	// Convert to pgtype.UUID
	var pgUserID pgtype.UUID
	pgUserID.Bytes = userUUID
	pgUserID.Valid = true

	// Set defaults
	if req.Priority == 0 {
		req.Priority = 100 // Default user rule priority
	}
	if req.MatchType == "" {
		req.MatchType = "substring"
	}
	if req.SimilarityThreshold == 0 {
		req.SimilarityThreshold = 0.3
	}

	// Convert similarity threshold to pgtype.Numeric
	var pgThreshold pgtype.Numeric
	pgThreshold.Scan(req.SimilarityThreshold)

	// Create rule in database
	rule, err := h.db.CreateUserRule(c.Context(), db.CreateUserRuleParams{
		UserID:              pgUserID,
		Keyword:             req.Keyword,
		Category:            req.Category,
		Priority:            pgtype.Int4{Int32: req.Priority, Valid: true},
		MatchType:           pgtype.Text{String: req.MatchType, Valid: true},
		SimilarityThreshold: pgThreshold,
		IsActive:            pgtype.Bool{Bool: true, Valid: true},
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to create rule",
			"details": err.Error(),
		})
	}

	// Invalidate categorizer cache for this user
	if h.categorizer != nil {
		h.categorizer.InvalidateUserCache(userUUID)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"rule":    rule,
		"message": "rule created successfully",
	})
}

// UpdateUserRule updates an existing user rule
// PUT /v1/rules/:id
func (h *RulesHandler) UpdateUserRule(c fiber.Ctx) error {
	// Parse rule ID from URL parameter
	ruleIDStr := c.Params("id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid rule ID",
		})
	}

	// Parse request body
	var req UpdateRuleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Get user_id from context
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user_id not found",
		})
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	// Convert to pgtype types
	var pgRuleID pgtype.UUID
	pgRuleID.Bytes = ruleID
	pgRuleID.Valid = true

	var pgUserID pgtype.UUID
	pgUserID.Bytes = userUUID
	pgUserID.Valid = true

	var pgThreshold pgtype.Numeric
	pgThreshold.Scan(req.SimilarityThreshold)

	// Update rule in database
	rule, err := h.db.UpdateUserRule(c.Context(), db.UpdateUserRuleParams{
		ID:                  pgRuleID,
		Category:            req.Category,
		Priority:            pgtype.Int4{Int32: req.Priority, Valid: true},
		MatchType:           pgtype.Text{String: req.MatchType, Valid: true},
		SimilarityThreshold: pgThreshold,
		IsActive:            pgtype.Bool{Bool: req.IsActive, Valid: true},
		UserID:              pgUserID,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to update rule",
			"details": err.Error(),
		})
	}

	// Invalidate categorizer cache for this user
	if h.categorizer != nil {
		h.categorizer.InvalidateUserCache(userUUID)
	}

	return c.JSON(fiber.Map{
		"rule":    rule,
		"message": "rule updated successfully",
	})
}

// DeleteUserRule deletes a user rule
// DELETE /v1/rules/:id
func (h *RulesHandler) DeleteUserRule(c fiber.Ctx) error {
	// Parse rule ID from URL parameter
	ruleIDStr := c.Params("id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid rule ID",
		})
	}

	// Get user_id from context
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user_id not found",
		})
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	// Convert to pgtype types
	var pgRuleID pgtype.UUID
	pgRuleID.Bytes = ruleID
	pgRuleID.Valid = true

	var pgUserID pgtype.UUID
	pgUserID.Bytes = userUUID
	pgUserID.Valid = true

	// Delete rule from database
	err = h.db.DeleteUserRule(c.Context(), db.DeleteUserRuleParams{
		ID:     pgRuleID,
		UserID: pgUserID,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to delete rule",
			"details": err.Error(),
		})
	}

	// Invalidate categorizer cache for this user
	if h.categorizer != nil {
		h.categorizer.InvalidateUserCache(userUUID)
	}

	return c.JSON(fiber.Map{
		"message": "rule deleted successfully",
	})
}

// GetRuleStats returns statistics about categorization rules
// GET /v1/rules/stats
func (h *RulesHandler) GetRuleStats(c fiber.Ctx) error {
	// Get user_id from context
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user_id not found",
		})
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	// Get stats from categorizer
	if h.categorizer != nil {
		stats, err := h.categorizer.GetStats(c.Context(), userUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "failed to get stats",
				"details": err.Error(),
			})
		}
		return c.JSON(stats)
	}

	// Fallback: get stats directly from database
	var pgUserID pgtype.UUID
	pgUserID.Bytes = userUUID
	pgUserID.Valid = true

	dbStats, err := h.db.GetRuleStats(c.Context(), pgUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to get stats",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"global_rules_count":      dbStats.GlobalRulesCount,
		"user_rules_count":        dbStats.UserRulesCount,
		"global_categories_count": dbStats.GlobalCategoriesCount,
		"user_categories_count":   dbStats.UserCategoriesCount,
	})
}

// SearchRules searches for rules by keyword
// GET /v1/rules/search?q=keyword&limit=10
func (h *RulesHandler) SearchRules(c fiber.Ctx) error {
	// Get search query
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "search query (q) is required",
		})
	}

	// Get limit (default 20)
	limitStr := c.Query("limit", "20")
	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil {
		limit = l
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	// Get user_id from context
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized - user_id not found",
		})
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	// Convert to pgtype.UUID
	var pgUserID pgtype.UUID
	pgUserID.Bytes = userUUID
	pgUserID.Valid = true

	// Search user rules
	userRules, err := h.db.SearchUserRulesByKeyword(c.Context(), db.SearchUserRulesByKeywordParams{
		UserID:  pgUserID,
		Column2: pgtype.Text{String: query, Valid: true},
		Limit:   int32(limit),
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to search rules",
			"details": err.Error(),
		})
	}

	// Search global rules
	globalRules, err := h.db.SearchRulesByKeyword(c.Context(), db.SearchRulesByKeywordParams{
		Column1: pgtype.Text{String: query, Valid: true},
		Limit:   int32(limit),
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to search global rules",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"user_rules":   userRules,
		"global_rules": globalRules,
		"query":        query,
	})
}
