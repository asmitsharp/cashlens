-- SQLC queries for upload_history table

-- name: CreateUploadHistory :one
-- Create a new upload history record
INSERT INTO upload_history (
    user_id,
    filename,
    file_key,
    file_size_bytes,
    file_hash,
    bank_type,
    status,
    total_rows
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetUploadHistoryByID :one
-- Get a single upload history record by ID
SELECT * FROM upload_history
WHERE id = $1
LIMIT 1;

-- name: GetUploadHistoryByUserAndID :one
-- Get upload history for a specific user and upload ID (for security)
SELECT * FROM upload_history
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: GetUserUploadHistory :many
-- Get paginated upload history for a user
SELECT * FROM upload_history
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserUploadHistoryWithStats :many
-- Get upload history with transaction counts
SELECT
    uh.*,
    COUNT(DISTINCT t.id) as transaction_count
FROM upload_history uh
LEFT JOIN transactions t ON uh.id = t.upload_id
WHERE uh.user_id = $1
GROUP BY uh.id
ORDER BY uh.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentUploads :many
-- Get recent uploads for a user (last 10)
SELECT * FROM upload_history
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 10;

-- name: GetUploadByFileHash :one
-- Check if file has already been uploaded (duplicate detection)
SELECT * FROM upload_history
WHERE user_id = $1 AND file_hash = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateUploadStatus :one
-- Update upload status (for state transitions)
UPDATE upload_history
SET
    status = $2,
    error_message = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: StartProcessingUpload :one
-- Mark upload as processing and set start time
UPDATE upload_history
SET
    status = 'processing',
    processing_started_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING *;

-- name: CompleteUploadProcessing :one
-- Mark upload as completed with statistics
UPDATE upload_history
SET
    status = $2,
    error_message = $3,
    processing_completed_at = NOW(),
    total_rows = $4,
    parsed_rows = $5,
    categorized_rows = $6,
    duplicate_rows = $7,
    error_rows = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUploadStatistics :one
-- Update only statistics (for partial updates during processing)
UPDATE upload_history
SET
    parsed_rows = COALESCE($2, parsed_rows),
    categorized_rows = COALESCE($3, categorized_rows),
    duplicate_rows = COALESCE($4, duplicate_rows),
    error_rows = COALESCE($5, error_rows),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUploadStatsByUser :one
-- Get aggregated upload statistics for a user
SELECT
    COUNT(*) as total_uploads,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_uploads,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_uploads,
    COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing_uploads,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_uploads,
    COALESCE(SUM(total_rows), 0) as total_rows_processed,
    COALESCE(SUM(categorized_rows), 0) as total_categorized,
    COALESCE(AVG(accuracy_percent), 0) as avg_accuracy_percent,
    COALESCE(AVG(processing_duration_ms), 0) as avg_processing_ms
FROM upload_history
WHERE user_id = $1;

-- name: GetProcessingUploads :many
-- Get all uploads currently being processed (for monitoring)
SELECT * FROM upload_history
WHERE status IN ('pending', 'processing')
ORDER BY created_at ASC;

-- name: GetStuckUploads :many
-- Find uploads stuck in processing state (older than 5 minutes)
SELECT * FROM upload_history
WHERE status = 'processing'
  AND processing_started_at < NOW() - INTERVAL '5 minutes'
ORDER BY processing_started_at ASC;

-- name: DeleteUploadHistory :exec
-- Delete an upload history record (cascade will handle transactions)
DELETE FROM upload_history
WHERE id = $1;

-- name: DeleteUserUploadHistory :exec
-- Delete all upload history for a user
DELETE FROM upload_history
WHERE user_id = $1;

-- ============================================
-- Bulk Transaction Insert Queries (Optimized)
-- ============================================

-- name: BulkInsertTransactions :copyfrom
-- Bulk insert transactions using COPY (most efficient for 1000+ rows)
INSERT INTO transactions (
    user_id,
    upload_id,
    txn_date,
    description,
    amount,
    txn_type,
    category,
    is_reviewed,
    raw_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: BatchInsertTransactions :many
-- Batch insert with ON CONFLICT for duplicate handling (for smaller batches)
INSERT INTO transactions (
    user_id,
    upload_id,
    txn_date,
    description,
    amount,
    txn_type,
    category,
    is_reviewed,
    raw_data
) VALUES (
    unnest(@user_ids::UUID[]),
    unnest(@upload_ids::UUID[]),
    unnest(@txn_dates::DATE[]),
    unnest(@descriptions::TEXT[]),
    unnest(@amounts::DECIMAL[]),
    unnest(@txn_types::VARCHAR[]),
    unnest(@categories::VARCHAR[]),
    unnest(@is_reviewed::BOOLEAN[]),
    unnest(@raw_data::TEXT[])
)
ON CONFLICT (user_id, txn_date, description, amount) DO NOTHING
RETURNING *;

-- name: InsertTransactionWithDuplicateCheck :one
-- Insert single transaction with duplicate check
INSERT INTO transactions (
    user_id,
    upload_id,
    txn_date,
    description,
    amount,
    txn_type,
    category,
    is_reviewed,
    raw_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (user_id, txn_date, description, amount) DO NOTHING
RETURNING *;

-- ============================================
-- Duplicate Detection Queries
-- ============================================

-- name: CheckTransactionDuplicate :one
-- Check if a transaction is a duplicate
SELECT check_transaction_duplicate($1, $2, $3, $4) as is_duplicate;

-- name: CheckBatchDuplicates :many
-- Check multiple transactions for duplicates (returns list of duplicates)
SELECT * FROM check_batch_duplicates($1, $2);

-- name: GetDuplicateTransactions :many
-- Find potential duplicate transactions for a user
SELECT
    t1.id as transaction_id,
    t1.txn_date,
    t1.description,
    t1.amount,
    t2.id as duplicate_id,
    t2.created_at as duplicate_created_at
FROM transactions t1
JOIN transactions t2 ON
    t1.user_id = t2.user_id AND
    t1.txn_date = t2.txn_date AND
    t1.description = t2.description AND
    t1.amount = t2.amount AND
    t1.id != t2.id
WHERE t1.user_id = $1
ORDER BY t1.txn_date DESC, t1.created_at DESC;

-- name: CountDuplicatesByUpload :one
-- Count duplicate transactions for an upload
SELECT COUNT(*) as duplicate_count
FROM (
    SELECT DISTINCT t1.id
    FROM transactions t1
    JOIN transactions t2 ON
        t1.user_id = t2.user_id AND
        t1.txn_date = t2.txn_date AND
        t1.description = t2.description AND
        t1.amount = t2.amount AND
        t1.id != t2.id
    WHERE t1.upload_id = $1 AND t2.upload_id != $1
) as duplicates;

-- ============================================
-- Performance Monitoring Queries
-- ============================================

-- name: GetUploadPerformanceMetrics :one
-- Get performance metrics for uploads
SELECT
    AVG(processing_duration_ms) as avg_duration_ms,
    MIN(processing_duration_ms) as min_duration_ms,
    MAX(processing_duration_ms) as max_duration_ms,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY processing_duration_ms) as median_duration_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY processing_duration_ms) as p95_duration_ms,
    AVG(total_rows) as avg_rows_per_upload,
    AVG(CASE WHEN total_rows > 0 THEN processing_duration_ms::FLOAT / total_rows ELSE 0 END) as avg_ms_per_row
FROM upload_history
WHERE user_id = $1
  AND status = 'completed'
  AND processing_duration_ms IS NOT NULL;

-- name: GetTransactionsByUpload :many
-- Get all transactions from a specific upload
SELECT * FROM transactions
WHERE upload_id = $1
ORDER BY txn_date DESC, created_at DESC;

-- name: CountTransactionsByUpload :one
-- Count transactions for an upload
SELECT
    COUNT(*) as total,
    COUNT(CASE WHEN category IS NOT NULL THEN 1 END) as categorized,
    COUNT(CASE WHEN category IS NULL THEN 1 END) as uncategorized
FROM transactions
WHERE upload_id = $1;

-- ============================================
-- Materialized View Refresh
-- ============================================

-- name: RefreshUserUploadStats :exec
-- Refresh materialized view for user statistics
REFRESH MATERIALIZED VIEW CONCURRENTLY user_upload_stats;

-- name: GetUserUploadStatsFromView :one
-- Get cached user statistics from materialized view
SELECT * FROM user_upload_stats
WHERE user_id = $1;