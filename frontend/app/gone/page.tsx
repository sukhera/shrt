import Link from "next/link"

import { Button } from "@/components/ui/button"

// Branded 410 page for expired links. The Go redirect server returns 410 Gone
// directly; this page exists so a deployment can point expired-link traffic at a
// consistent, on-brand page.
export default function GonePage() {
  return (
    <main className="mx-auto flex min-h-[calc(100vh-3.5rem)] max-w-4xl flex-col items-center justify-center gap-4 px-4 text-center">
      <p className="font-mono text-5xl font-bold text-primary">410</p>
      <h1 className="text-xl font-semibold">This link has expired</h1>
      <p className="max-w-md text-muted-foreground">
        The short link you followed was set to expire and is no longer active.
      </p>
      <Button asChild>
        <Link href="/">Shorten a new URL</Link>
      </Button>
    </main>
  )
}
