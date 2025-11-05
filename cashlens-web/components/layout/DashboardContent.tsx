"use client"

import { cn } from "@/lib/utils"
import { useSidebar } from "./SidebarContext"
import { Header } from "./Header"

export function DashboardContent({ children }: { children: React.ReactNode }) {
  const { collapsed } = useSidebar()

  return (
    <div
      className={cn(
        "transition-all duration-300",
        collapsed ? "pl-16" : "pl-64"
      )}
    >
      {/* Header */}
      <Header />

      {/* Page content */}
      <main className="p-6">
        <div className="mx-auto max-w-7xl">{children}</div>
      </main>
    </div>
  )
}
