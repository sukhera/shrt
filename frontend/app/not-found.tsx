import Link from "next/link"

import { Button } from "@/components/ui/button"

// Branded 404 — shown for unknown app routes and unknown slugs that fall through
// to the frontend.
export default function NotFound() {
  return (
    <main className="mx-auto flex min-h-[calc(100vh-3.5rem)] max-w-4xl flex-col items-center justify-center gap-4 px-4 text-center">
      <p className="font-mono text-5xl font-bold text-primary">404</p>
      <h1 className="text-xl font-semibold">This link doesn&apos;t exist</h1>
      <p className="max-w-md text-muted-foreground">
        The short link you followed is invalid or was never created.
      </p>
      <Button asChild>
        <Link href="/">Shorten a URL</Link>
      </Button>
    </main>
  )
}
