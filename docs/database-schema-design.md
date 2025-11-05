# Database Schema Design - Transactions & Users

**Database:** PostgreSQL 16
**Migration Files:** `internal/database/migrations/`
**Query Files:** `internal/database/queries/`
**Version:** 1.0.0
**Date:** 2025-11-05

---

## Table of Contents

1. [Overview](#overview)
2. [Schema Diagram](#schema-diagram)
3. [Users Table](#users-table)
4. [Transactions Table](#transactions-table)
5. [Index Strategy](#index-strategy)
6. [Design Decisions](#design-decisions)
7. [Query Patterns](#query-patterns)
8. [Performance Considerations](#performance-considerations)
9. [Future Schema Changes](#future-schema-changes)

---

## Overview

The Cashlens database schema is designed for:
- **High read performance** (frequent dashboard queries)
- **Efficient categorization queries** (filter by category, review status)
- **Audit trail** (raw_data field for debugging)
- **Scalability** (indexed for millions of transactions)

### Core Tables

1. **users** - User accounts (synced from Clerk)
2. **transactions** - Parsed bank transactions
3. **(future) categorization_rules** - Auto-categorization rules
4. **(future) upload_history** - File upload tracking

---

## Schema Diagram

```
┌─────────────────────────────────────┐
│             users                   │
├─────────────────────────────────────┤
│ id (PK)              UUID           │
│ clerk_user_id        TEXT UNIQUE    │
│ email                TEXT           │
│ full_name            TEXT           │
│ created_at           TIMESTAMPTZ    │
│ updated_at           TIMESTAMPTZ    │
└─────────────────┬───────────────────┘
                  │
                  │ 1:N
                  │
┌─────────────────▼───────────────────┐
│          transactions               │
├─────────────────────────────────────┤
│ id (PK)              UUID           │
│ user_id (FK)         UUID           │ → users.id
│ txn_date             DATE           │
│ description          TEXT           │
│ amount               DECIMAL(15,2)  │ (negative = debit)
│ txn_type             VARCHAR(10)    │ CHECK(credit|debit)
│ category             VARCHAR(100)   │ (nullable)
│ is_reviewed          BOOLEAN        │ DEFAULT false
│ raw_data             TEXT           │ (original CSV row)
│ created_at           TIMESTAMPTZ    │
│ updated_at           TIMESTAMPTZ    │
└─────────────────────────────────────┘

Indexes:
  - idx_transactions_user_id (user_id)
  - idx_transactions_txn_date (txn_date DESC)
  - idx_transactions_category (category)
  - idx_transactions_is_reviewed (is_reviewed)
  - idx_transactions_user_date (user_id, txn_date DESC)
  - idx_transactions_uncategorized (user_id, category) WHERE category IS NULL
```

---

## Users Table

### Schema

```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clerk_user_id TEXT UNIQUE NOT NULL,
    email TEXT NOT NULL,
    full_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_clerk_id ON users(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

### Columns

| Column         | Type         | Constraints      | Purpose                                |
|----------------|--------------|------------------|----------------------------------------|
| id             | UUID         | PRIMARY KEY      | Internal unique identifier             |
| clerk_user_id  | TEXT         | UNIQUE, NOT NULL | Clerk's user ID (external reference)   |
| email          | TEXT         | NOT NULL         | User's email (for communication)       |
| full_name      | TEXT         | nullable         | Display name                           |
| created_at     | TIMESTAMPTZ  | NOT NULL         | Account creation timestamp             |
| updated_at     | TIMESTAMPTZ  | NOT NULL         | Last update timestamp                  |

### Design Decisions

**Q: Why store both `id` and `clerk_user_id`?**

A: Separation of concerns:
- `id` - Our internal UUID (stable, never changes)
- `clerk_user_id` - External reference (could change if we switch auth providers)
- Foreign keys use `id` for stability

**Q: Why not use clerk_user_id as primary key?**

A:
- External IDs might have character limits
- Internal UUID gives us control
- Easier to test (can generate UUIDs locally)

**Q: Why TIMESTAMPTZ instead of TIMESTAMP?**

A:
- Stores timezone information
- Handles users in different timezones
- PostgreSQL best practice

### Triggers

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
```

**Purpose:** Automatically update `updated_at` on every row update.

---

## Transactions Table

### Schema

```sql
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    txn_date DATE NOT NULL,
    description TEXT NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    txn_type VARCHAR(10) NOT NULL CHECK (txn_type IN ('credit', 'debit')),
    category VARCHAR(100),
    is_reviewed BOOLEAN NOT NULL DEFAULT FALSE,
    raw_data TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Columns

| Column       | Type          | Constraints                      | Purpose                                |
|--------------|---------------|----------------------------------|----------------------------------------|
| id           | UUID          | PRIMARY KEY                      | Unique transaction identifier          |
| user_id      | UUID          | FK → users(id), NOT NULL, CASCADE| Owner of transaction                   |
| txn_date     | DATE          | NOT NULL                         | Transaction date (not time)            |
| description  | TEXT          | NOT NULL                         | Transaction description/narration      |
| amount       | DECIMAL(15,2) | NOT NULL                         | Amount (negative = debit, positive = credit) |
| txn_type     | VARCHAR(10)   | NOT NULL, CHECK                  | "credit" or "debit"                    |
| category     | VARCHAR(100)  | nullable                         | Auto-assigned category                 |
| is_reviewed  | BOOLEAN       | NOT NULL, DEFAULT FALSE          | User reviewed this transaction?        |
| raw_data     | TEXT          | nullable                         | Original CSV row (debugging/audit)     |
| created_at   | TIMESTAMPTZ   | NOT NULL                         | When transaction was imported          |
| updated_at   | TIMESTAMPTZ   | NOT NULL                         | Last modification time                 |

### Design Decisions

#### 1. Why DATE instead of TIMESTAMPTZ for txn_date?

```sql
-- Our choice: DATE
txn_date DATE  -- '2024-01-15'

-- Alternative: TIMESTAMPTZ
txn_date TIMESTAMPTZ  -- '2024-01-15 14:30:00+05:30'
```

**Reasoning:**
- Bank statements show dates, not times
- Time precision is irrelevant for transactions
- Simpler date range queries (`BETWEEN '2024-01-01' AND '2024-01-31'`)
- Saves 4 bytes per row (DATE = 4 bytes, TIMESTAMPTZ = 8 bytes)

**Impact at scale:**
- 1M transactions: ~4MB saved
- Cleaner dashboard queries (group by month/year)

#### 2. Why DECIMAL(15,2) instead of FLOAT for amount?

```sql
-- Our choice: DECIMAL(15,2)
amount DECIMAL(15, 2)  -- Exact: 1234567890123.45

-- Alternative: FLOAT
amount FLOAT  -- Approximate: 1234567890123.449999...
```

**Reasoning:**
- **Financial accuracy required** - no rounding errors
- DECIMAL is exact for monetary values
- FLOAT has precision issues:
  ```sql
  SELECT 0.1 + 0.2;  -- Returns 0.30000000000000004 (float)
  SELECT 0.1::DECIMAL + 0.2::DECIMAL;  -- Returns 0.3 (exact)
  ```

**Why (15,2)?**
- 15 total digits, 2 after decimal
- Max value: 9,999,999,999,999.99 (nearly 10 trillion)
- Sufficient for largest Indian transactions
- 2 decimal places match currency precision (paise)

#### 3. Why negative amounts for debits?

```sql
-- Our design:
debit:  amount = -3500.00
credit: amount = 50000.00

-- Balance calculation:
balance += amount  -- Simple!

-- Total expenses:
SELECT SUM(amount) WHERE amount < 0
```

**Benefits:**
1. Simpler SQL queries
2. Mathematical correctness (debit decreases balance)
3. Easy filtering by sign
4. Standard accounting convention

**Alternative (rejected):**
```sql
-- Two columns approach:
debit_amount DECIMAL(15,2)
credit_amount DECIMAL(15,2)

-- More complex queries:
balance = balance - debit_amount + credit_amount
```

#### 4. Why CHECK constraint on txn_type?

```sql
CHECK (txn_type IN ('credit', 'debit'))
```

**Benefits:**
- Database-level validation (can't insert invalid data)
- Prevents bugs (typos like "deb it" or "CREDIT")
- Self-documenting schema
- Faster than application-level validation

#### 5. Why nullable category?

```sql
category VARCHAR(100)  -- Can be NULL
```

**Reasoning:**
- Auto-categorization might fail (85% accuracy = 15% uncategorized)
- NULL indicates "needs review"
- Explicit NULL vs empty string is clearer:
  ```sql
  WHERE category IS NULL      -- Uncategorized
  WHERE category = 'Salaries' -- Specific category
  ```

#### 6. Why TEXT for raw_data?

```sql
raw_data TEXT  -- Unlimited length
```

**Purpose:** Store original CSV row for:
- Debugging parsing issues
- Compliance/audit trails
- Re-processing if parser logic changes

**Example:**
```
raw_data: "15/01/2024,AWS SERVICES,UPI/123456,15/01/2024,3500.00,,450000.00"
```

If user reports parsing error, we can:
1. Check raw_data
2. Fix parser
3. Re-parse without re-uploading

**Cost:** ~100 bytes per row (negligible)

#### 7. Why ON DELETE CASCADE?

```sql
user_id UUID REFERENCES users(id) ON DELETE CASCADE
```

**Behavior:** When user is deleted, all their transactions are also deleted.

**Why?**
- GDPR compliance (right to be forgotten)
- Data integrity (no orphaned transactions)
- Simpler deletion logic

**Alternative (rejected):**
```sql
ON DELETE SET NULL  -- Would create orphaned transactions
ON DELETE RESTRICT  -- Would prevent user deletion if they have transactions
```

---

## Index Strategy

### Index 1: User Lookup

```sql
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
```

**Query Pattern:**
```sql
SELECT * FROM transactions WHERE user_id = 'uuid-123';
```

**Benefit:** O(log n) instead of O(n) table scan

**When Used:** Every dashboard view, every user query

---

### Index 2: Date Sorting

```sql
CREATE INDEX idx_transactions_txn_date ON transactions(txn_date DESC);
```

**Query Pattern:**
```sql
SELECT * FROM transactions
WHERE user_id = 'uuid-123'
ORDER BY txn_date DESC;
```

**Why DESC?**
- Most queries want recent transactions first
- `ORDER BY txn_date DESC` uses index directly
- No need to reverse scan

**When Used:** Dashboard timeline, transaction list

---

### Index 3: Category Filtering

```sql
CREATE INDEX idx_transactions_category ON transactions(category);
```

**Query Pattern:**
```sql
SELECT * FROM transactions
WHERE category = 'Salaries';
```

**When Used:** Category-specific reports, expense breakdown

---

### Index 4: Review Status

```sql
CREATE INDEX idx_transactions_is_reviewed ON transactions(is_reviewed);
```

**Query Pattern:**
```sql
SELECT * FROM transactions
WHERE user_id = 'uuid-123' AND is_reviewed = false;
```

**When Used:** Review inbox (show uncategorized transactions)

---

### Index 5: Composite User + Date

```sql
CREATE INDEX idx_transactions_user_date ON transactions(user_id, txn_date DESC);
```

**Query Pattern:**
```sql
SELECT * FROM transactions
WHERE user_id = 'uuid-123'
ORDER BY txn_date DESC;
```

**Why Composite?**
- Single index serves both WHERE and ORDER BY
- More efficient than two separate indexes
- Column order matters: user_id first (filter), then txn_date (sort)

**When Used:** Main dashboard query (most common)

---

### Index 6: Partial Index for Uncategorized

```sql
CREATE INDEX idx_transactions_uncategorized
ON transactions(user_id, category)
WHERE category IS NULL;
```

**This is a PARTIAL INDEX!**

**Query Pattern:**
```sql
SELECT * FROM transactions
WHERE user_id = 'uuid-123' AND category IS NULL;
```

**Why Partial?**
- Only indexes rows WHERE `category IS NULL`
- Smaller index size (~15% of full table)
- Faster queries on review inbox
- Doesn't index already-categorized transactions

**Size Comparison:**
```
Full index:    1M rows × 16 bytes = 16MB
Partial index: 150K rows × 16 bytes = 2.4MB  (85% reduction!)
```

**When Used:** Review inbox (most frequent feature after dashboard)

---

## Design Decisions

### 1. UUID vs Auto-Increment ID

```sql
-- Our choice:
id UUID PRIMARY KEY DEFAULT gen_random_uuid()

-- Alternative:
id SERIAL PRIMARY KEY
```

**Why UUID?**

| Aspect          | UUID                     | SERIAL (INT)        |
|-----------------|--------------------------|---------------------|
| Uniqueness      | Globally unique          | Table-specific      |
| Merge-friendly  | Easy to merge databases  | Conflicts on merge  |
| Security        | Not guessable            | Sequential (1, 2, 3)|
| Size            | 16 bytes                 | 4 bytes             |
| Performance     | Slightly slower inserts  | Faster inserts      |

**Our priorities:** Security, merge-ability > size

**Drawback:** Larger indexes, but acceptable for our scale.

---

### 2. Triggers vs Application Logic

```sql
-- Database trigger (our choice):
CREATE TRIGGER update_transactions_updated_at
BEFORE UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Alternative: Application code
transaction.UpdatedAt = time.Now()
db.Save(transaction)
```

**Why Trigger?**

| Aspect       | Trigger                  | Application Code    |
|--------------|--------------------------|---------------------|
| Consistency  | Always fires             | Might forget        |
| Security     | Can't bypass             | Can be skipped      |
| Performance  | Database-native          | Extra round-trip    |
| Testing      | Works in all contexts    | Need to mock        |

**Our choice:** Triggers for automatic timestamps

---

### 3. Soft Delete vs Hard Delete

```sql
-- Soft delete (not used):
deleted_at TIMESTAMPTZ

-- Hard delete (our choice):
DELETE FROM transactions WHERE id = 'uuid';
```

**Why Hard Delete?**

- GDPR compliance (right to be forgotten)
- Simpler queries (no need to filter deleted rows)
- Smaller indexes
- Can use `ON DELETE CASCADE`

**When Soft Delete Makes Sense:**
- Need audit trail of deletions
- "Undo" feature required
- Regulatory requirement to keep records

**Our case:** Not needed for MVP

---

## Query Patterns

### Pattern 1: Get User's Recent Transactions

```sql
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY txn_date DESC
LIMIT 50;
```

**Indexes Used:**
- `idx_transactions_user_date` (composite)

**Performance:** O(log n + 50) = very fast

---

### Pattern 2: Get Uncategorized Transactions

```sql
SELECT * FROM transactions
WHERE user_id = $1 AND category IS NULL
ORDER BY txn_date DESC;
```

**Indexes Used:**
- `idx_transactions_uncategorized` (partial index)

**Performance:** O(log n) on 15% of data

---

### Pattern 3: Category Breakdown

```sql
SELECT
    category,
    COUNT(*) as txn_count,
    SUM(amount) as total_amount
FROM transactions
WHERE user_id = $1 AND txn_date >= $2
GROUP BY category;
```

**Indexes Used:**
- `idx_transactions_user_date` (for filtering)

**Performance:** O(n) but only on filtered subset

---

### Pattern 4: Monthly Cash Flow

```sql
SELECT
    DATE_TRUNC('month', txn_date) as month,
    SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE 0 END) as inflow,
    SUM(CASE WHEN txn_type = 'debit' THEN amount ELSE 0 END) as outflow
FROM transactions
WHERE user_id = $1
GROUP BY DATE_TRUNC('month', txn_date)
ORDER BY month DESC;
```

**Indexes Used:**
- `idx_transactions_user_id` (for filtering)

**Performance:** Sequential scan on user's data (acceptable for dashboards)

---

## Performance Considerations

### At Different Scales

| Users | Transactions | Storage    | Query Time (indexed) |
|-------|--------------|------------|----------------------|
| 100   | 10K          | ~2MB       | <1ms                 |
| 1K    | 100K         | ~20MB      | ~2ms                 |
| 10K   | 1M           | ~200MB     | ~5ms                 |
| 100K  | 10M          | ~2GB       | ~10ms                |

**Assumptions:**
- ~200 bytes per transaction row
- Proper indexes in place
- Queries filtered by user_id

### When to Optimize

**Watch these metrics:**
- Query time > 100ms consistently
- Index size > 2× table size
- Table size > 10GB

**Optimization Strategies:**
1. **Partitioning:** Split by date (yearly tables)
2. **Archival:** Move old transactions to archive table
3. **Read Replicas:** Separate analytics queries
4. **Materialized Views:** Pre-compute dashboard stats

---

## Future Schema Changes

### Upcoming Tables (Day 4+)

#### 1. categorization_rules

```sql
CREATE TABLE categorization_rules (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    keyword TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Purpose:** Store user-specific categorization overrides

---

#### 2. upload_history

```sql
CREATE TABLE upload_history (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    filename TEXT NOT NULL,
    file_key TEXT NOT NULL,
    total_rows INTEGER,
    categorized_rows INTEGER,
    accuracy_percent DECIMAL(5,2),
    status VARCHAR(20),  -- 'processing', 'completed', 'failed'
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Purpose:** Track upload jobs and accuracy metrics

---

## Conclusion

The schema is designed for:
- ✅ **Fast queries** (6 strategic indexes)
- ✅ **Data integrity** (foreign keys, check constraints)
- ✅ **Audit trail** (raw_data, timestamps)
- ✅ **Scalability** (partitioning-ready)
- ✅ **Type safety** (DECIMAL for money, UUID for IDs)

**Key Principles:**
1. Index for read performance (dashboard-first)
2. Validate at database level (constraints)
3. Keep history (raw_data, timestamps)
4. Design for GDPR (CASCADE deletes)
5. Use PostgreSQL features (partial indexes, triggers)

---

**Document Version:** 1.0.0
**Last Updated:** 2025-11-05
**Next Review:** After 10K transactions in production
