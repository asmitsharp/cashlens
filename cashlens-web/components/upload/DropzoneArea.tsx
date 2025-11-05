"use client"

import { useCallback } from "react"
import { useDropzone } from "react-dropzone"
import { Upload, FileText, AlertCircle } from "lucide-react"
import { cn } from "@/lib/utils"
import { ACCEPTED_FILE_TYPES, ACCEPTED_EXTENSIONS } from "@/types/upload"

interface DropzoneAreaProps {
  onFileSelect: (file: File) => void
  disabled?: boolean
  error?: string
}

export function DropzoneArea({
  onFileSelect,
  disabled = false,
  error,
}: DropzoneAreaProps) {
  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      if (acceptedFiles.length > 0) {
        onFileSelect(acceptedFiles[0])
      }
    },
    [onFileSelect]
  )

  const { getRootProps, getInputProps, isDragActive, fileRejections } =
    useDropzone({
      onDrop,
      accept: ACCEPTED_FILE_TYPES,
      maxFiles: 1,
      disabled,
      maxSize: 10 * 1024 * 1024, // 10MB
    })

  const rejectionError = fileRejections[0]?.errors[0]?.message

  return (
    <div
      {...getRootProps()}
      className={cn(
        "relative flex flex-col items-center justify-center rounded-2xl border-2 border-dashed border-border bg-background p-12 transition-colors",
        isDragActive && "border-primary bg-accent",
        disabled && "cursor-not-allowed opacity-60",
        !disabled && "cursor-pointer hover:border-primary hover:bg-accent/50",
        (error || rejectionError) && "border-destructive bg-destructive/5"
      )}
      role="button"
      aria-label="Upload file area"
      aria-disabled={disabled}
      tabIndex={disabled ? -1 : 0}
    >
      <input {...getInputProps()} aria-label="File upload input" />

      {/* Icon */}
      <div
        className={cn(
          "mb-4 rounded-full p-4",
          isDragActive ? "bg-primary/10" : "bg-muted"
        )}
      >
        {error || rejectionError ? (
          <AlertCircle className="h-8 w-8 text-destructive" aria-hidden="true" />
        ) : (
          <Upload
            className={cn(
              "h-8 w-8",
              isDragActive ? "text-primary" : "text-muted-foreground"
            )}
            aria-hidden="true"
          />
        )}
      </div>

      {/* Text */}
      <div className="text-center">
        {isDragActive ? (
          <p className="text-base font-medium text-primary">
            Drop your file here
          </p>
        ) : (
          <>
            <p className="text-base font-medium text-foreground">
              Drag and drop your bank statement
            </p>
            <p className="mt-2 text-sm text-muted-foreground">
              or click to browse files
            </p>
          </>
        )}

        {/* Accepted formats */}
        <div className="mt-4 flex items-center justify-center gap-2">
          <FileText className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          <p className="text-xs text-muted-foreground">
            Supports: {ACCEPTED_EXTENSIONS.join(", ").toUpperCase()} (Max 10MB)
          </p>
        </div>
      </div>

      {/* Error message */}
      {(error || rejectionError) && (
        <div
          className="mt-4 rounded-lg bg-destructive/10 px-4 py-2"
          role="alert"
          aria-live="assertive"
        >
          <p className="text-sm font-medium text-destructive">
            {error || rejectionError}
          </p>
        </div>
      )}
    </div>
  )
}
