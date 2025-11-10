"use client"

import { useState, useMemo } from "react"
import { Transaction } from "@/types/transaction"
import { Badge } from "@/components/ui/badge"
import { ArrowDownToLine, ArrowUpFromLine } from "lucide-react"

interface TransactionsTableProps {
  transactions: Transaction[]
  limit?: number
  showFilters?: boolean
}

type SortOption = "recent" | "oldest" | "highest" | "lowest"
type TypeFilter = "all" | "inflow" | "outflow"

export default function TransactionsTable({ transactions, limit, showFilters = false }: TransactionsTableProps) {
  const [sortBy, setSortBy] = useState<SortOption>("recent")
  const [typeFilter, setTypeFilter] = useState<TypeFilter>("all")
  const [categoryFilter, setCategoryFilter] = useState<string>("all")

  // Get unique categories
  const categories = useMemo(() => {
    const cats = new Set(
      transactions
        .map(t => t.category)
        .filter((c): c is string => c !== null && c !== undefined)
    )
    return Array.from(cats).sort()
  }, [transactions])

  // Apply filters and sorting
  const filteredAndSortedTransactions = useMemo(() => {
    let filtered = [...transactions]

    // Apply type filter
    if (typeFilter === "inflow") {
      filtered = filtered.filter(t => t.txn_type === "credit")
    } else if (typeFilter === "outflow") {
      filtered = filtered.filter(t => t.txn_type === "debit")
    }

    // Apply category filter
    if (categoryFilter !== "all") {
      filtered = filtered.filter(t => t.category === categoryFilter)
    }

    // Apply sorting
    filtered.sort((a, b) => {
      switch (sortBy) {
        case "recent":
          return new Date(b.txn_date).getTime() - new Date(a.txn_date).getTime()
        case "oldest":
          return new Date(a.txn_date).getTime() - new Date(b.txn_date).getTime()
        case "highest":
          return b.amount - a.amount
        case "lowest":
          return a.amount - b.amount
        default:
          return 0
      }
    })

    return filtered
  }, [transactions, sortBy, typeFilter, categoryFilter])

  const displayTransactions = limit
    ? filteredAndSortedTransactions.slice(0, limit)
    : filteredAndSortedTransactions

  if (!transactions || transactions.length === 0) {
    return (
      <div className="bg-white p-6 rounded-2xl shadow">
        <h2 className="text-xl font-semibold mb-4">Recent Transactions</h2>
        <div className="text-center py-8 text-muted-foreground">
          No transactions found
        </div>
      </div>
    )
  }

  return (
    <div className="bg-white p-6 rounded-2xl shadow">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">Recent Transactions</h2>
        <p className="text-sm text-muted-foreground">
          Showing {displayTransactions.length} of {transactions.length}
        </p>
      </div>

      {/* Filters */}
      {showFilters && (
        <div className="mb-6 p-4 bg-muted/30 rounded-lg">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Sort By */}
            <div>
              <label className="block text-sm font-medium text-muted-foreground mb-2">
                Sort By
              </label>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as SortOption)}
                className="w-full px-3 py-2 bg-white border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
              >
                <option value="recent">Most Recent</option>
                <option value="oldest">Oldest First</option>
                <option value="highest">Highest Amount</option>
                <option value="lowest">Lowest Amount</option>
              </select>
            </div>

            {/* Transaction Type */}
            <div>
              <label className="block text-sm font-medium text-muted-foreground mb-2">
                Type
              </label>
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value as TypeFilter)}
                className="w-full px-3 py-2 bg-white border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
              >
                <option value="all">All Transactions</option>
                <option value="inflow">Inflow Only</option>
                <option value="outflow">Outflow Only</option>
              </select>
            </div>

            {/* Category */}
            <div>
              <label className="block text-sm font-medium text-muted-foreground mb-2">
                Category
              </label>
              <select
                value={categoryFilter}
                onChange={(e) => setCategoryFilter(e.target.value)}
                className="w-full px-3 py-2 bg-white border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
              >
                <option value="all">All Categories</option>
                {categories.map((cat) => (
                  <option key={cat} value={cat}>
                    {cat}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Date</th>
              <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Description</th>
              <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Bank</th>
              <th className="text-left py-3 px-4 text-sm font-medium text-muted-foreground">Category</th>
              <th className="text-right py-3 px-4 text-sm font-medium text-muted-foreground">Amount</th>
            </tr>
          </thead>
          <tbody>
            {displayTransactions.map((txn) => {
              const isCredit = txn.txn_type === 'credit'
              const amount = Math.abs(txn.amount)

              return (
                <tr key={txn.id} className="border-b border-border last:border-0 hover:bg-muted/50 transition-colors">
                  <td className="py-3 px-4 text-sm">
                    {new Date(txn.txn_date).toLocaleDateString('en-US', {
                      month: 'short',
                      day: 'numeric',
                      year: 'numeric',
                    })}
                  </td>
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-2">
                      {isCredit ? (
                        <ArrowDownToLine className="w-4 h-4 text-success flex-shrink-0" />
                      ) : (
                        <ArrowUpFromLine className="w-4 h-4 text-destructive flex-shrink-0" />
                      )}
                      <span className="text-sm font-medium truncate max-w-md">
                        {txn.description}
                      </span>
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    {txn.bank_type && (
                      <span className="text-xs font-medium text-muted-foreground uppercase">
                        {txn.bank_type}
                      </span>
                    )}
                  </td>
                  <td className="py-3 px-4">
                    {txn.category && (
                      <Badge variant="secondary" className="text-xs">
                        {txn.category}
                      </Badge>
                    )}
                  </td>
                  <td className="py-3 px-4 text-right">
                    <span className={`text-sm font-semibold ${
                      isCredit ? 'text-success' : 'text-destructive'
                    }`}>
                      {isCredit ? '+' : '-'}₹{amount.toLocaleString('en-IN', {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })}
                    </span>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
