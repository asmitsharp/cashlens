-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id,
    txn_date,
    description,
    amount,
    txn_type,
    category,
    is_reviewed,
    raw_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: BulkCreateTransactions :copyfrom
INSERT INTO transactions (
    user_id,
    txn_date,
    description,
    amount,
    txn_type,
    category,
    is_reviewed,
    raw_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetTransactionByID :one
SELECT * FROM transactions
WHERE id = $1
LIMIT 1;

-- name: GetUserTransactions :many
SELECT
    t.*,
    uh.bank_type
FROM transactions t
LEFT JOIN upload_history uh ON t.upload_id = uh.id
WHERE t.user_id = $1
ORDER BY t.txn_date DESC
LIMIT $2 OFFSET $3;

-- name: GetAllTransactions :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY txn_date DESC;

-- name: GetCategorizedTransactions :many
SELECT
    t.*,
    uh.bank_type
FROM transactions t
LEFT JOIN upload_history uh ON t.upload_id = uh.id
WHERE t.user_id = $1
  AND t.category IS NOT NULL
ORDER BY t.txn_date DESC
LIMIT $2 OFFSET $3;

-- name: GetUncategorizedTransactions :many
SELECT
    t.*,
    uh.bank_type
FROM transactions t
LEFT JOIN upload_history uh ON t.upload_id = uh.id
WHERE t.user_id = $1
  AND t.category IS NULL
ORDER BY t.txn_date DESC
LIMIT $2 OFFSET $3;

-- name: GetTransactionsByDateRange :many
SELECT * FROM transactions
WHERE user_id = $1
  AND txn_date BETWEEN $2 AND $3
ORDER BY txn_date DESC;

-- name: GetTransactionsByCategory :many
SELECT * FROM transactions
WHERE user_id = $1
  AND category = $2
ORDER BY txn_date DESC
LIMIT $3 OFFSET $4;

-- name: UpdateTransactionCategory :one
UPDATE transactions
SET category = $2,
    is_reviewed = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTransaction :one
UPDATE transactions
SET description = COALESCE($2, description),
    amount = COALESCE($3, amount),
    txn_type = COALESCE($4, txn_type),
    category = COALESCE($5, category),
    is_reviewed = COALESCE($6, is_reviewed),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1;

-- name: DeleteUserTransactions :exec
DELETE FROM transactions
WHERE user_id = $1;

-- name: CountUserTransactions :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1;

-- name: CountCategorizedTransactions :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1
  AND category IS NOT NULL;

-- name: CountUncategorizedTransactions :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1
  AND category IS NULL;

-- name: GetTransactionStats :one
SELECT
    COUNT(*) AS total_count,
    COUNT(CASE WHEN category IS NOT NULL THEN 1 END) AS categorized_count,
    COUNT(CASE WHEN category IS NULL THEN 1 END) AS uncategorized_count,
    ROUND(
        CAST(COUNT(CASE WHEN category IS NOT NULL THEN 1 END) AS DECIMAL) /
        NULLIF(COUNT(*), 0) * 100,
        2
    ) AS accuracy_percent
FROM transactions
WHERE user_id = $1;
