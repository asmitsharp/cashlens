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

### Current Phase: Day 0 - Project Setup ✅ COMPLETE

- [x] Backend project structure
- [x] Frontend project structure
- [x] Docker Compose configuration (PostgreSQL, Redis, LocalStack)
- [x] Environment configuration
- [x] Go dependencies installed (Fiber, pgx, AWS SDK)
- [x] Node dependencies installed (Next.js, React, Tailwind)
- [x] Backend server running on port 8080
- [x] Frontend server running on port 3000
- [x] All infrastructure services healthy
- [x] LocalStack S3 configured and working

### Next Steps: Day 1 - Authentication

Follow `CLAUDE.md` instructions to implement Clerk authentication.

**Start here:**
1. Sign up for Clerk account at https://dashboard.clerk.com
2. Get API keys (CLERK_PUBLISHABLE_KEY, CLERK_SECRET_KEY)
3. Update .env files with Clerk credentials
4. Implement Clerk middleware in backend
5. Add authentication pages in frontend

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
