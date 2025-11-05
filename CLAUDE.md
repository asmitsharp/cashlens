# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

cashlens is a financial analytics SaaS for Indian SMBs that automatically categorizes bank transactions with 85%+ accuracy. The system follows strict Test-Driven Development (TDD) principles.

**Key Goals:**

- 85%+ auto-categorization accuracy
- 60-second time-to-dashboard
- Support for 5 major Indian banks (HDFC, ICICI, SBI, Axis, Kotak)

## Tech Stack

**Backend** (`cashlens-api/`):

- Go 1.23+ with Fiber v3 framework
- PostgreSQL 16 for data storage
- Redis for caching (future)
- AWS S3 for CSV file storage (LocalStack for local dev)

**Frontend** (`cashlens-web/`):

- Next.js 15 with App Router
- React 19 + TypeScript
- Tailwind CSS + shadcn/ui components
- Clerk for authentication

**Infrastructure:**

- Docker Compose for local development
- PostgreSQL, Redis, LocalStack (S3) services

## Development Commands

### Backend (Go API)

```bash
cd cashlens-api

# Run the API server
go run cmd/api/main.go

# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test -v ./internal/services

# Format code
go fmt ./...

# Lint code
go vet ./...

# Install dependencies
go mod download
go mod tidy
```

### Frontend (Next.js)

```bash
cd cashlens-web

# Start development server
npm run dev

# Build for production
npm run build

# Run production build
npm start

# Lint code
npm run lint

# Run Playwright E2E tests
npx playwright test

# Run specific test file
npx playwright test tests/e2e/upload-flow.spec.ts
```

### Infrastructure

```bash
# Start all services (PostgreSQL, Redis, LocalStack)
docker-compose up -d

# Stop all services
docker-compose down

# View logs
docker-compose logs -f

# Restart specific service
docker-compose restart db

# Check service health
docker-compose ps
```

### Database

```bash
# Connect to PostgreSQL
psql postgres://postgres:dev123@localhost:5432/cashlens

# Run migrations (when implemented)
# go run cmd/migrate/main.go up

# Create new migration
# migrate create -ext sql -dir internal/database/migrations -seq migration_name
```

## Project Architecture

### Backend Structure

```
cashlens-api/
├── cmd/api/main.go              # Application entry point
├── internal/
│   ├── config/                  # Environment configuration loader
│   ├── database/
│   │   ├── migrations/          # SQL migration files
│   │   └── queries/             # SQL queries (sqlc)
│   ├── handlers/                # HTTP request handlers
│   │   ├── auth.go             # Clerk JWT validation
│   │   ├── upload.go           # S3 presigned URLs
│   │   ├── transactions.go     # CRUD operations
│   │   └── summary.go          # Dashboard KPIs
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go             # Authentication
│   │   └── cors.go             # CORS handling
│   ├── models/                  # Domain models
│   ├── services/                # Business logic
│   │   ├── parser.go           # CSV parsing (5 bank formats)
│   │   ├── categorizer.go      # Rule-based categorization
│   │   └── storage.go          # S3 operations
│   └── utils/                   # Helper functions
│       ├── response.go         # JSON response helpers
│       └── errors.go           # Error types
├── testdata/                    # Sample CSV files for testing
└── go.mod
```

**Key Architectural Patterns:**

1. **Handler → Service → Database**: HTTP handlers delegate to services, which interact with the database
2. **Structured Error Handling**: All errors use `APIError` type with status codes
3. **Configuration**: Environment variables loaded via `config.LoadFromEnv()`
4. **Authentication**: Clerk JWT middleware validates all protected routes

### Frontend Structure

```
cashlens-web/
├── app/
│   ├── (auth)/                  # Public auth pages
│   │   ├── sign-in/
│   │   └── sign-up/
│   ├── (dashboard)/             # Protected pages
│   │   ├── layout.tsx          # Auth check + sidebar
│   │   ├── page.tsx            # Dashboard with KPIs
│   │   ├── upload/             # CSV upload page
│   │   └── review/             # Review uncategorized txns
│   ├── api/webhooks/clerk/     # User sync webhook
│   ├── layout.tsx              # Root layout with ClerkProvider
│   └── globals.css
├── components/
│   ├── ui/                     # shadcn/ui components
│   ├── charts/                 # Recharts visualizations
│   ├── upload/                 # File upload components
│   └── transactions/           # Transaction list components
├── lib/
│   └── api.ts                  # API client wrapper
├── stores/
│   └── useAuthStore.ts         # Zustand state (optional with Clerk)
└── types/
    └── index.ts                # TypeScript type definitions
```

**Key Frontend Patterns:**

1. **Route Groups**: `(auth)` and `(dashboard)` for layout organization
2. **Server Components by Default**: Client components marked with `"use client"`
3. **API Client**: Centralized fetch wrapper in `lib/api.ts` with error handling
4. **Authentication**: Clerk handles all auth UI and JWT management

### Database Schema (Key Tables)

```sql
-- Users (Clerk integration)
users (id, clerk_user_id, email, full_name, created_at, updated_at)

-- Transactions (main data)
transactions (id, user_id, txn_date, description, amount, txn_type, category, is_reviewed, raw_data, created_at, updated_at)

-- Global categorization rules
global_categorization_rules (id, keyword, category, priority, is_active, created_at)

-- User-specific rules (override global)
user_categorization_rules (id, user_id, keyword, category, priority, is_active, created_at, updated_at)

-- Upload tracking
upload_history (id, user_id, filename, file_key, total_rows, categorized_rows, accuracy_percent, status, error_message, created_at)
```

**Important Indexes:**

- `transactions(user_id)` - Fast user queries
- `transactions(txn_date DESC)` - Chronological sorting
- `transactions(category)` - Category filtering
- `transactions(is_reviewed)` - Review inbox

## Test-Driven Development (TDD)

**Primary Directive:** ALWAYS write tests BEFORE implementation.

### TDD Workflow

1. **RED**: Write a failing test
2. **GREEN**: Write minimal code to pass the test
3. **REFACTOR**: Improve code while keeping tests green
4. **COMMIT**: Commit after tests pass

### Testing Patterns

**Backend Tests:**

```go
// internal/services/parser_test.go
func TestParseCSV_HDFC(t *testing.T) {
    file, _ := os.Open("testdata/hdfc_sample.csv")
    defer file.Close()

    parser := NewParser()
    transactions, err := parser.ParseCSV(file)

    assert.NoError(t, err)
    assert.Greater(t, len(transactions), 0)
    assert.Equal(t, "AWS SERVICES", transactions[0].Description)
}
```

**Frontend E2E Tests (Playwright):**

```typescript
// tests/e2e/upload-flow.spec.ts
test("complete upload and categorization flow", async ({ page }) => {
  await page.goto("/upload")
  await page.setInputFiles('input[type="file"]', filePath)
  await expect(page.locator("text=Total Transactions")).toBeVisible()

  const accuracy = parseFloat(
    await page.locator('[data-testid="accuracy"]').textContent()
  )
  expect(accuracy).toBeGreaterThanOrEqual(85)
})
```

## CSV Parsing Architecture

The system supports 5 bank formats through schema detection:

1. **Schema Detection** (`DetectSchema()`): Identifies bank by header patterns
2. **Date Parsing** (`ParseDate()`): Handles multiple date formats (DD/MM/YYYY, YYYY-MM-DD, etc.)
3. **Amount Parsing** (`ParseAmount()`): Strips currency symbols (₹, Rs) and commas
4. **Transaction Normalization**: Converts to standard `ParsedTransaction` struct

**Supported Banks:**

- HDFC: "Date", "Narration", "Withdrawal Amt", "Deposit Amt"
- ICICI: "Transaction Date", "Transaction Remarks", "Withdrawal Amount", "Deposit Amount"
- SBI: "Txn Date", "Description", "Debit", "Credit"
- Axis: "Transaction Date", "Particulars", "Dr/Cr"
- Kotak: "Date", "Description", "Debit", "Credit"

## Categorization Engine

**How It Works:**

1. **Global Rules**: Pre-seeded keywords (e.g., "aws" → "Cloud & Hosting")
2. **User Rules**: User-specific overrides (higher priority)
3. **Matching**: Case-insensitive substring matching
4. **Caching**: User rules cached in memory, invalidated on updates

**Accuracy Requirement:** ≥85% on test datasets (5 banks × 100+ transactions each)

**Categories:**

- Cloud & Hosting
- Payment Processing
- Marketing
- Salaries
- Office Supplies
- Software & SaaS
- Travel
- Legal & Professional Services
- Utilities
- Team Meals

## Git Commit Convention

Use **Conventional Commits** with **Gitmoji**:

```
<gitmoji> <type>(<scope>): <subject>

<body>

<footer>
```

**Common Types:**

- ✨ `feat`: New feature
- 🐛 `fix`: Bug fix
- ♻️ `refactor`: Code refactoring
- ✅ `test`: Adding tests
- 📝 `docs`: Documentation
- 🔧 `chore`: Maintenance

**Scopes:** `auth`, `parser`, `categorizer`, `upload`, `dashboard`, `api`, `db`, `config`, `tests`, `infra`

**Example:**

```bash
✨ feat(parser): detect HDFC bank CSV schema

- Implement DetectSchema() function
- Support Date, Narration, Withdrawal Amt columns
- Add test with real HDFC sample CSV
- Achieves 100% detection accuracy on test file

Closes #12
```

**Commit Only When:**

1. ✅ All tests passing (`go test ./...` or `npm test`)
2. ✅ No linter errors (`go vet ./...` or `npm run lint`)
3. ✅ Single logical unit of work
4. ✅ Code formatted (`go fmt ./...`)

## 🎨 Design System & UI/UX

**Primary Directive:** All frontend development (UI/UX) _must_ strictly adhere to the specifications in the **`design-system.md`** file.

- **Theme:** The UI is a direct implementation of the "Pareto" theme. Do not deviate.
- **Source of Truth:** `design-system.md` contains all color, typography, border-radius, and component style definitions.
- **Component Library:** Use `shadcn/ui` components as the base. All components will be automatically styled by the theme definitions in `globals.css` and `tailwind.config.js`.
- **Fonts:** Use `font-sans` (Inter) for all UI and `font-serif` (Lora) _only_ for landing page display headlines.
- **Aesthetic:** Simple, minimal, spacious, and professional.

---

## Environment Configuration

**Backend (.env):**

```bash
# Server
PORT=8080
ENVIRONMENT=development

# Database
DATABASE_URL=postgres://postgres:dev123@localhost:5432/cashlens

# Clerk Auth
CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...

# S3 (LocalStack for local dev)
S3_BUCKET=cashlens-uploads
S3_REGION=ap-south-1
AWS_ENDPOINT=http://localhost:4566
```

**Frontend (.env.local):**

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080/v1
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
```

## Authentication Flow (Clerk)

1. **Frontend**: User signs in via Clerk pre-built UI
2. **JWT Issued**: Clerk issues JWT token
3. **API Requests**: Frontend includes `Authorization: Bearer <token>` header
4. **Backend Validation**: `ClerkAuth()` middleware verifies JWT
5. **User Context**: User ID extracted and stored in `c.Locals("user_id")`
6. **Webhook Sync**: Clerk webhook creates user in local DB on first sign-up

## Common Development Tasks

### Adding a New API Endpoint

1. Write handler test in `internal/handlers/*_test.go`
2. Implement handler in `internal/handlers/*.go`
3. Add route in `cmd/api/main.go`
4. Run tests: `go test -v ./internal/handlers`
5. Test manually: `curl -X POST http://localhost:8080/v1/endpoint`

### Adding a New Frontend Page

1. Create page in `app/(dashboard)/page-name/page.tsx`
2. Add navigation link in layout
3. Create API client function in `lib/api.ts`
4. Write Playwright test in `tests/e2e/page-name.spec.ts`
5. Run: `npm run dev` and test manually

### Adding a Database Migration

1. Create SQL file: `internal/database/migrations/00X_description.sql`
2. Write UP migration (CREATE TABLE, ALTER TABLE, etc.)
3. Test locally: Run SQL against local DB
4. Update models and queries as needed

### Debugging CSV Parsing Issues

1. Add test CSV to `cashlens-api/testdata/`
2. Write test in `internal/services/parser_test.go`
3. Run: `go test -v ./internal/services -run TestParseCSV`
4. Use `t.Logf()` to inspect parsed data
5. Fix `DetectSchema()` or `parseRow()` logic

## Critical Success Metrics

Monitor these throughout development:

- **Categorization Accuracy**: ≥85% (run: `go test -v ./internal/services -run TestCategorizerAccuracy`)
- **Backend Test Coverage**: ≥80% (run: `go test -cover ./...`)
- **Time-to-Dashboard**: ≤60 seconds (Playwright test with timing)
- **API Response Time**: p95 < 500ms for `/summary` endpoint
- **Zero Security Issues**: Run OWASP ZAP scan before production

## References

- **Implementation Plan**: See [plan.md](plan.md) for 10-day roadmap
- **Technical Spec**: See [techspec.md](techspec.md) for detailed architecture
- **README**: See [README.md](README.md) for quick start guide

## Development Best Practices

1. **Always Test First**: Write failing test → Implement → Refactor
2. **Keep Handlers Thin**: Move business logic to services
3. **Use Type Safety**: Leverage Go's type system and TypeScript
4. **Error Handling**: Always return structured errors with context
5. **Database Queries**: Use indexes for all frequent queries
6. **Security**: Never commit secrets, always validate user input
7. **Performance**: Measure before optimizing, profile with benchmarks

## Specialized Development Agents

The project has specialized AI agents available for specific tasks. Use these agents to improve code quality, architecture, and development velocity.

### Backend Development Agents

**TDD Orchestrator** (`backend-development:tdd-orchestrator` - Sonnet)

- **When to use**: For enforcing TDD workflow, coordinating test-first development
- **Best for**: Writing tests before implementation, ensuring red-green-refactor discipline
- **Example**: "Use TDD to implement CSV parser for HDFC bank format"

**Backend Architect** (`backend-development:backend-architect` - Sonnet)

- **When to use**: Designing new API endpoints, microservices architecture, system design
- **Best for**: API design, service boundaries, scalability planning
- **Example**: "Design the categorization service API with clean architecture principles"

**GraphQL Architect** (`backend-development:graphql-architect` - Sonnet)

- **When to use**: If/when implementing GraphQL endpoints (future enhancement)
- **Best for**: Schema design, federation, performance optimization
- **Example**: "Design GraphQL schema for transaction querying"

### Database Development

**Database Architect** (`database-architect` - Opus)

- **When to use**: Database schema design, migration planning, performance optimization
- **Best for**: Table design, indexing strategy, query optimization, data modeling
- **Example**: "Design optimized schema for storing 1M+ transactions with category filters"
- **Example**: "Review and optimize the transactions table indexes for performance"

### Frontend Development

**Frontend Developer** (`frontend-developer` - Sonnet)

- **When to use**: Building UI components, state management, responsive design
- **Best for**: React components, Next.js patterns, accessibility, performance
- **Example**: "Build a responsive transaction review component with keyboard shortcuts"
- **Example**: "Optimize dashboard charts for mobile viewport"
- **Mandate:** All UI must strictly follow `design-system.md`

### Go-Specific Development

**Golang Pro** (`golang-pro` - Sonnet)

- **When to use**: Writing idiomatic Go code, concurrency patterns, optimization
- **Best for**: Goroutines, channels, error handling, Go best practices
- **Example**: "Refactor CSV parser to use goroutines for parallel processing"
- **Example**: "Review and optimize CSV upload handler for better error handling"

### Code Quality & Documentation

**Code Reviewer** (`code-documentation:code-reviewer` - Sonnet)

- **When to use**: After completing features, before merging PRs
- **Best for**: Security vulnerabilities, performance issues, code quality
- **Example**: "Review the authentication middleware for security issues"
- **Mandate:** All frontend PRs _must_ be reviewed for compliance with `design-system.md`
- **Proactive**: Should be used automatically after significant code changes

**Docs Architect** (`code-documentation:docs-architect` - Sonnet)

- **When to use**: Creating technical documentation, architecture guides
- **Best for**: System documentation, API documentation, architecture deep-dives
- **Example**: "Generate comprehensive API documentation for all endpoints"

**Tutorial Engineer** (`code-documentation:tutorial-engineer` - Haiku)

- **When to use**: Creating onboarding guides, tutorials, educational content
- **Best for**: Step-by-step guides, feature tutorials, developer onboarding
- **Example**: "Create a tutorial for adding a new bank CSV format"

### Payment & Compliance

**Payment Integration** (`payment-processing:payment-integration` - Haiku)

- **When to use**: Future feature - subscription billing, payment processing
- **Best for**: Stripe integration, PayPal, payment flows, webhooks
- **Example**: "Implement Stripe subscription billing for premium tier"

### Senior Engineering

**Senior Engineer** (`senior-engineer` - Opus)

- **When to use**: Complex system design, architectural decisions, technical leadership
- **Best for**: High-level architecture, design patterns, technical strategy
- **Example**: "Review overall system architecture for scaling to 100k users"
- **Example**: "Design migration strategy from monolith to microservices"

## Agent Usage Patterns

### Pattern 1: Test-First Development

```bash
# Start with TDD orchestrator for new features
1. Use tdd-orchestrator to write tests first
2. Implement with golang-pro for idiomatic Go
3. Review with code-reviewer before commit
```

### Pattern 2: New Feature Development

```bash
# Full feature cycle
1. Use backend-architect to design API structure
2. Use database-architect for schema changes
3. Use tdd-orchestrator for test-first implementation
4. Use frontend-developer for UI components
5. Use code-reviewer for final review
```

### Pattern 3: Performance Optimization

```bash
# Optimization workflow
1. Use golang-pro to identify bottlenecks
2. Use database-architect for query optimization
3. Use frontend-developer for frontend performance
4. Use code-reviewer to validate improvements
```

### Pattern 4: Documentation Sprint

```bash
# Documentation generation
1. Use docs-architect for system documentation
2. Use tutorial-engineer for onboarding guides
3. Use code-reviewer to validate examples
```

## Common Agent Commands

When working with Claude Code, invoke agents using the Task tool with appropriate prompts:

**Example: TDD for CSV Parser**

```
"Use TDD approach to implement ICICI bank CSV parser. Write tests first,
then implement parser.ParseCSV() function to handle ICICI format from testdata/icici_sample.csv.
Ensure 100% test coverage."
```

**Example: Architecture Review**

```
"Review the current API architecture in cmd/api/main.go and internal/handlers/*.go.
Suggest improvements for scalability and maintainability following clean architecture principles."
```

**Example: Database Optimization**

```
"Analyze the transactions table schema and queries. Suggest optimal indexes for:
1. Filtering by user_id + category
2. Date range queries with pagination
3. Uncategorized transactions lookup
Provide migration SQL and performance estimates."
```

**Example: Frontend Component**

```
"Build a responsive transaction review component with:
1. Data table with sorting/filtering
2. Category dropdown with autocomplete
3. Keyboard shortcuts (Enter to save, arrows to navigate)
4. Optimistic UI updates
5. Accessibility (ARIA labels, keyboard navigation)
Use shadcn/ui components and Tailwind CSS."
```

**Example: Code Review**

```
"Review internal/handlers/upload.go for:
1. Security vulnerabilities (file upload attacks, path traversal)
2. Error handling completeness
3. Performance issues (memory leaks, blocking operations)
4. Go best practices and idioms
Provide specific fixes with code examples."
```

## Troubleshooting

**Database connection fails:**

```bash
docker-compose ps          # Check if db is running
docker-compose logs db     # View database logs
psql postgres://postgres:dev123@localhost:5432/cashlens  # Test connection
```

**Frontend can't reach API:**

- Check `NEXT_PUBLIC_API_URL` in `.env.local`
- Ensure backend is running: `curl http://localhost:8080/health`
- Check browser console for CORS errors

**Tests failing:**

- Run `go mod tidy` or `npm install` to ensure dependencies are up-to-date
- Check test database is empty: `docker-compose down -v && docker-compose up -d`
- Verify test fixtures exist in `testdata/`

**CSV parsing errors:**

- Check bank format matches expected schema
- Verify date format is supported
- Test with minimal CSV (3-5 rows) first
- Add debug logging in `parseRow()`
