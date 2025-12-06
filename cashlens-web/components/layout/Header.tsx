"use client"

import { UserButton } from "@clerk/nextjs"
import { Bell, Search } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useRouter, usePathname, useSearchParams } from "next/navigation"
import { useState, useEffect } from "react"

export function Header() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const [searchValue, setSearchValue] = useState(searchParams.get("search") || "")

  // Sync local state with URL param when it changes (e.g. back button)
  useEffect(() => {
    setSearchValue(searchParams.get("search") || "")
  }, [searchParams])

  // Debounced search effect
  useEffect(() => {
    const timer = setTimeout(() => {
      const currentSearch = searchParams.get("search") || ""
      if (searchValue === currentSearch) return

      const params = new URLSearchParams(searchParams.toString())
      if (searchValue) {
        params.set("search", searchValue)
      } else {
        params.delete("search")
      }

      // If not on transactions page, redirect there
      if (pathname !== "/transactions") {
        router.push(`/transactions?${params.toString()}`)
      } else {
        router.replace(`/transactions?${params.toString()}`)
      }
    }, 300)

    return () => clearTimeout(timer)
  }, [searchValue, router, pathname, searchParams])

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-4 border-b border-border bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/80 px-6">
      {/* Search */}
      <div className="flex-1 max-w-xl">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search transactions, categories..."
            className="w-full rounded-lg border-border bg-background pl-9"
            value={searchValue}
            onChange={(e) => setSearchValue(e.target.value)}
          />
        </div>
      </div>

      {/* Right side actions */}
      <div className="flex items-center gap-2">
        {/* Notifications button */}
        <Button
          variant="ghost"
          size="icon"
          className="relative rounded-lg"
          aria-label="Notifications"
        >
          <Bell className="h-5 w-5" />
          {/* Notification badge - hidden for now */}
          {/* <span className="absolute right-1 top-1 h-2 w-2 rounded-full bg-destructive" /> */}
        </Button>

        {/* User button */}
        <UserButton
          appearance={{
            elements: {
              avatarBox: "h-9 w-9",
            },
          }}
        />
      </div>
    </header>
  )
}
