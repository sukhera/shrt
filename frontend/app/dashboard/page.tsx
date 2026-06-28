"use client"

import { useRouter } from "next/navigation"
import { useEffect } from "react"

import { useAuth } from "@/hooks/use-auth"
import { LinksTable } from "@/components/app/links-table"
import { Skeleton } from "@/components/ui/skeleton"

export default function DashboardPage() {
  const { status } = useAuth()
  const router = useRouter()

  // Route guard: once auth state resolves, send anonymous users to /login.
  useEffect(() => {
    if (status === "anonymous") {
      router.replace("/login")
    }
  }, [status, router])

  if (status !== "authenticated") {
    // Loading or mid-redirect — show a light skeleton rather than flashing the table.
    return (
      <main className="mx-auto max-w-4xl space-y-4 px-4 py-10">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-64 w-full" />
      </main>
    )
  }

  return (
    <main className="mx-auto max-w-4xl px-4 py-10">
      <h1 className="mb-6 text-2xl font-semibold">Your links</h1>
      <LinksTable />
    </main>
  )
}
