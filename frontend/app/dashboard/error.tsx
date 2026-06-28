"use client"

import { Button } from "@/components/ui/button"

// Error boundary for the dashboard route segment (App Router convention).
export default function DashboardError({
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <main className="mx-auto flex max-w-4xl flex-col items-center justify-center gap-4 px-4 py-20 text-center">
      <h2 className="text-xl font-semibold">Something went wrong</h2>
      <p className="text-muted-foreground">
        We couldn&apos;t load your dashboard. Please try again.
      </p>
      <Button onClick={reset}>Try again</Button>
    </main>
  )
}
