const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1"

export interface KPIsResponse {
  total_inflow: number
  total_outflow: number
  net_cash_flow: number
  transaction_count: number
}

export interface NetFlowTrendPoint {
  period: string
  inflow: number
  outflow: number
  net_flow: number
}

export interface SummaryResponse {
  kpis: KPIsResponse
  net_flow_trend: NetFlowTrendPoint[]
  from_date: string
  to_date: string
  group_by: string
}

export async function getSummary(
  token: string,
  fromDate?: string,
  toDate?: string,
  groupBy: string = "month"
): Promise<SummaryResponse> {
  if (!token) {
    throw new Error("Not authenticated")
  }

  const params = new URLSearchParams()
  if (fromDate) params.append("from", fromDate)
  if (toDate) params.append("to", toDate)
  params.append("group_by", groupBy)

  const response = await fetch(`${API_BASE_URL}/summary?${params.toString()}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    cache: "no-store",
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: "Unknown error" }))
    throw new Error(error.error || `Failed to fetch summary: ${response.statusText}`)
  }

  return response.json()
}
