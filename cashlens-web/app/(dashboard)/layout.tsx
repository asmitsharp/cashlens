import { auth } from "@clerk/nextjs/server"
import { redirect } from "next/navigation"
import { SidebarProvider } from "@/components/layout/SidebarContext"
import { Sidebar } from "@/components/layout/Sidebar"
import { DashboardContent } from "@/components/layout/DashboardContent"

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { userId } = await auth()

  if (!userId) {
    redirect("/sign-in")
  }

  return (
    <SidebarProvider>
      <div className="min-h-screen bg-background">
        <Sidebar />
        <DashboardContent>{children}</DashboardContent>
      </div>
    </SidebarProvider>
  )
}
