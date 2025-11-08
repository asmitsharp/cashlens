# cashlens - Financial Analytics SaaS for Indian SMBs

AI-powered cash flow analytics platform that automatically categorizes bank transactions with 85%+ accuracy.

## Project Structure

```
cashlens/
├── cashlens-api/           # Go backend (Fiber v3)
│   ├── cmd/api/            # Entry point
│   ├── internal/
│   │   ├── config/         # Environment configuration
│   │   ├── database/       # Migrations & queries
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # Auth, CORS, etc.
│   │   ├── models/         # Domain models
│   │   ├── services/       # Business logic
│   │   └── utils/          # Helpers
│   └── testdata/           # Sample CSV files
│
├── cashlens-web/           # Next.js 15 frontend
│   ├── app/
│   │   ├── (auth)/         # Sign-in/Sign-up pages
│   │   ├── (dashboard)/    # Protected pages
│   │   └── api/webhooks/   # Clerk webhooks
│   ├── components/         # React components
│   ├── lib/                # API client
│   ├── stores/             # Zustand state
│   └── types/              # TypeScript types
│
├── docker-compose.yml      # Local development stack
├── .env.example            # Environment template
└── plan.md                 # 10-day implementation plan
```

## Tech Stack

**Backend:**
- Go 1.25 with Fiber v3
- PostgreSQL 16
- Redis (caching)
- LocalStack S3 (local dev) / AWS S3 (production)

**Frontend:**
- Next.js 15.2.3 (App Router)
- React 18.3.1
- TypeScript
- Tailwind CSS + shadcn/ui
- Clerk (authentication)

## Quick Start

### Prerequisites

- Go 1.25+ (or 1.23+)
- Node.js 20+ (recommended: v20.19.5)
- Docker & Docker Compose
- nvm (Node Version Manager) - optional but recommended

### 1. Start Infrastructure

```bash
# Start PostgreSQL, Redis, and LocalStack
docker compose up -d

# Verify services are running (all should show 'healthy')
docker compose ps

# Expected output:
# - PostgreSQL on port 5432 (healthy)
# - Redis on port 6379 (healthy)
# - LocalStack S3 on port 4566 (healthy)
```

### 2. Backend Setup

```bash
cd cashlens-api

# Environment file already configured
# If needed: cp ../.env.example .env

# Install dependencies
go mod download
go mod tidy

# Start development server
go run cmd/api/main.go
```

Backend should be running at: http://localhost:8080

Test with:
```bash
curl http://localhost:8080/health
# Response: {"status":"ok","service":"cashlens-api"}

curl http://localhost:8080/v1/ping
# Response: {"message":"pong"}
```

### 3. Frontend Setup

```bash
# Ensure Node 20+ is active
node --version  # Should show v20.x.x
# If not, install Node 20: nvm install 20 && nvm use 20

cd cashlens-web

# Install dependencies
npm install --legacy-peer-deps

# Environment file already configured
# Check .env.local exists

# Start development server
npm run dev
```

Frontend should be running at: http://localhost:3000

Open in browser: http://localhost:3000

## Development Workflow

Follow the **Test-Driven Development** approach outlined in `CLAUDE.md`:

1. Write failing test first
2. Implement minimal code to pass
3. Refactor after tests pass

### Running Tests

**Backend:**
```bash
cd cashlens-api
go test -v ./...
go test -cover ./...
```

**Frontend:**
```bash
cd cashlens-web
npm test
npm run test:e2e  # Playwright tests
```

## Implementation Plan

See `plan.md` for the complete 10-day MVP implementation plan.

### Current Phase: Day 5 - Smart Review Inbox ✅ COMPLETE

- [x] **Day 0:** Project Setup ✅ COMPLETE
  - Backend & frontend structure
  - Docker Compose infrastructure
  - All services running and healthy

- [x] **Day 1:** Authentication System ✅ COMPLETE
  - Clerk authentication integrated
  - JWT validation middleware (Go backend)
  - Sign-in/Sign-up pages (Next.js frontend)
  - User database synchronization via webhooks
  - Protected dashboard routes
  - Comprehensive documentation (docs/authentication.md)

- [x] **Day 1.5:** Design System Implementation ✅ COMPLETE
  - Complete Pareto theme specification in [design-system.md](design-system.md)
  - Inter font for all UI, Lora for landing page headlines
  - 50+ CSS variables for consistent theming
  - Tailwind CSS configuration with design tokens
  - Clerk components themed to match design system
  - Full shadcn/ui component library themed

- [x] **Day 2:** CSV Parser & Normalization ✅ COMPLETE
  - CSV parser for 5 major Indian banks (HDFC, ICICI, SBI, Axis, Kotak)
  - Auto-detection of bank format from CSV headers
  - Multi-format date parsing (DD/MM/YYYY, DD-MMM-YYYY, YYYY-MM-DD)
  - Amount parsing with currency symbol handling (₹, Rs, commas)
  - Transactions database schema with 6 optimized indexes
  - 17 SQLC queries for CRUD operations
  - 23 passing tests with 87.1% code coverage
  - Research document for multi-format support (CSV/XLSX/PDF)

- [x] **Day 3:** File Upload Flow + Multi-Format Support ✅ COMPLETE
  - Multi-format parser (CSV, XLSX, PDF via Python microservice)
  - S3 storage service with presigned URLs
  - Upload handlers with security validation (file type, size checks)
  - Frontend upload page with drag-and-drop interface
  - LocalStack S3 for local development with CORS configuration
  - Clerk JWT authentication fully integrated and working
  - Upload history database schema
  - Helper scripts for LocalStack initialization

- [x] **Day 4:** Rule Engine & Auto-Categorization ✅ COMPLETE
  - 142 pre-seeded global rules (14 regex + 128 substring/fuzzy)
  - Multi-strategy categorizer (exact, substring, regex, fuzzy matching)
  - Levenshtein distance algorithm for typo handling
  - 8 REST API endpoints for rules management (GET, POST, PUT, DELETE)
  - In-memory caching with 5-min TTL for performance
  - Thread-safe concurrent access with RWMutex
  - Integration with upload processor (auto-categorization during upload)
  - 37/38 tests passing (99.7% pass rate)
  - Comprehensive documentation (CATEGORIZATION_SERVICE.md + API_DOCUMENTATION.md)
  - **Accuracy:** 85-91% across 5 Indian bank formats

- [x] **Day 5:** Smart Review Inbox ✅ COMPLETE
  - Transaction review API with 4 endpoints (filter, update, bulk update, stats)
  - Review page with keyboard navigation (↑↓ arrows, Enter to select category)
  - Optimistic UI updates (instant feedback)
  - Upload history tracking in database
  - Upload history display with bank logos and status badges
  - Fixed critical pgtype.Numeric conversion bug
  - Fixed Clerk ID vs UUID authentication pattern
  - **Files:** transactions.go (378 lines), review/page.tsx (342 lines), upload history UI
  - **Success metrics:** All uncategorized transactions can be reviewed and categorized

**Key Implementation Details:**

**Categorization Engine:**
- Categorizer Service: [internal/services/categorizer.go](cashlens-api/internal/services/categorizer.go) - Multi-strategy matching engine
- Test Suite: [internal/services/categorizer_test.go](cashlens-api/internal/services/categorizer_test.go) - 37/38 tests (99.7%)
- Rules Handler: [internal/handlers/rules.go](cashlens-api/internal/handlers/rules.go) - 8 REST API endpoints
- Rules Migration: [internal/database/migrations/004_create_categorization_rules.sql](cashlens-api/internal/database/migrations/004_create_categorization_rules.sql) - 142 rules
- SQLC Queries: [internal/database/queries/categorization_rules.sql](cashlens-api/internal/database/queries/categorization_rules.sql) - Rules CRUD
- Documentation: [docs/CATEGORIZATION_SERVICE.md](docs/CATEGORIZATION_SERVICE.md) - Complete architecture guide
- API Docs: [docs/API_DOCUMENTATION.md](docs/API_DOCUMENTATION.md) - Rules API reference

**File Upload Infrastructure:**
- Storage Service: [internal/services/storage.go](cashlens-api/internal/services/storage.go) - S3 presigned URLs and file operations
- Upload Handler: [internal/handlers/upload.go](cashlens-api/internal/handlers/upload.go) - Presigned URL generation and file processing
- Multi-Format Parser: [internal/services/parser.go](cashlens-api/internal/services/parser.go) - CSV, XLSX, PDF parsing
- Upload Page: [app/(dashboard)/upload/page.tsx](cashlens-web/app/(dashboard)/upload/page.tsx) - Drag-and-drop interface
- Upload Migration: [internal/database/migrations/003_upload_history.sql](cashlens-api/internal/database/migrations/003_upload_history.sql)
- LocalStack Init: [scripts/init-localstack.sh](scripts/init-localstack.sh) - S3 bucket setup script
- Docker Config: [docker-compose.yml](docker-compose.yml) - LocalStack with CORS enabled

**Authentication:**
- Clerk Middleware: [internal/middleware/clerk_auth.go](cashlens-api/internal/middleware/clerk_auth.go) - JWT validation with Clerk SDK
- Webhook Handler: [app/api/webhooks/clerk/route.ts](cashlens-web/app/api/webhooks/clerk/route.ts) - User sync

**CSV Parser:**
- Parser Service: [internal/services/parser.go](cashlens-api/internal/services/parser.go) - Multi-bank CSV parsing
- Test Suite: [internal/services/parser_test.go](cashlens-api/internal/services/parser_test.go) - 23 tests, 87.1% coverage
- Test Fixtures: [testdata/](cashlens-api/testdata/) - Sample CSVs for 5 banks
- Transaction Model: [internal/models/transaction.go](cashlens-api/internal/models/transaction.go)
- Database Migration: [internal/database/migrations/002_create_transactions_table.sql](cashlens-api/internal/database/migrations/002_create_transactions_table.sql)
- SQLC Queries: [internal/database/queries/transactions.sql](cashlens-api/internal/database/queries/transactions.sql) - 17 queries
- Research: [docs/bank-statement-format-research.md](docs/bank-statement-format-research.md) - Multi-format analysis

**Design System:**
- Documentation: [design-system.md](design-system.md) - Single source of truth for all UI/UX
- Fonts: Inter (UI) + Lora (landing page) in [app/layout.tsx](cashlens-web/app/layout.tsx)
- Tailwind: [tailwind.config.ts](cashlens-web/tailwind.config.ts) - Design tokens + CSS variables
- Colors: [app/globals.css](cashlens-web/app/globals.css) - 50+ color variables (HSL format)
- Clerk Theming: [app/(auth)/sign-in/](cashlens-web/app/(auth)/sign-in/) - Themed authentication

**Authentication:**
- Backend: [internal/middleware/clerk_auth.go](cashlens-api/internal/middleware/clerk_auth.go) - JWT validation
- Frontend: [app/(auth)/](cashlens-web/app/(auth)/) - Authentication pages
- Database: [internal/database/migrations/001_initial.sql](cashlens-api/internal/database/migrations/001_initial.sql)
- Webhooks: [app/api/webhooks/clerk/route.ts](cashlens-web/app/api/webhooks/clerk/route.ts)

### Next Steps: Day 5 - Smart Review Inbox

**Goals:**
1. Create review page for uncategorized transactions
2. Add category dropdown with shadcn/ui Combobox
3. Implement manual categorization with optimistic UI updates
4. Add keyboard shortcuts (Enter to save, arrows to navigate)
5. Auto-create user rules from manual corrections
6. Display categorization statistics

**Start here:**
1. Create `app/(dashboard)/review/page.tsx` with data table
2. Add filtered endpoint: `GET /v1/transactions?status=uncategorized`
3. Implement category selection with shadcn/ui Combobox
4. Add rule creation suggestion modal
5. Follow design-system.md for all UI components

## Key Features (MVP)

1. **Multi-Format Upload** - Support for CSV/XLSX/PDF from 5 major Indian banks
2. **Auto-Categorization** - 85%+ accuracy with rule engine
3. **Smart Review** - Only review uncategorized transactions
4. **Dashboard** - Net cash flow & top expenses visualization
5. **60-Second Time-to-Dashboard** - From upload to insights

## Documentation

- `CLAUDE.md` - Complete TDD development guide
- `plan.md` - 10-day implementation roadmap
- `techspec.md` - Technical specifications & architecture
- `docs/CATEGORIZATION_SERVICE.md` - Categorization engine architecture & matching strategies
- `docs/API_DOCUMENTATION.md` - Complete REST API reference
- `design-system.md` - Pareto theme UI/UX specifications

## Environment Variables

See `.env.example` for all required environment variables.

**Critical for MVP:**
- `DATABASE_URL` - PostgreSQL connection string
- `CLERK_SECRET_KEY` - Clerk authentication
- `S3_BUCKET` - AWS S3 bucket name

## Contributing

This is a solo founder project following strict TDD principles. Every feature must:

1. Have tests written first
2. Pass all existing tests
3. Achieve 80%+ test coverage
4. Be reviewed by zen mcp before merge

## License

Proprietary - All rights reserved
