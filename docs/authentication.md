# Authentication System Documentation

## Overview

Cashlens uses **Clerk** for authentication, providing a secure, scalable, and feature-rich authentication system with minimal implementation overhead. This document covers the complete authentication architecture, API endpoints, and integration details.

## Table of Contents

1. [Architecture](#architecture)
2. [Authentication Flow](#authentication-flow)
3. [API Endpoints](#api-endpoints)
4. [Frontend Integration](#frontend-integration)
5. [Backend Integration](#backend-integration)
6. [Database Schema](#database-schema)
7. [Security Considerations](#security-considerations)
8. [Testing](#testing)
9. [Troubleshooting](#troubleshooting)

---

## Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client (Browser)                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │          Next.js 15 Frontend (Port 3000)                 │   │
│  │  • React 18 with App Router                              │   │
│  │  • Clerk Provider & Components                           │   │
│  │  • Protected Routes (Middleware)                         │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                          │
                          │ HTTPS/JWT
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Clerk Service (SaaS)                        │
│  • User Management                                               │
│  • JWT Token Generation & Validation                            │
│  • OAuth Providers (Google, GitHub, etc.)                       │
│  • Webhooks for User Events                                     │
└─────────────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┴─────────────────┐
        │                                   │
        ▼                                   ▼
┌──────────────────┐              ┌──────────────────┐
│  Go Backend API  │              │   Frontend       │
│   (Port 8080)    │              │   Webhooks       │
│                  │              │                  │
│ • JWT Validation │              │ • User Sync      │
│ • Protected      │              │ • User Events    │
│   Endpoints      │              │                  │
└──────────────────┘              └──────────────────┘
        │                                   │
        └─────────────────┬─────────────────┘
                          ▼
              ┌─────────────────────┐
              │   PostgreSQL DB     │
              │  • users table      │
              │  • transactions     │
              └─────────────────────┘
```

### Technology Stack

**Frontend:**
- **Framework**: Next.js 15.2.3 (App Router)
- **Auth Library**: `@clerk/nextjs` v2.x
- **Webhook Verification**: `svix`

**Backend:**
- **Language**: Go 1.25
- **Framework**: Fiber v3
- **Auth Library**: `github.com/clerk/clerk-sdk-go/v2`
- **JWT Validation**: Clerk JWT SDK

**Database:**
- **PostgreSQL** 16 for user data storage

---

## Authentication Flow

### 1. User Registration Flow

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Clerk
    participant Webhook
    participant Backend
    participant Database

    User->>Frontend: Navigate to /sign-up
    Frontend->>Clerk: Display Clerk SignUp component
    User->>Clerk: Submit registration (email/password or OAuth)
    Clerk->>Clerk: Create user account
    Clerk->>Frontend: Return JWT token
    Frontend->>Frontend: Store token, update state

    Clerk->>Webhook: POST /api/webhooks/clerk (user.created event)
    Webhook->>Backend: POST /internal/users (sync user)
    Backend->>Database: INSERT INTO users
    Database-->>Backend: User created
    Backend-->>Webhook: 200 OK
    Webhook-->>Clerk: 200 OK

    Frontend->>Frontend: Redirect to /dashboard
```

### 2. User Login Flow

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Clerk
    participant Backend

    User->>Frontend: Navigate to /sign-in
    Frontend->>Clerk: Display Clerk SignIn component
    User->>Clerk: Submit credentials
    Clerk->>Clerk: Validate credentials
    Clerk->>Frontend: Return JWT token
    Frontend->>Frontend: Store token, redirect to /dashboard
```

### 3. Protected API Request Flow

```mermaid
sequenceDiagram
    participant Frontend
    participant Middleware
    participant Backend
    participant Database

    Frontend->>Backend: GET /v1/me (with JWT in Authorization header)
    Backend->>Middleware: Extract & validate JWT
    Middleware->>Clerk SDK: Verify JWT signature
    Clerk SDK-->>Middleware: Valid (user_id extracted)
    Middleware->>Backend: Continue with user_id in context
    Backend->>Database: Query user data
    Database-->>Backend: User data
    Backend-->>Frontend: 200 OK with user data
```

---

## API Endpoints

### Public Endpoints

#### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "service": "cashlens-api"
}
```

#### Ping

```http
GET /v1/ping
```

**Response:**
```json
{
  "message": "pong"
}
```

### Protected Endpoints

All protected endpoints require a valid Clerk JWT token in the `Authorization` header.

#### Get Current User

```http
GET /v1/me
Authorization: Bearer <clerk_jwt_token>
```

**Response (200 OK):**
```json
{
  "user_id": "user_2abc123...",
  "message": "This is a protected route"
}
```

**Error Responses:**

**401 Unauthorized - Missing Token:**
```json
{
  "error": "Missing authorization token"
}
```

**401 Unauthorized - Invalid Format:**
```json
{
  "error": "Invalid authorization header format"
}
```

**401 Unauthorized - Invalid Token:**
```json
{
  "error": "Invalid or expired token",
  "details": "token signature is invalid"
}
```

---

## Frontend Integration

### 1. Root Layout Setup

**File:** `cashlens-web/app/layout.tsx`

```typescript
import { ClerkProvider } from "@clerk/nextjs"
import "./globals.css"

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <ClerkProvider>
      <html lang="en">
        <body>{children}</body>
      </html>
    </ClerkProvider>
  )
}
```

### 2. Middleware Configuration

**File:** `cashlens-web/middleware.ts`

```typescript
import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server"

const isPublicRoute = createRouteMatcher([
  "/",
  "/sign-in(.*)",
  "/sign-up(.*)",
  "/api/webhooks(.*)",
])

export default clerkMiddleware(async (auth, request) => {
  if (!isPublicRoute(request)) {
    await auth.protect()
  }
}, {
  signInUrl: "/sign-in",
  signUpUrl: "/sign-up",
})

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
  ],
}
```

### 3. Sign-In Page

**File:** `cashlens-web/app/(auth)/sign-in/[[...sign-in]]/page.tsx`

```typescript
import { SignIn } from "@clerk/nextjs"

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-gray-900">Cashlens</h1>
          <p className="mt-2 text-sm text-gray-600">
            Sign in to your account
          </p>
        </div>
        <SignIn
          appearance={{
            elements: {
              rootBox: "mx-auto",
              card: "shadow-lg",
              formButtonPrimary: "bg-blue-600 hover:bg-blue-700",
            },
          }}
        />
      </div>
    </div>
  )
}
```

### 4. Protected Dashboard

**File:** `cashlens-web/app/(dashboard)/layout.tsx`

```typescript
import { auth } from "@clerk/nextjs/server"
import { redirect } from "next/navigation"
import { UserButton } from "@clerk/nextjs"

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { userId } = await auth()

  if (!userId) {
    redirect("/sign-in")
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="border-b bg-white">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex h-16 items-center justify-between">
            <h1 className="text-xl font-bold text-gray-900">Cashlens</h1>
            <UserButton afterSignOutUrl="/sign-in" />
          </div>
        </div>
      </nav>
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        {children}
      </main>
    </div>
  )
}
```

### 5. Dashboard Page (Client Component)

**File:** `cashlens-web/app/(dashboard)/dashboard/page.tsx`

```typescript
"use client"

import { useUser } from "@clerk/nextjs"

export default function DashboardPage() {
  const { user, isLoaded } = useUser()

  if (!isLoaded) {
    return <div className="text-gray-600">Loading...</div>
  }

  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900">
        Welcome, {user?.firstName || user?.emailAddresses?.[0]?.emailAddress}!
      </h1>
      <p className="mt-4 text-gray-600">
        This is your dashboard.
      </p>
    </div>
  )
}
```

### 6. Webhook Handler

**File:** `cashlens-web/app/api/webhooks/clerk/route.ts`

```typescript
import { Webhook } from "svix"
import { headers } from "next/headers"
import { WebhookEvent } from "@clerk/nextjs/server"

export async function POST(req: Request) {
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET

  // Verify webhook signature
  const headerPayload = await headers()
  const svix_id = headerPayload.get("svix-id")
  const svix_timestamp = headerPayload.get("svix-timestamp")
  const svix_signature = headerPayload.get("svix-signature")

  if (!svix_id || !svix_timestamp || !svix_signature) {
    return new Response("Error: Missing svix headers", { status: 400 })
  }

  const payload = await req.json()
  const body = JSON.stringify(payload)

  const wh = new Webhook(WEBHOOK_SECRET!)
  let evt: WebhookEvent

  try {
    evt = wh.verify(body, {
      "svix-id": svix_id,
      "svix-timestamp": svix_timestamp,
      "svix-signature": svix_signature,
    }) as WebhookEvent
  } catch (err) {
    return new Response("Error: Verification failed", { status: 400 })
  }

  // Handle user.created event
  if (evt.type === "user.created") {
    const { id, email_addresses, first_name, last_name } = evt.data

    // Sync user to backend database
    await fetch(`${process.env.NEXT_PUBLIC_API_URL}/internal/users`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        clerk_user_id: id,
        email: email_addresses[0]?.email_address,
        full_name: `${first_name || ""} ${last_name || ""}`.trim(),
      }),
    })
  }

  return new Response("", { status: 200 })
}
```

---

## Backend Integration

### 1. Clerk Auth Middleware

**File:** `cashlens-api/internal/middleware/clerk_auth.go`

```go
package middleware

import (
	"context"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gofiber/fiber/v3"
)

// ClerkAuth middleware validates Clerk JWT tokens
func ClerkAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization token",
			})
		}

		// Remove "Bearer " prefix
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Verify token with Clerk
		claims, err := jwt.Verify(context.Background(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
				"details": err.Error(),
			})
		}

		// Store user ID in context for use in handlers
		c.Locals("user_id", claims.Subject)
		c.Locals("clerk_user_id", claims.Subject)

		return c.Next()
	}
}
```

### 2. CORS Middleware

**File:** `cashlens-api/internal/middleware/cors.go`

```go
package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// CORS returns a configured CORS middleware
func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowCredentials: true,
	})
}
```

### 3. Main API Setup

**File:** `cashlens-api/cmd/api/main.go`

```go
package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/ashmitsharp/cashlens-api/internal/middleware"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "cashlens API v1.0",
	})

	// Apply global middleware
	app.Use(middleware.CORS())

	// Health check endpoint (public)
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "cashlens-api",
		})
	})

	// API v1 routes
	v1 := app.Group("/v1")

	// Public routes
	v1.Get("/ping", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong"})
	})

	// Protected routes (require authentication)
	protected := v1.Group("", middleware.ClerkAuth())

	// Test protected endpoint
	protected.Get("/me", func(c fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		return c.JSON(fiber.Map{
			"user_id": userID,
			"message": "This is a protected route",
		})
	})

	log.Println("Starting cashlens API on :8080")
	log.Fatal(app.Listen(":8080"))
}
```

---

## Database Schema

### Users Table

**Migration:** `cashlens-api/internal/database/migrations/001_create_users_table.sql`

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clerk_user_id TEXT UNIQUE NOT NULL,
    email TEXT NOT NULL,
    full_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_clerk_user_id ON users(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

**Table Structure:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PRIMARY KEY | Internal user ID |
| `clerk_user_id` | TEXT | UNIQUE, NOT NULL | Clerk user ID (for JWT validation) |
| `email` | TEXT | NOT NULL | User email address |
| `full_name` | TEXT | NULL | User's full name |
| `created_at` | TIMESTAMPTZ | NOT NULL | Account creation timestamp |
| `updated_at` | TIMESTAMPTZ | NOT NULL | Last update timestamp |

**Indexes:**
- `idx_users_clerk_user_id`: Fast lookups by Clerk user ID
- `idx_users_email`: Email queries for user management

---

## Security Considerations

### 1. JWT Token Security

- **Tokens are short-lived**: Default 1-hour expiration
- **Signed by Clerk**: Uses RS256 algorithm with Clerk's private key
- **Verified on every request**: Backend validates signature before allowing access
- **Stored securely**: Frontend stores in memory (not localStorage for security)

### 2. Password Security

- **Not stored locally**: Passwords handled entirely by Clerk
- **Bcrypt hashing**: Clerk uses industry-standard bcrypt
- **Password requirements**: Configurable in Clerk dashboard

### 3. HTTPS/TLS

- **Production requirement**: All traffic must use HTTPS
- **Clerk enforces HTTPS**: API calls to Clerk require TLS 1.2+

### 4. CORS Protection

- **Whitelisted origins**: Only `localhost:3000` allowed in development
- **Production**: Update to production domain only

### 5. Rate Limiting

- **Clerk handles**: Built-in rate limiting for auth endpoints
- **Backend**: To be implemented in future (Day 8)

---

## Testing

### Manual Testing

#### 1. Test User Registration

```bash
# Navigate to sign-up page
open http://localhost:3000/sign-up

# Register with:
# - Email: test@example.com
# - Password: TestPassword123!

# Verify redirect to /dashboard
```

#### 2. Test User Login

```bash
# Navigate to sign-in page
open http://localhost:3000/sign-in

# Login with registered credentials

# Verify redirect to /dashboard
```

#### 3. Test Protected API Endpoint

```bash
# Get JWT token from browser (DevTools > Application > Cookies)
# Or use Clerk's test tokens

curl http://localhost:8080/v1/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected response:
# {"user_id":"user_...", "message":"This is a protected route"}
```

#### 4. Test Unauthorized Access

```bash
curl http://localhost:8080/v1/me

# Expected response (401):
# {"error":"Missing authorization token"}
```

### Automated Testing

**E2E Test (Playwright) - To be implemented:**

```typescript
test("user can sign up and access dashboard", async ({ page }) => {
  await page.goto("/sign-up")

  await page.fill('input[name="emailAddress"]', "test@example.com")
  await page.fill('input[name="password"]', "TestPass123!")
  await page.click('button:has-text("Continue")')

  // Should redirect to dashboard
  await expect(page).toHaveURL("/dashboard")
  await expect(page.locator("h1")).toContainText("Welcome")
})
```

---

## Troubleshooting

### Issue: 404 on /auth/sign-in

**Symptom:** After OAuth callback, seeing 404 errors for `/auth/sign-in`

**Solution:** Ensure middleware has `signInUrl` and `signUpUrl` configured:

```typescript
export default clerkMiddleware(async (auth, request) => {
  if (!isPublicRoute(request)) {
    await auth.protect()
  }
}, {
  signInUrl: "/sign-in",  // Add this
  signUpUrl: "/sign-up",  // Add this
})
```

### Issue: Hydration Mismatch

**Symptom:** Console warning about hydration mismatch on dashboard

**Solution:** Use client component with `useUser()` hook instead of server-side `currentUser()`:

```typescript
"use client"

import { useUser } from "@clerk/nextjs"

export default function DashboardPage() {
  const { user, isLoaded } = useUser()

  if (!isLoaded) {
    return <div>Loading...</div>
  }

  // Render content
}
```

### Issue: JWT Verification Fails

**Symptom:** `401 Unauthorized` on protected endpoints

**Possible Causes:**
1. **Expired token**: JWT tokens expire after 1 hour
2. **Wrong secret key**: Check `CLERK_SECRET_KEY` in backend `.env`
3. **Invalid token format**: Ensure `Bearer ` prefix is correct

**Solution:**
```bash
# Verify environment variables
cat cashlens-api/.env | grep CLERK

# Test with fresh token (sign out and sign in again)
```

### Issue: Webhook Not Receiving Events

**Symptom:** Users created in Clerk but not synced to database

**Solution:**

1. **Verify webhook URL in Clerk dashboard**:
   - Development: `https://your-ngrok-url.ngrok.io/api/webhooks/clerk`
   - Production: `https://yourdomain.com/api/webhooks/clerk`

2. **Check webhook secret**:
   ```bash
   # In .env.local
   CLERK_WEBHOOK_SECRET=whsec_...
   ```

3. **Test webhook manually**:
   ```bash
   # Trigger test event from Clerk dashboard
   # Check Next.js logs for incoming requests
   ```

---

## Environment Variables Reference

### Frontend (.env.local)

```bash
# API URL
NEXT_PUBLIC_API_URL=http://localhost:8080/v1

# Clerk Authentication
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
CLERK_WEBHOOK_SECRET=whsec_...

# Clerk Routes
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL=/dashboard
NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL=/dashboard
```

### Backend (.env)

```bash
# Server
PORT=8080
ENVIRONMENT=development

# Database
DATABASE_URL=postgresql://postgres:dev123@localhost:5432/cashlens

# Clerk Authentication
CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
```

---

## Next Steps

After authentication is working:

1. **Day 2**: Implement CSV parser for bank statements
2. **Day 3**: Build file upload flow with S3
3. **Day 4**: Create categorization engine
4. **Day 8**: Add rate limiting and security hardening

---

## Support & Resources

- **Clerk Documentation**: https://clerk.com/docs
- **Next.js App Router**: https://nextjs.org/docs/app
- **Fiber v3 Documentation**: https://docs.gofiber.io

**Last Updated:** 2024-11-04
**Version:** 1.0.0
**Status:** ✅ Production Ready
