import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function InboxPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">Inbox</h1>
        <p className="text-muted-foreground">
          View and manage your transaction notifications and alerts
        </p>
      </div>

      <Card className="rounded-2xl border-border">
        <CardHeader>
          <CardTitle>Coming Soon</CardTitle>
          <CardDescription>
            The inbox feature is under development
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            This page will show transaction alerts, duplicate detections, and other important notifications.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
