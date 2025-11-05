# Database Performance Guide - Upload History & Transactions

## Table of Contents
1. [Schema Overview](#schema-overview)
2. [Index Strategy](#index-strategy)
3. [Performance Estimates](#performance-estimates)
4. [Query Patterns](#query-patterns)
5. [Optimization Tips](#optimization-tips)

## Schema Overview

### Core Tables
- **upload_history**: Tracks CSV file uploads with processing status
- **transactions**: Stores parsed bank transactions (enhanced with upload_id)
- **duplicate_detection_rules**: Configurable duplicate detection logic
- **user_upload_stats**: Materialized view for dashboard metrics

### Key Features
- Automatic duplicate prevention via unique constraint
- Generated columns for calculated metrics (accuracy_percent, processing_duration_ms)
- Batch insert optimization with COPY and array operations
- Materialized view for expensive aggregations

## Index Strategy

### Primary Indexes (Created by Migration)

#### upload_history Table
```sql
-- Primary lookups
idx_upload_history_user_id            -- Filter by user
idx_upload_history_status              -- Monitor processing status
idx_upload_history_created_at          -- Time-based sorting
idx_upload_history_user_created        -- User's upload timeline
idx_upload_history_file_hash           -- Duplicate file detection
idx_upload_history_processing          -- Active processing monitoring
```

**Rationale:**
- `user_id`: Most queries are user-scoped (O(log n))
- `status`: Critical for monitoring active uploads
- `created_at DESC`: Default sorting for upload history
- `(user_id, created_at DESC)`: Composite for user timeline queries
- `file_hash` (partial): Only indexes non-null values for efficiency
- `status WHERE IN ('pending','processing')`: Partial index for queue monitoring

#### transactions Table (Enhanced)
```sql
-- Existing indexes
idx_transactions_user_id               -- User filtering
idx_transactions_txn_date              -- Date sorting
idx_transactions_category              -- Category filtering
idx_transactions_is_reviewed           -- Review queue
idx_transactions_user_date             -- User + date composite
idx_transactions_uncategorized         -- Uncategorized queue (partial)

-- New indexes from migration 003
idx_transactions_upload_id             -- Link to upload history
idx_transactions_bulk_ops              -- Bulk operation optimization
unique_user_transaction                -- Duplicate prevention
```

**Rationale:**
- `upload_id`: Fast join with upload_history (O(log n))
- `(user_id, upload_id, created_at)`: Optimizes bulk insert verification
- Unique constraint enforces data integrity at DB level

### Recommended Additional Indexes (Optional)

For high-volume scenarios (>1M transactions):

```sql
-- Hot path optimization for dashboard
CREATE INDEX CONCURRENTLY idx_upload_history_user_status
ON upload_history(user_id, status, created_at DESC)
WHERE status = 'completed';

-- Transaction search optimization
CREATE INDEX CONCURRENTLY idx_transactions_description_trgm
ON transactions USING gin(description gin_trgm_ops);

-- Time-series analysis
CREATE INDEX CONCURRENTLY idx_transactions_monthly
ON transactions(user_id, date_trunc('month', txn_date), category);
```

## Performance Estimates

### Bulk Insert Performance

#### Using COPY (BulkInsertTransactions)
- **1,000 rows**: ~50-100ms
- **10,000 rows**: ~400-600ms
- **100,000 rows**: ~3-5 seconds

#### Using Array Insert (BatchInsertTransactions)
- **100 rows**: ~20-40ms
- **500 rows**: ~100-200ms
- **1,000 rows**: ~200-400ms

#### Single Row Insert (with duplicate check)
- **Average**: 2-5ms per row
- **With conflict**: 3-7ms per row

### Query Performance

#### Common Operations
| Operation | Expected Time | Index Used |
|-----------|--------------|------------|
| Get user upload history (paginated) | <10ms | idx_upload_history_user_created |
| Check duplicate by hash | <5ms | idx_upload_history_file_hash |
| Get transactions by upload | <20ms | idx_transactions_upload_id |
| Count user transactions | <15ms | idx_transactions_user_id |
| Get uncategorized transactions | <25ms | idx_transactions_uncategorized |
| Update upload status | <10ms | Primary key |
| Dashboard stats (materialized) | <5ms | idx_user_upload_stats_user_id |

### Scalability Metrics

#### Per User Capacity
- **Transactions**: 100,000+ per user efficiently
- **Uploads**: 1,000+ upload history records
- **Query degradation**: <10% up to 1M rows per user

#### System-wide Capacity
- **Total transactions**: 10M+ with current indexes
- **Concurrent uploads**: 100+ simultaneous processing
- **Dashboard refresh**: <100ms for materialized view

## Query Patterns

### Handler Implementation Examples

#### 1. Upload Flow Handler
```go
func (h *UploadHandler) ProcessUpload(userID, fileKey string, fileData []byte) error {
    // 1. Create upload history record
    upload, err := h.queries.CreateUploadHistory(ctx, CreateUploadHistoryParams{
        UserID:        userID,
        Filename:      filename,
        FileKey:       fileKey,
        FileSizeBytes: int64(len(fileData)),
        FileHash:      calculateSHA256(fileData),
        Status:        "pending",
        TotalRows:     countCSVRows(fileData),
    })

    // 2. Check for duplicate file
    duplicate, err := h.queries.GetUploadByFileHash(ctx, GetUploadByFileHashParams{
        UserID:   userID,
        FileHash: upload.FileHash,
    })
    if duplicate != nil {
        return ErrDuplicateFile
    }

    // 3. Start processing
    upload, err = h.queries.StartProcessingUpload(ctx, upload.ID)

    // 4. Parse CSV and prepare batch
    transactions := parseCSV(fileData)

    // 5. Bulk insert with duplicate handling
    inserted, err := h.queries.BatchInsertTransactions(ctx, BatchInsertTransactionsParams{
        UserIds:      repeatValue(userID, len(transactions)),
        UploadIds:    repeatValue(upload.ID, len(transactions)),
        TxnDates:     extractDates(transactions),
        Descriptions: extractDescriptions(transactions),
        // ... other fields
    })

    // 6. Complete upload with statistics
    h.queries.CompleteUploadProcessing(ctx, CompleteUploadProcessingParams{
        ID:               upload.ID,
        Status:           "completed",
        ParsedRows:       len(transactions),
        CategorizedRows:  countCategorized(inserted),
        DuplicateRows:    len(transactions) - len(inserted),
    })
}
```

#### 2. Dashboard Handler
```go
func (h *DashboardHandler) GetDashboardStats(userID string) (*DashboardStats, error) {
    // Use materialized view for fast aggregates
    stats, err := h.queries.GetUserUploadStatsFromView(ctx, userID)
    if err != nil {
        // Fallback to real-time calculation
        stats, err = h.queries.GetUploadStatsByUser(ctx, userID)
    }

    // Get recent uploads for activity feed
    recentUploads, err := h.queries.GetRecentUploads(ctx, GetRecentUploadsParams{
        UserID: userID,
        Limit:  10,
    })

    return &DashboardStats{
        TotalTransactions:   stats.TotalTransactions,
        CategorizedCount:    stats.TotalCategorized,
        AverageAccuracy:     stats.AvgAccuracyPercent,
        RecentUploads:       recentUploads,
    }, nil
}
```

#### 3. Duplicate Detection Handler
```go
func (h *TransactionHandler) CheckDuplicates(userID string, txns []Transaction) ([]bool, error) {
    // Prepare JSON for batch check
    txnJSON, _ := json.Marshal(txns)

    // Batch duplicate check (single query)
    results, err := h.queries.CheckBatchDuplicates(ctx, CheckBatchDuplicatesParams{
        UserID:       userID,
        Transactions: txnJSON,
    })

    duplicates := make([]bool, len(results))
    for i, result := range results {
        duplicates[i] = result.IsDuplicate
    }

    return duplicates, nil
}
```

## Optimization Tips

### 1. Bulk Operations
- **Use COPY for >1000 rows**: 10x faster than individual inserts
- **Batch size**: Optimal batch size is 500-1000 rows for array operations
- **Transaction batching**: Wrap bulk inserts in single transaction

### 2. Connection Pooling
```go
// Recommended pool settings
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 3. Query Optimization
- **Use prepared statements**: Reduces parsing overhead
- **Limit pagination**: Max 100 rows per page for responsive UI
- **Selective columns**: Only SELECT needed columns
- **Avoid N+1**: Use JOIN or batch queries

### 4. Monitoring Queries
```sql
-- Check index usage
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;

-- Find slow queries
SELECT
    query,
    calls,
    mean_exec_time,
    total_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 100  -- queries slower than 100ms
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Check table bloat
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    n_live_tup,
    n_dead_tup,
    round(n_dead_tup::numeric / NULLIF(n_live_tup + n_dead_tup, 0) * 100, 2) AS dead_percent
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY dead_percent DESC;
```

### 5. Maintenance Tasks
```sql
-- Weekly maintenance
VACUUM ANALYZE transactions;
VACUUM ANALYZE upload_history;

-- Monthly refresh of materialized view
REFRESH MATERIALIZED VIEW CONCURRENTLY user_upload_stats;

-- Quarterly reindex for heavily updated tables
REINDEX INDEX CONCURRENTLY idx_transactions_user_id;
REINDEX INDEX CONCURRENTLY idx_upload_history_user_id;
```

### 6. Scaling Strategies

#### When to Consider Partitioning (>10M rows)
```sql
-- Partition transactions by month
CREATE TABLE transactions_2024_01 PARTITION OF transactions
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Partition by user_id hash (for very large user bases)
CREATE TABLE transactions_user_0 PARTITION OF transactions
FOR VALUES WITH (modulus 10, remainder 0);
```

#### Read Replica Strategy
- Dashboard queries → Read replica
- Upload processing → Primary
- Analytics → Read replica with delay tolerance

#### Caching Layer (Redis)
- Cache dashboard stats (TTL: 5 minutes)
- Cache category rules (TTL: 1 hour)
- Cache user upload history (TTL: 1 minute)

## Security Considerations

### Row-Level Security (Optional)
```sql
-- Enable RLS for multi-tenant isolation
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_transactions_policy ON transactions
    FOR ALL
    TO application_role
    USING (user_id = current_setting('app.user_id')::UUID);
```

### Audit Trail
```sql
-- Create audit log for sensitive operations
CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    user_id UUID,
    action VARCHAR(50),
    table_name VARCHAR(50),
    record_id UUID,
    old_data JSONB,
    new_data JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Backup and Recovery

### Backup Strategy
- **Daily**: Full backup of all user data
- **Hourly**: WAL archiving for point-in-time recovery
- **Real-time**: Streaming replication to standby

### Recovery Time Objectives
- **RPO (Recovery Point Objective)**: <1 hour
- **RTO (Recovery Time Objective)**: <4 hours

### Backup Commands
```bash
# Full backup
pg_dump -h localhost -U postgres -d cashlens -Fc > backup_$(date +%Y%m%d).dump

# Restore
pg_restore -h localhost -U postgres -d cashlens -c backup_20240101.dump

# Point-in-time recovery
pg_basebackup -h localhost -U replicator -D /backup/base -Fp -Xs -P
```