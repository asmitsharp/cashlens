package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "cashlens API v1.0",
	})

	// Health check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "cashlens-api",
		})
	})

	// API v1 routes
	v1 := app.Group("/v1")

	// Public routes (to be implemented)
	v1.Get("/ping", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})

	log.Println("Starting cashlens API on :8080")
	log.Fatal(app.Listen(":8080"))
}
