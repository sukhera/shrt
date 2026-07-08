import { statusLabel, type LinkStatus } from "@/lib/links"

// Status chip: shape + color + label — never color alone (WCAG 1.4.1).
//   active      → filled dot    + --ok
//   expires-soon→ triangle      + --warn
//   expired     → filled square + --bad
//   disabled    → ring (hollow) + --off

const STATUS_CONFIG: Record<LinkStatus, { color: string; bg: string }> = {
  active: { color: "var(--ok)", bg: "var(--ok)" },
  "expires-soon": { color: "var(--warn)", bg: "var(--warn)" },
  expired: { color: "var(--bad)", bg: "var(--bad)" },
  disabled: { color: "var(--off)", bg: "var(--off)" },
}

function StatusShape({ status }: { status: LinkStatus }) {
  const { color } = STATUS_CONFIG[status]

  switch (status) {
    case "active":
      // Filled dot
      return (
        <svg width="8" height="8" viewBox="0 0 8 8" aria-hidden="true">
          <circle cx="4" cy="4" r="3.5" fill={color} />
        </svg>
      )
    case "expires-soon":
      // Filled triangle
      return (
        <svg width="9" height="8" viewBox="0 0 9 8" aria-hidden="true">
          <path d="M4.5 0.5L8.5 7.5H0.5Z" fill={color} />
        </svg>
      )
    case "expired":
      // Filled square
      return (
        <svg width="8" height="8" viewBox="0 0 8 8" aria-hidden="true">
          <rect x="0.5" y="0.5" width="7" height="7" rx="1" fill={color} />
        </svg>
      )
    case "disabled":
      // Ring (hollow circle)
      return (
        <svg width="8" height="8" viewBox="0 0 8 8" aria-hidden="true">
          <circle cx="4" cy="4" r="3" fill="none" stroke={color} strokeWidth="1.5" />
        </svg>
      )
  }
}

export function StatusBadge({ status }: { status: LinkStatus }) {
  const { color, bg } = STATUS_CONFIG[status]

  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
      style={{
        color,
        backgroundColor: `color-mix(in srgb, ${bg} 12%, transparent)`,
      }}
    >
      <StatusShape status={status} />
      {statusLabel(status)}
    </span>
  )
}
