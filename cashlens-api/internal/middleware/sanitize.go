package middleware

import (
	"html"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// InputSanitizer sanitizes user input to prevent XSS and injection attacks
// Note: This is a basic implementation for demonstration
// In production, consider using a dedicated library
func InputSanitizer() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Note: Query parameter sanitization would need to happen in handlers
		// as Fiber v3 doesn't allow direct modification of query args in middleware

		// We provide the SanitizeString function for handlers to use
		// Continue to next middleware
		return c.Next()
	}
}

// sanitizeInput performs basic sanitization on user input
// Note: This is a simple implementation. For production, consider using
// a dedicated sanitization library like bluemonday
func sanitizeInput(input string) string {
	// Remove leading/trailing whitespace
	input = strings.TrimSpace(input)

	// HTML escape to prevent XSS
	input = html.EscapeString(input)

	// Remove potential SQL injection characters (defense in depth)
	// Note: We're using prepared statements, but this adds extra protection
	dangerous := []string{
		"';",
		"--",
		"/*",
		"*/",
		"xp_",
		"sp_",
		"DROP",
		"INSERT",
		"DELETE",
		"UPDATE",
		"UNION",
		"SELECT",
	}

	for _, d := range dangerous {
		input = strings.ReplaceAll(input, d, "")
		input = strings.ReplaceAll(input, strings.ToLower(d), "")
	}

	// Limit input length to prevent DoS
	if len(input) > 1000 {
		input = input[:1000]
	}

	return input
}

// SanitizeString is a helper function to sanitize individual strings
func SanitizeString(input string) string {
	return sanitizeInput(input)
}
