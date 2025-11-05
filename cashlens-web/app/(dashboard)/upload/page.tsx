"use client"

import { useState } from "react"
import { useAuth } from "@clerk/nextjs"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { DropzoneArea } from "@/components/upload/DropzoneArea"
import { UploadProgress } from "@/components/upload/UploadProgress"
import { UploadSummary } from "@/components/upload/UploadSummary"
import { AlertCircle, ArrowLeft } from "lucide-react"
import { uploadFile, validateFile } from "@/lib/upload-api"
import type { UploadState, ProcessResponse } from "@/types/upload"

export default function UploadPage() {
  const { getToken } = useAuth()
  const [uploadState, setUploadState] = useState<UploadState>({ status: "idle" })
  const [selectedFile, setSelectedFile] = useState<File | null>(null)

  const handleFileSelect = (file: File) => {
    // Validate file
    const validation = validateFile(file)
    if (!validation.valid) {
      setUploadState({
        status: "error",
        message: validation.error || "Invalid file",
      })
      return
    }

    setSelectedFile(file)
    setUploadState({ status: "selecting" })
  }

  const handleUpload = async () => {
    if (!selectedFile) return

    try {
      // Get auth token
      const token = await getToken()
      if (!token) {
        throw new Error("Authentication required. Please sign in again.")
      }

      // Start upload with progress tracking
      setUploadState({ status: "uploading", progress: 0 })

      const result = await uploadFile(
        selectedFile,
        token,
        (progress) => {
          setUploadState({ status: "uploading", progress })
        }
      )

      // Show processing state
      setUploadState({ status: "processing" })

      // Simulate processing delay (backend processes async)
      await new Promise(resolve => setTimeout(resolve, 1000))

      // Show success with results
      setUploadState({ status: "success", data: result })
    } catch (error) {
      console.error("Upload error:", error)
      setUploadState({
        status: "error",
        message:
          error instanceof Error
            ? error.message
            : "Failed to upload file. Please try again.",
      })
    }
  }

  const handleReset = () => {
    setSelectedFile(null)
    setUploadState({ status: "idle" })
  }

  const handleRetry = () => {
    setUploadState({ status: "selecting" })
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      {/* Page header */}
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => window.history.back()}
            className="rounded-lg"
            aria-label="Go back"
          >
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-foreground">
              Upload Bank Statement
            </h1>
            <p className="text-muted-foreground">
              Upload your bank statement to automatically categorize transactions
            </p>
          </div>
        </div>
      </div>

      {/* Main upload card */}
      <Card className="rounded-2xl border-border shadow">
        <CardHeader className="pb-4">
          <CardTitle className="text-2xl font-semibold">
            {uploadState.status === "idle" || uploadState.status === "selecting"
              ? "Select File"
              : uploadState.status === "uploading" || uploadState.status === "processing"
              ? "Uploading"
              : uploadState.status === "success"
              ? "Upload Complete"
              : "Upload Failed"}
          </CardTitle>
          <CardDescription>
            {uploadState.status === "idle" || uploadState.status === "selecting"
              ? "Choose a bank statement file to get started"
              : uploadState.status === "uploading"
              ? "Uploading file to secure storage"
              : uploadState.status === "processing"
              ? "Processing transactions and categorizing"
              : uploadState.status === "success"
              ? "Your transactions have been processed successfully"
              : "There was an error uploading your file"}
          </CardDescription>
        </CardHeader>

        <CardContent>
          {/* Idle state - show dropzone */}
          {uploadState.status === "idle" && (
            <DropzoneArea
              onFileSelect={handleFileSelect}
              disabled={false}
            />
          )}

          {/* File selected - show confirmation */}
          {uploadState.status === "selecting" && selectedFile && (
            <div className="space-y-6">
              <div className="rounded-2xl border-2 border-dashed border-border bg-accent/30 p-6">
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                    <svg
                      className="h-6 w-6 text-primary"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                      aria-hidden="true"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      />
                    </svg>
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="truncate text-base font-medium text-foreground">
                      {selectedFile.name}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {(selectedFile.size / 1024 / 1024).toFixed(2)} MB
                    </p>
                  </div>
                </div>
              </div>

              <div className="flex gap-3">
                <Button
                  onClick={handleUpload}
                  className="flex-1 rounded-lg"
                >
                  Upload and Process
                </Button>
                <Button
                  onClick={handleReset}
                  variant="outline"
                  className="rounded-lg"
                >
                  Cancel
                </Button>
              </div>
            </div>
          )}

          {/* Uploading or processing - show progress */}
          {(uploadState.status === "uploading" || uploadState.status === "processing") &&
            selectedFile && (
              <UploadProgress
                progress={uploadState.status === "uploading" ? uploadState.progress : 100}
                status={uploadState.status}
                fileName={selectedFile.name}
              />
            )}

          {/* Success - show summary */}
          {uploadState.status === "success" && (
            <UploadSummary
              data={uploadState.data}
              onUploadAnother={handleReset}
            />
          )}

          {/* Error state */}
          {uploadState.status === "error" && (
            <div className="space-y-6">
              <div
                className="flex items-start gap-3 rounded-2xl border-2 border-destructive bg-destructive/5 p-6"
                role="alert"
                aria-live="assertive"
              >
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
                  <AlertCircle className="h-6 w-6 text-destructive" aria-hidden="true" />
                </div>
                <div className="flex-1">
                  <h4 className="font-semibold text-foreground">Upload Failed</h4>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {uploadState.message}
                  </p>
                </div>
              </div>

              <div className="flex gap-3">
                <Button
                  onClick={handleRetry}
                  className="flex-1 rounded-lg"
                >
                  Try Again
                </Button>
                <Button
                  onClick={handleReset}
                  variant="outline"
                  className="rounded-lg"
                >
                  Choose Different File
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Information cards */}
      {uploadState.status === "idle" && (
        <div className="grid gap-4 sm:grid-cols-2">
          <Card className="rounded-2xl border-border">
            <CardHeader>
              <CardTitle className="text-lg font-semibold">
                Supported Banks
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li className="flex items-center gap-2">
                  <span className="h-1.5 w-1.5 rounded-full bg-primary" />
                  HDFC Bank
                </li>
                <li className="flex items-center gap-2">
                  <span className="h-1.5 w-1.5 rounded-full bg-primary" />
                  ICICI Bank
                </li>
                <li className="flex items-center gap-2">
                  <span className="h-1.5 w-1.5 rounded-full bg-primary" />
                  State Bank of India (SBI)
                </li>
                <li className="flex items-center gap-2">
                  <span className="h-1.5 w-1.5 rounded-full bg-primary" />
                  Axis Bank
                </li>
                <li className="flex items-center gap-2">
                  <span className="h-1.5 w-1.5 rounded-full bg-primary" />
                  Kotak Mahindra Bank
                </li>
              </ul>
            </CardContent>
          </Card>

          <Card className="rounded-2xl border-border">
            <CardHeader>
              <CardTitle className="text-lg font-semibold">
                What Happens Next?
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2 text-sm text-muted-foreground">
                <li className="flex items-center gap-2">
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                    1
                  </span>
                  File uploaded securely to S3
                </li>
                <li className="flex items-center gap-2">
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                    2
                  </span>
                  Transactions parsed and extracted
                </li>
                <li className="flex items-center gap-2">
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                    3
                  </span>
                  Auto-categorization applied (85%+ accuracy)
                </li>
                <li className="flex items-center gap-2">
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                    4
                  </span>
                  Ready to view in dashboard
                </li>
              </ul>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
