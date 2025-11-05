# Cashlens Testing Guide

This document provides comprehensive testing commands, strategies, and coverage requirements for the Cashlens project.

---

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Backend Testing (Go)](#backend-testing-go)
3. [Frontend Testing (Next.js)](#frontend-testing-nextjs)
4. [Integration Testing](#integration-testing)
5. [E2E Testing](#e2e-testing)
6. [Test Coverage Requirements](#test-coverage-requirements)
7. [CI/CD Testing](#cicd-testing)

---

## Testing Philosophy

Cashlens follows **Test-Driven Development (TDD)** principles:

1. **RED**: Write a failing test first
2. **GREEN**: Write minimal code to pass the test
3. **REFACTOR**: Improve code while keeping tests green
4. **COMMIT**: Commit only when all tests pass

**Coverage Targets:**
- Backend: ≥80% line coverage
- Frontend: ≥70% component coverage
- Categorization accuracy: ≥85%

---

## Backend Testing (Go)

### Quick Start

```bash
cd cashlens-api

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Run tests for specific package
go test -v ./internal/services
go test -v ./internal/handlers

# Run specific test function
go test -v ./internal/services -run TestParseCSV_HDFC

# Run tests with race detection
go test -race ./...

# Run tests with timeout
go test -timeout 30s ./...

# Clean test cache and re-run
go clean -testcache && go test ./...
```

---

### Test File Structure

**Naming Convention:** `*_test.go`

**Example Test File:**
```go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test data"

    // Act
    result := FunctionToTest(input)

    // Assert
    assert.Equal(t, expectedValue, result)
}
```

---

### Unit Tests by Package

#### 1. Parser Tests (`internal/services/parser_test.go`)

**Test Coverage:**
- CSV parsing for all 5 banks (HDFC, ICICI, SBI, Axis, Kotak)
- Date format detection and parsing
- Amount parsing (₹, Rs, commas)
- Transaction type detection (debit/credit)
- Schema detection
- Error handling (invalid files, malformed data)

**Run Parser Tests:**
```bash
# All parser tests
go test -v ./internal/services -run TestParse

# Specific bank
go test -v ./internal/services -run TestParseCSV_HDFC
go test -v ./internal/services -run TestParseCSV_ICICI

# Date parsing
go test -v ./internal/services -run TestParseDate

# Amount parsing
go test -v ./internal/services -run TestParseAmount

# Schema detection
go test -v ./internal/services -run TestDetectSchema
```

**Sample Test:**
```go
func TestParseCSV_HDFC(t *testing.T) {
    file, err := os.Open("../../testdata/hdfc_sample.csv")
    assert.NoError(t, err)
    defer file.Close()

    parser := NewParser()
    transactions, err := parser.ParseCSV(file)

    assert.NoError(t, err)
    assert.Greater(t, len(transactions), 0)
    assert.Equal(t, "AWS SERVICES", transactions[0].Description)
    assert.Equal(t, -1250.50, transactions[0].Amount)
}
```

**Test Data Location:** `cashlens-api/testdata/*.csv`

---

#### 2. XLSX Parser Tests (`internal/services/xlsx_parser_test.go`)

**Test Coverage:**
- XLSX file opening and reading
- Sheet detection and selection
- Row iteration and data extraction
- Same schema detection as CSV
- Error handling (corrupted files, empty sheets)

**Run XLSX Tests:**
```bash
# All XLSX tests
go test -v ./internal/services -run TestXLSX

# Specific tests
go test -v ./internal/services -run TestParseXLSX_HDFC
go test -v ./internal/services -run TestParseXLSX_EmptyFile
go test -v ./internal/services -run TestParseXLSX_InvalidFormat

# With coverage
go test -coverprofile=xlsx_coverage.out ./internal/services -run TestXLSX
go tool cover -func=xlsx_coverage.out
```

**Sample Test:**
```go
func TestParseXLSX_HDFC(t *testing.T) {
    file, err := os.Open("../../testdata/hdfc_sample.xlsx")
    assert.NoError(t, err)
    defer file.Close()

    parser := NewXLSXParser()
    transactions, err := parser.ParseXLSX(file)

    assert.NoError(t, err)
    assert.Greater(t, len(transactions), 0)
    assert.Equal(t, "debit", transactions[0].TxnType)
}
```

**Coverage Target:** 85%+ (achieved: 89.2%)

---

#### 3. Categorizer Tests (`internal/services/categorizer_test.go`)

**Test Coverage:**
- Rule matching (global and user-specific)
- Priority handling
- Accuracy measurement
- Case-insensitive matching
- Keyword substring matching
- Uncategorized transaction handling

**Run Categorizer Tests:**
```bash
# All categorizer tests
go test -v ./internal/services -run TestCategorizer

# Accuracy test (critical)
go test -v ./internal/services -run TestCategorizerAccuracy

# Rule priority
go test -v ./internal/services -run TestRulePriority

# Case sensitivity
go test -v ./internal/services -run TestCaseInsensitive
```

**Sample Accuracy Test:**
```go
func TestCategorizerAccuracy(t *testing.T) {
    categorizer := NewCategorizer(globalRules, userRules)

    testTransactions := []ParsedTransaction{
        {Description: "AWS SERVICES", Amount: -1000},
        {Description: "SALARY CREDIT", Amount: 50000},
        {Description: "GOOGLE ADS", Amount: -5000},
    }

    categorized := categorizer.CategorizeAll(testTransactions)
    accuracy := calculateAccuracy(categorized)

    assert.GreaterOrEqual(t, accuracy, 85.0,
        "Categorization accuracy must be >= 85%")
}
```

**Success Criteria:** ≥85% accuracy on test dataset (500+ transactions across 5 banks)

---

#### 4. Storage Service Tests (`internal/services/storage_test.go`)

**Test Coverage:**
- S3 presigned URL generation
- File upload/download
- File deletion
- LocalStack integration (local testing)
- Error handling (network failures, invalid keys)

**Run Storage Tests:**
```bash
# Start LocalStack first
docker-compose up -d localstack

# Run storage tests
go test -v ./internal/services -run TestStorage

# Specific tests
go test -v ./internal/services -run TestGeneratePresignedURL
go test -v ./internal/services -run TestUploadFile
go test -v ./internal/services -run TestDownloadFile
```

---

#### 5. Handler Tests (`internal/handlers/*_test.go`)

**Test Coverage:**
- HTTP request/response handling
- Authentication middleware
- Input validation
- Error responses
- Status codes

**Run Handler Tests:**
```bash
# All handler tests
go test -v ./internal/handlers

# Specific handlers
go test -v ./internal/handlers -run TestUploadHandlers
go test -v ./internal/handlers -run TestTransactionHandlers
go test -v ./internal/handlers -run TestSummaryHandler

# Test with mock database
go test -v ./internal/handlers -run TestGetTransactions
```

**Sample Handler Test:**
```go
func TestGeneratePresignedURL(t *testing.T) {
    app := fiber.New()
    app.Post("/presigned-url", GeneratePresignedURL)

    reqBody := `{"filename":"test.csv","file_type":"text/csv"}`
    req := httptest.NewRequest("POST", "/presigned-url",
        strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")

    resp, _ := app.Test(req)

    assert.Equal(t, 200, resp.StatusCode)
}
```

---

#### 6. Validation Tests (`internal/services/validation_test.go`)

**Test Coverage:**
- File type validation (magic bytes)
- CSV validation
- XLSX validation
- PDF validation
- Size limits

**Run Validation Tests:**
```bash
go test -v ./internal/services -run TestValidation

# Specific tests
go test -v ./internal/services -run TestValidateFileType
go test -v ./internal/services -run TestMagicBytes
```

---

### Test Data Management

**Location:** `cashlens-api/testdata/`

**Files:**
```
testdata/
├── hdfc_sample.csv          # 50 transactions
├── hdfc_sample.xlsx         # Same data in XLSX
├── icici_sample.csv         # 50 transactions
├── sbi_sample.csv           # 50 transactions
├── axis_sample.csv          # 50 transactions
├── kotak_sample.csv         # 50 transactions
├── invalid.csv              # Malformed CSV
├── empty.csv                # Empty file
└── mixed_bank.csv           # Multiple bank formats (error case)
```

**Creating Test Data:**
```bash
# Use real anonymized bank statements
# Replace sensitive data with fake data
# Ensure variety: different categories, amounts, dates
```

---

### Benchmarking

**Run Benchmarks:**
```bash
# Run all benchmarks
go test -bench=. ./...

# Specific benchmark
go test -bench=BenchmarkParseCSV ./internal/services

# With memory allocation stats
go test -bench=. -benchmem ./internal/services

# Profile CPU usage
go test -bench=BenchmarkParseCSV -cpuprofile=cpu.prof ./internal/services
go tool pprof cpu.prof
```

**Sample Benchmark:**
```go
func BenchmarkParseCSV(b *testing.B) {
    file, _ := os.Open("../../testdata/hdfc_sample.csv")
    defer file.Close()

    parser := NewParser()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        file.Seek(0, 0)
        parser.ParseCSV(file)
    }
}
```

---

### Integration Tests with Database

**Prerequisites:**
```bash
# Start PostgreSQL
docker-compose up -d db

# Run migrations (when implemented)
# go run cmd/migrate/main.go up
```

**Run Integration Tests:**
```bash
# Set test database URL
export DATABASE_URL="postgres://postgres:dev123@localhost:5432/cashlens_test"

# Run integration tests
go test -v ./internal/handlers -tags=integration

# Clean up test database after
psql $DATABASE_URL -c "TRUNCATE transactions, users, upload_history CASCADE;"
```

---

## Frontend Testing (Next.js)

### Quick Start

```bash
cd cashlens-web

# Install dependencies
npm install

# Run unit tests (when implemented)
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm test -- --coverage

# Type checking
npm run type-check

# Linting
npm run lint

# Fix linting issues
npm run lint -- --fix
```

---

### Component Testing

**Testing Library:** Jest + React Testing Library

**Example Component Test:**
```typescript
// components/upload/DropzoneArea.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { DropzoneArea } from './DropzoneArea'

describe('DropzoneArea', () => {
  it('renders upload area', () => {
    render(<DropzoneArea onFileSelect={jest.fn()} />)
    expect(screen.getByText(/drag.*drop/i)).toBeInTheDocument()
  })

  it('calls onFileSelect when file is dropped', async () => {
    const onFileSelect = jest.fn()
    render(<DropzoneArea onFileSelect={onFileSelect} />)

    const file = new File(['content'], 'test.csv', { type: 'text/csv' })
    const input = screen.getByLabelText(/upload/i)

    fireEvent.change(input, { target: { files: [file] } })

    expect(onFileSelect).toHaveBeenCalledWith(file)
  })
})
```

**Run Component Tests:**
```bash
# All component tests
npm test -- components/

# Specific component
npm test -- DropzoneArea

# Update snapshots
npm test -- -u
```

---

### API Client Testing

**Mock API Responses:**
```typescript
// lib/api.test.ts
import { apiClient } from './api'

global.fetch = jest.fn()

describe('apiClient', () => {
  beforeEach(() => {
    jest.resetAllMocks()
  })

  it('includes authorization header', async () => {
    (fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => ({ data: 'test' }),
    })

    await apiClient('/v1/transactions')

    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: expect.stringContaining('Bearer'),
        }),
      })
    )
  })
})
```

---

## E2E Testing

### Playwright Setup

**Installation:**
```bash
cd cashlens-web

# Install Playwright
npx playwright install

# Install browsers
npx playwright install chromium firefox webkit
```

**Run E2E Tests:**
```bash
# Run all E2E tests
npx playwright test

# Run in headed mode (see browser)
npx playwright test --headed

# Run specific test file
npx playwright test tests/e2e/upload-flow.spec.ts

# Run in debug mode
npx playwright test --debug

# Run in specific browser
npx playwright test --project=chromium

# Generate test report
npx playwright show-report
```

---

### E2E Test Examples

#### 1. Upload Flow Test

**File:** `tests/e2e/upload-flow.spec.ts`

```typescript
import { test, expect } from '@playwright/test'

test('complete CSV upload flow', async ({ page }) => {
  // Login
  await page.goto('/sign-in')
  await page.fill('input[name="email"]', 'test@example.com')
  await page.fill('input[name="password"]', 'password123')
  await page.click('button[type="submit"]')

  // Navigate to upload page
  await page.goto('/upload')
  await expect(page.locator('h1')).toContainText('Upload')

  // Upload file
  const filePath = 'testdata/hdfc_sample.csv'
  await page.setInputFiles('input[type="file"]', filePath)

  // Wait for processing
  await expect(page.locator('text=Processing')).toBeVisible()
  await expect(page.locator('text=Complete')).toBeVisible({ timeout: 30000 })

  // Verify accuracy
  const accuracy = await page.locator('[data-testid="accuracy"]').textContent()
  const accuracyValue = parseFloat(accuracy!)
  expect(accuracyValue).toBeGreaterThanOrEqual(85)

  // Verify transaction count
  const total = await page.locator('[data-testid="total-transactions"]').textContent()
  expect(parseInt(total!)).toBeGreaterThan(0)
})
```

#### 2. Dashboard Test

**File:** `tests/e2e/dashboard.spec.ts`

```typescript
test('dashboard displays KPIs', async ({ page }) => {
  await page.goto('/dashboard')

  // Check KPI cards
  await expect(page.locator('text=Total Income')).toBeVisible()
  await expect(page.locator('text=Total Expenses')).toBeVisible()
  await expect(page.locator('text=Net Balance')).toBeVisible()

  // Check charts render
  await expect(page.locator('[data-testid="expense-chart"]')).toBeVisible()
  await expect(page.locator('[data-testid="trend-chart"]')).toBeVisible()

  // Verify data loads
  const income = await page.locator('[data-testid="total-income"]').textContent()
  expect(income).toMatch(/₹\s*[\d,]+/)
})
```

#### 3. Transaction Review Test

**File:** `tests/e2e/review.spec.ts`

```typescript
test('review and categorize transaction', async ({ page }) => {
  await page.goto('/review')

  // Check uncategorized transactions appear
  const firstRow = page.locator('table tbody tr').first()
  await expect(firstRow).toBeVisible()

  // Open category dropdown
  await firstRow.locator('[data-testid="category-select"]').click()

  // Select category
  await page.locator('text=Cloud & Hosting').click()

  // Mark as reviewed
  await firstRow.locator('[data-testid="mark-reviewed"]').check()

  // Save changes
  await page.locator('button:has-text("Save")').click()

  // Verify success message
  await expect(page.locator('text=Saved successfully')).toBeVisible()

  // Verify transaction disappears from review list
  await expect(firstRow).not.toBeVisible()
})
```

---

### E2E Test Configuration

**File:** `playwright.config.ts`

```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 30000,
  retries: 2,
  workers: 4,

  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'mobile',
      use: { ...devices['iPhone 13'] },
    },
  ],

  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: true,
  },
})
```

---

## Test Coverage Requirements

### Backend Coverage Targets

| Package | Target | Current |
|---------|--------|---------|
| `internal/services` | 85% | 89.2% ✅ |
| `internal/handlers` | 80% | TBD |
| `internal/middleware` | 90% | TBD |
| `internal/models` | 70% | TBD |
| **Overall** | **80%** | **TBD** |

### Frontend Coverage Targets

| Category | Target | Current |
|---------|--------|---------|
| Components | 70% | TBD |
| Pages | 60% | TBD |
| Utilities | 80% | TBD |
| **Overall** | **70%** | **TBD** |

### Critical Path Coverage

**Must have 100% coverage:**
- Authentication logic
- File upload validation
- CSV/XLSX/PDF parsing
- Categorization engine
- Payment processing (future)

---

## CI/CD Testing

### GitHub Actions Workflow

**File:** `.github/workflows/test.yml`

```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  backend:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test123
          POSTGRES_DB: cashlens_test
        ports:
          - 5432:5432

      localstack:
        image: localstack/localstack
        ports:
          - 4566:4566

    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Run tests
        working-directory: ./cashlens-api
        run: |
          go test -v -cover -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out

      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80%"
            exit 1
          fi

  frontend:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '20'

      - name: Install dependencies
        working-directory: ./cashlens-web
        run: npm ci

      - name: Run tests
        working-directory: ./cashlens-web
        run: npm test -- --coverage

      - name: Run E2E tests
        working-directory: ./cashlens-web
        run: npx playwright test
```

---

## Pre-commit Checks

**Install pre-commit hook:**
```bash
# Create .git/hooks/pre-commit
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

echo "Running pre-commit tests..."

# Backend tests
cd cashlens-api
go test ./...
if [ $? -ne 0 ]; then
    echo "❌ Backend tests failed"
    exit 1
fi

go vet ./...
if [ $? -ne 0 ]; then
    echo "❌ Go vet failed"
    exit 1
fi

# Frontend lint
cd ../cashlens-web
npm run lint
if [ $? -ne 0 ]; then
    echo "❌ Frontend lint failed"
    exit 1
fi

echo "✅ All pre-commit checks passed"
EOF

chmod +x .git/hooks/pre-commit
```

---

## Troubleshooting Tests

### Backend Tests Failing

**Database connection errors:**
```bash
# Check PostgreSQL is running
docker-compose ps db

# Check connection
psql postgres://postgres:dev123@localhost:5432/cashlens
```

**S3/LocalStack errors:**
```bash
# Check LocalStack is running
docker-compose ps localstack

# Test S3 connectivity
aws --endpoint-url=http://localhost:4566 s3 ls s3://cashlens-uploads
```

**Test data missing:**
```bash
# Ensure testdata files exist
ls cashlens-api/testdata/

# Regenerate if needed
cp path/to/bank/statements cashlens-api/testdata/
```

### Frontend Tests Failing

**Port already in use:**
```bash
# Kill process on port 3000
lsof -ti:3000 | xargs kill -9
```

**Playwright browser issues:**
```bash
# Reinstall browsers
npx playwright install --force
```

---

## Summary of Test Commands

### Backend (Go)
```bash
cd cashlens-api

# Quick test
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test -v ./internal/services

# Generate HTML coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend (Next.js)
```bash
cd cashlens-web

# Unit tests
npm test

# E2E tests
npx playwright test

# Lint
npm run lint
```

### Full Test Suite
```bash
# Run everything from root
./scripts/test-all.sh  # (create this script)
```

---

## Next Steps

1. **Achieve 80%+ backend coverage** - Add handler and middleware tests
2. **Implement frontend unit tests** - Test all components
3. **Expand E2E test suite** - Cover all user flows
4. **Set up CI/CD** - Automate testing on every push
5. **Performance testing** - Load test API endpoints
6. **Security testing** - OWASP ZAP scans

---

For questions or issues with testing, see [CONTRIBUTING.md](CONTRIBUTING.md) or open a GitHub issue.
