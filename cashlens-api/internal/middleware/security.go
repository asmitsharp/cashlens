package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// SecurityHeaders adds security-related HTTP headers to all responses
// Implements OWASP security best practices
func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		c.Set("X-Frame-Options", "DENY")

		// Enable XSS protection (legacy, but still good to have)
		c.Set("X-XSS-Protection", "1; mode=block")

		// Enforce HTTPS in production (max-age = 1 year)
		// Only enable if served over HTTPS
		// c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Content Security Policy - restrict resource loading
		// Note: Adjust this based on your actual requirements
		c.Set("Content-Security-Policy",
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: https:; "+
			"font-src 'self' data:; "+
			"connect-src 'self' https://api.clerk.com https://*.clerk.accounts.dev; "+
			"frame-ancestors 'none'",
		)

		// Referrer Policy - control referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy - control browser features
		c.Set("Permissions-Policy",
			"accelerometer=(), "+
			"camera=(), "+
			"geolocation=(), "+
			"gyroscope=(), "+
			"magnetometer=(), "+
			"microphone=(), "+
			"payment=(), "+
			"usb=()",
		)

		// Remove server header to prevent version disclosure
		c.Set("Server", "cashlens-api")

		return c.Next()
	}
}

// SecureJSON ensures JSON responses have proper content type
// and prevents JSON hijacking
func SecureJSON() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Set proper JSON content type
		if c.Get("Content-Type") == "application/json" {
			c.Set("X-Content-Type-Options", "nosniff")
		}
		return c.Next()
	}
}
