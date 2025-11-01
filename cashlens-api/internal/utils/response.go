package utils

import "github.com/gofiber/fiber/v3"

// SuccessResponse sends a standardized success response
func SuccessResponse(c fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// ErrorResponse sends a standardized error response
func ErrorResponse(c fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

// PaginatedResponse sends a paginated response
func PaginatedResponse(c fiber.Ctx, data interface{}, page, pageSize, total int) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
		"pagination": fiber.Map{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + pageSize - 1) / pageSize,
		},
	})
}
