import { Webhook } from "svix"
import { headers } from "next/headers"
import { WebhookEvent } from "@clerk/nextjs/server"

export async function POST(req: Request) {
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET

  if (!WEBHOOK_SECRET) {
    throw new Error("CLERK_WEBHOOK_SECRET is not set")
  }

  // Get the headers
  const headerPayload = await headers()
  const svix_id = headerPayload.get("svix-id")
  const svix_timestamp = headerPayload.get("svix-timestamp")
  const svix_signature = headerPayload.get("svix-signature")

  // If there are no headers, error out
  if (!svix_id || !svix_timestamp || !svix_signature) {
    return new Response("Error: Missing svix headers", {
      status: 400,
    })
  }

  // Get the body
  const payload = await req.json()
  const body = JSON.stringify(payload)

  // Create a new Svix instance with your webhook secret
  const wh = new Webhook(WEBHOOK_SECRET)

  let evt: WebhookEvent

  // Verify the webhook signature
  try {
    evt = wh.verify(body, {
      "svix-id": svix_id,
      "svix-timestamp": svix_timestamp,
      "svix-signature": svix_signature,
    }) as WebhookEvent
  } catch (err) {
    console.error("Error verifying webhook:", err)
    return new Response("Error: Verification failed", {
      status: 400,
    })
  }

  // Handle the webhook event
  const eventType = evt.type
  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1"

  if (eventType === "user.created") {
    const { id, email_addresses, first_name, last_name } = evt.data

    try {
      console.log(`Syncing new user to backend: ${API_URL}/internal/users`)
      
      // Call your backend API to create user in database
      const response = await fetch(
        `${API_URL}/internal/users`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            clerk_user_id: id,
            email: email_addresses[0]?.email_address,
            full_name: `${first_name || ""} ${last_name || ""}`.trim(),
          }),
        }
      )

      if (!response.ok) {
        const errorText = await response.text()
        console.error(`Failed to create user in database (${response.status}):`, errorText)
      } else {
        console.log("Successfully synced user to backend")
      }
    } catch (error) {
      console.error("Error creating user:", error)
    }
  }

  if (eventType === "user.updated") {
    const { id, email_addresses, first_name, last_name } = evt.data

    try {
      console.log(`Syncing user update to backend: ${API_URL}/internal/users/${id}`)

      // Call your backend API to update user in database
      const response = await fetch(
        `${API_URL}/internal/users/${id}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: email_addresses[0]?.email_address,
            full_name: `${first_name || ""} ${last_name || ""}`.trim(),
          }),
        }
      )

      if (!response.ok) {
        const errorText = await response.text()
        console.error(`Failed to update user in database (${response.status}):`, errorText)
      } else {
        console.log("Successfully synced user update to backend")
      }
    } catch (error) {
      console.error("Error updating user:", error)
    }
  }

  return new Response("", { status: 200 })
}
