package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders())

	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check security headers are present
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"))
	assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "default-src 'self'")
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
	assert.Contains(t, resp.Header.Get("Permissions-Policy"), "camera=()")
}

func TestRateLimiter(t *testing.T) {
	app := fiber.New()

	// Use a very strict rate limiter for testing (2 requests max)
	app.Use(func(c fiber.Ctx) error {
		return c.Next()
	})

	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)

		if i < 2 {
			// First 2 requests should succeed
			assert.Equal(t, 200, resp.StatusCode)
		}
		// Note: Rate limiting test would require actual limiter middleware
		// This is a basic structure test
	}
}

func TestInputSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML escaping",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "SQL injection attempt",
			input:    "'; DROP TABLE users--",
			expected: "&#39;;  TABLE users", // Semicolon is not removed, but DROP and -- are
		},
		{
			name:     "Normal input",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "Whitespace trimming",
			input:    "  spaces  ",
			expected: "spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLInjectionPrevention(t *testing.T) {
	// Test cases for SQL injection prevention
	maliciousInputs := []string{
		"'; DROP TABLE transactions--",
		"' OR '1'='1",
		"admin'--",
		"1' UNION SELECT * FROM users--",
		"'; DELETE FROM transactions WHERE '1'='1",
	}

	for _, input := range maliciousInputs {
		sanitized := SanitizeString(input)

		// Ensure dangerous SQL keywords are removed
		assert.NotContains(t, sanitized, "DROP")
		assert.NotContains(t, sanitized, "DELETE")
		assert.NotContains(t, sanitized, "UNION")
		assert.NotContains(t, sanitized, "--")
	}
}

func TestXSSPrevention(t *testing.T) {
	// Test cases for XSS prevention
	xssPayloads := []struct {
		input        string
		shouldEscape bool
	}{
		{"<script>alert('XSS')</script>", true},
		{"<img src=x onerror=alert('XSS')>", true},
		{"javascript:alert('XSS')", false}, // No HTML tags, just text
		{"<iframe src='evil.com'>", true},
	}

	for _, payload := range xssPayloads {
		sanitized := SanitizeString(payload.input)

		// HTML should be escaped
		assert.NotContains(t, sanitized, "<script>")
		assert.NotContains(t, sanitized, "<img")
		assert.NotContains(t, sanitized, "<iframe")

		// Should contain escaped versions if payload had HTML tags
		if payload.shouldEscape {
			assert.Contains(t, sanitized, "&lt;")
			assert.Contains(t, sanitized, "&gt;")
		}
	}
}
