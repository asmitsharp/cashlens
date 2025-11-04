"use client"

import { useUser } from "@clerk/nextjs"

export default function DashboardPage() {
  const { user, isLoaded } = useUser()

  if (!isLoaded) {
    return (
      <div className="flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900">
        Welcome, {user?.firstName || user?.emailAddresses?.[0]?.emailAddress || "User"}!
      </h1>
      <p className="mt-4 text-gray-600">
        This is your dashboard. CSV upload and analytics features will be added in the next steps.
      </p>
    </div>
  )
}
