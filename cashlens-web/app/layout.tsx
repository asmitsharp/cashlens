import type { Metadata } from "next"
import { ClerkProvider } from "@clerk/nextjs"
import { Inter, Lora } from "next/font/google"
import "./globals.css"

// Inter font for all UI text
const inter = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
})

// Lora font for landing page headlines only
const lora = Lora({
  subsets: ["latin"],
  variable: "--font-serif",
  weight: ["600", "700"],
  display: "swap",
})

export const metadata: Metadata = {
  title: "Cashlens - Financial Analytics for SMBs",
  description: "AI-powered cash flow analytics platform for Indian SMBs",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <ClerkProvider>
      <html lang="en" className={`${inter.variable} ${lora.variable}`}>
        <body className={inter.className}>{children}</body>
      </html>
    </ClerkProvider>
  )
}
