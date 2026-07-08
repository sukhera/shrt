// Presentation helpers for links — status derivation and formatting. Kept out of
// components so the rules live in one place.

import type { Link } from "@/types/api"

export type LinkStatus = "active" | "expires-soon" | "expired" | "disabled"

const SEVEN_DAYS_MS = 7 * 24 * 60 * 60 * 1000

/** Derives a display status from a link's expiry (design system § 2.1). */
export function linkStatus(link: Link, now: number = Date.now()): LinkStatus {
  if (!link.expires_at) return "active"
  const expiry = new Date(link.expires_at).getTime()
  if (expiry <= now) return "expired"
  if (expiry - now < SEVEN_DAYS_MS) return "expires-soon"
  return "active"
}

const STATUS_LABELS: Record<LinkStatus, string> = {
  active: "Active",
  "expires-soon": "Expiring",
  expired: "Expired",
  disabled: "Disabled",
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

/** Formats an ISO timestamp as a relative string (e.g. "3d ago", "in 2h"). */
export function relativeTime(iso: string, now: number = Date.now()): string {
  const diff = new Date(iso).getTime() - now
  const absDiff = Math.abs(diff)
  const past = diff < 0

  if (absDiff < 60_000) return "just now"

  const minutes = Math.floor(absDiff / 60_000)
  if (minutes < 60) {
    const label = `${minutes}m`
    return past ? `${label} ago` : `in ${label}`
  }

  const hours = Math.floor(absDiff / 3_600_000)
  if (hours < 24) {
    const label = `${hours}h`
    return past ? `${label} ago` : `in ${label}`
  }

  const days = Math.floor(absDiff / 86_400_000)
  if (days < 30) {
    const label = `${days}d`
    return past ? `${label} ago` : `in ${label}`
  }

  const months = Math.floor(days / 30)
  if (months < 12) {
    const label = `${months}mo`
    return past ? `${label} ago` : `in ${label}`
  }

  const years = Math.floor(days / 365)
  const label = `${years}y`
  return past ? `${label} ago` : `in ${label}`
}
