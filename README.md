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
- Go 1.21+ with Fiber v3
- PostgreSQL 16
- Redis (caching)
- AWS S3 (file storage)

**Frontend:**
- Next.js 15 (App Router)
- TypeScript
- Tailwind CSS + shadcn/ui
- Clerk (authentication)

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 20+
- Docker & Docker Compose

### 1. Start Infrastructure

```bash
# Start PostgreSQL, Redis, and LocalStack
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 2. Backend Setup

```bash
cd cashlens-api

# Copy environment file
cp .env.example .env

# Install dependencies
go mod download

# Run migrations (coming soon)
# make migrate-up

# Start development server
go run cmd/api/main.go
```

Backend should be running at: http://localhost:8080

Test with: `curl http://localhost:8080/health`

### 3. Frontend Setup

```bash
cd cashlens-web

# Install dependencies
npm install

# Copy environment file
cp .env.example .env.local

# Start development server
npm run dev
```

Frontend should be running at: http://localhost:3000

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

### Current Phase: Day 0 - Project Setup ✅

- [x] Backend project structure
- [x] Frontend project structure
- [x] Docker Compose configuration
- [x] Environment configuration

### Next Steps: Day 1 - Authentication

Follow `CLAUDE.md` instructions to implement Clerk authentication.

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
