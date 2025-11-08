"use client"

import { useEffect, useState, useCallback } from "react"
import { useAuth } from "@clerk/nextjs"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { ArrowLeft, CheckCircle2, Loader2, AlertCircle } from "lucide-react"
import { getTransactions, updateTransaction } from "@/lib/transactions-api"
import { CATEGORIES, type Transaction } from "@/types/transaction"
import { useRouter } from "next/navigation"

export default function ReviewPage() {
  const { getToken } = useAuth()
  const router = useRouter()

  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedRow, setSelectedRow] = useState<number>(0)
  const [processingId, setProcessingId] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const loadTransactions = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)

      const token = await getToken()
      if (!token) {
        throw new Error("Authentication required")
      }

      const data = await getTransactions(token, {
        status: "uncategorized",
        limit: 50,
      })

      setTransactions(data.transactions || [])
    } catch (err) {
      console.error("Failed to load transactions:", err)
      setError(err instanceof Error ? err.message : "Failed to load transactions")
    } finally {
      setLoading(false)
    }
  }, [getToken])

  useEffect(() => {
    loadTransactions()
  }, [loadTransactions])

  const handleCategorySelect = async (transactionId: string, category: string) => {
    try {
      setProcessingId(transactionId)
      setError(null)
      setSuccessMessage(null)

      const token = await getToken()
      if (!token) {
        throw new Error("Authentication required")
      }

      await updateTransaction(token, transactionId, { category })

      // Remove from list (optimistic update)
      setTransactions((prev) => prev.filter((t) => t.id !== transactionId))
      setSuccessMessage(`Transaction categorized as "${category}"`)

      // Clear success message after 3 seconds
      setTimeout(() => setSuccessMessage(null), 3000)

      // Move to next row if available
      if (selectedRow >= transactions.length - 1) {
        setSelectedRow(Math.max(0, selectedRow - 1))
      }
    } catch (err) {
      console.error("Failed to update transaction:", err)
      setError(err instanceof Error ? err.message : "Failed to update transaction")
    } finally {
      setProcessingId(null)
    }
  }

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't handle if a select is open
      if (document.querySelector('[role="listbox"]')) {
        return
      }

      switch (e.key) {
        case "ArrowDown":
          e.preventDefault()
          setSelectedRow((prev) => Math.min(prev + 1, transactions.length - 1))
          break
        case "ArrowUp":
          e.preventDefault()
          setSelectedRow((prev) => Math.max(prev - 1, 0))
          break
        case "Enter":
          e.preventDefault()
          // Focus the select dropdown for the selected row
          document.querySelector(`[data-row="${selectedRow}"] button`)?.click()
          break
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [selectedRow, transactions.length])

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: "INR",
      minimumFractionDigits: 2,
    }).format(Math.abs(amount))
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-IN", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    })
  }

  if (loading) {
    return (
      <div className="flex min-h-[400px] items-center justify-center">
        <div className="flex items-center gap-2 text-muted-foreground">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span>Loading transactions...</span>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      {/* Page header */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => router.back()}
            className="rounded-lg"
            aria-label="Go back"
          >
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-foreground">
              Review Transactions
            </h1>
            <p className="text-muted-foreground">
              {transactions.length > 0
                ? `Categorize ${transactions.length} uncategorized transaction${transactions.length !== 1 ? "s" : ""}`
                : "All transactions have been categorized"}
            </p>
          </div>
        </div>
      </div>

      {/* Success message */}
      {successMessage && (
        <Alert className="border-success bg-success/10">
          <CheckCircle2 className="h-4 w-4 text-success" />
          <AlertDescription className="text-success">
            {successMessage}
          </AlertDescription>
        </Alert>
      )}

      {/* Error message */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Empty state */}
      {transactions.length === 0 && !loading && (
        <Card className="rounded-2xl border-border">
          <CardContent className="flex min-h-[400px] flex-col items-center justify-center gap-4 p-12">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-success/10">
              <CheckCircle2 className="h-8 w-8 text-success" />
            </div>
            <div className="text-center">
              <h3 className="text-lg font-semibold text-foreground">
                All Caught Up!
              </h3>
              <p className="mt-2 text-sm text-muted-foreground">
                All your transactions have been categorized.
              </p>
            </div>
            <Button
              onClick={() => router.push("/dashboard")}
              className="mt-4 rounded-lg"
            >
              View Dashboard
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Transactions table */}
      {transactions.length > 0 && (
        <Card className="rounded-2xl border-border">
          <CardHeader>
            <CardTitle className="text-2xl font-semibold">
              Uncategorized Transactions
            </CardTitle>
            <CardDescription>
              Use arrow keys to navigate, Enter to open dropdown, or click to
              select category
            </CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="border-b border-border bg-accent/30">
                  <tr>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-foreground">
                      Date
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-foreground">
                      Description
                    </th>
                    <th className="px-6 py-3 text-right text-sm font-semibold text-foreground">
                      Amount
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-foreground">
                      Type
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-foreground">
                      Category
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {transactions.map((transaction, index) => (
                    <tr
                      key={transaction.id}
                      data-row={index}
                      className={`transition-colors ${
                        selectedRow === index
                          ? "bg-primary/5 ring-2 ring-inset ring-primary/20"
                          : "hover:bg-accent/50"
                      }`}
                    >
                      <td className="whitespace-nowrap px-6 py-4 text-sm text-muted-foreground">
                        {formatDate(transaction.txn_date)}
                      </td>
                      <td className="max-w-md px-6 py-4 text-sm text-foreground">
                        <div className="truncate" title={transaction.description}>
                          {transaction.description}
                        </div>
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-right text-sm">
                        <span
                          className={
                            transaction.txn_type === "credit"
                              ? "font-medium text-success"
                              : "text-foreground"
                          }
                        >
                          {transaction.txn_type === "credit" ? "+" : "-"}
                          {formatCurrency(transaction.amount)}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <Badge
                          variant={
                            transaction.txn_type === "credit"
                              ? "default"
                              : "secondary"
                          }
                          className="rounded-lg"
                        >
                          {transaction.txn_type}
                        </Badge>
                      </td>
                      <td className="px-6 py-4">
                        <Select
                          onValueChange={(value) =>
                            handleCategorySelect(transaction.id, value)
                          }
                          disabled={processingId === transaction.id}
                        >
                          <SelectTrigger className="w-[200px] rounded-lg">
                            <SelectValue placeholder="Select category" />
                          </SelectTrigger>
                          <SelectContent>
                            {CATEGORIES.map((category) => (
                              <SelectItem key={category} value={category}>
                                {category}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Keyboard shortcuts help */}
      {transactions.length > 0 && (
        <Card className="rounded-2xl border-border bg-accent/30">
          <CardContent className="flex items-center gap-6 p-4">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <kbd className="rounded bg-background px-2 py-1 font-mono text-xs">↑</kbd>
              <kbd className="rounded bg-background px-2 py-1 font-mono text-xs">↓</kbd>
              <span>Navigate</span>
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <kbd className="rounded bg-background px-2 py-1 font-mono text-xs">Enter</kbd>
              <span>Open dropdown</span>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
