package middleware

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/storage/redis/v3"
)

var store *redis.Storage

func initRedis() {
	if store != nil {
		return
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	if redisHost == "" {
		redisHost = "localhost"
	}
	if redisPort == "" {
		redisPort = "6379"
	}

	store = redis.New(redis.Config{
		Host:      redisHost,
		Port:      parseInt(redisPort),
		Username:  "",
		Password:  "",
		Database:  0,
		Reset:     false,
		TLSConfig: nil,
		PoolSize:  10 * runtime.GOMAXPROCS(0),
	})

	log.Println("✓ Redis rate limiter storage initialized")
}

func parseInt(s string) int {
	var n int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			continue
		}
		n = n*10 + int(ch-'0')
	}
	if n == 0 {
		return 6379
	}
	return n
}

// RateLimiter returns a configured rate limiting middleware
// Limits requests to 100 per minute per user (or IP for unauthenticated requests)
func RateLimiter() fiber.Handler {
	initRedis()

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
				"error":       "Rate limit exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": 60,
			})
		},

		// Store rate limit data in Redis
		Storage: store,
	})
}

// StrictRateLimiter returns a more restrictive rate limiter for sensitive endpoints
// Limits to 20 requests per minute
func StrictRateLimiter() fiber.Handler {
	initRedis()

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
				"error":       "Rate limit exceeded",
				"message":     "Too many requests to sensitive endpoint. Please try again later.",
				"retry_after": 60,
			})
		},

		Storage: store,
	})
}
