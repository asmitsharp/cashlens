"use client"

import { Progress } from "@/components/ui/progress"
import { Loader2, Upload, CheckCircle2 } from "lucide-react"
import { cn } from "@/lib/utils"

interface UploadProgressProps {
  progress: number
  status: "uploading" | "processing" | "success"
  fileName: string
}

export function UploadProgress({
  progress,
  status,
  fileName,
}: UploadProgressProps) {
  return (
    <div
      className="space-y-4"
      role="status"
      aria-live="polite"
      aria-label={`Upload progress: ${progress}%`}
    >
      {/* File info */}
      <div className="flex items-center gap-3">
        <div
          className={cn(
            "flex h-12 w-12 items-center justify-center rounded-lg",
            status === "success"
              ? "bg-success/10"
              : status === "processing"
              ? "bg-primary/10"
              : "bg-muted"
          )}
        >
          {status === "success" ? (
            <CheckCircle2 className="h-6 w-6 text-success" aria-hidden="true" />
          ) : status === "processing" ? (
            <Loader2
              className="h-6 w-6 animate-spin text-primary"
              aria-hidden="true"
            />
          ) : (
            <Upload className="h-6 w-6 text-muted-foreground" aria-hidden="true" />
          )}
        </div>

        <div className="flex-1 min-w-0">
          <p className="truncate text-sm font-medium text-foreground">
            {fileName}
          </p>
          <p className="text-xs text-muted-foreground">
            {status === "success"
              ? "Upload complete"
              : status === "processing"
              ? "Processing file..."
              : `Uploading... ${progress}%`}
          </p>
        </div>
      </div>

      {/* Progress bar */}
      {status !== "success" && (
        <div className="space-y-2">
          <Progress
            value={status === "processing" ? 100 : progress}
            className="h-2"
            aria-label={`Upload progress: ${progress} percent`}
          />
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>{status === "processing" ? "Processing" : "Uploading"}</span>
            <span>{status === "processing" ? "Please wait" : `${progress}%`}</span>
          </div>
        </div>
      )}

      {/* Processing message */}
      {status === "processing" && (
        <div className="rounded-lg bg-muted px-4 py-3">
          <p className="text-sm text-muted-foreground">
            Parsing transactions and categorizing expenses. This may take a few
            moments...
          </p>
        </div>
      )}
    </div>
  )
}
