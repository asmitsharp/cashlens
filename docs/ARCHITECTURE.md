# Cashlens Architecture Documentation

This document provides a comprehensive overview of the Cashlens system architecture, code structure, design patterns, and implementation details.

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Backend Architecture](#backend-architecture)
4. [Frontend Architecture](#frontend-architecture)
5. [Data Flow](#data-flow)
6. [Database Design](#database-design)
7. [Design Patterns](#design-patterns)
8. [Security Architecture](#security-architecture)
9. [Scalability Considerations](#scalability-considerations)

---

## System Overview

Cashlens is a financial analytics SaaS platform designed for Indian SMBs. The system follows a modern microservices-inspired architecture with clear separation between frontend, backend, and data layers.

### Tech Stack Summary

**Backend:**
- **Language:** Go 1.23+
- **Web Framework:** Fiber v3 (Express-like API)
- **Database:** PostgreSQL 16
- **Cache:** Redis (future)
- **Storage:** AWS S3 (LocalStack for local dev)
- **PDF Processing:** Python microservice (Flask)

**Frontend:**
- **Framework:** Next.js 15 (App Router)
- **UI Library:** React 19
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **Components:** shadcn/ui (Radix UI primitives)
- **Authentication:** Clerk

**Infrastructure:**
- **Orchestration:** Docker Compose
- **CI/CD:** GitHub Actions
- **Deployment:** TBD (AWS/Vercel)

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
├─────────────────────────────────────────────────────────────────┤
│  Web Browser (Chrome, Firefox, Safari)                          │
│  ├─ Next.js App (React 19 + TypeScript)                         │
│  ├─ Clerk Authentication UI                                     │
│  └─ Tailwind CSS + shadcn/ui Components                         │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTPS
                           │ JWT Auth
┌──────────────────────────┴──────────────────────────────────────┐
│                      API GATEWAY LAYER                           │
├─────────────────────────────────────────────────────────────────┤
│  Go Fiber API (Port 8080)                                       │
│  ├─ CORS Middleware                                             │
│  ├─ Clerk JWT Validation Middleware                             │
│  ├─ Request Logging                                             │
│  └─ Error Handling                                              │
└──────────────────────────┬──────────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
┌───────┴─────┐   ┌────────┴───────┐   ┌─────┴────────┐
│  HANDLERS   │   │   SERVICES     │   │  MIDDLEWARE  │
├─────────────┤   ├────────────────┤   ├──────────────┤
│ • Upload    │   │ • Parser       │   │ • Auth       │
│ • Txns      │   │ • XLSX Parser  │   │ • CORS       │
│ • Summary   │   │ • Categorizer  │   │ • Logging    │
│ • Rules     │   │ • Storage (S3) │   │ • Errors     │
└─────────────┘   │ • Validation   │   └──────────────┘
                  └────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
┌───────┴─────┐   ┌────────┴───────┐   ┌─────┴────────┐
│ PostgreSQL  │   │  AWS S3        │   │ Python PDF   │
│   Database  │   │  (LocalStack)  │   │ Microservice │
├─────────────┤   ├────────────────┤   ├──────────────┤
│ • Users     │   │ • CSV files    │   │ • Tabula-py  │
│ • Txns      │   │ • XLSX files   │   │ • PDF parse  │
│ • Rules     │   │ • PDF files    │   │ • Flask API  │
│ • Upload    │   │ • Backups      │   │ Port: 5001   │
└─────────────┘   └────────────────┘   └──────────────┘
```

---

## Backend Architecture

### Project Structure

```
cashlens-api/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
│
├── internal/                       # Private application code
│   ├── config/
│   │   └── config.go              # Environment variable loader
│   │
│   ├── database/
│   │   ├── migrations/            # SQL migration files (future)
│   │   ├── queries/               # SQL queries (sqlc, future)
│   │   └── db.go                  # Database connection
│   │
│   ├── handlers/                  # HTTP request handlers
│   │   ├── auth.go               # Clerk JWT validation
│   │   ├── upload.go             # Presigned URLs + file processing
│   │   ├── transactions.go       # CRUD operations
│   │   ├── summary.go            # Dashboard KPIs
│   │   └── rules.go              # Categorization rules
│   │
│   ├── middleware/                # HTTP middleware
│   │   ├── auth.go               # JWT authentication
│   │   ├── cors.go               # CORS headers
│   │   └── logger.go             # Request logging
│   │
│   ├── models/                    # Domain models
│   │   ├── user.go
│   │   ├── transaction.go
│   │   ├── rule.go
│   │   └── upload.go
│   │
│   ├── services/                  # Business logic
│   │   ├── parser.go             # CSV parser (5 banks)
│   │   ├── xlsx_parser.go        # XLSX parser (excelize)
│   │   ├── categorizer.go        # Rule-based categorization
│   │   ├── storage.go            # S3 operations
│   │   ├── validation.go         # File validation (magic bytes)
│   │   └── pdf_client.go         # Python PDF service client
│   │
│   └── utils/                     # Helper functions
│       ├── response.go           # JSON response helpers
│       ├── errors.go             # Error types
│       └── logger.go             # Structured logging
│
├── testdata/                      # Test CSV/XLSX files
│   ├── hdfc_sample.csv
│   ├── hdfc_sample.xlsx
│   ├── icici_sample.csv
│   └── ...
│
├── go.mod                         # Go dependencies
├── go.sum                         # Dependency checksums
└── Dockerfile                     # Container image
```

---

### Handler Layer

**Purpose:** HTTP request/response handling, input validation, delegation to services

**File:** `internal/handlers/upload.go`

```go
// GeneratePresignedURL handles POST /v1/upload/presigned-url
func GeneratePresignedURL(c *fiber.Ctx) error {
    // 1. Extract user ID from JWT (set by auth middleware)
    userID := c.Locals("user_id").(string)

    // 2. Parse request body
    var req PresignedURLRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(ErrorResponse{
            Error: "Invalid request body",
        })
    }

    // 3. Validate file type
    if !isValidFileType(req.FileType) {
        return c.Status(400).JSON(ErrorResponse{
            Error: "Invalid file type. Supported: CSV, XLSX, PDF",
        })
    }

    // 4. Delegate to storage service
    storage := services.NewStorageService()
    url, fileKey, err := storage.GeneratePresignedURL(
        userID, req.Filename, req.FileType,
    )
    if err != nil {
        return c.Status(500).JSON(ErrorResponse{
            Error: "Failed to generate presigned URL",
        })
    }

    // 5. Return response
    return c.JSON(PresignedURLResponse{
        UploadURL: url,
        FileKey:   fileKey,
        ExpiresIn: 300, // 5 minutes
    })
}
```

**Key Principles:**
- **Thin handlers:** Minimal logic, delegate to services
- **Input validation:** Always validate before passing to services
- **Error handling:** Use structured error responses
- **User context:** Extract user ID from JWT via middleware

---

### Service Layer

**Purpose:** Business logic, data processing, external integrations

#### 1. Parser Service (`internal/services/parser.go`)

**Responsibility:** Parse CSV files from 5 Indian banks into normalized transactions

**Key Functions:**

```go
type Parser struct {
    // No state - stateless service
}

// ParseCSV parses a CSV file and returns normalized transactions
func (p *Parser) ParseCSV(file io.Reader) ([]ParsedTransaction, error) {
    // 1. Read CSV into memory
    reader := csv.NewReader(file)
    rows, err := reader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("failed to read CSV: %w", err)
    }

    // 2. Detect bank schema from headers
    schema := p.DetectSchema(rows[0])
    if schema == SchemaUnknown {
        return nil, errors.New("unsupported bank format")
    }

    // 3. Parse each row according to schema
    var transactions []ParsedTransaction
    for i, row := range rows[1:] { // Skip header
        txn, err := p.parseRow(row, schema)
        if err != nil {
            log.Printf("Failed to parse row %d: %v", i+1, err)
            continue // Skip invalid rows
        }
        transactions = append(transactions, txn)
    }

    return transactions, nil
}

// DetectSchema identifies bank format from headers
func (p *Parser) DetectSchema(headers []string) BankSchema {
    headerStr := strings.ToLower(strings.Join(headers, "|"))

    if strings.Contains(headerStr, "narration") &&
       strings.Contains(headerStr, "withdrawal amt") {
        return SchemaHDFC
    }

    if strings.Contains(headerStr, "transaction remarks") &&
       strings.Contains(headerStr, "deposit amount") {
        return SchemaICICI
    }

    // ... other banks

    return SchemaUnknown
}

// parseRow extracts transaction data from a CSV row
func (p *Parser) parseRow(row []string, schema BankSchema) (ParsedTransaction, error) {
    var txn ParsedTransaction

    switch schema {
    case SchemaHDFC:
        // HDFC format: Date, Narration, Withdrawal Amt, Deposit Amt, Balance
        txn.Date = p.ParseDate(row[0])
        txn.Description = strings.TrimSpace(row[1])
        txn.Amount = p.ParseAmount(row[2], row[3]) // Withdrawal or Deposit
        txn.TxnType = p.DetermineTxnType(txn.Amount)

    case SchemaICICI:
        // ICICI format: Transaction Date, Value Date, Transaction Remarks, ...
        txn.Date = p.ParseDate(row[0])
        txn.Description = strings.TrimSpace(row[2])
        txn.Amount = p.ParseAmount(row[3], row[4])
        txn.TxnType = p.DetermineTxnType(txn.Amount)

    // ... other schemas
    }

    return txn, nil
}

// ParseDate handles multiple date formats
func (p *Parser) ParseDate(dateStr string) time.Time {
    formats := []string{
        "02/01/2006",    // DD/MM/YYYY (Indian standard)
        "2006-01-02",    // YYYY-MM-DD (ISO)
        "02-Jan-2006",   // DD-Mon-YYYY
        "Jan 02, 2006",  // Mon DD, YYYY
    }

    for _, format := range formats {
        if t, err := time.Parse(format, dateStr); err == nil {
            return t
        }
    }

    return time.Time{} // Zero time if parsing fails
}

// ParseAmount extracts numeric amount from string
func (p *Parser) ParseAmount(debit, credit string) float64 {
    // Try debit column first
    if amount := p.extractAmount(debit); amount != 0 {
        return -amount // Debits are negative
    }

    // Try credit column
    if amount := p.extractAmount(credit); amount != 0 {
        return amount // Credits are positive
    }

    return 0
}

func (p *Parser) extractAmount(s string) float64 {
    // Remove currency symbols: ₹, Rs, INR
    s = strings.ReplaceAll(s, "₹", "")
    s = strings.ReplaceAll(s, "Rs", "")
    s = strings.ReplaceAll(s, "INR", "")
    s = strings.ReplaceAll(s, ",", "") // Remove thousand separators
    s = strings.TrimSpace(s)

    amount, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0
    }

    return amount
}
```

**Design Decisions:**
- **Stateless:** Parser has no state, can be reused
- **Schema detection:** Automatic bank format detection from headers
- **Error tolerance:** Skip invalid rows instead of failing entire file
- **Multiple date formats:** Supports all common Indian date formats
- **Currency normalization:** Handles ₹, Rs, INR, commas

**Supported Bank Formats:**

| Bank | Header Patterns | Date Format | Amount Columns |
|------|----------------|-------------|----------------|
| HDFC | "Narration", "Withdrawal Amt" | DD/MM/YYYY | Withdrawal Amt, Deposit Amt |
| ICICI | "Transaction Remarks", "Deposit Amount" | DD/MM/YYYY | Withdrawal Amount, Deposit Amount |
| SBI | "Description", "Debit", "Credit" | DD-Mon-YYYY | Debit, Credit |
| Axis | "Particulars", "Dr/Cr" | YYYY-MM-DD | Dr/Cr (combined) |
| Kotak | "Description", "Debit", "Credit" | DD/MM/YYYY | Debit, Credit |

---

#### 2. XLSX Parser Service (`internal/services/xlsx_parser.go`)

**Responsibility:** Parse Excel files using excelize library

**Key Implementation:**

```go
import "github.com/xuri/excelize/v2"

type XLSXParser struct{}

func (p *XLSXParser) ParseXLSX(file io.Reader) ([]ParsedTransaction, error) {
    // 1. Open XLSX file
    xlFile, err := excelize.OpenReader(file)
    if err != nil {
        return nil, fmt.Errorf("failed to open XLSX: %w", err)
    }
    defer xlFile.Close()

    // 2. Get first sheet
    sheets := xlFile.GetSheetList()
    if len(sheets) == 0 {
        return nil, errors.New("no sheets found in XLSX")
    }

    // 3. Read all rows
    rows, err := xlFile.GetRows(sheets[0])
    if err != nil {
        return nil, fmt.Errorf("failed to read rows: %w", err)
    }

    // 4. Use CSV parser for schema detection and parsing
    csvParser := &Parser{}
    schema := csvParser.DetectSchema(rows[0])

    var transactions []ParsedTransaction
    for i, row := range rows[1:] {
        txn, err := csvParser.parseRow(row, schema)
        if err != nil {
            log.Printf("Failed to parse XLSX row %d: %v", i+1, err)
            continue
        }
        transactions = append(transactions, txn)
    }

    return transactions, nil
}
```

**Design Decision:** Reuse CSV parser logic for schema detection and row parsing. XLSX parser just handles Excel-specific file I/O.

---

#### 3. Categorizer Service (`internal/services/categorizer.go`)

**Responsibility:** Auto-categorize transactions using keyword matching rules

**Algorithm:**

```go
type Categorizer struct {
    globalRules []CategorizationRule
    userRules   []CategorizationRule
}

func (c *Categorizer) Categorize(txn ParsedTransaction) string {
    description := strings.ToLower(txn.Description)

    // 1. Check user rules first (higher priority)
    for _, rule := range c.userRules {
        if !rule.IsActive {
            continue
        }
        keyword := strings.ToLower(rule.Keyword)
        if strings.Contains(description, keyword) {
            return rule.Category
        }
    }

    // 2. Check global rules
    for _, rule := range c.globalRules {
        if !rule.IsActive {
            continue
        }
        keyword := strings.ToLower(rule.Keyword)
        if strings.Contains(description, keyword) {
            return rule.Category
        }
    }

    // 3. No match found
    return "Uncategorized"
}

func (c *Categorizer) CategorizeAll(txns []ParsedTransaction) []CategorizedTransaction {
    var categorized []CategorizedTransaction

    for _, txn := range txns {
        category := c.Categorize(txn)
        categorized = append(categorized, CategorizedTransaction{
            Transaction: txn,
            Category:    category,
        })
    }

    return categorized
}
```

**Categorization Rules:**

```go
// Global rules (pre-seeded in database)
var globalRules = []Rule{
    {Keyword: "aws", Category: "Cloud & Hosting", Priority: 100},
    {Keyword: "google cloud", Category: "Cloud & Hosting", Priority: 100},
    {Keyword: "azure", Category: "Cloud & Hosting", Priority: 100},
    {Keyword: "digitalocean", Category: "Cloud & Hosting", Priority: 100},

    {Keyword: "stripe", Category: "Payment Processing", Priority: 90},
    {Keyword: "razorpay", Category: "Payment Processing", Priority: 90},

    {Keyword: "salary", Category: "Salaries", Priority: 100},
    {Keyword: "payroll", Category: "Salaries", Priority: 100},

    {Keyword: "google ads", Category: "Marketing", Priority: 80},
    {Keyword: "facebook ads", Category: "Marketing", Priority: 80},

    {Keyword: "office supplies", Category: "Office Supplies", Priority: 70},
    {Keyword: "stationery", Category: "Office Supplies", Priority: 70},

    // ... more rules
}
```

**Accuracy Measurement:**

```go
func (c *Categorizer) CalculateAccuracy(txns []CategorizedTransaction) float64 {
    if len(txns) == 0 {
        return 0.0
    }

    categorizedCount := 0
    for _, txn := range txns {
        if txn.Category != "Uncategorized" {
            categorizedCount++
        }
    }

    return (float64(categorizedCount) / float64(len(txns))) * 100.0
}
```

**Target:** ≥85% categorization accuracy

---

#### 4. Storage Service (`internal/services/storage.go`)

**Responsibility:** S3 file operations (upload, download, presigned URLs)

**Key Functions:**

```go
import "github.com/aws/aws-sdk-go/service/s3"

type StorageService struct {
    s3Client *s3.S3
    bucket   string
}

func (s *StorageService) GeneratePresignedURL(
    userID, filename, fileType string,
) (string, string, error) {
    // 1. Generate unique file key
    timestamp := time.Now().Unix()
    fileKey := fmt.Sprintf("%s/%d_%s", userID, timestamp, filename)

    // 2. Create presigned PUT request
    req, _ := s.s3Client.PutObjectRequest(&s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(fileKey),
        ContentType: aws.String(fileType),
    })

    // 3. Sign URL (valid for 5 minutes)
    url, err := req.Presign(5 * time.Minute)
    if err != nil {
        return "", "", err
    }

    return url, fileKey, nil
}

func (s *StorageService) DownloadFile(fileKey string) ([]byte, error) {
    // Download file from S3
    result, err := s.s3Client.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(fileKey),
    })
    if err != nil {
        return nil, err
    }
    defer result.Body.Close()

    // Read into memory
    return io.ReadAll(result.Body)
}
```

**LocalStack Configuration (Local Development):**

```go
// Use LocalStack endpoint for local dev
endpoint := os.Getenv("AWS_ENDPOINT") // http://localhost:4566

s3Config := &aws.Config{
    Region:           aws.String("ap-south-1"),
    Endpoint:         aws.String(endpoint),
    S3ForcePathStyle: aws.Bool(true), // Required for LocalStack
    Credentials: credentials.NewStaticCredentials(
        "test", "test", "", // Dummy credentials for LocalStack
    ),
}
```

---

### Middleware Layer

**Purpose:** Cross-cutting concerns (auth, logging, CORS, errors)

#### Authentication Middleware (`internal/middleware/auth.go`)

```go
func ClerkAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Extract JWT from Authorization header
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(401).JSON(fiber.Map{
                "error": "Missing authorization header",
            })
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        // 2. Verify JWT with Clerk
        clerkClient := clerk.NewClient(os.Getenv("CLERK_SECRET_KEY"))
        claims, err := clerkClient.VerifyToken(token)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid token",
            })
        }

        // 3. Store user ID in request context
        c.Locals("user_id", claims.Subject)

        // 4. Continue to next handler
        return c.Next()
    }
}
```

**Usage in routes:**
```go
app.Use("/v1", ClerkAuth()) // Protect all /v1/* routes
```

---

## Frontend Architecture

### Project Structure

```
cashlens-web/
├── app/                            # Next.js App Router
│   ├── (auth)/                    # Auth route group
│   │   ├── sign-in/
│   │   │   └── [[...sign-in]]/
│   │   │       └── page.tsx       # Clerk sign-in page
│   │   └── sign-up/
│   │       └── [[...sign-up]]/
│   │           └── page.tsx       # Clerk sign-up page
│   │
│   ├── (dashboard)/               # Protected dashboard routes
│   │   ├── layout.tsx            # Auth check + sidebar layout
│   │   ├── page.tsx              # Dashboard (KPIs, charts)
│   │   ├── upload/
│   │   │   └── page.tsx          # CSV/XLSX/PDF upload
│   │   ├── review/
│   │   │   └── page.tsx          # Review uncategorized txns
│   │   ├── inbox/
│   │   │   └── page.tsx          # Inbox (future)
│   │   ├── notifications/
│   │   │   └── page.tsx          # Notifications (future)
│   │   └── settings/
│   │       └── page.tsx          # User settings
│   │
│   ├── api/
│   │   └── webhooks/
│   │       └── clerk/
│   │           └── route.ts      # User sync webhook
│   │
│   ├── layout.tsx                # Root layout (ClerkProvider)
│   ├── globals.css               # Tailwind + theme CSS variables
│   └── error.tsx                 # Global error boundary
│
├── components/
│   ├── ui/                       # shadcn/ui components
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── input.tsx
│   │   ├── switch.tsx
│   │   ├── label.tsx
│   │   ├── separator.tsx
│   │   └── ...
│   │
│   ├── layout/                   # Layout components
│   │   ├── Sidebar.tsx          # Collapsible sidebar
│   │   ├── Header.tsx           # Top header bar
│   │   ├── SidebarContext.tsx   # Sidebar state context
│   │   ├── DashboardContent.tsx # Main content wrapper
│   │   └── index.ts             # Barrel exports
│   │
│   ├── charts/                   # Chart components
│   │   ├── ExpenseChart.tsx     # Category breakdown pie chart
│   │   ├── TrendChart.tsx       # Monthly trend line chart
│   │   └── index.ts
│   │
│   ├── upload/                   # Upload flow components
│   │   ├── DropzoneArea.tsx     # Drag-and-drop file input
│   │   ├── UploadProgress.tsx   # Progress bar
│   │   ├── UploadSummary.tsx    # Results summary
│   │   └── index.ts
│   │
│   └── transactions/             # Transaction components
│       ├── TransactionList.tsx  # Table with filters
│       ├── TransactionRow.tsx   # Individual row
│       ├── CategoryBadge.tsx    # Category pill
│       └── index.ts
│
├── lib/
│   ├── api.ts                    # API client wrapper
│   ├── utils.ts                  # cn() utility
│   └── constants.ts              # App constants
│
├── types/
│   ├── index.ts                  # TypeScript types
│   └── api.ts                    # API response types
│
├── hooks/                        # Custom React hooks
│   ├── useTransactions.ts       # Fetch transactions
│   ├── useSummary.ts            # Fetch dashboard data
│   └── useUpload.ts             # Upload flow logic
│
├── public/                       # Static assets
│   ├── logo.svg
│   └── favicon.ico
│
├── tailwind.config.ts            # Tailwind configuration
├── tsconfig.json                 # TypeScript configuration
├── next.config.js                # Next.js configuration
└── package.json                  # Dependencies
```

---

### Component Architecture

#### Layout System

**Collapsible Sidebar Pattern:**

```
┌─────────────────────────────────────────────────┐
│  SidebarProvider (React Context)                │
│  ├─ collapsed: boolean                          │
│  └─ setCollapsed: (value: boolean) => void      │
│                                                  │
│  ┌────────────────────────────────────────────┐ │
│  │  Sidebar Component                         │ │
│  │  ├─ Reads: collapsed state                 │ │
│  │  ├─ Width: collapsed ? w-16 : w-64         │ │
│  │  └─ Toggle button: setCollapsed(!collapsed)│ │
│  └────────────────────────────────────────────┘ │
│                                                  │
│  ┌────────────────────────────────────────────┐ │
│  │  DashboardContent Component                │ │
│  │  ├─ Reads: collapsed state                 │ │
│  │  ├─ Padding: collapsed ? pl-16 : pl-64     │ │
│  │  └─ Children: page content                 │ │
│  └────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

**Implementation:**

```typescript
// components/layout/SidebarContext.tsx
const SidebarContext = createContext<SidebarContextType | undefined>(undefined)

export function SidebarProvider({ children }: { children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  return (
    <SidebarContext.Provider value={{ collapsed, setCollapsed }}>
      {children}
    </SidebarContext.Provider>
  )
}

// components/layout/Sidebar.tsx
export function Sidebar() {
  const { collapsed, setCollapsed } = useSidebar()
  return (
    <aside className={cn(
      "fixed left-0 top-0 h-screen transition-all",
      collapsed ? "w-16" : "w-64"
    )}>
      {/* Navigation items */}
    </aside>
  )
}

// components/layout/DashboardContent.tsx
export function DashboardContent({ children }) {
  const { collapsed } = useSidebar()
  return (
    <div className={cn(
      "transition-all",
      collapsed ? "pl-16" : "pl-64"
    )}>
      <Header />
      <main>{children}</main>
    </div>
  )
}

// app/(dashboard)/layout.tsx
export default async function DashboardLayout({ children }) {
  return (
    <SidebarProvider>
      <Sidebar />
      <DashboardContent>{children}</DashboardContent>
    </SidebarProvider>
  )
}
```

---

### API Client Pattern

**File:** `lib/api.ts`

```typescript
import { auth } from '@clerk/nextjs/server'

export async function apiClient<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const { getToken } = auth()
  const token = await getToken()

  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const error = await response.json()
    throw new APIError(error.error, response.status)
  }

  return response.json()
}

// Usage in components
const transactions = await apiClient<TransactionsResponse>('/v1/transactions')
```

---

## Data Flow

### Complete Upload Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. USER UPLOADS FILE                                             │
├─────────────────────────────────────────────────────────────────┤
│ User selects CSV/XLSX/PDF file in browser                       │
│ Frontend: DropzoneArea component                                │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. REQUEST PRESIGNED URL                                         │
├─────────────────────────────────────────────────────────────────┤
│ POST /v1/upload/presigned-url                                    │
│ Body: { filename, file_type }                                    │
│ Backend: GeneratePresignedURL handler                            │
│ ├─ Validate file type (CSV/XLSX/PDF)                            │
│ ├─ Generate unique file key: user123/1234567890_file.csv        │
│ └─ Return presigned S3 PUT URL (expires in 5 min)               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. UPLOAD FILE TO S3                                             │
├─────────────────────────────────────────────────────────────────┤
│ PUT {presigned_url}                                              │
│ Body: file binary data                                           │
│ Direct upload from browser to S3 (no backend involved)          │
│ File stored: s3://cashlens-uploads/user123/1234567890_file.csv  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. PROCESS FILE                                                  │
├─────────────────────────────────────────────────────────────────┤
│ POST /v1/upload/process                                          │
│ Body: { file_key, filename }                                     │
│ Backend: ProcessUpload handler                                   │
│ ├─ Download file from S3                                         │
│ ├─ Validate file (magic bytes)                                   │
│ ├─ Detect file type (CSV/XLSX/PDF)                               │
│ ├─ Parse transactions                                            │
│ │   ├─ CSV: Parser.ParseCSV()                                    │
│ │   ├─ XLSX: XLSXParser.ParseXLSX()                              │
│ │   └─ PDF: Python microservice HTTP call                        │
│ ├─ Categorize transactions (Categorizer.CategorizeAll())         │
│ ├─ Insert into database (transactions table)                     │
│ ├─ Create upload history record                                  │
│ └─ Return summary (total, categorized, accuracy)                 │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. DISPLAY RESULTS                                               │
├─────────────────────────────────────────────────────────────────┤
│ Frontend: UploadSummary component                                │
│ ├─ Total Transactions: 156                                       │
│ ├─ Categorized: 142 (91.03%)                                     │
│ ├─ Uncategorized: 14                                             │
│ └─ Link to Review page for manual categorization                │
└─────────────────────────────────────────────────────────────────┘
```

---

## Database Design

### Entity Relationship Diagram

```sql
┌─────────────────────────────────────────────────────────────────┐
│                            users                                 │
├─────────────────────────────────────────────────────────────────┤
│ id                UUID PRIMARY KEY                               │
│ clerk_user_id     VARCHAR(255) UNIQUE NOT NULL                   │
│ email             VARCHAR(255) NOT NULL                          │
│ full_name         VARCHAR(255)                                   │
│ created_at        TIMESTAMP DEFAULT NOW()                        │
│ updated_at        TIMESTAMP DEFAULT NOW()                        │
└────────────┬────────────────────────────────────────────────────┘
             │
             │ 1:N
             │
┌────────────┴────────────────────────────────────────────────────┐
│                       transactions                               │
├─────────────────────────────────────────────────────────────────┤
│ id                UUID PRIMARY KEY                               │
│ user_id           UUID REFERENCES users(id) ON DELETE CASCADE    │
│ txn_date          DATE NOT NULL                                  │
│ description       TEXT NOT NULL                                  │
│ amount            DECIMAL(15,2) NOT NULL                         │
│ txn_type          VARCHAR(10) CHECK (txn_type IN ('debit','credit'))│
│ category          VARCHAR(100)                                   │
│ is_reviewed       BOOLEAN DEFAULT FALSE                          │
│ raw_data          JSONB                                          │
│ created_at        TIMESTAMP DEFAULT NOW()                        │
│ updated_at        TIMESTAMP DEFAULT NOW()                        │
│                                                                  │
│ INDEX idx_user_id (user_id)                                      │
│ INDEX idx_txn_date (txn_date DESC)                               │
│ INDEX idx_category (category)                                    │
│ INDEX idx_is_reviewed (is_reviewed) WHERE is_reviewed = FALSE    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│              global_categorization_rules                         │
├─────────────────────────────────────────────────────────────────┤
│ id                UUID PRIMARY KEY                               │
│ keyword           VARCHAR(255) NOT NULL                          │
│ category          VARCHAR(100) NOT NULL                          │
│ priority          INT DEFAULT 50                                 │
│ is_active         BOOLEAN DEFAULT TRUE                           │
│ created_at        TIMESTAMP DEFAULT NOW()                        │
│                                                                  │
│ INDEX idx_keyword (keyword)                                      │
│ INDEX idx_is_active (is_active) WHERE is_active = TRUE           │
└─────────────────────────────────────────────────────────────────┘

             ┌───────────────────────────────────────────────────┐
             │                                                   │
             │ 1:N                                               │
             │                                                   │
┌────────────┴────────────────────────────────────────────────────┐
│               user_categorization_rules                          │
├─────────────────────────────────────────────────────────────────┤
│ id                UUID PRIMARY KEY                               │
│ user_id           UUID REFERENCES users(id) ON DELETE CASCADE    │
│ keyword           VARCHAR(255) NOT NULL                          │
│ category          VARCHAR(100) NOT NULL                          │
│ priority          INT DEFAULT 50                                 │
│ is_active         BOOLEAN DEFAULT TRUE                           │
│ created_at        TIMESTAMP DEFAULT NOW()                        │
│ updated_at        TIMESTAMP DEFAULT NOW()                        │
│                                                                  │
│ INDEX idx_user_id (user_id)                                      │
│ INDEX idx_keyword (keyword)                                      │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                       upload_history                             │
├─────────────────────────────────────────────────────────────────┤
│ id                UUID PRIMARY KEY                               │
│ user_id           UUID REFERENCES users(id) ON DELETE CASCADE    │
│ filename          VARCHAR(255) NOT NULL                          │
│ file_key          VARCHAR(512) NOT NULL                          │
│ total_rows        INT NOT NULL                                   │
│ categorized_rows  INT NOT NULL                                   │
│ accuracy_percent  DECIMAL(5,2)                                   │
│ status            VARCHAR(50) CHECK (status IN ('processing','completed','failed'))│
│ error_message     TEXT                                           │
│ created_at        TIMESTAMP DEFAULT NOW()                        │
│                                                                  │
│ INDEX idx_user_id (user_id)                                      │
│ INDEX idx_created_at (created_at DESC)                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Design Patterns

### 1. Handler → Service → Database Pattern

**Separation of Concerns:**
- **Handlers:** HTTP layer, input validation, response formatting
- **Services:** Business logic, data processing, external integrations
- **Database:** Data persistence, queries

**Benefits:**
- Testable (mock services in handler tests)
- Maintainable (change business logic without touching HTTP layer)
- Reusable (services can be called from multiple handlers or CLI)

---

### 2. Repository Pattern (Future)

**Current:** Direct database queries in handlers
**Future:** Abstract database operations into repositories

```go
type TransactionRepository interface {
    Create(txn Transaction) error
    FindByID(id string) (Transaction, error)
    FindByUserID(userID string, filters Filters) ([]Transaction, error)
    Update(id string, updates map[string]interface{}) error
    Delete(id string) error
}

type PostgresTransactionRepository struct {
    db *sql.DB
}

func (r *PostgresTransactionRepository) Create(txn Transaction) error {
    // SQL implementation
}
```

---

### 3. Strategy Pattern (Parsers)

**Current:** Switch statement in parser
**Better:** Strategy pattern with interface

```go
type BankParser interface {
    CanParse(headers []string) bool
    Parse(row []string) (ParsedTransaction, error)
}

type HDFCParser struct{}
type ICICIParser struct{}
type SBIParser struct{}

func (p *Parser) ParseCSV(file io.Reader) ([]ParsedTransaction, error) {
    parsers := []BankParser{
        &HDFCParser{},
        &ICICIParser{},
        &SBIParser{},
    }

    for _, parser := range parsers {
        if parser.CanParse(headers) {
            return parser.Parse(rows)
        }
    }

    return nil, errors.New("unsupported format")
}
```

---

## Security Architecture

### Authentication Flow

```
1. User signs in via Clerk UI
2. Clerk issues JWT token (HS256/RS256)
3. Frontend stores token (httpOnly cookie or localStorage)
4. Every API request includes: Authorization: Bearer <token>
5. Backend middleware validates token with Clerk public key
6. If valid, extract user_id from JWT claims
7. Store user_id in request context (c.Locals("user_id"))
8. Handler accesses user_id for database queries
```

### Authorization

**Current:** User can only access their own data
```go
userID := c.Locals("user_id").(string)
transactions := db.FindByUserID(userID) // Automatic tenant isolation
```

**Future:** Role-Based Access Control (RBAC)
- Admin role: Can view all users' data
- Accountant role: Can view but not edit
- Owner role: Full access

---

### File Validation

**Magic Bytes Detection:**
```go
func ValidateFile(data []byte) (FileType, error) {
    // CSV: starts with text (no magic bytes, check valid UTF-8)
    if utf8.Valid(data) {
        return FileTypeCSV, nil
    }

    // XLSX: PK\x03\x04 (ZIP archive)
    if bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04}) {
        return FileTypeXLSX, nil
    }

    // PDF: %PDF-
    if bytes.HasPrefix(data, []byte("%PDF-")) {
        return FileTypePDF, nil
    }

    return FileTypeUnknown, errors.New("invalid file type")
}
```

**Why:** Prevents file extension spoofing (e.g., malware.exe renamed to malware.csv)

---

## Scalability Considerations

### Current Architecture (MVP)

- **Monolithic Go API:** Single binary, simple deployment
- **Synchronous processing:** Upload → parse → categorize → respond
- **Local S3 (LocalStack):** For development

### Future Optimizations

**1. Async Processing:**
```
User uploads file → API returns job ID → Background worker processes file → WebSocket notifies completion
```

**2. Caching Layer:**
```
Redis cache for:
- User categorization rules (TTL: 1 hour)
- Dashboard summaries (TTL: 5 minutes)
- API responses (HTTP caching headers)
```

**3. Database Scaling:**
```
- Read replicas for analytics queries
- Partitioning transactions table by date (monthly partitions)
- Connection pooling (pgBouncer)
```

**4. CDN + Edge Functions:**
```
- Next.js static assets → Vercel Edge Network
- API endpoints → AWS Lambda / Cloudflare Workers
```

**5. Microservices (if needed):**
```
- Auth service (Clerk)
- Upload service (Go)
- Parsing service (Go + Python)
- Categorization service (Go, future: ML model)
- Analytics service (Go + TimescaleDB)
```

---

## Summary

This architecture prioritizes:
1. **Simplicity:** Monolithic Go API for fast development
2. **Testability:** TDD with >80% coverage
3. **Security:** Clerk JWT auth, file validation, input sanitization
4. **Scalability:** Designed for future enhancements (caching, async, microservices)
5. **Maintainability:** Clear separation of concerns (handlers → services → DB)

For implementation details, see:
- [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - API endpoints and integration
- [TESTING.md](TESTING.md) - Testing strategies and commands
- [CODE_EXPLANATION.md](CODE_EXPLANATION.md) - Line-by-line code walkthrough
