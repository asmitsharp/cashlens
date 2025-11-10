"use client"

import { useUser, useAuth } from "@clerk/nextjs"
import { useEffect, useState } from "react"
import { getSummary, SummaryResponse } from "@/lib/summary-api"
import { getTransactions } from "@/lib/transactions-api"
import { Transaction } from "@/types/transaction"
import NetCashFlowChart from "@/components/charts/NetCashFlowChart"
import TransactionsTable from "@/components/transactions/TransactionsTable"
import { TrendingUp, TrendingDown, ArrowDownToLine, ArrowUpFromLine } from "lucide-react"

export default function DashboardPage() {
  const { user, isLoaded } = useUser()
  const { getToken } = useAuth()
  const [summary, setSummary] = useState<SummaryResponse | null>(null)
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isLoaded) {
      loadDashboardData()
    }
  }, [isLoaded])

  const loadDashboardData = async () => {
    try {
      setLoading(true)
      setError(null)

      // Get Clerk token
      const token = await getToken()
      if (!token) {
        throw new Error("Not authenticated")
      }

      // Get all-time data by using a wide date range
      // TODO: Add date range picker in future
      const toDate = new Date()
      const fromDate = new Date('2020-01-01') // Start from 2020 to catch all historical data

      // Load summary data
      const summaryData = await getSummary(
        token,
        fromDate.toISOString().split('T')[0],
        toDate.toISOString().split('T')[0],
        'day'
      )
      setSummary(summaryData)

      // Load recent transactions
      const transactionsData = await getTransactions(token, {
        limit: 10,
        status: 'all'
      })
      setTransactions(transactionsData.transactions || [])
    } catch (err) {
      console.error("Failed to load dashboard data:", err)
      setError(err instanceof Error ? err.message : "Failed to load dashboard data")
    } finally {
      setLoading(false)
    }
  }

  if (!isLoaded || loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-muted-foreground">Loading dashboard...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-destructive">
          <p className="font-semibold">Error loading dashboard</p>
          <p className="text-sm mt-1">{error}</p>
        </div>
      </div>
    )
  }

  if (!summary) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-muted-foreground">No data available</div>
      </div>
    )
  }

  const isNetPositive = summary.kpis.net_cash_flow >= 0

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">
          Welcome back, {user?.firstName || "there"}!
        </h1>
        <p className="mt-2 text-muted-foreground">
          Here's your complete financial overview
        </p>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Net Cash Flow Card */}
        <div className="bg-white p-6 rounded-2xl shadow">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium text-muted-foreground">Net Cash Flow</p>
            {isNetPositive ? (
              <TrendingUp className="w-5 h-5 text-success" />
            ) : (
              <TrendingDown className="w-5 h-5 text-destructive" />
            )}
          </div>
          <p className={`text-3xl font-bold ${
            isNetPositive ? 'text-success' : 'text-destructive'
          }`}>
            ₹{Math.abs(summary.kpis.net_cash_flow).toLocaleString('en-IN', {
              minimumFractionDigits: 2,
              maximumFractionDigits: 2,
            })}
          </p>
          <p className="text-xs text-muted-foreground mt-1">
            {summary.kpis.transaction_count.toLocaleString()} transactions
          </p>
        </div>

        {/* Total Inflow Card */}
        <div className="bg-white p-6 rounded-2xl shadow">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium text-muted-foreground">Total Inflow</p>
            <ArrowDownToLine className="w-5 h-5 text-success" />
          </div>
          <p className="text-3xl font-bold text-success">
            ₹{summary.kpis.total_inflow.toLocaleString('en-IN', {
              minimumFractionDigits: 2,
              maximumFractionDigits: 2,
            })}
          </p>
          <p className="text-xs text-muted-foreground mt-1">
            Money received
          </p>
        </div>

        {/* Total Outflow Card */}
        <div className="bg-white p-6 rounded-2xl shadow">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium text-muted-foreground">Total Outflow</p>
            <ArrowUpFromLine className="w-5 h-5 text-destructive" />
          </div>
          <p className="text-3xl font-bold text-destructive">
            ₹{summary.kpis.total_outflow.toLocaleString('en-IN', {
              minimumFractionDigits: 2,
              maximumFractionDigits: 2,
            })}
          </p>
          <p className="text-xs text-muted-foreground mt-1">
            Money spent
          </p>
        </div>
      </div>

      {/* Net Flow Chart */}
      <NetCashFlowChart data={summary.net_flow_trend} />

      {/* Recent Transactions */}
      <TransactionsTable transactions={transactions} limit={10} />
    </div>
  )
}
