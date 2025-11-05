-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    txn_date DATE NOT NULL,
    description TEXT NOT NULL,
    amount DECIMAL(15, 2) NOT NULL, -- Negative for debit, positive for credit
    txn_type VARCHAR(10) NOT NULL CHECK (txn_type IN ('credit', 'debit')),
    category VARCHAR(100),
    is_reviewed BOOLEAN NOT NULL DEFAULT FALSE,
    raw_data TEXT, -- Original CSV row for debugging
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_txn_date ON transactions(txn_date DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);
CREATE INDEX IF NOT EXISTS idx_transactions_is_reviewed ON transactions(is_reviewed);
CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, txn_date DESC);

-- Create composite index for uncategorized transactions
CREATE INDEX IF NOT EXISTS idx_transactions_uncategorized ON transactions(user_id, category) WHERE category IS NULL;

-- Add updated_at trigger (reusing function from 001_initial.sql)
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
CREATE TRIGGER update_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment
COMMENT ON TABLE transactions IS 'Stores all parsed bank transactions from CSV uploads';
COMMENT ON COLUMN transactions.amount IS 'Negative for debit (money out), positive for credit (money in)';
COMMENT ON COLUMN transactions.raw_data IS 'Original CSV row for debugging and audit trail';
