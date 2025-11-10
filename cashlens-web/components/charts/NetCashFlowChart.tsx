"use client"

import { useState } from "react"
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from "recharts"

interface NetFlowTrendPoint {
  period: string
  inflow: number
  outflow: number
  net_flow: number
}

interface NetCashFlowChartProps {
  data: NetFlowTrendPoint[]
}

// Format date for display
function formatDate(dateStr: string): string {
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  } catch {
    return dateStr
  }
}

export default function NetCashFlowChart({ data }: NetCashFlowChartProps) {
  const [visibleLines, setVisibleLines] = useState({
    inflow: true,
    outflow: true,
    net: true,
  })

  const toggleLine = (line: 'inflow' | 'outflow' | 'net') => {
    setVisibleLines(prev => ({
      ...prev,
      [line]: !prev[line],
    }))
  }

  if (!data || data.length === 0) {
    return (
      <div className="bg-white p-6 rounded-2xl shadow">
        <h2 className="text-xl font-semibold mb-4">Net Cash Flow Trend</h2>
        <div className="h-[300px] flex items-center justify-center text-muted-foreground">
          No data available
        </div>
      </div>
    )
  }

  // Transform data for better display
  const chartData = data.map(point => ({
    ...point,
    formattedDate: formatDate(point.period),
  }))

  return (
    <div className="bg-white p-6 rounded-2xl shadow">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">Cash Flow Trend</h2>
        <div className="flex items-center gap-4 text-sm">
          <button
            onClick={() => toggleLine('inflow')}
            className={`flex items-center gap-2 transition-opacity ${
              visibleLines.inflow ? 'opacity-100' : 'opacity-40'
            } hover:opacity-100 cursor-pointer`}
          >
            <div className="w-3 h-3 rounded-full bg-[hsl(var(--chart-green))]"></div>
            <span className="text-muted-foreground">Inflow</span>
          </button>
          <button
            onClick={() => toggleLine('outflow')}
            className={`flex items-center gap-2 transition-opacity ${
              visibleLines.outflow ? 'opacity-100' : 'opacity-40'
            } hover:opacity-100 cursor-pointer`}
          >
            <div className="w-3 h-3 rounded-full bg-[hsl(var(--chart-red))]"></div>
            <span className="text-muted-foreground">Outflow</span>
          </button>
          <button
            onClick={() => toggleLine('net')}
            className={`flex items-center gap-2 transition-opacity ${
              visibleLines.net ? 'opacity-100' : 'opacity-40'
            } hover:opacity-100 cursor-pointer`}
          >
            <div className="w-3 h-3 rounded-full bg-[hsl(var(--chart-blue))]"></div>
            <span className="text-muted-foreground">Net</span>
          </button>
        </div>
      </div>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={chartData}>
          <CartesianGrid
            strokeDasharray="3 3"
            stroke="hsl(var(--border))"
            vertical={false}
          />
          <XAxis
            dataKey="formattedDate"
            stroke="hsl(var(--muted-foreground))"
            fontSize={12}
            tickLine={false}
            axisLine={false}
            interval="preserveStartEnd"
          />
          <YAxis
            stroke="hsl(var(--muted-foreground))"
            fontSize={12}
            tickLine={false}
            axisLine={false}
            tickFormatter={(value) => `₹${(value / 1000).toFixed(0)}k`}
          />
          <Tooltip
            content={({ active, payload }) => {
              if (!active || !payload || !payload.length) return null

              const data = payload[0].payload

              // Safely extract values with fallbacks
              const inflow = typeof data.inflow === 'number' ? data.inflow : 0
              const outflow = typeof data.outflow === 'number' ? data.outflow : 0
              const netFlow = typeof data.net_flow === 'number' ? data.net_flow : 0

              return (
                <div className="bg-white p-3 rounded-lg shadow-lg border border-border">
                  <p className="text-sm font-medium mb-2">{data.formattedDate}</p>
                  <div className="space-y-1">
                    <div className="flex items-center justify-between gap-4">
                      <span className="text-xs text-muted-foreground">Inflow:</span>
                      <span className="text-sm font-semibold text-success">
                        +₹{inflow.toLocaleString('en-IN', {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </span>
                    </div>
                    <div className="flex items-center justify-between gap-4">
                      <span className="text-xs text-muted-foreground">Outflow:</span>
                      <span className="text-sm font-semibold text-destructive">
                        -₹{outflow.toLocaleString('en-IN', {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </span>
                    </div>
                    <div className="flex items-center justify-between gap-4 pt-1 border-t border-border">
                      <span className="text-xs text-muted-foreground">Net:</span>
                      <span className={`text-sm font-semibold ${
                        netFlow >= 0 ? 'text-success' : 'text-destructive'
                      }`}>
                        {netFlow >= 0 ? '+' : ''}₹{netFlow.toLocaleString('en-IN', {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </span>
                    </div>
                  </div>
                </div>
              )
            }}
          />
          <ReferenceLine
            y={0}
            stroke="hsl(var(--border))"
            strokeWidth={1}
            strokeDasharray="3 3"
          />
          {visibleLines.inflow && (
            <Line
              type="monotone"
              dataKey="inflow"
              stroke="hsl(var(--chart-green))"
              strokeWidth={2}
              dot={{
                fill: 'hsl(var(--chart-green))',
                strokeWidth: 2,
                r: 3,
              }}
              activeDot={{
                r: 5,
                strokeWidth: 2,
              }}
            />
          )}
          {visibleLines.outflow && (
            <Line
              type="monotone"
              dataKey="outflow"
              stroke="hsl(var(--chart-red))"
              strokeWidth={2}
              dot={{
                fill: 'hsl(var(--chart-red))',
                strokeWidth: 2,
                r: 3,
              }}
              activeDot={{
                r: 5,
                strokeWidth: 2,
              }}
            />
          )}
          {visibleLines.net && (
            <Line
              type="monotone"
              dataKey="net_flow"
              stroke="hsl(var(--chart-blue))"
              strokeWidth={2}
              dot={{
                fill: 'hsl(var(--chart-blue))',
                strokeWidth: 2,
                r: 3,
              }}
              activeDot={{
                r: 5,
                strokeWidth: 2,
              }}
            />
          )}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
