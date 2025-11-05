import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">Settings</h1>
        <p className="text-muted-foreground">
          Manage your account settings and preferences
        </p>
      </div>

      {/* Account Settings */}
      <Card className="rounded-2xl border-border">
        <CardHeader>
          <CardTitle>Account Settings</CardTitle>
          <CardDescription>
            Manage your account information and security
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="email-notifications">Email Notifications</Label>
              <p className="text-sm text-muted-foreground">
                Receive email updates about your transactions
              </p>
            </div>
            <Switch id="email-notifications" />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="auto-categorization">Auto-Categorization</Label>
              <p className="text-sm text-muted-foreground">
                Automatically categorize new transactions
              </p>
            </div>
            <Switch id="auto-categorization" defaultChecked />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="dark-mode">Dark Mode</Label>
              <p className="text-sm text-muted-foreground">
                Switch to dark theme (coming soon)
              </p>
            </div>
            <Switch id="dark-mode" disabled />
          </div>
        </CardContent>
      </Card>

      {/* Data & Privacy */}
      <Card className="rounded-2xl border-border">
        <CardHeader>
          <CardTitle>Data & Privacy</CardTitle>
          <CardDescription>
            Manage your data and privacy settings
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Export your data, delete your account, or review privacy settings.
          </p>
          <p className="mt-2 text-sm text-muted-foreground">
            (Feature under development)
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
