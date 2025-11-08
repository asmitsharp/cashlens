package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/ashmitsharp/cashlens-api/internal/database"
	"github.com/ashmitsharp/cashlens-api/internal/database/db"
	"github.com/ashmitsharp/cashlens-api/internal/handlers"
	"github.com/ashmitsharp/cashlens-api/internal/middleware"
	"github.com/ashmitsharp/cashlens-api/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to database
	pool, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("✓ Connected to database successfully")

	// Create database queries instance
	queries := db.New(pool)

	// Initialize services
	// Storage service for S3 operations
	storageService, err := services.NewStorageService(
		os.Getenv("S3_BUCKET"),      // e.g., "cashlens-uploads"
		os.Getenv("S3_REGION"),      // e.g., "ap-south-1"
		os.Getenv("AWS_ENDPOINT"),   // e.g., "http://localhost:4566" for LocalStack
	)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}
	log.Println("✓ Storage service initialized successfully")

	// Parser service for CSV/XLSX/PDF parsing
	parser := services.NewParser()
	log.Println("✓ Parser service initialized successfully")

	// Categorizer service for transaction categorization
	categorizer := services.NewCategorizer(queries)
	log.Println("✓ Categorizer service initialized successfully")

	// File validator service (ready for future integration)
	_ = services.NewFileValidator(10 * 1024 * 1024) // 10MB max
	log.Println("✓ File validator service initialized successfully")

	// Initialize handlers
	usersHandler := handlers.NewUsersHandler(pool) // UsersHandler uses pool directly
	uploadHandler := handlers.NewUploadHandlerFull(storageService, parser, categorizer, queries)
	transactionHandler := handlers.NewTransactionHandler(queries, categorizer)
	rulesHandler := handlers.NewRulesHandler(queries, categorizer)

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

	// Upload routes
	protected.Get("/upload/presigned-url", uploadHandler.GetPresignedURL)
	protected.Post("/upload/process", uploadHandler.ProcessUpload)
	protected.Get("/upload/history", uploadHandler.GetUploadHistory)

	// Transaction routes
	protected.Get("/transactions", transactionHandler.GetTransactions)
	protected.Get("/transactions/stats", transactionHandler.GetTransactionStats)
	protected.Put("/transactions/:id", transactionHandler.UpdateTransaction)
	protected.Put("/transactions/bulk", transactionHandler.BulkUpdateTransactions)

	// Categorization rules routes
	protected.Get("/rules", rulesHandler.GetUserRules)
	protected.Get("/rules/global", rulesHandler.GetGlobalRules)
	protected.Get("/rules/stats", rulesHandler.GetRuleStats)
	protected.Get("/rules/search", rulesHandler.SearchRules)
	protected.Post("/rules", rulesHandler.CreateUserRule)
	protected.Put("/rules/:id", rulesHandler.UpdateUserRule)
	protected.Delete("/rules/:id", rulesHandler.DeleteUserRule)

	log.Println("✓ All routes configured successfully")
	log.Println("")
	log.Println("🚀 cashlens API is running on :8080")
	log.Println("   Health check: http://localhost:8080/health")
	log.Println("   API base: http://localhost:8080/v1")
	log.Fatal(app.Listen(":8080"))
}
