# techspec.md - Auth Section (REVISED)

## Authentication Strategy - Phase 1 (MVP with Clerk)

### Why Clerk for MVP?

1. **Time-to-market:** 1-2 hours vs 3-4 days for custom auth
2. **Security:** Battle-tested, handles edge cases you'll miss
3. **Cost:** Free up to 10,000 MAU (Monthly Active Users)
4. **UX:** Pre-built components, multiple auth methods
5. **Indian market fit:** Email/password + Google OAuth + phone support

### Implementation (Next.js 15 + Clerk)

#### 1. Installation

```bash
npm install @clerk/nextjs
```

#### 2. Environment Setup

```bash
# .env.local
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL=/dashboard
NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL=/upload
```

#### 3. Root Layout Setup

```typescript
// app/layout.tsx
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

#### 4. Auth Pages (Pre-built UI)

```typescript
// app/(auth)/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from "@clerk/nextjs"

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <SignIn
        appearance={{
          elements: {
            rootBox: "mx-auto",
            card: "shadow-lg",
          },
        }}
      />
    </div>
  )
}

// app/(auth)/sign-up/[[...sign-up]]/page.tsx
import { SignUp } from "@clerk/nextjs"

export default function SignUpPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <SignUp />
    </div>
  )
}
```

#### 5. Protected Routes (Server Components)

```typescript
// app/(dashboard)/layout.tsx
import { auth } from "@clerk/nextjs/server"
import { redirect } from "next/navigation"

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { userId } = await auth()

  if (!userId) {
    redirect("/sign-in")
  }

  return <div className="dashboard-layout">{children}</div>
}
```

#### 6. Getting User Data

```typescript
// app/(dashboard)/page.tsx
import { currentUser } from "@clerk/nextjs/server"

export default async function DashboardPage() {
  const user = await currentUser()

  return (
    <div>
      <h1>
        Welcome, {user?.firstName || user?.emailAddresses[0].emailAddress}!
      </h1>
    </div>
  )
}
```

#### 7. API Route Protection

```typescript
// app/api/transactions/route.ts
import { auth } from "@clerk/nextjs/server"
import { NextResponse } from "next/server"

export async function GET(request: Request) {
  const { userId } = await auth()

  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  // Your API logic here
  // userId is the Clerk user ID (use as user_id in your DB)

  return NextResponse.json({ data: [] })
}
```

---

### Backend Integration (Go API + Clerk)

#### 1. Verify Clerk JWT in Go

```go
// internal/middleware/clerk_auth.go
package middleware

import (
    "github.com/clerk/clerk-sdk-go/v2"
    "github.com/clerk/clerk-sdk-go/v2/jwt"
    "github.com/gofiber/fiber/v3"
)

func ClerkAuth() fiber.Handler {
    return func(c fiber.Ctx) error {
        // Get token from Authorization header
        token := c.Get("Authorization")
        if token == "" {
            return c.Status(401).JSON(fiber.Map{
                "error": "Missing authorization token",
            })
        }

        // Remove "Bearer " prefix
        token = token[7:]

        // Verify with Clerk
        claims, err := jwt.Verify(c.Context(), &jwt.VerifyParams{
            Token: token,
        })
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid token",
            })
        }

        // Store user ID in context
        c.Locals("user_id", claims.Subject)

        return c.Next()
    }
}
```

#### 2. Install Clerk Go SDK

```bash
go get github.com/clerk/clerk-sdk-go/v2
```

#### 3. Use in Routes

```go
// cmd/api/main.go
package main

import (
    "github.com/gofiber/fiber/v3"
    "yourapp/internal/middleware"
    "yourapp/internal/handlers"
)

func main() {
    app := fiber.New()

    // Public routes
    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    // Protected routes
    api := app.Group("/v1", middleware.ClerkAuth())
    api.Post("/upload/process", handlers.ProcessUpload)
    api.Get("/transactions", handlers.GetTransactions)
    api.Get("/summary", handlers.GetSummary)

    app.Listen(":8080")
}
```

---

### Database Schema Changes

**CRITICAL:** Your `users` table simplifies to just store Clerk user IDs:

```sql
-- Simplified users table (Clerk handles auth data)
CREATE TABLE users (
    id UUID PRIMARY KEY, -- This is the Clerk user ID
    clerk_user_id TEXT UNIQUE NOT NULL, -- Redundant but explicit
    email TEXT NOT NULL,
    full_name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Webhook handler creates user on first login
```

**No more storing:**

- ❌ password_hash (Clerk handles)
- ❌ email_verified (Clerk handles)
- ❌ reset_tokens (Clerk handles)
- ❌ session_tokens (Clerk handles)

---

### Clerk Webhook Setup (Sync Users to Your DB)

```typescript
// app/api/webhooks/clerk/route.ts
import { Webhook } from "svix"
import { headers } from "next/headers"
import { WebhookEvent } from "@clerk/nextjs/server"

export async function POST(req: Request) {
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET

  const headerPayload = await headers()
  const svix_id = headerPayload.get("svix-id")
  const svix_timestamp = headerPayload.get("svix-timestamp")
  const svix_signature = headerPayload.get("svix-signature")

  const body = await req.text()

  const wh = new Webhook(WEBHOOK_SECRET)
  const evt = wh.verify(body, {
    "svix-id": svix_id!,
    "svix-timestamp": svix_timestamp!,
    "svix-signature": svix_signature!,
  }) as WebhookEvent

  if (evt.type === "user.created") {
    // Create user in your database
    await fetch("http://localhost:8080/v1/internal/users", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        id: evt.data.id,
        email: evt.data.email_addresses[0].email_address,
        full_name: `${evt.data.first_name} ${evt.data.last_name}`,
      }),
    })
  }

  return new Response("", { status: 200 })
}
```

---

### Cost Projection (First 2 Years)

| Timeframe   | Users (MAU) | Clerk Cost  | Your Revenue | Cost % |
| ----------- | ----------- | ----------- | ------------ | ------ |
| Month 1-3   | 10-100      | **$0**      | $0           | 0%     |
| Month 4-6   | 200-500     | **$0**      | $0           | 0%     |
| Month 7-9   | 1,000-3,000 | **$0**      | ~$500/mo     | 0%     |
| Month 10-12 | 5,000-8,000 | **$0**      | ~$2,000/mo   | 0%     |
| Year 2 Q1   | 12,000      | **$25/mo**  | ~$5,000/mo   | 0.5%   |
| Year 2 Q2   | 20,000      | **$99/mo**  | ~$10,000/mo  | 1%     |
| Year 2 Q3   | 35,000      | **$199/mo** | ~$20,000/mo  | 1%     |

**You won't pay Clerk for at least 9-12 months.**

---

### Migration Strategy (Future: v2.0)

When Clerk costs become significant (50k+ users, ~$500/month):

#### Option A: Stay with Clerk (Recommended if revenue is strong)

- They handle security updates
- Enterprise SSO support
- SAML for corporate customers
- **Your time >> $500/month**

#### Option B: Migrate to Auth.js (NextAuth)

```bash
# Clerk provides export API
GET /v1/users → Export all users

# Implement Auth.js
npm install next-auth
# Gradually migrate users (support both systems for 3 months)
# Send password reset emails to all users
```

#### Option C: Custom Auth (Last Resort)

- Only if you need very specific workflows
- Budget 2-3 weeks of dev time
- Hire a security consultant for audit

---

### Testing Strategy with Clerk

```typescript
// tests/e2e/auth.spec.ts
import { test, expect } from "@playwright/test"

test("user can sign up and access dashboard", async ({ page }) => {
  await page.goto("/sign-up")

  // Clerk's test accounts work in dev mode
  await page.fill('input[name="emailAddress"]', "test@example.com")
  await page.fill('input[name="password"]', "TestPass123!")
  await page.click('button:has-text("Continue")')

  // Verify email step (skip in dev with test mode)

  // Should redirect to dashboard
  await expect(page).toHaveURL("/dashboard")
  await expect(page.locator("h1")).toContainText("Welcome")
})
```

---

### Clerk Dashboard Configuration (5 minutes)

1. **Sign up at clerk.com**
2. **Create application** → Choose "Next.js"
3. **Enable auth methods:**
   - ✅ Email/Password
   - ✅ Google OAuth
   - ✅ Magic Links (optional, but great UX)
4. **Copy API keys** → Add to `.env.local`
5. **Configure branding** (your logo, colors)
6. **Set up webhook** (for user sync)

**Done.** You have production-ready auth.

---

### Updated Project Structure

```
cashlens-web/
├── app/
│   ├── (auth)/           # Public auth pages
│   │   ├── sign-in/[[...sign-in]]/page.tsx
│   │   └── sign-up/[[...sign-up]]/page.tsx
│   ├── (dashboard)/      # Protected pages
│   │   ├── layout.tsx    # Auth check here
│   │   ├── page.tsx
│   │   ├── upload/
│   │   └── review/
│   ├── api/
│   │   └── webhooks/
│   │       └── clerk/route.ts
│   └── layout.tsx        # ClerkProvider
├── middleware.ts         # Clerk middleware (auto-routing)
└── .env.local
```

---

## Summary: Why This is the Right Choice

**For a solo founder:**

1. ⏱️ **2 hours** vs 4 days → Ship faster
2. 🔒 **Bank-grade security** → Don't mess up financial data
3. 💰 **Free for 12+ months** → No cost during MVP
4. 🚀 **Better UX** → Multiple login options
5. 🧘 **Peace of mind** → Focus on core product

**Save auth implementation for v2.0 when you have:**

- Revenue to justify dev time
- Team to maintain security
- Clear requirements (enterprise SSO? SAML?)

---

## Next Steps (Immediate)

1. **Sign up for Clerk** (5 min) → clerk.com
2. **Replace auth section in TechSpec.md** with this
3. **Update Plan.md Day 1** to use Clerk instead of custom auth
4. **Ship MVP 3 days faster** 🚀

**Questions?** This is the battle-tested path for indie founders. You can always migrate later when it makes financial sense.

---

## Specialized Development Agents for Technical Implementation

### Agent Usage for Key Technical Components

**Authentication (Clerk Integration):**
- Use `senior-engineer` to validate Clerk vs custom auth decision
- Use `code-reviewer` to audit JWT validation middleware
- Use `golang-pro` for idiomatic Clerk SDK integration in Go

**Database Schema Design:**
- Use `database-architect` for schema design and indexing strategy
- Example: "Design optimal schema for transactions table supporting 1M+ records with efficient filtering by user_id, category, date range, and is_reviewed status"
- Use for migration planning and performance optimization

**Backend API Development:**
- Use `backend-development:backend-architect` for API endpoint design
- Use `backend-development:tdd-orchestrator` for test-first implementation
- Use `golang-pro` for Go-specific optimizations and best practices
- Use `code-reviewer` for security and performance validation

**Frontend Development:**
- Use `frontend-developer` for all React/Next.js components
- Example: "Build responsive dashboard with KPI cards and charts using shadcn/ui"
- Use for accessibility, performance, and responsive design

**Future: Payment Integration:**
- Use `payment-processing:payment-integration` when implementing subscription billing
- Example: "Implement Stripe subscription with webhook handling for premium tier"

### Recommended Agent Workflow

For any new feature implementation:

1. **Design Phase**: Use `backend-architect` or `senior-engineer` for architecture
2. **Schema Phase**: Use `database-architect` for database changes
3. **Implementation Phase**: Use `tdd-orchestrator` + `golang-pro` or `frontend-developer`
4. **Review Phase**: Use `code-reviewer` for security, performance, and quality
5. **Documentation Phase**: Use `docs-architect` or `tutorial-engineer`

This ensures consistent quality and adherence to best practices throughout development.
