package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// RateLimiter returns a configured rate limiting middleware
// Limits requests to 100 per minute per user (or IP for unauthenticated requests)
func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100, // Maximum 100 requests
		Expiration: 1 * time.Minute,

		// Use user ID for authenticated requests, IP for unauthenticated
		KeyGenerator: func(c fiber.Ctx) string {
			// Try to get user ID from auth middleware
			if userID, ok := c.Locals("user_id").(string); ok && userID != "" {
				return "user:" + userID
			}
			// Fall back to IP address for unauthenticated requests
			return c.IP()
		},

		// Custom response when rate limit is exceeded
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
				"retry_after": 60,
			})
		},

		// Store rate limit data in memory (use Redis in production)
		Storage: nil, // nil uses default in-memory storage
	})
}

// StrictRateLimiter returns a more restrictive rate limiter for sensitive endpoints
// Limits to 20 requests per minute
func StrictRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,

		KeyGenerator: func(c fiber.Ctx) string {
			if userID, ok := c.Locals("user_id").(string); ok && userID != "" {
				return "user:strict:" + userID
			}
			return "ip:strict:" + c.IP()
		},

		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": "Too many requests to sensitive endpoint. Please try again later.",
				"retry_after": 60,
			})
		},

		Storage: nil,
	})
}
