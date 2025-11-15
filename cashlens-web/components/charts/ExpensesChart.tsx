"use client"

import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts"
import type { TopExpenseItem } from "@/lib/summary-api"

interface ExpensesChartProps {
  data: TopExpenseItem[]
}

// Chart colors from design system - gradient from red to amber for expenses
const EXPENSE_COLORS = [
  "hsl(0, 84.2%, 60.2%)",    // Red (destructive) - highest expense
  "hsl(10, 84%, 60%)",        // Red-orange
  "hsl(20, 84%, 60%)",        // Orange-red
  "hsl(25, 84%, 60%)",        // Orange
  "hsl(32, 95%, 44%)",        // Deep orange
  "hsl(38, 92%, 50%)",        // Amber (warning)
  "hsl(43, 90%, 50%)",        // Light amber
  "hsl(48, 85%, 55%)",        // Yellow-amber
  "hsl(53, 80%, 60%)",        // Yellow
  "hsl(58, 75%, 65%)",        // Light yellow
]

export default function ExpensesChart({ data }: ExpensesChartProps) {
  if (!data || data.length === 0) {
    return (
      <div className="bg-white p-6 rounded-2xl shadow">
        <h2 className="text-xl font-semibold mb-4 text-foreground">
          Top Expense Categories
        </h2>
        <div className="flex items-center justify-center h-64 text-muted-foreground">
          No expense data available
        </div>
      </div>
    )
  }

  // Custom tooltip to show formatted values
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload
      return (
        <div className="bg-white p-4 rounded-lg shadow-lg border border-border">
          <p className="font-semibold text-foreground mb-2">{data.category}</p>
          <p className="text-sm text-muted-foreground">
            Amount: <span className="font-medium text-foreground">₹{data.total_amount.toLocaleString("en-IN")}</span>
          </p>
          <p className="text-sm text-muted-foreground">
            Transactions: <span className="font-medium text-foreground">{data.txn_count}</span>
          </p>
          <p className="text-sm text-muted-foreground">
            Percentage: <span className="font-medium text-foreground">{data.percentage.toFixed(1)}%</span>
          </p>
        </div>
      )
    }
    return null
  }

  return (
    <div className="bg-white p-6 rounded-2xl shadow">
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-foreground">
          Top Expense Categories
        </h2>
        <p className="text-sm text-muted-foreground mt-1">
          Your highest spending categories
        </p>
      </div>

      <ResponsiveContainer width="100%" height={400}>
        <BarChart
          data={data}
          layout="vertical"
          margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" opacity={0.3} />
          <XAxis
            type="number"
            tickFormatter={(value) => `₹${(value / 1000).toFixed(0)}k`}
            stroke="hsl(var(--muted-foreground))"
            fontSize={12}
          />
          <YAxis
            dataKey="category"
            type="category"
            width={150}
            stroke="hsl(var(--muted-foreground))"
            fontSize={12}
          />
          <Tooltip content={<CustomTooltip />} cursor={{ fill: "hsl(var(--muted))", opacity: 0.1 }} />
          <Bar dataKey="total_amount" radius={[0, 8, 8, 0]}>
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={EXPENSE_COLORS[index % EXPENSE_COLORS.length]} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>

      {/* Legend showing percentage breakdown */}
      <div className="mt-4 grid grid-cols-2 gap-2">
        {data.slice(0, 6).map((item, index) => (
          <div key={item.category} className="flex items-center gap-2">
            <div
              className="w-3 h-3 rounded-sm flex-shrink-0"
              style={{ backgroundColor: EXPENSE_COLORS[index] }}
            />
            <span className="text-xs text-muted-foreground truncate">
              {item.category} ({item.percentage.toFixed(1)}%)
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
