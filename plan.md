# Two-Week MVP Implementation Plan

**Duration:** 10 working days (2 calendar weeks)  
**Team:** 2 full-stack engineers  
**Target:** Ship production-ready MVP with ≥85% auto-categorization accuracy  
**Board:** GitHub Projects (Kanban)  
**Standups:** Daily 10:00 AM IST (15 min)

---

## Success Criteria (Definition of Done)

Before marking Day 10 as complete, ALL of these must be true:

- [ ] All code merged to `main` branch
- [ ] Backend test coverage ≥80% (`go test -cover`)
- [ ] All Playwright E2E tests passing
- [ ] Deployed to staging: `https://staging.cashlens.in`
- [ ] Accuracy experiment completed: ≥85% on 5 real bank CSVs
- [ ] p95 time-to-dashboard ≤60s (measured via Lighthouse)
- [ ] Zero high-severity security issues (OWASP ZAP scan)
- [ ] Docker Compose working on clean Ubuntu 24 machine

---

## Day-by-Day Breakdown

### **Day 0: Project Setup & Infrastructure** ✅ COMPLETE

**Goal:** Runnable local dev environment for both frontend and backend

**Status:** All tasks completed successfully
- [x] Backend project structure created
- [x] Frontend project structure created
- [x] Docker Compose configuration (PostgreSQL, Redis, LocalStack)
- [x] Environment configuration files created
- [x] Go dependencies installed (Fiber v3, pgx, AWS SDK)
- [x] Node dependencies installed (Next.js 15, React 18, Tailwind)
- [x] Backend server running on port 8080
- [x] Frontend server running on port 3000
- [x] All infrastructure services healthy
- [x] LocalStack S3 configured and working

#### Backend Setup (2h) ✅

```bash
# 1. Initialize Go project
mkdir cashlens-api && cd cashlens-api
go mod init github.com/yourorg/cashlens-api

# 2. Install core dependencies
go get github.com/gofiber/fiber/v3
go get github.com/jackc/pgx/v5
go get github.com/golang-jwt/jwt/v5
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/go-playground/validator/v10

# 3. Create project structure
mkdir -p cmd/api internal/{config,database/{migrations,queries},handlers,middleware,models,services,utils}

# 4. Create main.go
cat > cmd/api/main.go << 'EOF'
package main

import (
    "log"
    "github.com/gofiber/fiber/v3"
)

func main() {
    app := fiber.New()

    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    log.Fatal(app.Listen(":8080"))
}
EOF

# 5. Run to verify
go run cmd/api/main.go
# Should see: Fiber v3.0.0 listening on :8080
```

**Files to create:**

- `cmd/api/main.go` - Entry point
- `internal/config/config.go` - Environment loader
- `internal/utils/response.go` - JSON helpers
- `internal/utils/errors.go` - Error types
- `.env.example` - Template config
- `Dockerfile` - Production image
- `docker-compose.yml` - Local stack (Postgres + Redis + API)

#### Frontend Setup (2h)

```bash
# 1. Create Next.js app
npx create-next-app@latest cashlens-web --typescript --tailwind --app --no-src-dir

cd cashlens-web

# 2. Install dependencies
npm install zustand react-hook-form react-dropzone recharts
npm install -D @playwright/test

# 3. Setup shadcn/ui
npx shadcn@latest init -d

# 4. Create folder structure
mkdir -p app/{(auth)/{login,register},(dashboard)/{upload,review}} components/{ui,charts,upload,transactions} lib stores types

# 5. Create API client wrapper
cat > lib/api.ts << 'EOF'
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1';

export async function apiRequest(endpoint: string, options?: RequestInit) {
  const token = localStorage.getItem('token');

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`API Error: ${response.statusText}`);
  }

  return response.json();
}
EOF

# 6. Run dev server
npm run dev
# Should see: Local: http://localhost:3000
```

**Files to create:**

- `lib/api.ts` - API client
- `stores/useAuthStore.ts` - Zustand auth state
- `types/index.ts` - TypeScript definitions
- `.env.local` - Local config
- `next.config.js` - API proxy config

#### Database Setup (1h)

```bash
# 1. Start PostgreSQL via Docker
docker run --name cashlens-db \
  -e POSTGRES_PASSWORD=dev123 \
  -e POSTGRES_DB=cashlens \
  -p 5432:5432 \
  -d postgres:16-alpine

# 2. Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# 3. Create first migration
mkdir -p internal/database/migrations
cat > internal/database/migrations/001_initial.sql << 'EOF'
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
EOF

# 4. Apply migration
psql postgres://postgres:dev123@localhost:5432/cashlens < internal/database/migrations/001_initial.sql
```

**Deliverable:** `main` branch with runnable backend + frontend + database ✅ COMPLETE

**Actual Completion Notes:**
- Go upgraded to 1.25.0 automatically
- Node.js upgraded from 19.4.0 to 20.19.5 for compatibility
- React downgraded from 19 to 18.3.1 for Next.js 15 compatibility
- LocalStack volume configuration fixed (device busy issue resolved)
- Created missing Next.js app files (layout.tsx, page.tsx, globals.css)
- All services verified working: PostgreSQL (5432), Redis (6379), LocalStack (4566)
- Backend API endpoints tested: /health and /v1/ping responding correctly
- Frontend rendering successfully at localhost:3000

---

### **Day 1: Authentication System** (Monday)

**Goal:** User can register, login, and receive JWT token

#### Backend Tasks (4h)

1. **Create auth handlers** (`internal/handlers/auth.go`):

   - `POST /auth/register` - Hash password with bcrypt, insert user, return JWT
   - `POST /auth/login` - Verify password, return JWT

2. **Create JWT middleware** (`internal/middleware/auth.go`):

   - Extract Bearer token from header
   - Validate signature and expiry
   - Inject `user_id` into context

3. **Write unit tests** (`internal/handlers/auth_test.go`):
   - Test successful registration
   - Test duplicate email rejection
   - Test invalid login
   - Test JWT validation

**Key code snippet (auth.go):**

```go
func (h *AuthHandler) Register(c fiber.Ctx) error {
    var req struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=8"`
        FullName string `json:"full_name"`
    }

    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }

    if err := h.validator.Struct(req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)

    user, err := h.db.CreateUser(c.Context(), db.CreateUserParams{
        Email:        req.Email,
        PasswordHash: string(hash),
        FullName:     req.FullName,
    })
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "User creation failed"})
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":   user.ID,
        "email": user.Email,
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, _ := token.SignedString([]byte(h.config.JWTSecret))

    return c.Status(201).JSON(fiber.Map{
        "user":  user,
        "token": tokenString,
    })
}
```

#### Frontend Tasks (3h)

1. **Create login page** (`app/(auth)/login/page.tsx`):

   - Form with email + password
   - Call `/auth/login` API
   - Store JWT in localStorage + Zustand
   - Redirect to `/dashboard`

2. **Create register page** (`app/(auth)/register/page.tsx`):

   - Form with email + password + name
   - Call `/auth/register` API
   - Auto-login after registration

3. **Create auth store** (`stores/useAuthStore.ts`):

```typescript
import { create } from "zustand"

interface AuthState {
  user: { id: string; email: string } | null
  token: string | null
  login: (token: string, user: any) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  login: (token, user) => {
    localStorage.setItem("token", token)
    set({ token, user })
  },
  logout: () => {
    localStorage.removeItem("token")
    set({ token: null, user: null })
  },
}))
```

#### Testing (1h)

- **E2E test:** Register → Login → See dashboard (Playwright)
- **Manual test:** cURL commands for both endpoints

**Deliverable:** Working auth flow with `/auth/register` and `/auth/login`

---

### **Day 2: CSV Parser & Normalization** (Tuesday)

**Goal:** Backend can parse and normalize 5 major Indian bank CSV formats

#### Backend Tasks (5h)

1. **Create parser service** (`internal/services/parser.go`):

   - Implement `DetectSchema()` function (see TechSpec §7.1)
   - Implement `ParseDate()` with multiple format support
   - Handle edge cases: empty rows, malformed amounts

2. **Create test fixtures** (`internal/services/testdata/`):

   - `hdfc_sample.csv`
   - `icici_sample.csv`
   - `sbi_sample.csv`
   - `axis_sample.csv`
   - `kotak_sample.csv`

3. **Write comprehensive tests** (`internal/services/parser_test.go`):
   - Test each bank format
   - Test invalid formats
   - Test date parsing edge cases

**Key test case:**

```go
func TestParseCSV_HDFC(t *testing.T) {
    file, _ := os.Open("testdata/hdfc_sample.csv")
    defer file.Close()

    transactions, err := ParseCSV(file, "HDFC")

    assert.NoError(t, err)
    assert.Equal(t, 50, len(transactions))
    assert.Equal(t, "2024-01-15", transactions[0].Date.Format("2006-01-02"))
    assert.Equal(t, "AWS SERVICES", transactions[0].Description)
    assert.Equal(t, -3500.00, transactions[0].Amount)
}
```

4. **Create transaction model** (`internal/models/transaction.go`):

```go
type Transaction struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    TxnDate     time.Time
    Description string
    Amount      float64
    TxnType     string // "credit" or "debit"
    Category    string
    IsReviewed  bool
}
```

#### Database Tasks (2h)

1. **Create transactions table migration** (`internal/database/migrations/002_transactions.sql`)
2. **Create sqlc queries** (`internal/database/queries/transactions.sql`):

```sql
-- name: CreateTransaction :one
INSERT INTO transactions (user_id, txn_date, description, amount, txn_type, category)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserTransactions :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY txn_date DESC
LIMIT $2 OFFSET $3;
```

3. **Generate Go code:** `sqlc generate`

**Deliverable:** Parser with 100% unit test coverage across 5 bank formats

---

### **Day 3: File Upload Flow** (Wednesday)

**Goal:** User can upload CSV to S3, API processes it into database

#### Backend Tasks (5h)

1. **Implement S3 presigned URL** (`internal/handlers/upload.go`):

```go
func (h *UploadHandler) GetPresignedURL(c fiber.Ctx) error {
    filename := c.Query("filename")
    contentType := c.Query("content_type")

    userID := c.Locals("user_id").(string)
    key := fmt.Sprintf("uploads/%s/%d-%s", userID, time.Now().Unix(), filename)

    presignClient := s3.NewPresignClient(h.s3Client)
    request, _ := presignClient.PresignPutObject(c.Context(), &s3.PutObjectInput{
        Bucket:      aws.String(h.config.S3Bucket),
        Key:         aws.String(key),
        ContentType: aws.String(contentType),
    }, s3.WithPresignExpires(5*time.Minute))

    return c.JSON(fiber.Map{
        "upload_url": request.URL,
        "file_key":   key,
        "expires_in": 300,
    })
}
```

2. **Implement CSV processing endpoint** (`POST /upload/process`):

   - Download file from S3
   - Parse using `parser.ParseCSV()`
   - Bulk insert into `transactions` table
   - Return summary stats

3. **Add file validation**:

   - Check MIME type
   - Limit file size to 5MB
   - Verify extension

4. **Create upload history tracking** (migration + queries)

#### Frontend Tasks (3h)

1. **Create upload page** (`app/(dashboard)/upload/page.tsx`):

```tsx
"use client"

import { useDropzone } from "react-dropzone"
import { useState } from "react"
import { apiRequest } from "@/lib/api"

export default function UploadPage() {
  const [uploading, setUploading] = useState(false)
  const [result, setResult] = useState(null)

  const onDrop = async (acceptedFiles: File[]) => {
    const file = acceptedFiles[0]
    setUploading(true)

    try {
      // 1. Get presigned URL
      const { upload_url, file_key } = await apiRequest(
        `/upload/presign?filename=${file.name}&content_type=${file.type}`
      )

      // 2. Upload directly to S3
      await fetch(upload_url, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": file.type },
      })

      // 3. Trigger backend processing
      const result = await apiRequest("/upload/process", {
        method: "POST",
        body: JSON.stringify({ file_key }),
      })

      setResult(result)
    } catch (error) {
      console.error(error)
    } finally {
      setUploading(false)
    }
  }

  const { getRootProps, getInputProps } = useDropzone({
    onDrop,
    accept: { "text/csv": [".csv"], "application/vnd.ms-excel": [".xlsx"] },
    maxSize: 5 * 1024 * 1024,
  })

  return (
    <div className="max-w-2xl mx-auto p-8">
      <div
        {...getRootProps()}
        className="border-2 border-dashed p-12 text-center"
      >
        <input {...getInputProps()} />
        <p>Drag & drop your bank statement, or click to select</p>
      </div>

      {uploading && <p>Processing...</p>}

      {result && (
        <div className="mt-4">
          <p>Total: {result.total_rows}</p>
          <p>
            Categorized: {result.categorized} ({result.accuracy_percent}%)
          </p>
          <p>Need review: {result.uncategorized}</p>
        </div>
      )}
    </div>
  )
}
```

#### Testing (1h)

- **E2E test:** Upload real CSV → Verify transactions in DB
- **Load test:** 10 concurrent uploads (k6)

**Deliverable:** Working upload flow from browser to database

---

### **Day 4: Rule Engine & Auto-Categorization** (Thursday)

**Goal:** Implement intelligent categorization with 85%+ accuracy

#### Backend Tasks (6h)

1. **Create global rules migration** (`003_global_rules.sql`):

   - Insert seed data from TechSpec §5.2 (50+ rules)

2. **Implement categorizer service** (`internal/services/categorizer.go`):

   - Load global rules into memory
   - Implement `Categorize()` function (see TechSpec §7.2)
   - Add caching for user-specific rules

3. **Integrate with upload processor**:

   - After parsing, run every transaction through categorizer
   - Save category in database
   - Track accuracy metrics

4. **Create categorization_rules table** (migration):

```sql
CREATE TABLE categorization_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    keyword TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, keyword)
);
```

5. **Add endpoint for creating user rules** (`POST /rules`):

```go
func (h *RulesHandler) CreateRule(c fiber.Ctx) error {
    var req struct {
        Keyword  string `json:"keyword" validate:"required"`
        Category string `json:"category" validate:"required"`
    }

    c.Bind().JSON(&req)
    userID := c.Locals("user_id").(string)

    rule, err := h.db.CreateRule(c.Context(), db.CreateRuleParams{
        UserID:   uuid.MustParse(userID),
        Keyword:  strings.ToLower(req.Keyword),
        Category: req.Category,
    })

    // Invalidate cache
    h.categorizer.InvalidateUserCache(uuid.MustParse(userID))

    return c.Status(201).JSON(rule)
}
```

#### Testing Tasks (2h)

1. **Accuracy benchmark test**:

```go
func TestCategorizer_AccuracyBenchmark(t *testing.T) {
    // Load 5 real bank CSVs
    files := []string{"hdfc.csv", "icici.csv", "sbi.csv", "axis.csv", "kotak.csv"}

    totalTxns := 0
    categorizedTxns := 0

    for _, file := range files {
        txns := parseTestFile(file)
        totalTxns += len(txns)

        for _, txn := range txns {
            category := categorizer.Categorize(txn.Description, testUserID)
            if category != "" {
                categorizedTxns++
            }
        }
    }

    accuracy := float64(categorizedTxns) / float64(totalTxns) * 100
    assert.GreaterOrEqual(t, accuracy, 85.0, "Accuracy must be ≥85%")
}
```

2. **Performance test:** Categorize 10,000 transactions < 1 second

**Deliverable:** Rule engine achieving 85%+ accuracy on test data

---

### **Day 5: Smart Review Inbox** (Friday)

**Goal:** User sees only uncategorized transactions, can tag them

#### Backend Tasks (3h)

1. **Create filtered endpoint** (`GET /transactions?status=uncategorized`):

```go
func (h *TransactionHandler) GetTransactions(c fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    status := c.Query("status", "all") // all | categorized | uncategorized

    var txns []db.Transaction

    switch status {
    case "uncategorized":
        txns, _ = h.db.GetUncategorizedTransactions(c.Context(), uuid.MustParse(userID))
    case "categorized":
        txns, _ = h.db.GetCategorizedTransactions(c.Context(), uuid.MustParse(userID))
    default:
        txns, _ = h.db.GetAllTransactions(c.Context(), uuid.MustParse(userID))
    }

    return c.JSON(fiber.Map{"transactions": txns})
}
```

2. **Create update endpoint** (`PUT /transactions/:id`):

   - Validate category exists
   - Update transaction
   - Create user rule if keyword not exists
   - Set `is_reviewed = true`

3. **Add bulk update endpoint** (`PUT /transactions/bulk`)

#### Frontend Tasks (4h)

1. **Create review page** (`app/(dashboard)/review/page.tsx`):

```tsx
"use client"

import { useEffect, useState } from "react"
import { apiRequest } from "@/lib/api"

export default function ReviewPage() {
  const [transactions, setTransactions] = useState([])
  const [categories] = useState([
    "Cloud & Hosting",
    "Payment Processing",
    "Marketing",
    "Salaries",
    "Office Supplies",
    "Team Meals",
  ])

  useEffect(() => {
    loadTransactions()
  }, [])

  const loadTransactions = async () => {
    const data = await apiRequest("/transactions?status=uncategorized")
    setTransactions(data.transactions)
  }

  const updateCategory = async (id: string, category: string) => {
    await apiRequest(`/transactions/${id}`, {
      method: "PUT",
      body: JSON.stringify({ category }),
    })

    // Remove from list
    setTransactions((txns) => txns.filter((t) => t.id !== id))
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">
        Review {transactions.length} Uncategorized Transactions
      </h1>

      <table className="w-full">
        <thead>
          <tr>
            <th>Date</th>
            <th>Description</th>
            <th>Amount</th>
            <th>Category</th>
          </tr>
        </thead>
        <tbody>
          {transactions.map((txn) => (
            <tr key={txn.id}>
              <td>{new Date(txn.txn_date).toLocaleDateString()}</td>
              <td>{txn.description}</td>
              <td>{txn.amount}</td>
              <td>
                <select
                  onChange={(e) => updateCategory(txn.id, e.target.value)}
                >
                  <option value="">Select category</option>
                  {categories.map((cat) => (
                    <option key={cat} value={cat}>
                      {cat}
                    </option>
                  ))}
                </select>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
```

2. **Add category suggestions dropdown** (shadcn Combobox)
3. **Add keyboard shortcuts** (Enter to submit, Arrow keys to navigate)

**Deliverable:** Functional review screen with category tagging

---

### **Day 6: Dashboard KPIs & Net Flow Chart** (Monday)

**Goal:** User sees main dashboard with cash flow metrics

#### Backend Tasks (3h)

1. **Create summary endpoint** (`GET /summary`):

```go
func (h *SummaryHandler) GetSummary(c fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    from := c.Query("from")
    to := c.Query("to")
    groupBy := c.Query("group_by", "month")

    // Get KPIs
    kpis, _ := h.db.GetKPIs(c.Context(), db.GetKPIsParams{
        UserID: uuid.MustParse(userID),
        From:   parseDate(from),
        To:     parseDate(to),
    })

    // Get net flow trend
    trend, _ := h.db.GetNetFlowTrend(c.Context(), db.GetNetFlowTrendParams{
        UserID:  uuid.MustParse(userID),
        From:    parseDate(from),
        To:      parseDate(to),
        GroupBy: groupBy,
    })

    return c.JSON(fiber.Map{
        "kpis":           kpis,
        "net_flow_trend": trend,
    })
}
```

2. **Create SQL aggregation queries** (`internal/database/queries/summary.sql`):

```sql
-- name: GetKPIs :one
SELECT
    SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE 0 END) AS total_inflow,
    SUM(CASE WHEN txn_type = 'debit' THEN amount ELSE 0 END) AS total_outflow,
    SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE -amount END) AS net_cash_flow,
    COUNT(*) AS transaction_count
FROM transactions
WHERE user_id = $1
  AND txn_date BETWEEN $2 AND $3;

-- name: GetNetFlowTrend :many
SELECT
    DATE_TRUNC($4, txn_date) AS period,
    SUM(CASE WHEN txn_type = 'credit' THEN amount ELSE -amount END) AS net_flow
FROM transactions
WHERE user_id = $1
  AND txn_date BETWEEN $2 AND $3
GROUP BY DATE_TRUNC($4, txn_date)
ORDER BY period;
```

#### Frontend Tasks (4h)

1. **Create dashboard page** (`app/(dashboard)/page.tsx`):

```tsx
"use client"

import { useEffect, useState } from "react"
import { apiRequest } from "@/lib/api"
import NetCashFlowChart from "@/components/charts/NetCashFlowChart"

export default function DashboardPage() {
  const [summary, setSummary] = useState(null)

  useEffect(() => {
    loadSummary()
  }, [])

  const loadSummary = async () => {
    const data = await apiRequest("/summary?from=2024-01-01&to=2024-12-31")
    setSummary(data)
  }

  if (!summary) return <div>Loading...</div>

  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold mb-8">Dashboard</h1>

      {/* KPI Cards */}
      <div className="grid grid-cols-3 gap-4 mb-8">
        <div className="p-6 bg-white rounded-lg shadow">
          <p className="text-sm text-gray-600">Net Cash Flow</p>
          <p
            className={`text-3xl font-bold ${
              summary.kpis.net_cash_flow >= 0
                ? "text-green-600"
                : "text-red-600"
            }`}
          >
            ₹{summary.kpis.net_cash_flow.toLocaleString()}
          </p>
        </div>

        <div className="p-6 bg-white rounded-lg shadow">
          <p className="text-sm text-gray-600">Total Inflow</p>
          <p className="text-3xl font-bold text-green-600">
            ₹{summary.kpis.total_inflow.toLocaleString()}
          </p>
        </div>

        <div className="p-6 bg-white rounded-lg shadow">
          <p className="text-sm text-gray-600">Total Outflow</p>
          <p className="text-3xl font-bold text-red-600">
            ₹{summary.kpis.total_outflow.toLocaleString()}
          </p>
        </div>
      </div>

      {/* Net Flow Chart */}
      <NetCashFlowChart data={summary.net_flow_trend} />
    </div>
  )
}
```

2. **Create chart component** (`components/charts/NetCashFlowChart.tsx`):

```tsx
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts"

export default function NetCashFlowChart({ data }) {
  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h2 className="text-xl font-bold mb-4">Net Cash Flow Trend</h2>
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="period" />
          <YAxis />
          <Tooltip />
          <Bar
            dataKey="net_flow"
            fill={(entry) => (entry.net_flow >= 0 ? "#10b981" : "#ef4444")}
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
```

**Deliverable:** Dashboard showing KPIs and net cash flow chart

---

### **Day 7: Top Expenses Chart** (Tuesday)

**Goal:** Add horizontal bar chart showing top expense categories

#### Backend Tasks (2h)

1. **Create top expenses query** (`queries/summary.sql`):

```sql
-- name: GetTopExpenses :many
SELECT
    category,
    SUM(amount) AS total_amount,
    COUNT(*) AS txn_count,
    ROUND((SUM(amount) / total_outflow.sum) * 100, 2) AS percent
FROM transactions
CROSS JOIN (
    SELECT SUM(amount) AS sum
    FROM transactions
    WHERE user_id = $1 AND txn_type = 'debit' AND txn_date BETWEEN $2 AND $3
) AS total_outflow
WHERE user_id = $1
  AND txn_type = 'debit'
  AND category IS NOT NULL
  AND txn_date BETWEEN $2 AND $3
GROUP BY category, total_outflow.sum
ORDER BY total_amount DESC
LIMIT 10;
```

2. **Add to summary endpoint response**

#### Frontend Tasks (3h)

1. **Create expenses chart** (`components/charts/ExpensesChart.tsx`):

```tsx
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts"

const COLORS = ["#ef4444", "#f97316", "#f59e0b", "#eab308", "#84cc16"]

export default function ExpensesChart({ data }) {
  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h2 className="text-xl font-bold mb-4">Top Expense Categories</h2>
      <ResponsiveContainer width="100%" height={400}>
        <BarChart data={data} layout="vertical">
          <XAxis type="number" />
          <YAxis dataKey="category" type="category" width={150} />
          <Tooltip />
          <Bar dataKey="total_amount" radius={[0, 8, 8, 0]}>
            {data.map((entry, index) => (
              <Cell key={index} fill={COLORS[index % COLORS.length]} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
```

2. **Add to dashboard page** (below net flow chart)

**Deliverable:** Complete dashboard with both charts

---

### **Day 8: Security Hardening** (Wednesday)

**Goal:** Pass OWASP ZAP scan with zero high-severity issues

#### Backend Tasks (4h)

1. **Add rate limiting** (`internal/middleware/ratelimit.go`):

```go
import "github.com/gofiber/fiber/v3/middleware/limiter"

func RateLimiter() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        20,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c fiber.Ctx) string {
            return c.Locals("user_id").(string)
        },
        LimitReached: func(c fiber.Ctx) error {
            return c.Status(429).JSON(fiber.Map{
                "error": "Too many requests",
            })
        },
    })
}
```

2. **Add security headers middleware**:

```go
func SecurityHeaders() fiber.Handler {
    return func(c fiber.Ctx) error {
        c.Set("X-Content-Type-Options", "nosniff")
        c.Set("X-Frame-Options", "DENY")
        c.Set("X-XSS-Protection", "1; mode=block")
        c.Set("Strict-Transport-Security", "max-age=31536000")
        c.Set("Content-Security-Policy", "default-src 'self'")
        return c.Next()
    }
}
```

3. **Implement file validation** (enhance upload handler):

```go
func ValidateFile(fileHeader *multipart.FileHeader) error {
    // Size check
    if fileHeader.Size > 5*1024*1024 {
        return errors.New("file too large")
    }

    // Extension check
    ext := filepath.Ext(fileHeader.Filename)
    if ext != ".csv" && ext != ".xlsx" {
        return errors.New("invalid file type")
    }

    // Magic byte check
    file, _ := fileHeader.Open()
    buffer := make([]byte, 512)
    file.Read(buffer)
    mimeType := http.DetectContentType(buffer)

    if !strings.HasPrefix(mimeType, "text/") && !strings.Contains(mimeType, "spreadsheet") {
        return errors.New("file content mismatch")
    }

    return nil
}
```

4. **Add SQL injection tests**:

```go
func TestSQLInjection_LoginEndpoint(t *testing.T) {
    maliciousPayload := `{"email": "admin'--", "password": "anything"}`

    resp := makeRequest("POST", "/auth/login", maliciousPayload)

    // Should return 401, not 500 or 200
    assert.Equal(t, 401, resp.StatusCode)
}
```

#### Security Scanning (3h)

1. **Run OWASP ZAP scan:**

```bash
docker run -t ghcr.io/zaproxy/zaproxy:stable \
  zap-baseline.py -t http://localhost:8080 \
  -r zap-report.html
```

2. **Fix all high/medium issues**

3. **Run trivy container scan:**

```bash
trivy image cashlens-api:latest
```

**Deliverable:** Clean security scan reports

---

### **Day 9: Performance & Accuracy Testing** (Thursday)

**Goal:** Validate 85%+ accuracy and <60s time-to-dashboard

#### Accuracy Experiment (3h)

1. **Collect 5 real bank CSVs** (anonymized):

   - HDFC (100 rows)
   - ICICI (120 rows)
   - SBI (90 rows)
   - Axis (110 rows)
   - Kotak (80 rows)

2. **Run categorization test:**

```bash
# Upload each file via API
for file in hdfc.csv icici.csv sbi.csv axis.csv kotak.csv; do
  curl -X POST http://localhost:8080/v1/upload/process \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@$file"
done

# Query accuracy
psql -c "
SELECT
  COUNT(*) AS total,
  COUNT(category) AS categorized,
  ROUND(COUNT(category)::NUMERIC / COUNT(*) * 100, 2) AS accuracy
FROM transactions
WHERE user_id = '$USER_ID';
"
```

3. **Document results:**
   - Mean accuracy: \_\_\_%
   - Min accuracy: \_\_\_%
   - Max accuracy: \_\_\_%
   - Pass/Fail: \_\_\_

#### Performance Testing (3h)

1. **Time-to-dashboard E2E test:**

```typescript
// tests/e2e/performance.spec.ts
test("time to dashboard < 60s", async ({ page }) => {
  const startTime = Date.now()

  await page.goto("/login")
  await page.fill('[name="email"]', "test@example.com")
  await page.fill('[name="password"]', "password")
  await page.click('button[type="submit"]')

  await page.goto("/upload")
  await page.setInputFiles('input[type="file"]', "./fixtures/hdfc-500rows.csv")
  await page.click('button:has-text("Upload")')

  await page.waitForSelector("text=Dashboard")
  await page.waitForSelector('[data-testid="net-cash-flow"]')

  const elapsed = Date.now() - startTime
  console.log(`Time to dashboard: ${elapsed}ms`)

  expect(elapsed).toBeLessThan(60000)
})
```

2. **Run k6 load test:**

```bash
k6 run tests/load/upload-load.js
```

3. **Lighthouse audit:**

```bash
lighthouse http://localhost:3000/dashboard --output=html
```

**Deliverable:** Performance report with all metrics green

---

### **Day 10: Demo Video & Sprint Review** (Friday)

**Goal:** Production deployment + investor-ready demo

#### Deployment Tasks (3h)

1. **Create production docker-compose.yml:**

```yaml
version: "3.8"
services:
  api:
    image: cashlens-api:latest
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: ${DATABASE_URL}
      JWT_SECRET: ${JWT_SECRET}
      AWS_REGION: ap-south-1
    depends_on:
      - db

  web:
    image: cashlens-web:latest
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: https://api.cashlens.dev/v1

  db:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}

volumes:
  postgres_data:
```

2. **Deploy to staging:**

```bash
# SSH into EC2
ssh ubuntu@staging.cashlens.dev

# Pull latest code
git pull origin main

# Deploy
docker-compose up -d --build

# Run migrations
docker exec cashlens-api ./migrate up

# Verify health
curl https://api.cashlens.dev/health
```

#### Demo Recording (2h)

Create 3-minute demo video showing:

1. **Registration** (10s)
2. **Upload CSV** (20s)
3. **Auto-categorization result** (15s) - "412 of 487 categorized (84.6%)"
4. **Smart review** (30s) - Tag 5 transactions
5. **Dashboard reveal** (45s) - Emphasize net cash flow
6. **Conclusion** (20s) - "From CSV to insights in 60 seconds"

#### Sprint Review (2h)

**Attendees:** Team + 1 external advisor

**Agenda:**

1. Live demo (5 min)
2. Accuracy results (3 min)
3. Performance metrics (3 min)
4. Security scan results (2 min)
5. Architecture walkthrough (5 min)
6. Open discussion (10 min)

**Go/No-Go Decision:**

- [ ] All acceptance criteria met?
- [ ] Accuracy ≥85%?
- [ ] Performance <60s?
- [ ] Security scan clean?
- [ ] **Decision:** Proceed to beta / 1-week remediation

**Deliverable:** Deployed MVP + demo video + go-live decision

---

## Specialized Development Agents

The project has access to specialized AI agents that should be used throughout the development process:

### When to Use Agents in Daily Workflow

**Day 1 (Authentication):**
- Use `senior-engineer` to review Clerk integration architecture
- Use `code-reviewer` after implementing Clerk middleware

**Day 2 (CSV Parser):**
- Use `backend-development:tdd-orchestrator` to implement parser with tests first
- Use `golang-pro` to ensure idiomatic Go code with proper error handling
- Use `code-reviewer` before committing parser implementation

**Day 3 (File Upload):**
- Use `backend-architect` to design S3 presigned URL flow
- Use `golang-pro` for S3 client implementation
- Use `code-reviewer` to check for security vulnerabilities in upload handler

**Day 4 (Categorization Engine):**
- Use `database-architect` to optimize rules table schema
- Use `backend-development:tdd-orchestrator` for categorizer tests
- Use `golang-pro` to optimize rule matching algorithm
- Use `code-reviewer` for accuracy validation

**Day 5 (Review Inbox):**
- Use `backend-architect` to design filtered transactions API
- Use `frontend-developer` for review UI components
- Use `code-reviewer` for security check on update endpoints

**Day 6-7 (Dashboard):**
- Use `database-architect` to optimize aggregation queries
- Use `frontend-developer` for chart components and responsive design
- Use `code-reviewer` for performance validation

**Day 8 (Security):**
- Use `code-reviewer` proactively for security audit
- Use `senior-engineer` to review overall security architecture
- Use `golang-pro` to ensure proper error handling and input validation

**Day 9 (Performance Testing):**
- Use `database-architect` for index optimization
- Use `golang-pro` for concurrency optimization
- Use `senior-engineer` for architecture review

**Day 10 (Documentation):**
- Use `code-documentation:docs-architect` for API documentation
- Use `code-documentation:tutorial-engineer` for user onboarding guides
- Use `senior-engineer` for final architecture review

### Agent Best Practices

1. **Always use TDD orchestrator** for new features to ensure tests are written first
2. **Use code-reviewer proactively** after completing any significant code (don't wait to be asked)
3. **Use database-architect** before creating migrations to ensure optimal schema design
4. **Use golang-pro** for any Go code to ensure idiomatic patterns
5. **Use frontend-developer** for all React/Next.js components
6. **Use senior-engineer** for major architectural decisions

## Risk Mitigation Strategies

### R1: CSV Format Variance (Probability: HIGH, Impact: HIGH)

**Mitigation:**

- Build normalization matrix on Day 2
- Use `backend-development:tdd-orchestrator` to ensure comprehensive test coverage
- Test with real CSVs early
- Use `golang-pro` to optimize parser performance
- Fallback: Manual column mapping UI (not implemented in MVP, but schema-ready)

### R2: Rule Engine Performance (Probability: MEDIUM, Impact: MEDIUM)

**Mitigation:**

- Cache user rules in memory (TTL 5 min)
- Limit rule count to 500 per user
- Monitor query performance with indexes

### R3: <85% Accuracy (Probability: MEDIUM, Impact: HIGH)

**Mitigation:**

- Expand global keyword list by Day 8
- Implement fuzzy matching (Levenshtein distance ≤2)
- Priority: Focus on top 10 expense categories

### R4: Security Scan Failures (Probability: LOW, Impact: HIGH)

**Mitigation:**

- Day 9 buffer allocated
- Run trivy + ZAP in CI on every PR
- Have backup: Use managed WAF (CloudFlare)

---

## Post-MVP Checklist (Beta Readiness)

After Day 10, before opening private beta:

- [ ] Privacy policy published (Delhi law firm template)
- [ ] Terms of service published
- [ ] Intercom widget embedded for support
- [ ] Error tracking setup (Sentry)
- [ ] Analytics tracking (Plausible)
- [ ] Email onboarding sequence drafted
- [ ] Wait-list ≥50 signups
- [ ] Backup/restore procedure documented
- [ ] Incident response runbook created
- [ ] Beta feedback form embedded

---

## Daily Standup Template

**Yesterday:**

- What did I complete?
- What blockers did I face?

**Today:**

- What will I complete?
- What help do I need?

**Blockers:**

- What's preventing progress?

**Metrics:**

- Test coverage: \_\_\_%
- Lines of code: \_\_\_
- Open PRs: \_\_\_

---

## Success Metrics Dashboard

Track these daily on GitHub Projects:

| Metric                       | Target | Current | Status |
| ---------------------------- | ------ | ------- | ------ |
| Backend test coverage        | ≥80%   | \_\_\_% | 🟡     |
| Frontend test coverage       | ≥70%   | \_\_\_% | 🟡     |
| Auto-categorization accuracy | ≥85%   | \_\_\_% | 🟡     |
| p95 time-to-dashboard        | ≤60s   | \_\_\_s | 🟡     |
| OWASP high-severity issues   | 0      | \_\_\_  | 🟡     |
| Playwright E2E pass rate     | 100%   | \_\_\_% | 🟡     |

---

## Emergency Contacts

- **AWS Support:** support.aws.amazon.com
- **Database Issues:** DBA on-call
- **Security Incidents:** security@yourcompany.com
- **Deployment Issues:** DevOps Slack channel

---

**Plan Status:** Ready for Execution  
**Last Updated:** 2025-10-31  
**Next Review:** After Day 5 (mid-sprint checkpoint)
