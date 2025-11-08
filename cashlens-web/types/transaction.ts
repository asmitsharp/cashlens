// Transaction-related TypeScript types

export interface Transaction {
  id: string
  user_id: string
  txn_date: string
  description: string
  amount: number
  txn_type: "credit" | "debit"
  category: string | null
  is_reviewed: boolean
  raw_data: string | null
  created_at: string
  updated_at: string
}

export interface TransactionListResponse {
  transactions: Transaction[]
  total: number
  limit: number
  offset: number
}

export interface TransactionStats {
  total_transactions: number
  categorized_count: number
  uncategorized_count: number
  accuracy_percent: number
}

export interface UpdateTransactionRequest {
  category: string
}

export interface BulkUpdateRequest {
  transaction_ids: string[]
  category: string
}

// Predefined categories
export const CATEGORIES = [
  "Cloud & Hosting",
  "Payment Processing",
  "Marketing",
  "Salaries",
  "Office Supplies",
  "Software & SaaS",
  "Travel",
  "Legal & Professional Services",
  "Utilities",
  "Team Meals",
  "Rent",
  "Insurance",
  "Bank Charges",
  "Taxes",
  "Other",
] as const

export type Category = typeof CATEGORIES[number]
