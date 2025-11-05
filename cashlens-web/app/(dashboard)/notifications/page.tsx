import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function NotificationsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">Notifications</h1>
        <p className="text-muted-foreground">
          Manage your notification preferences and view recent activity
        </p>
      </div>

      <Card className="rounded-2xl border-border">
        <CardHeader>
          <CardTitle>Coming Soon</CardTitle>
          <CardDescription>
            Notification preferences and history
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Configure email and in-app notifications for uploads, categorizations, and important events.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
