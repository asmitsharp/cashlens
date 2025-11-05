package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersHandler struct {
	db *pgxpool.Pool
}

func NewUsersHandler(db *pgxpool.Pool) *UsersHandler {
	return &UsersHandler{db: db}
}

type CreateUserRequest struct {
	ClerkUserID string `json:"clerk_user_id"`
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// CreateUser creates a new user in the database (called by Clerk webhook)
func (h *UsersHandler) CreateUser(c fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.ClerkUserID == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "clerk_user_id and email are required",
		})
	}

	// Generate UUID for user
	userID := uuid.New()

	// Insert user into database
	query := `
		INSERT INTO users (id, clerk_user_id, email, full_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (clerk_user_id) DO UPDATE
		SET email = EXCLUDED.email,
		    full_name = EXCLUDED.full_name,
		    updated_at = EXCLUDED.updated_at
		RETURNING id, clerk_user_id, email, full_name, created_at, updated_at
	`

	now := time.Now()
	var user struct {
		ID          uuid.UUID `json:"id"`
		ClerkUserID string    `json:"clerk_user_id"`
		Email       string    `json:"email"`
		FullName    *string   `json:"full_name"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	err := h.db.QueryRow(
		context.Background(),
		query,
		userID,
		req.ClerkUserID,
		req.Email,
		req.FullName,
		now,
		now,
	).Scan(
		&user.ID,
		&user.ClerkUserID,
		&user.Email,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// UpdateUser updates an existing user (called by Clerk webhook)
func (h *UsersHandler) UpdateUser(c fiber.Ctx) error {
	clerkUserID := c.Params("id")
	if clerkUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	var req UpdateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	query := `
		UPDATE users
		SET email = $1,
		    full_name = $2,
		    updated_at = $3
		WHERE clerk_user_id = $4
		RETURNING id, clerk_user_id, email, full_name, created_at, updated_at
	`

	var user struct {
		ID          uuid.UUID `json:"id"`
		ClerkUserID string    `json:"clerk_user_id"`
		Email       string    `json:"email"`
		FullName    *string   `json:"full_name"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	err := h.db.QueryRow(
		context.Background(),
		query,
		req.Email,
		req.FullName,
		time.Now(),
		clerkUserID,
	).Scan(
		&user.ID,
		&user.ClerkUserID,
		&user.Email,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update user",
			"details": err.Error(),
		})
	}

	return c.JSON(user)
}

// GetUser retrieves a user by Clerk user ID
func (h *UsersHandler) GetUser(c fiber.Ctx) error {
	clerkUserID := c.Locals("clerk_user_id").(string)

	query := `
		SELECT id, clerk_user_id, email, full_name, created_at, updated_at
		FROM users
		WHERE clerk_user_id = $1
	`

	var user struct {
		ID          uuid.UUID `json:"id"`
		ClerkUserID string    `json:"clerk_user_id"`
		Email       string    `json:"email"`
		FullName    *string   `json:"full_name"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	err := h.db.QueryRow(context.Background(), query, clerkUserID).Scan(
		&user.ID,
		&user.ClerkUserID,
		&user.Email,
		&user.FullName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}
