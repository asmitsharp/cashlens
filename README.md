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

### Current Phase: Day 1.5 - Design System ✅ COMPLETE

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

**Key Implementation Details:**

**Design System:**
- Documentation: [design-system.md](design-system.md) - Single source of truth for all UI/UX
- Fonts: Inter (UI) + Lora (landing page) in [app/layout.tsx](cashlens-web/app/layout.tsx)
- Tailwind: [tailwind.config.ts](cashlens-web/tailwind.config.ts) - Design tokens + CSS variables
- Colors: [app/globals.css](cashlens-web/app/globals.css) - 50+ color variables (HSL format)
- Clerk Theming: [app/(auth)/sign-in/](cashlens-web/app/(auth)/sign-in/) - Themed authentication

**Authentication:**
- Backend: [internal/middleware/clerk_auth.go](cashlens-api/internal/middleware/clerk_auth.go) - JWT validation
- Frontend: [app/(auth)/](cashlens-web/app/(auth)/) - Authentication pages
- Database: [internal/database/migrations/001_create_users_table.sql](cashlens-api/internal/database/migrations/001_create_users_table.sql)
- Webhooks: [app/api/webhooks/clerk/route.ts](cashlens-web/app/api/webhooks/clerk/route.ts)

### Next Steps: Day 2 - CSV Parser & Normalization

**Goals:**
1. Implement CSV parser for 5 major Indian banks (HDFC, ICICI, SBI, Axis, Kotak)
2. Create bank schema detection
3. Normalize transactions to standard format
4. Achieve 100% unit test coverage

**Start here:**
1. Create test fixtures in `cashlens-api/testdata/`
2. Use TDD to implement parser service
3. Focus on date parsing and amount normalization

## Key Features (MVP)

1. **CSV Upload** - Support for 5 major Indian banks
2. **Auto-Categorization** - 85%+ accuracy with rule engine
3. **Smart Review** - Only review uncategorized transactions
4. **Dashboard** - Net cash flow & top expenses visualization
5. **60-Second Time-to-Dashboard** - From upload to insights

## Documentation

- `CLAUDE.md` - Complete TDD development guide
- `plan.md` - 10-day implementation roadmap
- `techspec.md` - Technical specifications & architecture

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
