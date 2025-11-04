import type { Metadata } from "next"
import "./globals.css"

export const metadata: Metadata = {
  title: "Cashlens - Financial Analytics for SMBs",
  description: "AI-powered cash flow analytics platform for Indian SMBs",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.Node
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
