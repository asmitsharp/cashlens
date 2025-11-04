package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/ashmitsharp/cashlens-api/internal/database"
	"github.com/ashmitsharp/cashlens-api/internal/handlers"
	"github.com/ashmitsharp/cashlens-api/internal/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Connected to database successfully")

	// Initialize handlers
	usersHandler := handlers.NewUsersHandler(db)

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

	// Internal routes (webhook callbacks - should be secured with webhook secret in production)
	internal := v1.Group("/internal")
	internal.Post("/users", usersHandler.CreateUser)
	internal.Put("/users/:id", usersHandler.UpdateUser)

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

	// Get current user details
	protected.Get("/user", usersHandler.GetUser)

	log.Println("Starting cashlens API on :8080")
	log.Fatal(app.Listen(":8080"))
}
