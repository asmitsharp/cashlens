"use client"

import { CheckCircle2, FileText, Tag, AlertCircle } from "lucide-react"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import Link from "next/link"
import type { ProcessResponse } from "@/types/upload"

interface UploadSummaryProps {
  data: ProcessResponse
  onUploadAnother: () => void
}

export function UploadSummary({ data, onUploadAnother }: UploadSummaryProps) {
  const uncategorizedCount =
    data.total_transactions - data.categorized_count

  return (
    <div className="space-y-6" role="region" aria-label="Upload summary">
      {/* Success header */}
      <div className="flex items-center gap-3 rounded-2xl bg-success/10 p-6">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-success/20">
          <CheckCircle2 className="h-6 w-6 text-success" aria-hidden="true" />
        </div>
        <div>
          <h3 className="text-lg font-semibold text-foreground">
            Upload Successful!
          </h3>
          <p className="text-sm text-muted-foreground">
            Your transactions have been processed
          </p>
        </div>
      </div>

      {/* Statistics grid */}
      <div className="grid gap-4 sm:grid-cols-3">
        {/* Total transactions */}
        <Card className="rounded-2xl border-border">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <FileText className="h-6 w-6 text-primary" aria-hidden="true" />
            </div>
            <div>
              <p className="text-2xl font-bold text-foreground">
                {data.total_transactions}
              </p>
              <p className="text-sm text-muted-foreground">
                Total Transactions
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Categorized */}
        <Card className="rounded-2xl border-border">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-success/10">
              <Tag className="h-6 w-6 text-success" aria-hidden="true" />
            </div>
            <div>
              <p className="text-2xl font-bold text-foreground">
                {data.categorized_count}
              </p>
              <p className="text-sm text-muted-foreground">
                Auto-Categorized
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Accuracy */}
        <Card className="rounded-2xl border-border">
          <CardContent className="flex items-center gap-4 p-6">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <CheckCircle2 className="h-6 w-6 text-primary" aria-hidden="true" />
            </div>
            <div>
              <p
                className="text-2xl font-bold text-foreground"
                data-testid="accuracy"
              >
                {data.accuracy_percent.toFixed(1)}%
              </p>
              <p className="text-sm text-muted-foreground">
                Accuracy Rate
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Review prompt if needed */}
      {uncategorizedCount > 0 && (
        <div
          className="flex items-start gap-3 rounded-lg bg-warning/10 p-4"
          role="alert"
          aria-live="polite"
        >
          <AlertCircle
            className="h-5 w-5 text-warning mt-0.5"
            aria-hidden="true"
          />
          <div className="flex-1">
            <p className="text-sm font-medium text-foreground">
              {uncategorizedCount} transaction{uncategorizedCount !== 1 ? "s" : ""}{" "}
              need{uncategorizedCount === 1 ? "s" : ""} review
            </p>
            <p className="mt-1 text-sm text-muted-foreground">
              Some transactions could not be automatically categorized. Review
              them to improve accuracy.
            </p>
          </div>
        </div>
      )}

      {/* Action buttons */}
      <div className="flex flex-col gap-3 sm:flex-row">
        <Button asChild className="flex-1 rounded-lg">
          <Link href="/dashboard">
            View Dashboard
          </Link>
        </Button>

        {uncategorizedCount > 0 && (
          <Button asChild variant="secondary" className="flex-1 rounded-lg">
            <Link href="/review">
              Review Transactions
            </Link>
          </Button>
        )}

        <Button
          variant="outline"
          onClick={onUploadAnother}
          className="rounded-lg"
        >
          Upload Another
        </Button>
      </div>
    </div>
  )
}
