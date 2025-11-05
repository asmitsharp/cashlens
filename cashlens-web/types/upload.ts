// Upload-related TypeScript types

export interface PresignedUrlResponse {
  upload_url: string
  file_key: string
}

export interface ProcessResponse {
  upload_id: string
  total_transactions: number
  categorized_count: number
  accuracy_percent: number
  status: "success" | "processing" | "failed"
  error_message?: string
}

export interface UploadSummary {
  total_transactions: number
  categorized_count: number
  uncategorized_count: number
  accuracy_percent: number
  upload_time: string
}

export type UploadState =
  | { status: "idle" }
  | { status: "selecting" }
  | { status: "uploading"; progress: number }
  | { status: "processing" }
  | { status: "success"; data: ProcessResponse }
  | { status: "error"; message: string }

export interface FileValidation {
  valid: boolean
  error?: string
}

// File size limit: 10MB
export const MAX_FILE_SIZE = 10 * 1024 * 1024

// Accepted file types
export const ACCEPTED_FILE_TYPES = {
  "text/csv": [".csv"],
  "application/vnd.ms-excel": [".xls"],
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
    ".xlsx",
  ],
  "application/pdf": [".pdf"],
}

export const ACCEPTED_EXTENSIONS = [".csv", ".xls", ".xlsx", ".pdf"]
