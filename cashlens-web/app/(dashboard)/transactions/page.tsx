"use client"

import { useAuth } from "@clerk/nextjs"
import { useEffect, useState } from "react"
import { useSearchParams } from "next/navigation"
import { getTransactions } from "@/lib/transactions-api"
import { Transaction } from "@/types/transaction"
import TransactionsTable from "@/components/transactions/TransactionsTable"

export default function TransactionsPage() {
  const { getToken } = useAuth()
  const searchParams = useSearchParams()
  const search = searchParams.get("search") || ""
  
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const ITEMS_PER_PAGE = 50

  useEffect(() => {
    setPage(1) // Reset page on search change
    loadTransactions(1)
  }, [search])

  const loadTransactions = async (pageNum: number) => {
    try {
      setLoading(true)
      setError(null)

      const token = await getToken()
      if (!token) {
        throw new Error("Not authenticated")
      }

      const data = await getTransactions(token, {
        limit: ITEMS_PER_PAGE * pageNum,
        status: "all",
        search: search,
      })

      setTransactions(data.transactions || [])
      setHasMore(data.transactions.length >= ITEMS_PER_PAGE * pageNum)
    } catch (err) {
      console.error("Failed to load transactions:", err)
      setError(err instanceof Error ? err.message : "Failed to load transactions")
    } finally {
      setLoading(false)
    }
  }

  const loadMore = async () => {
    try {
      const token = await getToken()
      if (!token) return

      const nextPage = page + 1
      const data = await getTransactions(token, {
        limit: ITEMS_PER_PAGE * nextPage,
        status: "all",
        search: search,
      })

      setTransactions(data.transactions || [])
      setHasMore(data.transactions.length >= ITEMS_PER_PAGE * nextPage)
      setPage(nextPage)
    } catch (err) {
      console.error("Failed to load more transactions:", err)
    }
  }

  if (loading && transactions.length === 0) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-muted-foreground">Loading transactions...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-destructive">
          <p className="font-semibold">Error loading transactions</p>
          <p className="text-sm mt-1">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">All Transactions</h1>
        <p className="mt-2 text-muted-foreground">
          {search ? `Showing results for "${search}"` : "View and filter all your bank transactions"}
        </p>
      </div>

      {/* Transactions Table with Filters */}
      <TransactionsTable transactions={transactions} showFilters={true} />

      {/* Load More Button */}
      {hasMore && transactions.length > 0 && (
        <div className="flex justify-center">
          <button
            onClick={loadMore}
            className="px-6 py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:bg-primary/90 transition-colors"
          >
            Load More Transactions
          </button>
        </div>
      )}

      {!hasMore && transactions.length > 0 && (
        <div className="text-center py-4 text-sm text-muted-foreground">
          No more transactions to load
        </div>
      )}
    </div>
  )
}
