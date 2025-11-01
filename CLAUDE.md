# CLAUDE.md - cashlens Test-Driven Development Guide

Read techspec.md before conducting your work.

Always follow the instructions in plan.md(...).
Please always consult with techspec.md(...)

## Primary Directive

When you say "go", I will find the next unmarked checkbox in plan.md, implement the test for that feature first, then implement only enough code to make that test pass. This is the fundamental rule that overrides all other considerations for the cashlens MVP development.

Add `ultrathink` keyword to every single request when complex decision-making is required.

---

## Table of Contents

1. [Role Definition and Core Responsibilities](#1-role-definition-and-core-responsibilities)
2. [Test-Driven Development for cashlens](#2-test-driven-development-for-cashlens)
3. [Commit Discipline](#3-commit-discipline)
4. [Go Backend Development Guidelines](#4-go-backend-development-guidelines)
5. [Next.js Frontend Patterns](#5-nextjs-frontend-patterns)
6. [Database and Migration Patterns](#6-database-and-migration-patterns)
7. [CSV Processing and Categorization](#7-csv-processing-and-categorization)
8. [Integration Testing Strategy](#8-integration-testing-strategy)
9. [Performance and Accuracy Validation](#9-performance-and-accuracy-validation)

---

## 1. Role Definition and Core Responsibilities

You are working as a senior full-stack engineer building cashlens, a financial analytics SaaS for Indian SMBs. Your primary responsibility is to implement features using strict Test-Driven Development, ensuring:

1. **Every feature starts with a failing test**
2. **Minimum code to pass the test**
3. **Refactor only after tests pass**
4. **85%+ auto-categorization accuracy**
5. **60-second time-to-dashboard**

### Core Technical Stack

- **Backend**: Go (Fiber v3) + PostgreSQL + Redis
- **Frontend**: Next.js 15 + TypeScript + Tailwind + shadcn/ui
- **Auth**: Clerk (Phase 1 MVP)
- **Storage**: AWS S3 (presigned URLs)
- **Testing**: Go testing + Playwright E2E

---

## 2. Test-Driven Development for cashlens

### The Three Laws Applied to cashlens

**First Law**: Before implementing any API endpoint, write the test that calls it and asserts expected behavior.

**Second Law**: Before implementing any CSV parser logic, write tests with sample bank CSVs that should pass.

**Third Law**: Before implementing categorization rules, write tests asserting 85%+ accuracy on known transaction sets.

### Red-Green-Refactor for Financial Features

#### Example: CSV Upload Endpoint

**Step 1 - RED (Write Failing Test)**

```go
// internal/handlers/upload_test.go
func TestGetPresignedURL_Success(t *testing.T) {
    // Setup
    app := setupTestApp(t)

    // Test request
    req := httptest.NewRequest("GET", "/v1/upload/presign?filename=hdfc.csv&content_type=text/csv", nil)
    req.Header.Set("Authorization", "Bearer "+validTestToken())

    resp, err := app.Test(req)
    require.NoError(t, err)

    // Assertions
    assert.Equal(t, fiber.StatusOK, resp.StatusCode)

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    assert.NotEmpty(t, result["upload_url"])
    assert.NotEmpty(t, result["file_key"])
    assert.Equal(t, 300.0, result["expires_in"])
}
```

**Step 2 - GREEN (Minimal Implementation)**

```go
// internal/handlers/upload.go
type UploadHandler struct {
    s3Client *s3.Client
    config   *config.Config
}

func (h *UploadHandler) GetPresignedURL(c fiber.Ctx) error {
    filename := c.Query("filename")
    contentType := c.Query("content_type")

    if filename == "" || contentType == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "filename and content_type required",
        })
    }

    userID := c.Locals("user_id").(string)
    key := fmt.Sprintf("uploads/%s/%d-%s", userID, time.Now().Unix(), filename)

    presignClient := s3.NewPresignClient(h.s3Client)
    request, err := presignClient.PresignPutObject(c.Context(), &s3.PutObjectInput{
        Bucket:      aws.String(h.config.S3Bucket),
        Key:         aws.String(key),
        ContentType: aws.String(contentType),
    }, s3.WithPresignExpires(5*time.Minute))

    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "failed to generate presigned URL",
        })
    }

    return c.JSON(fiber.Map{
        "upload_url": request.URL,
        "file_key":   key,
        "expires_in": 300,
    })
}
```

**Step 3 - REFACTOR (After Tests Pass)**

Only after all tests pass, extract common logic:

```go
// internal/utils/s3.go
type S3URLGenerator struct {
    client *s3.Client
    bucket string
}

func (g *S3URLGenerator) GenerateUploadURL(ctx context.Context, key, contentType string, expires time.Duration) (string, error) {
    presignClient := s3.NewPresignClient(g.client)
    request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(g.bucket),
        Key:         aws.String(key),
        ContentType: aws.String(contentType),
    }, s3.WithPresignExpires(expires))

    if err != nil {
        return "", fmt.Errorf("presign failed: %w", err)
    }

    return request.URL, nil
}
```

### Testing the Categorization Engine (Critical for MVP)

The categorization engine must achieve 85%+ accuracy. Test first:

```go
// internal/services/categorizer_test.go
func TestCategorizer_AccuracyBenchmark(t *testing.T) {
    categorizer := NewCategorizer()

    testFiles := []string{
        "testdata/hdfc_100rows.csv",
        "testdata/icici_120rows.csv",
        "testdata/sbi_90rows.csv",
        "testdata/axis_110rows.csv",
        "testdata/kotak_80rows.csv",
    }

    totalTxns := 0
    categorizedTxns := 0

    for _, file := range testFiles {
        transactions := parseTestFile(t, file)
        totalTxns += len(transactions)

        for _, txn := range transactions {
            category := categorizer.Categorize(txn.Description)
            if category != "" {
                categorizedTxns++
            }
        }
    }

    accuracy := float64(categorizedTxns) / float64(totalTxns) * 100

    // CRITICAL: Must be >= 85%
    assert.GreaterOrEqual(t, accuracy, 85.0,
        "Categorization accuracy must be at least 85%%, got %.2f%%", accuracy)

    t.Logf("Accuracy: %.2f%% (%d/%d transactions)", accuracy, categorizedTxns, totalTxns)
}
```

Implement with global rules (techspec §5.2):

```go
// internal/services/categorizer.go
type Categorizer struct {
    globalRules map[string]string // keyword -> category
    userRules   map[string]map[string]string // userID -> (keyword -> category)
    mu          sync.RWMutex
}

func NewCategorizer() *Categorizer {
    return &Categorizer{
        globalRules: loadGlobalRules(),
        userRules:   make(map[string]map[string]string),
    }
}

func loadGlobalRules() map[string]string {
    return map[string]string{
        // Cloud & Hosting
        "aws":           "Cloud & Hosting",
        "amazon web":    "Cloud & Hosting",
        "digitalocean":  "Cloud & Hosting",
        "heroku":        "Cloud & Hosting",
        "vercel":        "Cloud & Hosting",

        // Payment Processing
        "razorpay":      "Payment Processing",
        "stripe":        "Payment Processing",
        "paytm":         "Payment Processing",
        "phonepe":       "Payment Processing",

        // Marketing
        "google ads":    "Marketing",
        "facebook ads":  "Marketing",
        "linkedin ads":  "Marketing",
        "instagram":     "Marketing",

        // Salaries
        "salary":        "Salaries",
        "payroll":       "Salaries",
        "emp salary":    "Salaries",

        // Office Supplies
        "amazon":        "Office Supplies",
        "flipkart":      "Office Supplies",
        "office depot":  "Office Supplies",

        // Add more rules from techspec §5.2...
    }
}

func (c *Categorizer) Categorize(description string) string {
    desc := strings.ToLower(description)

    c.mu.RLock()
    defer c.mu.RUnlock()

    // Check global rules
    for keyword, category := range c.globalRules {
        if strings.Contains(desc, keyword) {
            return category
        }
    }

    return "" // Uncategorized
}
```

### File Operation Protocol for cashlens

#### Before Any File Modification

Before modifying, adding, or deleting any files:

1. **Use serena mcp** to accurately pinpoint the file to edit
2. **Run zen mcp** with Gemini 2.5 Pro for code review
3. **Repeat** until zen mcp gives OK

#### After Any File Modification

After any file modification:

1. **Run tests**: `go test -v ./...` (backend) or `npm test` (frontend)
2. **Run linters**: `go fmt ./...` && `go vet ./...`
3. **Run zen mcp** with Gemini 2.5 Pro for review
4. **Repeat** until all tests pass and zen mcp gives OK

---

## 3. Commit Discipline

### Commit Requirements

**Only commit when ALL of the following are true:**

1. ✅ **ALL tests are passing**
   - Backend: `go test -v ./...` exits with code 0
   - Frontend: `npm test` passes
   - E2E: Playwright tests pass (if applicable)

2. ✅ **ALL compiler/linter warnings resolved**
   - Go: `go fmt ./...` && `go vet ./...` clean
   - TypeScript: `npm run lint` passes
   - No build warnings

3. ✅ **Single logical unit of work**
   - One feature, one fix, one refactor
   - Not mixing multiple unrelated changes
   - Atomic and reversible

4. ✅ **Clear structural vs behavioral change indication**
   - Commit message states what changed (structure) and why (behavior)

### Conventional Commits with Gitmoji

We follow [Conventional Commits](https://www.conventionalcommits.org/) with [Gitmoji](https://gitmoji.dev/) for visual clarity.

#### Format

```
<gitmoji> <type>(<scope>): <subject>

<body>

<footer>
```

#### Types

| Type | Gitmoji | Description | Example |
|------|---------|-------------|---------|
| **feat** | ✨ | New feature | `✨ feat(auth): add Clerk JWT validation` |
| **fix** | 🐛 | Bug fix | `🐛 fix(parser): handle empty CSV rows` |
| **docs** | 📝 | Documentation only | `📝 docs(readme): add quick start guide` |
| **style** | 💄 | Code style/formatting | `💄 style(api): format with gofmt` |
| **refactor** | ♻️ | Code refactor | `♻️ refactor(categorizer): extract rule loader` |
| **perf** | ⚡️ | Performance improvement | `⚡️ perf(db): add index on transactions.user_id` |
| **test** | ✅ | Adding/updating tests | `✅ test(parser): add HDFC CSV test cases` |
| **build** | 👷 | Build system/dependencies | `👷 build(deps): update fiber to v3.1.0` |
| **ci** | 💚 | CI/CD changes | `💚 ci(github): add test workflow` |
| **chore** | 🔧 | Maintenance tasks | `🔧 chore(env): update .env.example` |
| **revert** | ⏪ | Revert previous commit | `⏪ revert: remove broken migration` |
| **init** | 🎉 | Initial commit | `🎉 init: project setup` |

#### Scopes (cashlens-specific)

- `auth` - Authentication & authorization
- `parser` - CSV parsing logic
- `categorizer` - Transaction categorization
- `upload` - File upload handling
- `dashboard` - Dashboard & KPIs
- `api` - API endpoints
- `db` - Database & migrations
- `config` - Configuration
- `tests` - Testing infrastructure
- `infra` - Infrastructure (Docker, etc.)

#### Commit Message Examples

**Good Examples:**

```bash
# New feature with tests
✨ feat(parser): detect HDFC bank CSV schema

- Implement DetectSchema() function
- Support Date, Narration, Withdrawal Amt columns
- Add test with real HDFC sample CSV
- Achieves 100% detection accuracy on test file

Closes #12
```

```bash
# Bug fix
🐛 fix(categorizer): prevent case-sensitive keyword matching

- Convert descriptions to lowercase before matching
- Update global rules to use lowercase keywords
- Add test case for "AWS" vs "aws"

Fixes accuracy drop from 87% to 91%
```

```bash
# Refactoring
♻️ refactor(api): extract error handling to middleware

- Move error response logic to utils/errors.go
- Create APIError type with status codes
- Update all handlers to use new error types

No behavioral changes. All tests passing.
```

```bash
# Documentation
📝 docs(claude): add commit discipline guidelines

- Document conventional commit format
- Add gitmoji reference table
- Include cashlens-specific scopes
- Provide good/bad examples
```

**Bad Examples (Don't do this):**

```bash
# ❌ Too vague
fix: stuff

# ❌ No gitmoji
feat(auth): add login

# ❌ Mixed concerns
✨ feat: add parser + fix bugs + update docs

# ❌ No body for complex change
♻️ refactor(db): rewrite entire schema
```

#### Breaking Changes

If a commit introduces breaking changes, add `BREAKING CHANGE:` in the footer:

```bash
✨ feat(api): change transaction response format

- Rename `txn_date` to `date`
- Rename `txn_type` to `type`
- Add `formatted_amount` field

BREAKING CHANGE: API response format changed. Frontend must update.
```

#### Commit Frequency

- **Prefer small, frequent commits** over large, infrequent ones
- Commit after each passing test in TDD cycle (Red → Green → Commit → Refactor → Commit)
- Typical workflow:
  1. Write test (don't commit yet)
  2. Implement code to pass test
  3. Run all tests → **Commit** ✅
  4. Refactor code
  5. Run all tests → **Commit** ✅

#### Git Workflow for cashlens

```bash
# 1. Make changes following TDD
# 2. Run tests
go test -v ./...

# 3. Stage changes
git add .

# 4. Commit with gitmoji and conventional format
git commit -m "✨ feat(parser): detect ICICI bank CSV schema

- Implement schema detection for ICICI format
- Add test with real ICICI sample CSV
- Update DetectSchema() to handle Transaction Date column

Tests passing: 15/15
Coverage: 87%"

# 5. Push to remote
git push origin main
```

#### Pre-commit Checklist

Before running `git commit`, verify:

- [ ] All tests passing (`go test -v ./...` or `npm test`)
- [ ] No linter errors (`go vet ./...` or `npm run lint`)
- [ ] Code formatted (`go fmt ./...` or `npm run format`)
- [ ] Commit message follows format: `<gitmoji> <type>(<scope>): <subject>`
- [ ] Single logical change (not mixing features/fixes)
- [ ] Body explains "why" not just "what"

#### Integration with CI/CD (Future)

When CI/CD is set up, commits will automatically:

1. Run all tests
2. Run linters
3. Check commit message format
4. Block merge if any check fails

---

## 4. Go Backend Development Guidelines

### Project Structure (Aligned with plan.md)

```
cashlens-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Environment configuration
│   ├── database/
│   │   ├── migrations/          # SQL migrations
│   │   └── queries/             # sqlc queries
│   ├── handlers/
│   │   ├── auth.go              # Clerk JWT validation (MVP)
│   │   ├── upload.go            # S3 presigned URLs
│   │   ├── transactions.go      # CRUD operations
│   │   └── summary.go           # Dashboard KPIs
│   ├── middleware/
│   │   ├── auth.go              # Clerk JWT middleware
│   │   └── cors.go              # CORS configuration
│   ├── models/
│   │   └── transaction.go       # Domain models
│   ├── services/
│   │   ├── parser.go            # CSV parsing (§7.1)
│   │   ├── categorizer.go       # Rule engine (§7.2)
│   │   └── storage.go           # S3 operations
│   └── utils/
│       ├── response.go          # JSON helpers
│       └── errors.go            # Error handling
├── testdata/                    # Sample CSVs
├── go.mod
└── go.sum
```

### Error Handling Pattern

Always use structured errors with context:

```go
// internal/utils/errors.go
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

// Middleware to handle APIError
func ErrorHandler(c fiber.Ctx, err error) error {
    apiErr, ok := err.(*APIError)
    if !ok {
        apiErr = NewInternalError(err)
    }

    return c.Status(apiErr.StatusCode).JSON(apiErr)
}
```

### Configuration Management

```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    // Server
    Port            int
    Environment     string
    ShutdownTimeout time.Duration

    // Database
    DatabaseURL         string
    DBMaxConnections    int
    DBConnectionTimeout time.Duration

    // Clerk Auth
    ClerkPublishableKey string
    ClerkSecretKey      string

    // S3
    S3Bucket    string
    S3Region    string
    AWSEndpoint string // For LocalStack

    // Feature Flags
    EnableRateLimiting bool
}

func LoadFromEnv() (*Config, error) {
    cfg := &Config{
        Port:                getEnvInt("PORT", 8080),
        Environment:         getEnv("ENVIRONMENT", "development"),
        ShutdownTimeout:     getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
        DatabaseURL:         getEnv("DATABASE_URL", ""),
        DBMaxConnections:    getEnvInt("DB_MAX_CONNECTIONS", 25),
        DBConnectionTimeout: getEnvDuration("DB_CONNECTION_TIMEOUT", 30*time.Second),
        ClerkPublishableKey: getEnv("CLERK_PUBLISHABLE_KEY", ""),
        ClerkSecretKey:      getEnv("CLERK_SECRET_KEY", ""),
        S3Bucket:            getEnv("S3_BUCKET", ""),
        S3Region:            getEnv("S3_REGION", "ap-south-1"),
        AWSEndpoint:         getEnv("AWS_ENDPOINT", ""),
        EnableRateLimiting:  getEnvBool("ENABLE_RATE_LIMITING", false),
    }

    // Validate required fields
    if cfg.DatabaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }
    if cfg.ClerkSecretKey == "" {
        return nil, fmt.Errorf("CLERK_SECRET_KEY is required")
    }
    if cfg.S3Bucket == "" {
        return nil, fmt.Errorf("S3_BUCKET is required")
    }

    return cfg, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

### Clerk Authentication Middleware

```go
// internal/middleware/auth.go
package middleware

import (
    "fmt"
    "strings"

    "github.com/clerk/clerk-sdk-go/v2"
    "github.com/clerk/clerk-sdk-go/v2/jwt"
    "github.com/gofiber/fiber/v3"
)

func ClerkAuth(secretKey string) fiber.Handler {
    clerk.SetKey(secretKey)

    return func(c fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "missing authorization header",
            })
        }

        // Extract Bearer token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid authorization header format",
            })
        }

        token := parts[1]

        // Verify JWT with Clerk
        claims, err := jwt.Verify(c.Context(), &jwt.VerifyParams{
            Token: token,
        })
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid token",
            })
        }

        // Store user ID in context
        c.Locals("user_id", claims.Subject)

        return c.Next()
    }
}

// Optional: Middleware to require specific roles
func RequireRole(role string) fiber.Handler {
    return func(c fiber.Ctx) error {
        // Get claims from previous middleware
        claims := c.Locals("claims")
        // Check role...
        return c.Next()
    }
}
```

---

## 4. Next.js Frontend Patterns

### Project Structure (Aligned with plan.md)

```
cashlens-web/
├── app/
│   ├── (auth)/
│   │   ├── sign-in/[[...sign-in]]/page.tsx
│   │   └── sign-up/[[...sign-up]]/page.tsx
│   ├── (dashboard)/
│   │   ├── layout.tsx          # Auth check + sidebar
│   │   ├── page.tsx            # Dashboard with KPIs
│   │   ├── upload/page.tsx     # CSV upload
│   │   └── review/page.tsx     # Smart review inbox
│   ├── api/
│   │   └── webhooks/
│   │       └── clerk/route.ts  # User sync webhook
│   ├── layout.tsx              # ClerkProvider
│   └── globals.css
├── components/
│   ├── ui/                     # shadcn components
│   ├── charts/
│   │   ├── NetCashFlowChart.tsx
│   │   └── ExpensesChart.tsx
│   └── upload/
│       └── FileDropzone.tsx
├── lib/
│   └── api.ts                  # API client wrapper
├── stores/
│   └── useAuthStore.ts         # Zustand (optional with Clerk)
├── types/
│   └── index.ts                # TypeScript definitions
└── middleware.ts               # Clerk middleware
```

### API Client Pattern

```typescript
// lib/api.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1"

export class APIError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: any
  ) {
    super(message)
    this.name = "APIError"
  }
}

export async function apiRequest<T = any>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const token = localStorage.getItem("clerk-token") // Or use Clerk's useAuth hook

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({
      code: "UNKNOWN_ERROR",
      message: response.statusText,
    }))

    throw new APIError(
      response.status,
      error.code,
      error.message,
      error.details
    )
  }

  return response.json()
}

// Typed API functions
export const api = {
  // Upload
  getPresignedURL: (filename: string, contentType: string) =>
    apiRequest<{ upload_url: string; file_key: string; expires_in: number }>(
      `/upload/presign?filename=${filename}&content_type=${contentType}`
    ),

  processUpload: (fileKey: string) =>
    apiRequest<{
      total_rows: number
      categorized: number
      accuracy_percent: number
    }>("/upload/process", {
      method: "POST",
      body: JSON.stringify({ file_key: fileKey }),
    }),

  // Transactions
  getTransactions: (status?: "all" | "categorized" | "uncategorized") =>
    apiRequest<{ transactions: Transaction[] }>(
      `/transactions${status ? `?status=${status}` : ""}`
    ),

  updateTransaction: (id: string, category: string) =>
    apiRequest(`/transactions/${id}`, {
      method: "PUT",
      body: JSON.stringify({ category }),
    }),

  // Summary
  getSummary: (
    from: string,
    to: string,
    groupBy: "day" | "week" | "month" = "month"
  ) =>
    apiRequest<SummaryResponse>(
      `/summary?from=${from}&to=${to}&group_by=${groupBy}`
    ),
}

// Types
export interface Transaction {
  id: string
  user_id: string
  txn_date: string
  description: string
  amount: number
  txn_type: "credit" | "debit"
  category?: string
  is_reviewed: boolean
  created_at: string
}

export interface SummaryResponse {
  kpis: {
    total_inflow: number
    total_outflow: number
    net_cash_flow: number
    transaction_count: number
  }
  net_flow_trend: Array<{
    period: string
    net_flow: number
  }>
  top_expenses?: Array<{
    category: string
    total_amount: number
    txn_count: number
    percent: number
  }>
}
```

### File Upload Component (Day 3)

```typescript
// app/(dashboard)/upload/page.tsx
"use client"

import { useState } from "react"
import { useDropzone } from "react-dropzone"
import { api } from "@/lib/api"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"

export default function UploadPage() {
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState(0)
  const [result, setResult] = useState<{
    total_rows: number
    categorized: number
    accuracy_percent: number
  } | null>(null)
  const [error, setError] = useState<string | null>(null)

  const onDrop = async (acceptedFiles: File[]) => {
    const file = acceptedFiles[0]
    if (!file) return

    setUploading(true)
    setProgress(0)
    setError(null)

    try {
      // Step 1: Get presigned URL (10%)
      setProgress(10)
      const { upload_url, file_key } = await api.getPresignedURL(
        file.name,
        file.type
      )

      // Step 2: Upload to S3 (30%)
      setProgress(30)
      await fetch(upload_url, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": file.type },
      })

      // Step 3: Process CSV (60%)
      setProgress(60)
      const processResult = await api.processUpload(file_key)

      // Step 4: Done (100%)
      setProgress(100)
      setResult(processResult)
    } catch (err: any) {
      setError(err.message || "Upload failed")
    } finally {
      setUploading(false)
    }
  }

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      "text/csv": [".csv"],
      "application/vnd.ms-excel": [".csv"],
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
        ".xlsx",
      ],
    },
    maxSize: 5 * 1024 * 1024, // 5MB
    multiple: false,
  })

  return (
    <div className="max-w-4xl mx-auto p-8">
      <h1 className="text-3xl font-bold mb-8">Upload Bank Statement</h1>

      <Card className="mb-8">
        <CardContent className="pt-6">
          <div
            {...getRootProps()}
            className={`
              border-2 border-dashed rounded-lg p-12 text-center cursor-pointer
              transition-colors
              ${
                isDragActive
                  ? "border-blue-500 bg-blue-50"
                  : "border-gray-300 hover:border-gray-400"
              }
            `}
          >
            <input {...getInputProps()} />
            <div className="flex flex-col items-center gap-2">
              <svg
                className="w-12 h-12 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                />
              </svg>
              {isDragActive ? (
                <p className="text-lg">Drop the file here...</p>
              ) : (
                <>
                  <p className="text-lg">
                    Drag & drop your bank statement, or click to select
                  </p>
                  <p className="text-sm text-gray-500">
                    Supports CSV and XLSX (max 5MB)
                  </p>
                </>
              )}
            </div>
          </div>

          {uploading && (
            <div className="mt-6">
              <Progress value={progress} className="mb-2" />
              <p className="text-sm text-center text-gray-600">
                {progress < 30
                  ? "Uploading..."
                  : progress < 60
                  ? "Processing..."
                  : "Almost done..."}
              </p>
            </div>
          )}

          {error && (
            <div className="mt-6 p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-red-800">{error}</p>
            </div>
          )}

          {result && (
            <div className="mt-6 grid grid-cols-3 gap-4">
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm text-gray-600">
                    Total Transactions
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold">{result.total_rows}</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm text-gray-600">
                    Auto-Categorized
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold text-green-600">
                    {result.categorized}
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm text-gray-600">
                    Accuracy
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-bold text-blue-600">
                    {result.accuracy_percent.toFixed(1)}%
                  </p>
                </CardContent>
              </Card>
            </div>
          )}
        </CardContent>
      </Card>

      <div className="text-sm text-gray-600">
        <h3 className="font-semibold mb-2">Supported Banks:</h3>
        <ul className="list-disc list-inside space-y-1">
          <li>HDFC Bank</li>
          <li>ICICI Bank</li>
          <li>State Bank of India (SBI)</li>
          <li>Axis Bank</li>
          <li>Kotak Mahindra Bank</li>
        </ul>
      </div>
    </div>
  )
}
```

---

## 5. Database and Migration Patterns

### Migration Structure

```
internal/database/migrations/
├── 001_initial.sql
├── 002_transactions.sql
├── 003_global_rules.sql
└── 004_categorization_rules.sql
```

### Example Migration (Day 1)

```sql
-- internal/database/migrations/001_initial.sql
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### Transactions Table Migration (Day 2)

```sql
-- internal/database/migrations/002_transactions.sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    txn_date DATE NOT NULL,
    description TEXT NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    txn_type VARCHAR(10) NOT NULL CHECK (txn_type IN ('credit', 'debit')),
    category VARCHAR(100),
    is_reviewed BOOLEAN DEFAULT FALSE,
    raw_data JSONB, -- Store original CSV row for debugging
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(txn_date DESC);
CREATE INDEX idx_transactions_category ON transactions(category) WHERE category IS NOT NULL;
CREATE INDEX idx_transactions_reviewed ON transactions(is_reviewed) WHERE is_reviewed = FALSE;

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Upload tracking table
CREATE TABLE upload_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    file_key TEXT NOT NULL,
    total_rows INTEGER NOT NULL,
    categorized_rows INTEGER NOT NULL,
    accuracy_percent DECIMAL(5, 2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'processing',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_upload_history_user_id ON upload_history(user_id);
```

### Global Rules Migration (Day 4)

```sql
-- internal/database/migrations/003_global_rules.sql
CREATE TABLE global_categorization_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    keyword TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_global_rules_keyword ON global_categorization_rules(keyword);
CREATE INDEX idx_global_rules_category ON global_categorization_rules(category);

-- Seed global rules (from techspec §5.2)
INSERT INTO global_categorization_rules (keyword, category, priority) VALUES
-- Cloud & Hosting
('aws', 'Cloud & Hosting', 100),
('amazon web services', 'Cloud & Hosting', 100),
('digitalocean', 'Cloud & Hosting', 90),
('heroku', 'Cloud & Hosting', 90),
('vercel', 'Cloud & Hosting', 90),
('netlify', 'Cloud & Hosting', 90),
('gcp', 'Cloud & Hosting', 100),
('google cloud', 'Cloud & Hosting', 100),
('azure', 'Cloud & Hosting', 100),
('cloudflare', 'Cloud & Hosting', 80),

-- Payment Processing
('razorpay', 'Payment Processing', 100),
('stripe', 'Payment Processing', 100),
('paytm', 'Payment Processing', 90),
('phonepe', 'Payment Processing', 90),
('instamojo', 'Payment Processing', 80),
('cashfree', 'Payment Processing', 80),

-- Marketing & Advertising
('google ads', 'Marketing', 100),
('facebook ads', 'Marketing', 100),
('linkedin ads', 'Marketing', 90),
('instagram', 'Marketing', 80),
('twitter ads', 'Marketing', 80),

-- Salaries
('salary', 'Salaries', 100),
('payroll', 'Salaries', 100),
('emp salary', 'Salaries', 100),
('employee payment', 'Salaries', 90),

-- Office Supplies
('amazon', 'Office Supplies', 50), -- Lower priority, common word
('flipkart', 'Office Supplies', 60),
('office depot', 'Office Supplies', 80),
('staples', 'Office Supplies', 80),

-- Software & SaaS
('github', 'Software & SaaS', 90),
('slack', 'Software & SaaS', 90),
('zoom', 'Software & SaaS', 90),
('notion', 'Software & SaaS', 80),
('figma', 'Software & SaaS', 80),
('jira', 'Software & SaaS', 80),

-- Travel
('flight', 'Travel', 80),
('hotel', 'Travel', 80),
('makemytrip', 'Travel', 90),
('goibibo', 'Travel', 90),
('uber', 'Travel', 70),
('ola', 'Travel', 70),

-- Legal & Professional Services
('lawyer', 'Legal & Professional Services', 80),
('legal', 'Legal & Professional Services', 70),
('consultant', 'Legal & Professional Services', 70),
('accounting', 'Legal & Professional Services', 80),

-- Utilities
('electricity', 'Utilities', 80),
('water bill', 'Utilities', 80),
('internet', 'Utilities', 80),
('broadband', 'Utilities', 80),

-- Team Meals
('zomato', 'Team Meals', 80),
('swiggy', 'Team Meals', 80),
('restaurant', 'Team Meals', 60),
('food', 'Team Meals', 40);
```

### User-Specific Rules Migration (Day 4)

```sql
-- internal/database/migrations/004_categorization_rules.sql
CREATE TABLE user_categorization_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    keyword TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, keyword)
);

CREATE INDEX idx_user_rules_user_id ON user_categorization_rules(user_id);
CREATE INDEX idx_user_rules_keyword ON user_categorization_rules(keyword);

CREATE TRIGGER update_user_rules_updated_at BEFORE UPDATE ON user_categorization_rules
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

---

## 6. CSV Processing and Categorization

### CSV Parser (Day 2 - techspec §7.1)

Test first approach:

```go
// internal/services/parser_test.go
package services

import (
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestDetectSchema_HDFC(t *testing.T) {
    file, err := os.Open("testdata/hdfc_sample.csv")
    require.NoError(t, err)
    defer file.Close()

    schema, err := DetectSchema(file)
    require.NoError(t, err)

    assert.Equal(t, BankHDFC, schema.BankType)
    assert.Equal(t, "Date", schema.DateColumn)
    assert.Equal(t, "Narration", schema.DescriptionColumn)
    assert.Equal(t, "Withdrawal Amt", schema.DebitColumn)
    assert.Equal(t, "Deposit Amt", schema.CreditColumn)
}

func TestDetectSchema_ICICI(t *testing.T) {
    file, err := os.Open("testdata/icici_sample.csv")
    require.NoError(t, err)
    defer file.Close()

    schema, err := DetectSchema(file)
    require.NoError(t, err)

    assert.Equal(t, BankICICI, schema.BankType)
}

func TestParseCSV_HDFC(t *testing.T) {
    file, err := os.Open("testdata/hdfc_sample.csv")
    require.NoError(t, err)
    defer file.Close()

    parser := NewParser()
    transactions, err := parser.ParseCSV(file)

    require.NoError(t, err)
    assert.Greater(t, len(transactions), 0)

    // Verify first transaction
    txn := transactions[0]
    assert.NotEmpty(t, txn.Description)
    assert.NotZero(t, txn.Amount)
    assert.Contains(t, []string{"credit", "debit"}, txn.TxnType)
    assert.False(t, txn.TxnDate.IsZero())
}

func TestParseDate_MultipleFormats(t *testing.T) {
    testCases := []struct {
        input    string
        expected time.Time
    }{
        {"15/01/2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
        {"01-15-2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
        {"2024-01-15", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
        {"15 Jan 2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
    }

    for _, tc := range testCases {
        t.Run(tc.input, func(t *testing.T) {
            result, err := ParseDate(tc.input)
            require.NoError(t, err)
            assert.Equal(t, tc.expected.Format("2006-01-02"), result.Format("2006-01-02"))
        })
    }
}

func TestParseAmount(t *testing.T) {
    testCases := []struct {
        input    string
        expected float64
    }{
        {"1,234.56", 1234.56},
        {"1234.56", 1234.56},
        {"₹1,234.56", 1234.56},
        {"Rs 1234", 1234.00},
        {"-500", -500.00},
    }

    for _, tc := range testCases {
        t.Run(tc.input, func(t *testing.T) {
            result, err := ParseAmount(tc.input)
            require.NoError(t, err)
            assert.InDelta(t, tc.expected, result, 0.01)
        })
    }
}
```

Implementation:

```go
// internal/services/parser.go
package services

import (
    "encoding/csv"
    "fmt"
    "io"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type BankType string

const (
    BankHDFC  BankType = "HDFC"
    BankICICI BankType = "ICICI"
    BankSBI   BankType = "SBI"
    BankAxis  BankType = "AXIS"
    BankKotak BankType = "KOTAK"
)

type CSVSchema struct {
    BankType          BankType
    DateColumn        string
    DescriptionColumn string
    DebitColumn       string
    CreditColumn      string
    BalanceColumn     string
    HeaderRow         int
}

type ParsedTransaction struct {
    TxnDate     time.Time
    Description string
    Amount      float64
    TxnType     string // "credit" or "debit"
    RawData     map[string]string
}

type Parser struct {
    dateFormats []string
}

func NewParser() *Parser {
    return &Parser{
        dateFormats: []string{
            "02/01/2006",      // DD/MM/YYYY (India)
            "02-01-2006",      // DD-MM-YYYY
            "2006-01-02",      // YYYY-MM-DD (ISO)
            "01/02/2006",      // MM/DD/YYYY (US)
            "02 Jan 2006",     // DD Mon YYYY
            "02 January 2006", // DD Month YYYY
        },
    }
}

func DetectSchema(r io.Reader) (*CSVSchema, error) {
    reader := csv.NewReader(r)
    reader.TrimLeadingSpace = true

    // Read first few rows to detect schema
    rows := make([][]string, 0, 10)
    for i := 0; i < 10; i++ {
        row, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("failed to read CSV: %w", err)
        }
        rows = append(rows, row)
    }

    if len(rows) == 0 {
        return nil, fmt.Errorf("empty CSV file")
    }

    // Detect bank type and schema by checking headers
    for i, row := range rows {
        headerStr := strings.Join(row, ",")
        headerLower := strings.ToLower(headerStr)

        // HDFC Detection
        if strings.Contains(headerLower, "narration") && strings.Contains(headerLower, "withdrawal") {
            return &CSVSchema{
                BankType:          BankHDFC,
                DateColumn:        "Date",
                DescriptionColumn: "Narration",
                DebitColumn:       "Withdrawal Amt",
                CreditColumn:      "Deposit Amt",
                BalanceColumn:     "Closing Balance",
                HeaderRow:         i,
            }, nil
        }

        // ICICI Detection
        if strings.Contains(headerLower, "transaction date") && strings.Contains(headerLower, "cheque number") {
            return &CSVSchema{
                BankType:          BankICICI,
                DateColumn:        "Transaction Date",
                DescriptionColumn: "Transaction Remarks",
                DebitColumn:       "Withdrawal Amount",
                CreditColumn:      "Deposit Amount",
                BalanceColumn:     "Balance",
                HeaderRow:         i,
            }, nil
        }

        // SBI Detection
        if strings.Contains(headerLower, "txn date") && strings.Contains(headerLower, "description") {
            return &CSVSchema{
                BankType:          BankSBI,
                DateColumn:        "Txn Date",
                DescriptionColumn: "Description",
                DebitColumn:       "Debit",
                CreditColumn:      "Credit",
                BalanceColumn:     "Balance",
                HeaderRow:         i,
            }, nil
        }

        // Axis Detection
        if strings.Contains(headerLower, "transaction date") && strings.Contains(headerLower, "particulars") {
            return &CSVSchema{
                BankType:          BankAxis,
                DateColumn:        "Transaction Date",
                DescriptionColumn: "Particulars",
                DebitColumn:       "Dr/Cr",
                CreditColumn:      "Dr/Cr",
                BalanceColumn:     "Balance",
                HeaderRow:         i,
            }, nil
        }

        // Kotak Detection
        if strings.Contains(headerLower, "date") && strings.Contains(headerLower, "description") && strings.Contains(headerLower, "amount") {
            return &CSVSchema{
                BankType:          BankKotak,
                DateColumn:        "Date",
                DescriptionColumn: "Description",
                DebitColumn:       "Debit",
                CreditColumn:      "Credit",
                BalanceColumn:     "Balance",
                HeaderRow:         i,
            }, nil
        }
    }

    return nil, fmt.Errorf("unable to detect bank type from CSV headers")
}

func (p *Parser) ParseCSV(r io.Reader) ([]ParsedTransaction, error) {
    // Detect schema first
    schema, err := DetectSchema(r)
    if err != nil {
        return nil, err
    }

    // Reset reader (in production, you'd need to re-open the file)
    // For now, assuming we can seek back

    reader := csv.NewReader(r)
    reader.TrimLeadingSpace = true

    var transactions []ParsedTransaction
    var headers []string
    lineNum := 0

    for {
        row, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("error reading row %d: %w", lineNum, err)
        }

        lineNum++

        // Skip until header row
        if lineNum <= schema.HeaderRow {
            continue
        }

        // First row after skip is headers
        if len(headers) == 0 {
            headers = row
            continue
        }

        // Parse transaction
        txn, err := p.parseRow(row, headers, schema)
        if err != nil {
            // Log and skip invalid rows
            continue
        }

        if txn != nil {
            transactions = append(transactions, *txn)
        }
    }

    return transactions, nil
}

func (p *Parser) parseRow(row, headers []string, schema *CSVSchema) (*ParsedTransaction, error) {
    if len(row) != len(headers) {
        return nil, fmt.Errorf("row length mismatch")
    }

    // Create map of column name to value
    data := make(map[string]string)
    for i, header := range headers {
        if i < len(row) {
            data[header] = strings.TrimSpace(row[i])
        }
    }

    // Skip empty rows
    if data[schema.DescriptionColumn] == "" {
        return nil, nil
    }

    // Parse date
    dateStr := data[schema.DateColumn]
    txnDate, err := p.ParseDate(dateStr)
    if err != nil {
        return nil, fmt.Errorf("invalid date %s: %w", dateStr, err)
    }

    // Parse amounts
    debitStr := data[schema.DebitColumn]
    creditStr := data[schema.CreditColumn]

    var amount float64
    var txnType string

    if debitStr != "" && debitStr != "0" && debitStr != "0.00" {
        amount, err = ParseAmount(debitStr)
        if err != nil {
            return nil, fmt.Errorf("invalid debit amount %s: %w", debitStr, err)
        }
        txnType = "debit"
    } else if creditStr != "" && creditStr != "0" && creditStr != "0.00" {
        amount, err = ParseAmount(creditStr)
        if err != nil {
            return nil, fmt.Errorf("invalid credit amount %s: %w", creditStr, err)
        }
        txnType = "credit"
    } else {
        // Skip zero-amount transactions
        return nil, nil
    }

    return &ParsedTransaction{
        TxnDate:     txnDate,
        Description: data[schema.DescriptionColumn],
        Amount:      amount,
        TxnType:     txnType,
        RawData:     data,
    }, nil
}

func (p *Parser) ParseDate(dateStr string) (time.Time, error) {
    dateStr = strings.TrimSpace(dateStr)

    for _, format := range p.dateFormats {
        if t, err := time.Parse(format, dateStr); err == nil {
            return t, nil
        }
    }

    return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func ParseAmount(amountStr string) (float64, error) {
    // Remove currency symbols and whitespace
    re := regexp.MustCompile(`[₹Rs,\s]`)
    cleaned := re.ReplaceAllString(amountStr, "")

    // Parse as float
    amount, err := strconv.ParseFloat(cleaned, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid amount: %s", amountStr)
    }

    return amount, nil
}
```

### Categorizer with Rule Engine (Day 4 - techspec §7.2)

```go
// internal/services/categorizer.go
package services

import (
    "context"
    "strings"
    "sync"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Categorizer struct {
    db          *pgxpool.Pool
    globalRules map[string]string
    userRules   map[string]map[string]string // userID -> (keyword -> category)
    mu          sync.RWMutex
}

func NewCategorizer(db *pgxpool.Pool) *Categorizer {
    c := &Categorizer{
        db:        db,
        userRules: make(map[string]map[string]string),
    }

    // Load global rules from database
    c.loadGlobalRules()

    return c
}

func (c *Categorizer) loadGlobalRules() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    rows, err := c.db.Query(context.Background(), `
        SELECT keyword, category
        FROM global_categorization_rules
        WHERE is_active = TRUE
        ORDER BY priority DESC
    `)
    if err != nil {
        return err
    }
    defer rows.Close()

    c.globalRules = make(map[string]string)

    for rows.Next() {
        var keyword, category string
        if err := rows.Scan(&keyword, &category); err != nil {
            return err
        }
        c.globalRules[strings.ToLower(keyword)] = category
    }

    return nil
}

func (c *Categorizer) LoadUserRules(userID string) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    rows, err := c.db.Query(context.Background(), `
        SELECT keyword, category
        FROM user_categorization_rules
        WHERE user_id = $1 AND is_active = TRUE
        ORDER BY priority DESC
    `, userID)
    if err != nil {
        return err
    }
    defer rows.Close()

    rules := make(map[string]string)

    for rows.Next() {
        var keyword, category string
        if err := rows.Scan(&keyword, &category); err != nil {
            return err
        }
        rules[strings.ToLower(keyword)] = category
    }

    c.userRules[userID] = rules

    return nil
}

func (c *Categorizer) Categorize(description string, userID string) string {
    desc := strings.ToLower(description)

    c.mu.RLock()
    defer c.mu.RUnlock()

    // Check user-specific rules first (higher priority)
    if userRuleMap, exists := c.userRules[userID]; exists {
        for keyword, category := range userRuleMap {
            if strings.Contains(desc, keyword) {
                return category
            }
        }
    }

    // Check global rules
    for keyword, category := range c.globalRules {
        if strings.Contains(desc, keyword) {
            return category
        }
    }

    return "" // Uncategorized
}

func (c *Categorizer) InvalidateUserCache(userID string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    delete(c.userRules, userID)
}
```

---

## 7. Integration Testing Strategy

### E2E Test with Playwright (Day 1, 3, 5)

```typescript
// tests/e2e/upload-flow.spec.ts
import { test, expect } from "@playwright/test"
import path from "path"

test.describe("CSV Upload Flow", () => {
  test.beforeEach(async ({ page }) => {
    // Login via Clerk
    await page.goto("/sign-in")
    await page.fill('input[name="identifier"]', "test@example.com")
    await page.fill('input[name="password"]', "TestPassword123!")
    await page.click('button[type="submit"]')
    await page.waitForURL("/dashboard")
  })

  test("complete upload and categorization flow", async ({ page }) => {
    // Navigate to upload page
    await page.goto("/upload")

    // Upload CSV file
    const filePath = path.join(__dirname, "../fixtures/hdfc_sample.csv")
    await page.setInputFiles('input[type="file"]', filePath)

    // Wait for upload to complete
    await expect(page.locator("text=Total Transactions")).toBeVisible({
      timeout: 30000,
    })

    // Verify results
    const totalText = await page
      .locator('[data-testid="total-transactions"]')
      .textContent()
    const total = parseInt(totalText || "0")
    expect(total).toBeGreaterThan(0)

    const accuracyText = await page
      .locator('[data-testid="accuracy"]')
      .textContent()
    const accuracy = parseFloat(accuracyText?.replace("%", "") || "0")
    expect(accuracy).toBeGreaterThanOrEqual(85)
  })

  test("review uncategorized transactions", async ({ page }) => {
    // Go to review page
    await page.goto("/review")

    // Wait for transactions to load
    await page.waitForSelector("table tbody tr")

    // Get first uncategorized transaction
    const firstRow = page.locator("table tbody tr").first()

    // Select category
    await firstRow.locator("select").selectOption("Cloud & Hosting")

    // Wait for update
    await page.waitForTimeout(1000)

    // Verify transaction was removed from list
    const remainingRows = await page.locator("table tbody tr").count()
    // Should be one less than before
  })
})
```

### API Integration Tests (Go)

```go
// tests/integration/api_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/yourorg/cashlens-api/cmd/api"
)

func setupTestApp(t *testing.T) *fiber.App {
    // Setup test database
    testDB := setupTestDB(t)

    // Setup test S3
    testS3 := setupMockS3(t)

    // Create app with test dependencies
    app := api.CreateApp(testDB, testS3)

    return app
}

func TestUploadFlow_EndToEnd(t *testing.T) {
    app := setupTestApp(t)
    token := getTestAuthToken(t)

    // Step 1: Get presigned URL
    req1 := httptest.NewRequest("GET", "/v1/upload/presign?filename=test.csv&content_type=text/csv", nil)
    req1.Header.Set("Authorization", "Bearer "+token)

    resp1, err := app.Test(req1)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp1.StatusCode)

    var presignResp struct {
        UploadURL string `json:"upload_url"`
        FileKey   string `json:"file_key"`
    }
    json.NewDecoder(resp1.Body).Decode(&presignResp)

    assert.NotEmpty(t, presignResp.UploadURL)
    assert.NotEmpty(t, presignResp.FileKey)

    // Step 2: Simulate S3 upload (in real test, use mocked S3)
    // ...

    // Step 3: Process uploaded file
    processReq := map[string]string{
        "file_key": presignResp.FileKey,
    }
    body, _ := json.Marshal(processReq)

    req2 := httptest.NewRequest("POST", "/v1/upload/process", bytes.NewReader(body))
    req2.Header.Set("Authorization", "Bearer "+token)
    req2.Header.Set("Content-Type", "application/json")

    resp2, err := app.Test(req2, 60000) // 60 second timeout
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp2.StatusCode)

    var processResp struct {
        TotalRows        int     `json:"total_rows"`
        Categorized      int     `json:"categorized"`
        AccuracyPercent  float64 `json:"accuracy_percent"`
    }
    json.NewDecoder(resp2.Body).Decode(&processResp)

    assert.Greater(t, processResp.TotalRows, 0)
    assert.GreaterOrEqual(t, processResp.AccuracyPercent, 85.0)
}
```

---

## 8. Performance and Accuracy Validation

### Accuracy Benchmark Test (Day 9)

```go
// internal/services/categorizer_benchmark_test.go
package services

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestCategorizerAccuracy_Production(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping accuracy benchmark in short mode")
    }

    // Setup
    db := setupTestDB(t)
    defer db.Close()

    categorizer := NewCategorizer(db)

    // Test files with known ground truth
    testCases := []struct {
        filename         string
        expectedAccuracy float64
    }{
        {"testdata/hdfc_100rows.csv", 85.0},
        {"testdata/icici_120rows.csv", 85.0},
        {"testdata/sbi_90rows.csv", 85.0},
        {"testdata/axis_110rows.csv", 85.0},
        {"testdata/kotak_80rows.csv", 85.0},
    }

    overallTotal := 0
    overallCategorized := 0

    for _, tc := range testCases {
        t.Run(tc.filename, func(t *testing.T) {
            transactions := loadTestFile(t, tc.filename)

            categorized := 0
            for _, txn := range transactions {
                category := categorizer.Categorize(txn.Description, "test-user")
                if category != "" {
                    categorized++
                }
            }

            accuracy := float64(categorized) / float64(len(transactions)) * 100

            t.Logf("%s: %.2f%% (%d/%d)", tc.filename, accuracy, categorized, len(transactions))

            assert.GreaterOrEqual(t, accuracy, tc.expectedAccuracy,
                "Accuracy below threshold for %s", tc.filename)
             EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table (simplified for Clerk)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clerk_user_id TEXT UNIQUE NOT NULL,
    email TEXT NOT NULL,
    full_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_clerk_id ON users(clerk_user_id);
CREATE INDEX idx_users_email ON users(email);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE
```
