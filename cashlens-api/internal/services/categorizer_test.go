package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test exact matching
func TestCategorizer_MatchExact(t *testing.T) {
	c := &Categorizer{}

	tests := []struct {
		name        string
		description string
		keyword     string
		wantMatch   bool
		wantScore   float64
	}{
		{
			name:        "Exact match",
			description: "aws",
			keyword:     "aws",
			wantMatch:   true,
			wantScore:   1.0,
		},
		{
			name:        "No match - different case",
			description: "AWS",
			keyword:     "aws",
			wantMatch:   false,
			wantScore:   0.0,
		},
		{
			name:        "No match - substring",
			description: "aws services",
			keyword:     "aws",
			wantMatch:   false,
			wantScore:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotScore := c.matchExact(tt.description, tt.keyword)
			assert.Equal(t, tt.wantMatch, gotMatch)
			assert.Equal(t, tt.wantScore, gotScore)
		})
	}
}

// Test substring matching
func TestCategorizer_MatchSubstring(t *testing.T) {
	c := &Categorizer{}

	tests := []struct {
		name        string
		description string
		keyword     string
		wantMatch   bool
	}{
		{
			name:        "Substring match - exact",
			description: "aws",
			keyword:     "aws",
			wantMatch:   true,
		},
		{
			name:        "Substring match - middle",
			description: "amazon aws services",
			keyword:     "aws",
			wantMatch:   true,
		},
		{
			name:        "Substring match - start",
			description: "aws invoice payment",
			keyword:     "aws",
			wantMatch:   true,
		},
		{
			name:        "Substring match - end",
			description: "payment to aws",
			keyword:     "aws",
			wantMatch:   true,
		},
		{
			name:        "No match",
			description: "google cloud platform",
			keyword:     "aws",
			wantMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, _ := c.matchSubstring(tt.description, tt.keyword)
			assert.Equal(t, tt.wantMatch, gotMatch)
		})
	}
}

// Test regex matching
func TestCategorizer_MatchRegex(t *testing.T) {
	c := &Categorizer{}

	tests := []struct {
		name        string
		description string
		pattern     string
		wantMatch   bool
	}{
		{
			name:        "NEFT salary pattern",
			description: "NEFT SALARY CREDIT",
			pattern:     "^(NEFT|IMPS|RTGS).*(SALARY|SAL|EMP|PAYROLL)",
			wantMatch:   true,
		},
		{
			name:        "IMPS salary pattern",
			description: "IMPS-EMP-12345-SALARY",
			pattern:     "^(NEFT|IMPS|RTGS).*(SALARY|SAL|EMP|PAYROLL)",
			wantMatch:   true,
		},
		{
			name:        "RTGS payroll pattern",
			description: "RTGS PAYROLL TRANSFER",
			pattern:     "^(NEFT|IMPS|RTGS).*(SALARY|SAL|EMP|PAYROLL)",
			wantMatch:   true,
		},
		{
			name:        "No match - wrong prefix",
			description: "UPI SALARY CREDIT",
			pattern:     "^(NEFT|IMPS|RTGS).*(SALARY|SAL|EMP|PAYROLL)",
			wantMatch:   false,
		},
		{
			name:        "UPI Zomato pattern",
			description: "UPI/123456789/ZOMATO/PAYMENT",
			pattern:     "^UPI/.*/(ZOMATO|SWIGGY)",
			wantMatch:   true,
		},
		{
			name:        "UPI Swiggy pattern",
			description: "UPI/987654321/SWIGGY/ORDER",
			pattern:     "^UPI/.*/(ZOMATO|SWIGGY)",
			wantMatch:   true,
		},
		{
			name:        "Tax pattern",
			description: "TDS PAYABLE Q4",
			pattern:     ".*TDS.*PAYABLE",
			wantMatch:   true,
		},
		{
			name:        "Invalid regex",
			description: "test",
			pattern:     "[invalid(",
			wantMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, _ := c.matchRegex(tt.description, tt.pattern)
			assert.Equal(t, tt.wantMatch, gotMatch)
		})
	}
}

// Test fuzzy matching
func TestCategorizer_MatchFuzzy(t *testing.T) {
	c := &Categorizer{}

	tests := []struct {
		name        string
		description string
		keyword     string
		threshold   float64
		wantMatch   bool
	}{
		{
			name:        "Exact match via substring",
			description: "paytm payment",
			keyword:     "paytm",
			threshold:   0.7,
			wantMatch:   true,
		},
		{
			name:        "Fuzzy match - one character difference",
			description: "paytmm payment gateway",
			keyword:     "paytm",
			threshold:   0.7,
			wantMatch:   true,
		},
		{
			name:        "Fuzzy match - transposition",
			description: "patym service",
			keyword:     "paytm",
			threshold:   0.7,
			wantMatch:   true,
		},
		{
			name:        "Fuzzy match - Razorpay misspelling",
			description: "razorpay payment",
			keyword:     "razorpay",
			threshold:   0.7,
			wantMatch:   true,
		},
		{
			name:        "Fuzzy match - Flipkart misspelling",
			description: "flipcart order",
			keyword:     "flipkart",
			threshold:   0.7,
			wantMatch:   true,
		},
		{
			name:        "No match - too different",
			description: "google services",
			keyword:     "stripe",
			threshold:   0.7,
			wantMatch:   false,
		},
		{
			name:        "Match with lower threshold",
			description: "strpe payment",
			keyword:     "stripe",
			threshold:   0.6,
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, _ := c.matchFuzzy(tt.description, tt.keyword, tt.threshold)
			assert.Equal(t, tt.wantMatch, gotMatch)
		})
	}
}

// Test Levenshtein distance calculation
func TestCategorizer_LevenshteinDistance(t *testing.T) {
	c := &Categorizer{}

	tests := []struct {
		name     string
		s1       string
		s2       string
		wantDist int
	}{
		{
			name:     "Identical strings",
			s1:       "paytm",
			s2:       "paytm",
			wantDist: 0,
		},
		{
			name:     "One insertion",
			s1:       "paytm",
			s2:       "paytmm",
			wantDist: 1,
		},
		{
			name:     "One deletion",
			s1:       "paytm",
			s2:       "paym",
			wantDist: 1,
		},
		{
			name:     "One substitution",
			s1:       "paytm",
			s2:       "paytx",
			wantDist: 1,
		},
		{
			name:     "Transposition (2 operations)",
			s1:       "paytm",
			s2:       "patym",
			wantDist: 2,
		},
		{
			name:     "Multiple differences",
			s1:       "stripe",
			s2:       "strpe",
			wantDist: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDist := c.levenshteinDistance(tt.s1, tt.s2)
			assert.Equal(t, tt.wantDist, gotDist)
		})
	}
}

// Test similarity calculation
func TestCategorizer_CalculateSimilarity(t *testing.T) {
	c := &Categorizer{}

	tests := []struct {
		name           string
		s1             string
		s2             string
		wantSimilarity float64 // Approximate
	}{
		{
			name:           "Identical strings",
			s1:             "paytm",
			s2:             "paytm",
			wantSimilarity: 1.0,
		},
		{
			name:           "One character difference in 5-char word",
			s1:             "paytm",
			s2:             "paytx",
			wantSimilarity: 0.8, // 1 - (1/5)
		},
		{
			name:           "Empty strings",
			s1:             "",
			s2:             "",
			wantSimilarity: 1.0,
		},
		{
			name:           "One empty string",
			s1:             "test",
			s2:             "",
			wantSimilarity: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSimilarity := c.calculateSimilarity(tt.s1, tt.s2)
			assert.InDelta(t, tt.wantSimilarity, gotSimilarity, 0.01)
		})
	}
}

// Test matchDescription with priority handling
func TestCategorizer_MatchDescription(t *testing.T) {
	c := &Categorizer{}

	tests := []struct {
		name        string
		description string
		rules       []Rule
		wantCategory string
	}{
		{
			name:        "Single match",
			description: "aws invoice payment",
			rules: []Rule{
				{Keyword: "aws", Category: "Cloud & Hosting", Priority: 10, MatchType: "substring"},
			},
			wantCategory: "Cloud & Hosting",
		},
		{
			name:        "Multiple matches - higher priority wins",
			description: "aws salary payment",
			rules: []Rule{
				{Keyword: "aws", Category: "Cloud & Hosting", Priority: 5, MatchType: "substring"},
				{Keyword: "salary", Category: "Salaries", Priority: 10, MatchType: "substring"},
			},
			wantCategory: "Salaries",
		},
		{
			name:        "Multiple matches - same priority, first wins",
			description: "office rent payment",
			rules: []Rule{
				{Keyword: "office", Category: "Office Supplies", Priority: 5, MatchType: "substring"},
				{Keyword: "rent", Category: "Rent & Lease", Priority: 5, MatchType: "substring"},
			},
			wantCategory: "Office Supplies",
		},
		{
			name:        "Regex match",
			description: "NEFT SALARY CREDIT",
			rules: []Rule{
				{Keyword: "^(NEFT|IMPS|RTGS).*(SALARY|SAL)", Category: "Salaries", Priority: 10, MatchType: "regex"},
			},
			wantCategory: "Salaries",
		},
		{
			name:        "Fuzzy match",
			description: "paytmm payment gateway",
			rules: []Rule{
				{Keyword: "paytm", Category: "Payment Processing", Priority: 9, MatchType: "fuzzy", SimilarityThreshold: 0.7},
			},
			wantCategory: "Payment Processing",
		},
		{
			name:        "No match",
			description: "unknown transaction xyz",
			rules: []Rule{
				{Keyword: "aws", Category: "Cloud & Hosting", Priority: 10, MatchType: "substring"},
			},
			wantCategory: "",
		},
		{
			name:        "User rule overrides global rule",
			description: "aws payment",
			rules: []Rule{
				{Keyword: "aws", Category: "Custom Category", Priority: 100, MatchType: "substring", RuleType: "user"},
				{Keyword: "aws", Category: "Cloud & Hosting", Priority: 10, MatchType: "substring", RuleType: "global"},
			},
			wantCategory: "Custom Category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCategory := c.matchDescription(tt.description, tt.rules)
			assert.Equal(t, tt.wantCategory, gotCategory)
		})
	}
}

// Test realistic Indian bank transaction patterns
func TestCategorizer_IndianTransactions(t *testing.T) {
	c := &Categorizer{}

	// Realistic rule set
	rules := []Rule{
		// Salary patterns (regex)
		{Keyword: "^(NEFT|IMPS|RTGS).*(SALARY|SAL|EMP|PAYROLL)", Category: "Salaries", Priority: 10, MatchType: "regex"},
		{Keyword: "^UPI.*(SALARY|SAL|PAYROLL)", Category: "Salaries", Priority: 10, MatchType: "regex"},

		// UPI patterns (regex)
		{Keyword: "^UPI/.*/(ZOMATO|SWIGGY)", Category: "Team Meals", Priority: 4, MatchType: "regex"},
		{Keyword: "^UPI/.*/(OLA|UBER|RAPIDO)", Category: "Travel", Priority: 6, MatchType: "regex"},
		{Keyword: "^UPI/.*/(PAYTM|PHONEPE|GPAY)", Category: "Payment Processing", Priority: 9, MatchType: "regex"},

		// Fuzzy matching for common vendors
		{Keyword: "zomato", Category: "Team Meals", Priority: 4, MatchType: "fuzzy", SimilarityThreshold: 0.7},
		{Keyword: "swiggy", Category: "Team Meals", Priority: 4, MatchType: "fuzzy", SimilarityThreshold: 0.7},
		{Keyword: "razorpay", Category: "Payment Processing", Priority: 9, MatchType: "fuzzy", SimilarityThreshold: 0.7},
		{Keyword: "paytm", Category: "Payment Processing", Priority: 9, MatchType: "fuzzy", SimilarityThreshold: 0.7},

		// Substring matching
		{Keyword: "aws", Category: "Cloud & Hosting", Priority: 10, MatchType: "substring"},
		{Keyword: "electricity", Category: "Utilities", Priority: 7, MatchType: "substring"},
		{Keyword: "rent", Category: "Rent & Lease", Priority: 9, MatchType: "substring"},
	}

	tests := []struct {
		name         string
		description  string
		wantCategory string
	}{
		{
			name:         "NEFT salary",
			description:  "NEFT SALARY CREDIT EMP123",
			wantCategory: "Salaries",
		},
		{
			name:         "IMPS salary",
			description:  "IMPS-SALARY-TRANSFER-456",
			wantCategory: "Salaries",
		},
		{
			name:         "UPI Zomato",
			description:  "UPI/123456789/ZOMATO/FOOD-ORDER",
			wantCategory: "Team Meals",
		},
		{
			name:         "UPI Swiggy",
			description:  "UPI/987654/SWIGGY/DELIVERY",
			wantCategory: "Team Meals",
		},
		{
			name:         "UPI Uber",
			description:  "UPI/555/UBER/RIDE",
			wantCategory: "Travel",
		},
		{
			name:         "UPI PayTM",
			description:  "UPI/777/PAYTM/WALLET",
			wantCategory: "Payment Processing",
		},
		{
			name:         "Zomato with misspelling",
			description:  "zomatto food delivery",
			wantCategory: "Team Meals",
		},
		{
			name:         "PayTM with extra m",
			description:  "paytmm payment gateway",
			wantCategory: "Payment Processing",
		},
		{
			name:         "AWS services",
			description:  "AWS SERVICES INVOICE MARCH",
			wantCategory: "Cloud & Hosting",
		},
		{
			name:         "Electricity bill",
			description:  "ELECTRICITY BILL PAYMENT",
			wantCategory: "Utilities",
		},
		{
			name:         "Office rent",
			description:  "OFFICE RENT PAYMENT",
			wantCategory: "Rent & Lease",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descLower := strings.ToLower(tt.description)
			gotCategory := c.matchDescription(descLower, rules)
			assert.Equal(t, tt.wantCategory, gotCategory, "Failed to categorize: %s", tt.description)
		})
	}
}

// Benchmark test for performance
func BenchmarkCategorizer_MatchDescription(b *testing.B) {
	c := &Categorizer{}

	// Create 100 rules
	rules := make([]Rule, 100)
	for i := 0; i < 100; i++ {
		rules[i] = Rule{
			Keyword:   "keyword" + string(rune(i)),
			Category:  "Category" + string(rune(i)),
			Priority:  int32(i),
			MatchType: "substring",
		}
	}

	description := "test transaction with keyword50"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.matchDescription(description, rules)
	}
}
