# Local Webhook Development Guide

You do **not** need to create a new webhook endpoint and update your `.env` file every time you restart ngrok. Here are better ways to handle local webhooks.

## Option 1: The "Clerk CLI" Method (Recommended)

The easiest way to develop locally is to use the Clerk CLI, which forwards webhooks directly to your localhost without needing ngrok or public URLs.

1.  **Install Clerk CLI** (if not already installed):
    ```bash
    npm install -g @clerk/clerk-sdk-node
    # OR just run with npx
    ```
    *Actually, checking the latest docs, Clerk recommends using `svix` directly or their dashboard's "Test" feature, but for local dev tunneling:*

    Go to the **Clerk Dashboard > Webhooks**.
    Click "Add Endpoint" (or use your existing one).
    
    **Better yet:** Use the built-in testing tool in the Clerk Dashboard to "replay" events to your local ngrok URL if you just need to test one-offs.

## Option 2: The "Ngrok" Method (Optimized)

If you prefer ngrok, you can avoid changing the secret every time.

1.  **Start ngrok**:
    ```bash
    ngrok http 3000
    ```
2.  **Copy the URL** (e.g., `https://random-id.ngrok-free.app`).
3.  **Update Clerk**:
    *   Go to Clerk Dashboard > Webhooks.
    *   Click on your **existing** local development endpoint.
    *   Click **Edit** (pencil icon) next to the URL.
    *   Paste the new ngrok URL (e.g., `https://random-id.ngrok-free.app/api/webhooks/clerk`).
    *   **Save**.
4.  **Do NOT change the Secret**:
    *   The `Signing Secret` (whsec_...) does **not** change when you update the URL.
    *   You do **not** need to update your `.env` file.

## Option 3: Static Domain (Free Alternatives)

If you want a URL that stays the same (so you don't even have to update Clerk), try `localtunnel`.

1.  **Install**:
    ```bash
    npm install -g localtunnel
    ```
2.  **Run**:
    ```bash
    lt --port 3000 --subdomain my-cashlens-dev
    ```
3.  **Configure Clerk**:
    *   Set the webhook URL to `https://my-cashlens-dev.loca.lt/api/webhooks/clerk`.
    *   This URL will stay the same as long as no one else takes that subdomain.

## Summary
- **Stop creating new endpoints.** Just update the URL of the existing one.
- **Your `.env` secret stays the same.**
