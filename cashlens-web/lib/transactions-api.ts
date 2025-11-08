// API client functions for transactions

import type {
  Transaction,
  TransactionListResponse,
  TransactionStats,
  UpdateTransactionRequest,
  BulkUpdateRequest,
} from "@/types/transaction"

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1"

/**
 * Get transactions with optional filtering
 */
export async function getTransactions(
  token: string,
  params?: {
    status?: "all" | "categorized" | "uncategorized"
    limit?: number
    offset?: number
  }
): Promise<TransactionListResponse> {
  const queryParams = new URLSearchParams()
  if (params?.status) queryParams.append("status", params.status)
  if (params?.limit) queryParams.append("limit", params.limit.toString())
  if (params?.offset) queryParams.append("offset", params.offset.toString())

  const url = `${API_URL}/transactions?${queryParams.toString()}`

  const response = await fetch(url, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || `Failed to fetch transactions: ${response.statusText}`)
  }

  return response.json()
}

/**
 * Get transaction statistics
 */
export async function getTransactionStats(
  token: string
): Promise<TransactionStats> {
  const response = await fetch(`${API_URL}/transactions/stats`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || `Failed to fetch stats: ${response.statusText}`)
  }

  return response.json()
}

/**
 * Update a single transaction's category
 */
export async function updateTransaction(
  token: string,
  transactionId: string,
  data: UpdateTransactionRequest
): Promise<{ transaction: Transaction; message: string }> {
  const response = await fetch(`${API_URL}/transactions/${transactionId}`, {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || `Failed to update transaction: ${response.statusText}`)
  }

  return response.json()
}

/**
 * Bulk update multiple transactions
 */
export async function bulkUpdateTransactions(
  token: string,
  data: BulkUpdateRequest
): Promise<{
  updated_count: number
  total_count: number
  message: string
  failed_ids?: string[]
  failed_count?: number
}> {
  const response = await fetch(`${API_URL}/transactions/bulk`, {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || `Failed to bulk update: ${response.statusText}`)
  }

  return response.json()
}
