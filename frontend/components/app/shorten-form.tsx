"use client"

import { ChevronDown } from "lucide-react"
import { useState } from "react"

import { ApiError, links } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { CopyButton } from "@/components/app/copy-button"
import type { Link } from "@/types/api"

// Home-page shortener. Works for both anonymous and authenticated users — the
// access token (if any) is attached by the api layer. The advanced toggle
// reveals optional alias and expiry fields.
export function ShortenForm() {
  const [url, setUrl] = useState("")
  const [alias, setAlias] = useState("")
  const [expiresAt, setExpiresAt] = useState("")
  const [advanced, setAdvanced] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<Link | null>(null)

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setResult(null)
    setSubmitting(true)
    try {
      const link = await links.create({
        url: url.trim(),
        alias: alias.trim() || undefined,
        // datetime-local yields "YYYY-MM-DDTHH:mm"; convert to RFC3339/ISO.
        expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
      })
      setResult(link)
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Something went wrong. Please try again.",
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="w-full max-w-2xl">
      <form onSubmit={onSubmit} className="space-y-4">
        <div className="flex flex-col gap-2 sm:flex-row">
          <Input
            type="url"
            required
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com/your/long/url"
            className="h-12 text-base"
            aria-label="URL to shorten"
          />
          <Button type="submit" size="lg" disabled={submitting} className="h-12">
            {submitting ? "Shortening…" : "Shorten"}
          </Button>
        </div>

        <button
          type="button"
          onClick={() => setAdvanced((v) => !v)}
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
          aria-expanded={advanced}
        >
          <ChevronDown
            className={`h-4 w-4 transition-transform ${advanced ? "rotate-180" : ""}`}
          />
          Advanced options
        </button>

        {advanced ? (
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="alias">Custom alias</Label>
              <Input
                id="alias"
                value={alias}
                onChange={(e) => setAlias(e.target.value)}
                placeholder="my-link"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="expires">Expires at</Label>
              <Input
                id="expires"
                type="datetime-local"
                value={expiresAt}
                onChange={(e) => setExpiresAt(e.target.value)}
              />
            </div>
          </div>
        ) : null}

        {error ? (
          <p className="text-sm text-destructive" role="alert">
            {error}
          </p>
        ) : null}
      </form>

      {result ? (
        <Card className="mt-6">
          <CardContent className="flex items-center justify-between gap-3 py-4">
            <a
              href={result.short_url}
              target="_blank"
              rel="noopener noreferrer"
              className="truncate font-mono text-sm text-primary hover:underline"
            >
              {result.short_url}
            </a>
            <CopyButton value={result.short_url} label="Copy short URL" />
          </CardContent>
        </Card>
      ) : null}
    </div>
  )
}
