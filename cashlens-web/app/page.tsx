import Link from "next/link"

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-24">
      <main className="flex flex-col items-center gap-8">
        <h1 className="text-4xl font-bold">Welcome to Cashlens</h1>
        <p className="text-xl text-gray-600">
          Financial Analytics SaaS for Indian SMBs
        </p>
        <div className="flex gap-4">
          <Link
            href="/auth/sign-in"
            className="rounded-lg bg-blue-600 px-6 py-3 text-white hover:bg-blue-700"
          >
            Sign In
          </Link>
          <Link
            href="/dashboard"
            className="rounded-lg border border-gray-300 px-6 py-3 hover:bg-gray-50"
          >
            Dashboard
          </Link>
        </div>
        <div className="mt-8 text-sm text-gray-500">
          <p>Backend API: <a href="http://localhost:8080/health" className="text-blue-600 underline" target="_blank">http://localhost:8080/health</a></p>
        </div>
      </main>
    </div>
  )
}
