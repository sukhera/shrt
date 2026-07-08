"use client"

import type { Link } from "@/types/api"
import { linkStatus } from "@/lib/links"

interface StatStripProps {
  links: Link[]
  total: number
}

// Dashboard stat strip — big mono numerals in the ping pattern.
// Shows: Total links (account-wide) · Active + Top link (current page only,
// since the backend has no account-wide equivalent yet — labelled as such so
// the numbers don't imply more than they measure).
export function StatStrip({ links, total }: StatStripProps) {
  const hasClicks = links.some((l) => l.click_count !== undefined)

  // Top link by clicks (current page only)
  const topLink = hasClicks
    ? links.reduce<Link | null>(
        (best, l) => (!best || (l.click_count ?? 0) > (best.click_count ?? 0) ? l : best),
        null,
      )
    : null

  // Active links count (current page only)
  const activeCount = links.filter((l) => linkStatus(l) === "active").length

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
      <StatCard label="Total links" value={total} />
      <StatCard label="Active (page)" value={activeCount} />
      {hasClicks ? (
        <StatCard
          label="Top link (page)"
          value={topLink?.click_count ?? 0}
          suffix={topLink ? topLink.slug : undefined}
        />
      ) : (
        <StatCard label="Clicks" value="—" />
      )}
    </div>
  )
}

function StatCard({
  label,
  value,
  suffix,
}: {
  label: string
  value: number | string
  suffix?: string
}) {
  return (
    <div className="rounded-lg border border-[var(--border)] bg-[var(--surface)] px-4 py-3">
      <p className="text-xs font-medium uppercase tracking-wider text-faint">{label}</p>
      <p className="mt-1 font-mono text-2xl font-semibold tabular-nums">{value}</p>
      {suffix ? (
        <p className="mt-0.5 truncate font-mono text-xs text-dim">{suffix}</p>
      ) : null}
    </div>
  )
}
