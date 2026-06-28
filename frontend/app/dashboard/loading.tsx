import { Skeleton } from "@/components/ui/skeleton"

// Shown while the dashboard route segment loads (App Router convention).
export default function DashboardLoading() {
  return (
    <main className="mx-auto max-w-4xl space-y-4 px-4 py-10">
      <Skeleton className="h-8 w-40" />
      <Skeleton className="h-10 w-full max-w-sm" />
      <Skeleton className="h-64 w-full" />
    </main>
  )
}
