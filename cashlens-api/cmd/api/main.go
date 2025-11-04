package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/ashmitsharp/cashlens-api/internal/middleware"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "cashlens API v1.0",
	})

	// Apply global middleware
	app.Use(middleware.CORS())

	// Health check endpoint (public)
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "cashlens-api",
		})
	})

	// API v1 routes
	v1 := app.Group("/v1")

	// Public routes
	v1.Get("/ping", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})

	// Protected routes (require authentication)
	protected := v1.Group("", middleware.ClerkAuth())

	// Test protected endpoint
	protected.Get("/me", func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		return c.JSON(fiber.Map{
			"user_id": userID,
			"message": "This is a protected route",
		})
	})

	log.Println("Starting cashlens API on :8080")
	log.Fatal(app.Listen(":8080"))
}
