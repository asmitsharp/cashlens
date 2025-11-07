package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ashmitsharp/cashlens-api/internal/database/db"
)

// Rule represents a categorization rule
type Rule struct {
	ID                  uuid.UUID
	Keyword             string
	Category            string
	Priority            int32
	MatchType           string  // substring, regex, exact, fuzzy
	SimilarityThreshold float64 // For fuzzy matching (0-1)
	RuleType            string  // global or user
}

// Categorizer handles transaction categorization
type Categorizer struct {
	db          *db.Queries
	globalRules []Rule
	userRules   map[uuid.UUID][]Rule // Cache of user rules by user_id
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration
	lastLoaded  time.Time
}

// NewCategorizer creates a new categorizer instance
func NewCategorizer(database *db.Queries) *Categorizer {
	return &Categorizer{
		db:         database,
		userRules:  make(map[uuid.UUID][]Rule),
		cacheTTL:   5 * time.Minute, // Cache rules for 5 minutes
		lastLoaded: time.Time{},
	}
}

// LoadGlobalRules loads all global rules from database into memory
func (c *Categorizer) LoadGlobalRules(ctx context.Context) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	// Check if cache is still valid
	if time.Since(c.lastLoaded) < c.cacheTTL && len(c.globalRules) > 0 {
		return nil
	}

	dbRules, err := c.db.GetAllGlobalRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to load global rules: %w", err)
	}

	c.globalRules = make([]Rule, 0, len(dbRules))
	for _, r := range dbRules {
		// Convert pgtype.UUID to uuid.UUID
		var id uuid.UUID
		copy(id[:], r.ID.Bytes[:])

		// Convert pgtype.Numeric to float64
		similarity, _ := r.SimilarityThreshold.Float64Value()

		c.globalRules = append(c.globalRules, Rule{
			ID:                  id,
			Keyword:             r.Keyword,
			Category:            r.Category,
			Priority:            r.Priority.Int32,
			MatchType:           r.MatchType.String,
			SimilarityThreshold: similarity.Float64,
			RuleType:            "global",
		})
	}

	c.lastLoaded = time.Now()
	return nil
}

// LoadUserRules loads user-specific rules and caches them
func (c *Categorizer) LoadUserRules(ctx context.Context, userID uuid.UUID) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	// Convert uuid.UUID to pgtype.UUID
	var pgUserID pgtype.UUID
	pgUserID.Bytes = userID
	pgUserID.Valid = true

	dbRules, err := c.db.GetUserRules(ctx, pgUserID)
	if err != nil {
		return fmt.Errorf("failed to load user rules: %w", err)
	}

	rules := make([]Rule, 0, len(dbRules))
	for _, r := range dbRules {
		// Convert pgtype.UUID to uuid.UUID
		var id uuid.UUID
		copy(id[:], r.ID.Bytes[:])

		// Convert pgtype.Numeric to float64
		similarity, _ := r.SimilarityThreshold.Float64Value()

		rules = append(rules, Rule{
			ID:                  id,
			Keyword:             r.Keyword,
			Category:            r.Category,
			Priority:            r.Priority.Int32,
			MatchType:           r.MatchType.String,
			SimilarityThreshold: similarity.Float64,
			RuleType:            "user",
		})
	}

	c.userRules[userID] = rules
	return nil
}

// InvalidateUserCache clears the cached rules for a specific user
func (c *Categorizer) InvalidateUserCache(userID uuid.UUID) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.userRules, userID)
}

// Categorize attempts to categorize a transaction description
// Returns category string or empty string if no match
func (c *Categorizer) Categorize(ctx context.Context, description string, userID uuid.UUID) (string, error) {
	// Ensure global rules are loaded
	if err := c.LoadGlobalRules(ctx); err != nil {
		return "", err
	}

	// Load user rules if not cached
	c.cacheMutex.RLock()
	_, userRulesExist := c.userRules[userID]
	c.cacheMutex.RUnlock()

	if !userRulesExist {
		if err := c.LoadUserRules(ctx, userID); err != nil {
			return "", err
		}
	}

	// Get combined rules (user rules have higher priority)
	c.cacheMutex.RLock()
	userRules := c.userRules[userID]
	globalRules := c.globalRules
	c.cacheMutex.RUnlock()

	// Combine rules: user rules first (higher priority)
	allRules := append(userRules, globalRules...)

	// Match description against rules
	return c.matchDescription(description, allRules), nil
}

// matchDescription finds the best matching rule for a description
func (c *Categorizer) matchDescription(description string, rules []Rule) string {
	descUpper := strings.ToUpper(strings.TrimSpace(description))
	descLower := strings.ToLower(descUpper)

	var bestMatch string
	highestPriority := int32(-1)
	highestScore := 0.0

	for _, rule := range rules {
		var matched bool
		var score float64

		// For regex, use uppercase (Indian bank CSVs are usually uppercase)
		// For other match types, use lowercase
		if rule.MatchType == "regex" {
			matched, score = c.matchRule(descUpper, rule)
		} else {
			matched, score = c.matchRule(descLower, rule)
		}

		if matched {
			// Higher priority wins
			if rule.Priority > highestPriority {
				bestMatch = rule.Category
				highestPriority = rule.Priority
				highestScore = score
			} else if rule.Priority == highestPriority && score > highestScore {
				// If priority is equal, higher score wins
				bestMatch = rule.Category
				highestScore = score
			}
		}
	}

	return bestMatch
}

// matchRule checks if a description matches a rule based on match_type
func (c *Categorizer) matchRule(description string, rule Rule) (bool, float64) {
	switch rule.MatchType {
	case "exact":
		keywordLower := strings.ToLower(rule.Keyword)
		return c.matchExact(description, keywordLower)
	case "substring":
		keywordLower := strings.ToLower(rule.Keyword)
		return c.matchSubstring(description, keywordLower)
	case "regex":
		// Don't lowercase regex patterns
		return c.matchRegex(description, rule.Keyword)
	case "fuzzy":
		keywordLower := strings.ToLower(rule.Keyword)
		return c.matchFuzzy(description, keywordLower, rule.SimilarityThreshold)
	default:
		// Default to substring matching
		keywordLower := strings.ToLower(rule.Keyword)
		return c.matchSubstring(description, keywordLower)
	}
}

// matchExact performs exact string matching
func (c *Categorizer) matchExact(description, keyword string) (bool, float64) {
	if description == keyword {
		return true, 1.0
	}
	return false, 0.0
}

// matchSubstring performs case-insensitive substring matching
func (c *Categorizer) matchSubstring(description, keyword string) (bool, float64) {
	if strings.Contains(description, keyword) {
		// Calculate score based on match position and length
		score := float64(len(keyword)) / float64(len(description))
		return true, score
	}
	return false, 0.0
}

// matchRegex performs regular expression matching
func (c *Categorizer) matchRegex(description, pattern string) (bool, float64) {
	// Compile regex (in production, cache compiled regexes)
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Invalid regex, skip this rule
		return false, 0.0
	}

	if re.MatchString(description) {
		// For regex, score is 0.8 (slightly lower than exact match)
		return true, 0.8
	}
	return false, 0.0
}

// matchFuzzy performs fuzzy string matching using Levenshtein distance
func (c *Categorizer) matchFuzzy(description, keyword string, threshold float64) (bool, float64) {
	// Check if keyword exists as substring first (fast path)
	if strings.Contains(description, keyword) {
		return true, 1.0
	}

	// Calculate similarity for the entire description
	similarity := c.calculateSimilarity(description, keyword)
	if similarity >= threshold {
		return true, similarity
	}

	// Also check against individual words in description
	words := strings.Fields(description)
	maxSimilarity := 0.0
	for _, word := range words {
		wordSimilarity := c.calculateSimilarity(word, keyword)
		if wordSimilarity > maxSimilarity {
			maxSimilarity = wordSimilarity
		}
		if wordSimilarity >= threshold {
			return true, wordSimilarity
		}
	}

	return false, maxSimilarity
}

// calculateSimilarity computes similarity score using Levenshtein distance
// Returns a value between 0 and 1, where 1 is identical
func (c *Categorizer) calculateSimilarity(s1, s2 string) float64 {
	// Handle empty strings
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Calculate Levenshtein distance
	distance := c.levenshteinDistance(s1, s2)

	// Convert distance to similarity (0-1 scale)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	similarity := 1.0 - (float64(distance) / float64(maxLen))
	return similarity
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (c *Categorizer) levenshteinDistance(s1, s2 string) int {
	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// GetStats returns categorization statistics
func (c *Categorizer) GetStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	// Convert uuid.UUID to pgtype.UUID
	var pgUserID pgtype.UUID
	pgUserID.Bytes = userID
	pgUserID.Valid = true

	stats, err := c.db.GetRuleStats(ctx, pgUserID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"global_rules_count":      stats.GlobalRulesCount,
		"user_rules_count":        stats.UserRulesCount,
		"global_categories_count": stats.GlobalCategoriesCount,
		"user_categories_count":   stats.UserCategoriesCount,
		"cache_size":              len(c.userRules),
	}, nil
}
