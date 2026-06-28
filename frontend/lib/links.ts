// Presentation helpers for links — status derivation and formatting. Kept out of
// components so the rules live in one place.

import type { Link } from "@/types/api"

export type LinkStatus = "active" | "expires-soon" | "expired"

const SEVEN_DAYS_MS = 7 * 24 * 60 * 60 * 1000

/** Derives a display status from a link's expiry (design system § 1.6). */
export function linkStatus(link: Link, now: number = Date.now()): LinkStatus {
  if (!link.expires_at) return "active"
  const expiry = new Date(link.expires_at).getTime()
  if (expiry <= now) return "expired"
  if (expiry - now < SEVEN_DAYS_MS) return "expires-soon"
  return "active"
}

const STATUS_LABELS: Record<LinkStatus, string> = {
  active: "Active",
  "expires-soon": "Expires soon",
  expired: "Expired",
}

export function statusLabel(status: LinkStatus): string {
  return STATUS_LABELS[status]
}

/** Formats an ISO timestamp as a short, locale-aware date. */
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}
