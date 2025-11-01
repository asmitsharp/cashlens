package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewBadRequestError(message string, details any) *APIError {
	return &APIError{
		StatusCode: fiber.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
		Details:    details,
	}
}

func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		StatusCode: fiber.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

func NewNotFoundError(resource string) *APIError {
	return &APIError{
		StatusCode: fiber.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
	}
}

func NewInternalError(err error) *APIError {
	return &APIError{
		StatusCode: fiber.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    "An internal error occurred",
		Details:    err.Error(), // Only in development
	}
}

// ErrorHandler is a middleware to handle APIError
func ErrorHandler(c fiber.Ctx, err error) error {
	apiErr, ok := err.(*APIError)
	if !ok {
		apiErr = NewInternalError(err)
	}

	return c.Status(apiErr.StatusCode).JSON(apiErr)
}
