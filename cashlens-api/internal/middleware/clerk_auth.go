package middleware

import (
	"context"
	"os"
	"strings"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gofiber/fiber/v3"
)

// ClerkAuth middleware validates Clerk JWT tokens
func ClerkAuth() fiber.Handler {
	// Initialize Clerk with secret key
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	return func(c fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization token",
			})
		}

		// Remove "Bearer " prefix
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Verify token with Clerk
		// The SDK automatically uses CLERK_SECRET_KEY from environment
		claims, err := jwt.Verify(context.Background(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			// Log the actual error for debugging
			secretKey := os.Getenv("CLERK_SECRET_KEY")
			if secretKey == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Server misconfiguration: CLERK_SECRET_KEY not set",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
				"details": err.Error(),
			})
		}

		// Store user ID in context for use in handlers
		c.Locals("user_id", claims.Subject)
		c.Locals("clerk_user_id", claims.Subject)

		return c.Next()
	}
}
