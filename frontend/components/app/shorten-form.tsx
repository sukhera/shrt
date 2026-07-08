"use client"

import { ChevronDown } from "lucide-react"
import { useRef, useState } from "react"

import { ApiError, links } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { CopyButton } from "@/components/app/copy-button"
import type { Link } from "@/types/api"

// Home-page shortener with the "result moment" — on success the form morphs
// into a result card with a large mono link, copy feedback, and an expanding-
// ring pulse (600ms, prefers-reduced-motion aware).
export function ShortenForm() {
  const [url, setUrl] = useState("")
  const [alias, setAlias] = useState("")
  const [expiresAt, setExpiresAt] = useState("")
  const [advanced, setAdvanced] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<Link | null>(null)
  const [showPulse, setShowPulse] = useState(false)
  const cardRef = useRef<HTMLDivElement>(null)

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setResult(null)
    setSubmitting(true)
    try {
      const link = await links.create({
        url: url.trim(),
        alias: alias.trim() || undefined,
        expires_at: expiresAt ? new Date(expiresAt).toISOString() : undefined,
      })
      setResult(link)
      // Trigger the pulse animation
      setShowPulse(true)
      setTimeout(() => setShowPulse(false), 700)
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Something went wrong. Please try again.",
      )
    } finally {
      setSubmitting(false)
    }
  }

  function reset() {
    setResult(null)
    setUrl("")
    setAlias("")
    setExpiresAt("")
    setAdvanced(false)
  }

  // ── Result moment ──────────────────────────────────────────────
  if (result) {
    return (
      <div className="w-full max-w-2xl">
        <div
          ref={cardRef}
          className="relative overflow-hidden rounded-lg border border-[var(--border)] bg-[var(--surface)] p-6"
        >
          {/* Pulse ring — expanding border glow on success */}
          {showPulse && (
            <span
              className="pointer-events-none absolute inset-0 rounded-lg border-2 border-primary motion-reduce:hidden"
              style={{
                animation: "pulse-ring 0.6s ease-out forwards",
              }}
            />
          )}

          {/* Large mono link */}
          <div className="flex items-center justify-between gap-3">
            <a
              href={result.short_url}
              target="_blank"
              rel="noopener noreferrer"
              className="truncate font-mono text-xl font-semibold text-primary hover:underline"
            >
              {result.short_url}
            </a>
            <CopyButton value={result.short_url} label="Copy short URL" />
          </div>

          {/* Destination preview */}
          <p className="mt-2 truncate font-mono text-sm text-faint">
            &#x2192; {result.original_url}
          </p>
        </div>

        {/* Shorten another */}
        <button
          type="button"
          onClick={reset}
          className="mt-4 text-sm text-dim hover:text-primary"
        >
          Shorten another
        </button>
      </div>
    )
  }

  // ── Form ───────────────────────────────────────────────────────
  return (
    <div className="w-full max-w-2xl">
      <form onSubmit={onSubmit}>
        <div className="flex flex-col gap-2 sm:flex-row">
          <Input
            type="url"
            required
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://…"
            className="h-14 font-mono text-base placeholder:text-faint"
            aria-label="URL to shorten"
          />
          <Button
            type="submit"
            size="lg"
            disabled={submitting}
            className="h-14 px-8 font-semibold"
          >
            {submitting ? "Shortening…" : "Shorten"}
          </Button>
        </div>

        <button
          type="button"
          onClick={() => setAdvanced((v) => !v)}
          className="mt-3 flex items-center gap-1 text-sm text-dim hover:text-foreground"
          aria-expanded={advanced}
        >
          <ChevronDown
            className={`h-4 w-4 transition-transform ${advanced ? "rotate-180" : ""}`}
          />
          Advanced options
        </button>

        {advanced ? (
          <div className="mt-3 grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="alias">Custom alias</Label>
              <Input
                id="alias"
                value={alias}
                onChange={(e) => setAlias(e.target.value)}
                placeholder="my-link"
                className="font-mono"
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
          <p className="mt-3 text-sm text-[var(--bad)]" role="alert">
            {error}
          </p>
        ) : null}
      </form>
    </div>
  )
}
