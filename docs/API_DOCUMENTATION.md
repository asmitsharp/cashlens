# Cashlens API Documentation

## Overview

The Cashlens API is a RESTful service built with Go 1.23+ and Fiber v3 framework. It provides endpoints for CSV/XLSX/PDF file upload, transaction parsing, categorization, and financial analytics for Indian SMBs.

**Base URL:** `http://localhost:8080/v1`

**Authentication:** All protected endpoints require Clerk JWT token in the `Authorization` header.

---

## Table of Contents

1. [Authentication](#authentication)
2. [API Endpoints](#api-endpoints)
   - [Health Check](#health-check)
   - [Upload Management](#upload-management)
   - [Transaction Operations](#transaction-operations)
   - [Dashboard & Analytics](#dashboard--analytics)
   - [Categorization Rules](#categorization-rules)
3. [Request/Response Formats](#requestresponse-formats)
4. [Error Handling](#error-handling)
5. [Frontend Integration](#frontend-integration)
6. [Testing the API](#testing-the-api)

---

## Authentication

### Clerk JWT Authentication

All protected endpoints use Clerk JWT tokens for authentication.

**Header Format:**
```http
Authorization: Bearer <clerk_jwt_token>
```

**How it works:**
1. User signs in via Clerk on the frontend
2. Clerk issues a JWT token
3. Frontend includes token in `Authorization` header
4. Backend middleware validates token and extracts `user_id`
5. User ID is stored in request context: `c.Locals("user_id")`

**Middleware:** `internal/middleware/auth.go` - `ClerkAuth()`

**Example:**
```bash
curl -X GET http://localhost:8080/v1/transactions \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## API Endpoints

### Health Check

#### `GET /health`
Check if the API server is running.

**Authentication:** None

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

---

### Upload Management

#### `POST /v1/upload/presigned-url`
Generate a presigned S3 URL for direct file upload.

**Authentication:** Required

**Request Body:**
```json
{
  "filename": "hdfc_statement.csv",
  "file_type": "text/csv"
}
```

**Supported File Types:**
- `text/csv` - CSV files
- `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` - XLSX files
- `application/pdf` - PDF files

**Response:**
```json
{
  "upload_url": "http://localhost:4566/cashlens-uploads/user123/1234567890_hdfc_statement.csv?X-Amz-Algorithm=...",
  "file_key": "user123/1234567890_hdfc_statement.csv",
  "expires_in": 300
}
```

**Implementation:** `internal/handlers/upload.go` - `GeneratePresignedURL()`

**Flow:**
1. Validate file type (CSV, XLSX, or PDF only)
2. Generate unique file key: `{user_id}/{timestamp}_{filename}`
3. Create presigned PUT URL valid for 5 minutes
4. Return URL to frontend for direct upload

**Example:**
```bash
curl -X POST http://localhost:8080/v1/upload/presigned-url \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filename":"statement.csv","file_type":"text/csv"}'
```

---

#### `POST /v1/upload/process`
Process an uploaded file (CSV/XLSX/PDF) and import transactions.

**Authentication:** Required

**Request Body:**
```json
{
  "file_key": "user123/1234567890_hdfc_statement.csv",
  "filename": "hdfc_statement.csv"
}
```

**Response:**
```json
{
  "upload_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_rows": 156,
  "categorized_rows": 142,
  "uncategorized_rows": 14,
  "accuracy_percent": 91.03,
  "status": "completed",
  "message": "File processed successfully"
}
```

**Implementation:** `internal/handlers/upload.go` - `ProcessUpload()`

**Processing Flow:**
1. Download file from S3 using `file_key`
2. Validate file using magic bytes detection
3. Detect file format (CSV/XLSX/PDF)
4. Parse transactions using appropriate parser:
   - **CSV:** `internal/services/parser.go`
   - **XLSX:** `internal/services/xlsx_parser.go`
   - **PDF:** Python microservice via HTTP
5. Categorize transactions using rule engine
6. Insert transactions into database
7. Create upload history record
8. Return summary statistics

**Error Responses:**
- `400` - Invalid file format, missing file_key
- `404` - File not found in S3
- `500` - Parsing error, database error

**Example:**
```bash
curl -X POST http://localhost:8080/v1/upload/process \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"file_key":"user123/1234567890_statement.csv","filename":"statement.csv"}'
```

---

### Transaction Operations

#### `GET /v1/transactions`
Retrieve user's transactions with filtering and pagination.

**Authentication:** Required

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `limit` (int, default: 50, max: 100) - Items per page
- `category` (string, optional) - Filter by category
- `start_date` (string, optional) - ISO 8601 date (YYYY-MM-DD)
- `end_date` (string, optional) - ISO 8601 date (YYYY-MM-DD)
- `search` (string, optional) - Search in description
- `is_reviewed` (bool, optional) - Filter by review status

**Response:**
```json
{
  "transactions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "user_2abc123",
      "txn_date": "2024-01-15T00:00:00Z",
      "description": "AWS SERVICES INDIA",
      "amount": -1250.50,
      "txn_type": "debit",
      "category": "Cloud & Hosting",
      "is_reviewed": true,
      "raw_data": {...},
      "created_at": "2024-01-16T10:30:00Z",
      "updated_at": "2024-01-16T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 156,
    "total_pages": 4
  }
}
```

**Implementation:** `internal/handlers/transactions.go` - `GetTransactions()`

**Example:**
```bash
# Get all transactions
curl "http://localhost:8080/v1/transactions" \
  -H "Authorization: Bearer $TOKEN"

# Filter by category and date range
curl "http://localhost:8080/v1/transactions?category=Cloud%20%26%20Hosting&start_date=2024-01-01&end_date=2024-01-31" \
  -H "Authorization: Bearer $TOKEN"

# Search for AWS transactions
curl "http://localhost:8080/v1/transactions?search=AWS&page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

---

#### `GET /v1/transactions/:id`
Get a specific transaction by ID.

**Authentication:** Required

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user_2abc123",
  "txn_date": "2024-01-15T00:00:00Z",
  "description": "AWS SERVICES INDIA",
  "amount": -1250.50,
  "txn_type": "debit",
  "category": "Cloud & Hosting",
  "is_reviewed": true,
  "raw_data": {...},
  "created_at": "2024-01-16T10:30:00Z",
  "updated_at": "2024-01-16T10:30:00Z"
}
```

**Example:**
```bash
curl "http://localhost:8080/v1/transactions/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $TOKEN"
```

---

#### `PUT /v1/transactions/:id`
Update a transaction (typically for manual categorization).

**Authentication:** Required

**Request Body:**
```json
{
  "category": "Software & SaaS",
  "is_reviewed": true
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "category": "Software & SaaS",
  "is_reviewed": true,
  "updated_at": "2024-01-16T11:00:00Z"
}
```

**Implementation:** `internal/handlers/transactions.go` - `UpdateTransaction()`

**Example:**
```bash
curl -X PUT "http://localhost:8080/v1/transactions/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"category":"Software & SaaS","is_reviewed":true}'
```

---

#### `DELETE /v1/transactions/:id`
Delete a transaction.

**Authentication:** Required

**Response:**
```json
{
  "message": "Transaction deleted successfully"
}
```

**Example:**
```bash
curl -X DELETE "http://localhost:8080/v1/transactions/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $TOKEN"
```

---

### Dashboard & Analytics

#### `GET /v1/summary`
Get dashboard KPIs and financial summary.

**Authentication:** Required

**Query Parameters:**
- `start_date` (string, optional) - ISO 8601 date
- `end_date` (string, optional) - ISO 8601 date
- `period` (string, optional) - "month", "quarter", "year"

**Response:**
```json
{
  "total_income": 250000.00,
  "total_expenses": 185432.50,
  "net_balance": 64567.50,
  "transaction_count": 156,
  "categorized_count": 142,
  "uncategorized_count": 14,
  "accuracy_percent": 91.03,
  "category_breakdown": [
    {
      "category": "Cloud & Hosting",
      "amount": 45000.00,
      "percentage": 24.27,
      "transaction_count": 12
    },
    {
      "category": "Salaries",
      "amount": 80000.00,
      "percentage": 43.14,
      "transaction_count": 4
    }
  ],
  "monthly_trend": [
    {
      "month": "2024-01",
      "income": 85000.00,
      "expenses": 62000.00,
      "net": 23000.00
    }
  ],
  "top_expenses": [
    {
      "description": "AWS SERVICES INDIA",
      "amount": 12500.00,
      "date": "2024-01-15",
      "category": "Cloud & Hosting"
    }
  ]
}
```

**Implementation:** `internal/handlers/summary.go` - `GetSummary()`

**Example:**
```bash
# Get current month summary
curl "http://localhost:8080/v1/summary?period=month" \
  -H "Authorization: Bearer $TOKEN"

# Get custom date range
curl "http://localhost:8080/v1/summary?start_date=2024-01-01&end_date=2024-03-31" \
  -H "Authorization: Bearer $TOKEN"
```

---

### Categorization Rules

#### `GET /v1/rules/global`
Get all active global categorization rules.

**Authentication:** Required

**Response:**
```json
{
  "rules": [
    {
      "id": "uuid",
      "keyword": "aws",
      "category": "Cloud & Hosting",
      "priority": 100,
      "is_active": true
    },
    {
      "id": "uuid",
      "keyword": "salary",
      "category": "Salaries",
      "priority": 90,
      "is_active": true
    }
  ]
}
```

**Example:**
```bash
curl "http://localhost:8080/v1/rules/global" \
  -H "Authorization: Bearer $TOKEN"
```

---

#### `GET /v1/rules/user`
Get user-specific categorization rules (higher priority than global).

**Authentication:** Required

**Response:**
```json
{
  "rules": [
    {
      "id": "uuid",
      "user_id": "user_2abc123",
      "keyword": "digitalocean",
      "category": "Cloud & Hosting",
      "priority": 100,
      "is_active": true
    }
  ]
}
```

**Example:**
```bash
curl "http://localhost:8080/v1/rules/user" \
  -H "Authorization: Bearer $TOKEN"
```

---

#### `POST /v1/rules/user`
Create a new user-specific categorization rule.

**Authentication:** Required

**Request Body:**
```json
{
  "keyword": "digitalocean",
  "category": "Cloud & Hosting",
  "priority": 100
}
```

**Response:**
```json
{
  "id": "uuid",
  "user_id": "user_2abc123",
  "keyword": "digitalocean",
  "category": "Cloud & Hosting",
  "priority": 100,
  "is_active": true,
  "created_at": "2024-01-16T10:30:00Z"
}
```

**Example:**
```bash
curl -X POST "http://localhost:8080/v1/rules/user" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"keyword":"digitalocean","category":"Cloud & Hosting","priority":100}'
```

---

#### `DELETE /v1/rules/user/:id`
Delete a user-specific categorization rule.

**Authentication:** Required

**Response:**
```json
{
  "message": "Rule deleted successfully"
}
```

**Example:**
```bash
curl -X DELETE "http://localhost:8080/v1/rules/user/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Request/Response Formats

### Standard Success Response
```json
{
  "data": {...},
  "message": "Success message"
}
```

### Standard Error Response
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {...}
}
```

### Pagination Format
```json
{
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 156,
    "total_pages": 4
  }
}
```

---

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing/invalid token |
| 403 | Forbidden - Access denied |
| 404 | Not Found - Resource doesn't exist |
| 422 | Unprocessable Entity - Validation error |
| 500 | Internal Server Error |

### Error Response Examples

**400 Bad Request:**
```json
{
  "error": "Invalid file type. Supported formats: CSV, XLSX, PDF",
  "code": "INVALID_FILE_TYPE"
}
```

**401 Unauthorized:**
```json
{
  "error": "Missing or invalid authorization token",
  "code": "UNAUTHORIZED"
}
```

**404 Not Found:**
```json
{
  "error": "Transaction not found",
  "code": "NOT_FOUND"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Failed to process file",
  "code": "INTERNAL_ERROR",
  "details": "Database connection failed"
}
```

---

## Frontend Integration

### Complete Upload Flow

**Step 1: Get Presigned URL**
```typescript
const response = await fetch('http://localhost:8080/v1/upload/presigned-url', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${clerkToken}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    filename: file.name,
    file_type: file.type,
  }),
})

const { upload_url, file_key } = await response.json()
```

**Step 2: Upload File to S3**
```typescript
await fetch(upload_url, {
  method: 'PUT',
  headers: {
    'Content-Type': file.type,
  },
  body: file,
})
```

**Step 3: Process File**
```typescript
const response = await fetch('http://localhost:8080/v1/upload/process', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${clerkToken}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    file_key: file_key,
    filename: file.name,
  }),
})

const result = await response.json()
console.log(`Processed ${result.total_rows} transactions`)
console.log(`Accuracy: ${result.accuracy_percent}%`)
```

### API Client Wrapper (Next.js)

**File:** `cashlens-web/lib/api.ts`

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
    throw new Error(error.error || 'API request failed')
  }

  return response.json()
}

// Usage
const transactions = await apiClient<TransactionsResponse>('/v1/transactions')
```

---

## Testing the API

### Manual Testing with cURL

**1. Health Check:**
```bash
curl http://localhost:8080/health
```

**2. Get Transactions (with auth):**
```bash
# Set your Clerk token
export TOKEN="your_clerk_jwt_token"

curl "http://localhost:8080/v1/transactions?limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

**3. Upload Flow:**
```bash
# Step 1: Get presigned URL
curl -X POST http://localhost:8080/v1/upload/presigned-url \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filename":"test.csv","file_type":"text/csv"}' \
  | jq -r '.upload_url' > url.txt

# Step 2: Upload file
curl -X PUT "$(cat url.txt)" \
  --upload-file testdata/hdfc_sample.csv \
  -H "Content-Type: text/csv"

# Step 3: Process file
curl -X POST http://localhost:8080/v1/upload/process \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"file_key":"user123/1234567890_test.csv","filename":"test.csv"}'
```

### Using Postman

1. **Import Collection**: Create a new collection "Cashlens API"
2. **Set Environment Variables**:
   - `base_url`: http://localhost:8080/v1
   - `token`: Your Clerk JWT token
3. **Add Authorization Header**: `Authorization: Bearer {{token}}`

### Integration Testing

See [TESTING.md](TESTING.md) for automated testing commands.

---

## Rate Limiting

Currently not implemented. Future enhancement will add rate limiting per user:
- 100 requests per minute per user
- 1000 requests per hour per user

---

## CORS Configuration

**Allowed Origins:**
- `http://localhost:3000` (Next.js dev)
- Production frontend domain (configured via env)

**Allowed Methods:** GET, POST, PUT, DELETE, OPTIONS

**Allowed Headers:** Authorization, Content-Type

**Implementation:** `internal/middleware/cors.go`

---

## API Versioning

Current version: **v1**

All endpoints are prefixed with `/v1/` to support future API versioning without breaking changes.

---

## Security Best Practices

1. **Always use HTTPS in production**
2. **Validate JWT tokens on every request**
3. **Sanitize user input** (SQL injection prevention)
4. **Rate limit API endpoints**
5. **Use presigned URLs** for S3 uploads (no direct credentials exposure)
6. **Validate file types** with magic bytes (not just extensions)
7. **Log security events** (failed auth, suspicious activity)

---

## Support

For API issues or questions:
- GitHub Issues: https://github.com/yourusername/cashlens/issues
- Email: support@cashlens.com

---

## Changelog

### v1.0.0 (2024-01-16)
- Initial API release
- CSV/XLSX/PDF upload and parsing
- Transaction CRUD operations
- Dashboard analytics
- Rule-based auto-categorization
- Clerk authentication integration
