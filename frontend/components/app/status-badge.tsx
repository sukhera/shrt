import { Badge } from "@/components/ui/badge"
import { statusLabel, type LinkStatus } from "@/lib/links"

// Maps a link status to a coloured badge. Uses semantic-ish utility colours that
// read in both themes (design system § 1.6).
const STYLES: Record<LinkStatus, string> = {
  active: "border-transparent bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300",
  "expires-soon":
    "border-transparent bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300",
  expired: "border-transparent bg-muted text-muted-foreground",
}

export function StatusBadge({ status }: { status: LinkStatus }) {
  return (
    <Badge variant="outline" className={STYLES[status]}>
      {statusLabel(status)}
    </Badge>
  )
}
