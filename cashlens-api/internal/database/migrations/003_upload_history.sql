-- Migration 003: Upload History and Transaction Improvements
-- This migration adds upload tracking, duplicate prevention, and performance optimizations

-- Create upload status enum
CREATE TYPE upload_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'partial');

-- Create upload_history table for tracking file uploads
CREATE TABLE IF NOT EXISTS upload_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- File information
    filename TEXT NOT NULL,
    file_key TEXT NOT NULL, -- S3 key or file path
    file_size_bytes BIGINT CHECK (file_size_bytes > 0),
    file_hash TEXT, -- SHA256 hash for duplicate detection
    bank_type VARCHAR(50), -- HDFC, ICICI, SBI, Axis, Kotak

    -- Processing status
    status upload_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    processing_started_at TIMESTAMPTZ,
    processing_completed_at TIMESTAMPTZ,

    -- Statistics
    total_rows INTEGER DEFAULT 0 CHECK (total_rows >= 0),
    parsed_rows INTEGER DEFAULT 0 CHECK (parsed_rows >= 0),
    categorized_rows INTEGER DEFAULT 0 CHECK (categorized_rows >= 0),
    duplicate_rows INTEGER DEFAULT 0 CHECK (duplicate_rows >= 0),
    error_rows INTEGER DEFAULT 0 CHECK (error_rows >= 0),

    -- Calculated metrics
    accuracy_percent DECIMAL(5, 2) GENERATED ALWAYS AS (
        CASE
            WHEN parsed_rows > 0 THEN ROUND((categorized_rows::DECIMAL / parsed_rows) * 100, 2)
            ELSE 0
        END
    ) STORED,

    processing_duration_ms INTEGER GENERATED ALWAYS AS (
        CASE
            WHEN processing_completed_at IS NOT NULL AND processing_started_at IS NOT NULL
            THEN EXTRACT(MILLISECONDS FROM (processing_completed_at - processing_started_at))::INTEGER
            ELSE NULL
        END
    ) STORED,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT valid_parsed_rows CHECK (parsed_rows <= total_rows),
    CONSTRAINT valid_categorized_rows CHECK (categorized_rows <= parsed_rows),
    CONSTRAINT valid_duplicate_rows CHECK (duplicate_rows <= total_rows),
    CONSTRAINT valid_error_rows CHECK (error_rows <= total_rows),
    CONSTRAINT valid_processing_times CHECK (
        (processing_started_at IS NULL AND processing_completed_at IS NULL) OR
        (processing_started_at IS NOT NULL AND (processing_completed_at IS NULL OR processing_completed_at >= processing_started_at))
    )
);

-- Add upload_id to transactions table for linking
ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS upload_id UUID REFERENCES upload_history(id) ON DELETE SET NULL;

-- Add unique constraint for duplicate detection (user + date + description + amount)
-- This prevents duplicate transactions from being imported
ALTER TABLE transactions
ADD CONSTRAINT unique_user_transaction
UNIQUE (user_id, txn_date, description, amount);

-- Create indexes for upload_history
CREATE INDEX IF NOT EXISTS idx_upload_history_user_id ON upload_history(user_id);
CREATE INDEX IF NOT EXISTS idx_upload_history_status ON upload_history(status);
CREATE INDEX IF NOT EXISTS idx_upload_history_created_at ON upload_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_upload_history_user_created ON upload_history(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_upload_history_file_hash ON upload_history(file_hash) WHERE file_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_upload_history_processing ON upload_history(status)
    WHERE status IN ('pending', 'processing');

-- Create index for transactions.upload_id
CREATE INDEX IF NOT EXISTS idx_transactions_upload_id ON transactions(upload_id);

-- Create composite index for bulk operations
CREATE INDEX IF NOT EXISTS idx_transactions_bulk_ops ON transactions(user_id, upload_id, created_at DESC);

-- Create a table for tracking duplicate detection rules
CREATE TABLE IF NOT EXISTS duplicate_detection_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    rule_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    -- Rules can be global (user_id NULL) or user-specific
    match_fields TEXT[] NOT NULL DEFAULT ARRAY['txn_date', 'description', 'amount'],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_match_fields CHECK (
        match_fields <@ ARRAY['txn_date', 'description', 'amount', 'category', 'txn_type']::TEXT[]
    )
);

-- Insert default global duplicate detection rule
INSERT INTO duplicate_detection_rules (user_id, rule_name, match_fields, is_active)
VALUES (NULL, 'Default Global Rule', ARRAY['txn_date', 'description', 'amount'], true)
ON CONFLICT DO NOTHING;

-- Create materialized view for user upload statistics (refreshed periodically)
CREATE MATERIALIZED VIEW IF NOT EXISTS user_upload_stats AS
SELECT
    u.id as user_id,
    u.email,
    COUNT(DISTINCT uh.id) as total_uploads,
    COUNT(DISTINCT CASE WHEN uh.status = 'completed' THEN uh.id END) as successful_uploads,
    COUNT(DISTINCT CASE WHEN uh.status = 'failed' THEN uh.id END) as failed_uploads,
    SUM(uh.total_rows) as total_rows_processed,
    SUM(uh.categorized_rows) as total_categorized,
    AVG(uh.accuracy_percent) as avg_accuracy_percent,
    AVG(uh.processing_duration_ms) as avg_processing_ms,
    MAX(uh.created_at) as last_upload_at,
    COUNT(DISTINCT t.id) as total_transactions
FROM users u
LEFT JOIN upload_history uh ON u.id = uh.user_id
LEFT JOIN transactions t ON u.id = t.user_id
GROUP BY u.id, u.email;

-- Create unique index on materialized view
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_upload_stats_user_id ON user_upload_stats(user_id);

-- Add trigger for upload_history updated_at
DROP TRIGGER IF EXISTS update_upload_history_updated_at ON upload_history;
CREATE TRIGGER update_upload_history_updated_at
    BEFORE UPDATE ON upload_history
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add trigger for duplicate_detection_rules updated_at
DROP TRIGGER IF EXISTS update_duplicate_detection_rules_updated_at ON duplicate_detection_rules;
CREATE TRIGGER update_duplicate_detection_rules_updated_at
    BEFORE UPDATE ON duplicate_detection_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create function for checking duplicates before insert
CREATE OR REPLACE FUNCTION check_transaction_duplicate(
    p_user_id UUID,
    p_txn_date DATE,
    p_description TEXT,
    p_amount DECIMAL
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM transactions
        WHERE user_id = p_user_id
          AND txn_date = p_txn_date
          AND description = p_description
          AND amount = p_amount
    );
END;
$$ LANGUAGE plpgsql;

-- Create function for batch duplicate checking (for performance)
CREATE OR REPLACE FUNCTION check_batch_duplicates(
    p_user_id UUID,
    p_transactions JSONB
) RETURNS TABLE(
    txn_date DATE,
    description TEXT,
    amount DECIMAL,
    is_duplicate BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    WITH input_txns AS (
        SELECT
            (t->>'txn_date')::DATE as txn_date,
            t->>'description' as description,
            (t->>'amount')::DECIMAL as amount
        FROM jsonb_array_elements(p_transactions) as t
    )
    SELECT
        i.txn_date,
        i.description,
        i.amount,
        EXISTS (
            SELECT 1 FROM transactions t
            WHERE t.user_id = p_user_id
              AND t.txn_date = i.txn_date
              AND t.description = i.description
              AND t.amount = i.amount
        ) as is_duplicate
    FROM input_txns i;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE upload_history IS 'Tracks all CSV file uploads with processing status and statistics';
COMMENT ON COLUMN upload_history.file_hash IS 'SHA256 hash of file content for duplicate file detection';
COMMENT ON COLUMN upload_history.status IS 'Current processing status: pending, processing, completed, failed, partial';
COMMENT ON COLUMN upload_history.accuracy_percent IS 'Percentage of successfully categorized transactions';
COMMENT ON COLUMN upload_history.processing_duration_ms IS 'Total processing time in milliseconds';

COMMENT ON TABLE duplicate_detection_rules IS 'Configurable rules for detecting duplicate transactions';
COMMENT ON COLUMN duplicate_detection_rules.match_fields IS 'Fields to match when detecting duplicates';

COMMENT ON FUNCTION check_transaction_duplicate IS 'Check if a single transaction is a duplicate';
COMMENT ON FUNCTION check_batch_duplicates IS 'Efficiently check multiple transactions for duplicates';

COMMENT ON MATERIALIZED VIEW user_upload_stats IS 'Aggregated statistics per user for dashboard display';