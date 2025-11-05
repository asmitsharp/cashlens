"use client"

import { UserButton } from "@clerk/nextjs"
import { Bell, Search } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

export function Header() {
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
