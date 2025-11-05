// API client functions for upload flow

import { auth } from "@clerk/nextjs/server"
import type {
  PresignedUrlResponse,
  ProcessResponse,
  FileValidation,
} from "@/types/upload"
import { MAX_FILE_SIZE, ACCEPTED_EXTENSIONS } from "@/types/upload"

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1"

/**
 * Get authentication token from Clerk
 * This function should be called from client components
 */
export async function getAuthToken(): Promise<string> {
  if (typeof window === "undefined") {
    throw new Error("getAuthToken must be called from client component")
  }

  // On client side, we'll use Clerk's useAuth hook instead
  throw new Error("Use getAuthTokenClient from client component")
}

/**
 * Validate file before upload
 */
export function validateFile(file: File): FileValidation {
  // Check file size
  if (file.size > MAX_FILE_SIZE) {
    return {
      valid: false,
      error: `File size exceeds 10MB limit. Your file is ${(
        file.size /
        1024 /
        1024
      ).toFixed(2)}MB.`,
    }
  }

  // Check file extension
  const extension = file.name.toLowerCase().match(/\.[^.]+$/)?.[0]
  if (!extension || !ACCEPTED_EXTENSIONS.includes(extension)) {
    return {
      valid: false,
      error: `Invalid file type. Please upload CSV, XLS, XLSX, or PDF files only.`,
    }
  }

  return { valid: true }
}

/**
 * Get presigned URL for S3 upload
 */
export async function getPresignedUrl(
  filename: string,
  contentType: string,
  token: string
): Promise<PresignedUrlResponse> {
  const url = `${API_URL}/upload/presigned-url?filename=${encodeURIComponent(
    filename
  )}&content_type=${encodeURIComponent(contentType)}`

  const response = await fetch(url, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(
      error.message || `Failed to get upload URL: ${response.statusText}`
    )
  }

  return response.json()
}

/**
 * Upload file directly to S3 using presigned URL
 */
export async function uploadToS3(
  file: File,
  uploadUrl: string,
  onProgress?: (progress: number) => void
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()

    // Track upload progress
    if (onProgress) {
      xhr.upload.addEventListener("progress", (event) => {
        if (event.lengthComputable) {
          const progress = Math.round((event.loaded / event.total) * 100)
          onProgress(progress)
        }
      })
    }

    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve()
      } else {
        reject(new Error(`Upload failed: ${xhr.statusText}`))
      }
    })

    xhr.addEventListener("error", () => {
      reject(new Error("Network error during upload"))
    })

    xhr.addEventListener("abort", () => {
      reject(new Error("Upload cancelled"))
    })

    xhr.open("PUT", uploadUrl)
    xhr.setRequestHeader("Content-Type", file.type)
    xhr.send(file)
  })
}

/**
 * Trigger backend processing of uploaded file
 */
export async function processUploadedFile(
  fileKey: string,
  token: string
): Promise<ProcessResponse> {
  const response = await fetch(`${API_URL}/upload/process`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ file_key: fileKey }),
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(
      error.message || `Processing failed: ${response.statusText}`
    )
  }

  return response.json()
}

/**
 * Complete upload flow: get presigned URL, upload to S3, process file
 */
export async function uploadFile(
  file: File,
  token: string,
  onProgress?: (progress: number) => void
): Promise<ProcessResponse> {
  // Step 1: Validate file
  const validation = validateFile(file)
  if (!validation.valid) {
    throw new Error(validation.error)
  }

  // Step 2: Get presigned URL
  const { upload_url, file_key } = await getPresignedUrl(
    file.name,
    file.type,
    token
  )

  // Step 3: Upload to S3 with progress tracking
  await uploadToS3(file, upload_url, onProgress)

  // Step 4: Trigger backend processing
  const result = await processUploadedFile(file_key, token)

  return result
}
