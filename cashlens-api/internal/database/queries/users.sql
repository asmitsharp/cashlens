-- name: GetUserByClerkID :one
SELECT * FROM users
WHERE clerk_user_id = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (id, clerk_user_id, email, full_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET email = $2,
    full_name = $3,
    updated_at = NOW()
WHERE clerk_user_id = $1
RETURNING *;
