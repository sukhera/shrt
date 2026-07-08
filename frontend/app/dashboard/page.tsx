"use client"

import { useRouter } from "next/navigation"
import { Suspense, useEffect } from "react"

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
    return (
      <main className="mx-auto max-w-5xl space-y-6 px-4 py-10">
        {/* Stat strip skeleton */}
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-20 rounded-lg" />
          ))}
        </div>
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-64 w-full" />
      </main>
    )
  }

  return (
    <main className="mx-auto max-w-5xl px-4 py-10">
      <h1 className="mb-6 text-2xl font-semibold">Your links</h1>
      <Suspense fallback={<Skeleton className="h-64 w-full" />}>
        <LinksTable />
      </Suspense>
    </main>
  )
}
