"use client"

import { useEffect, useMemo, useState } from "react"
import NextLink from "next/link"
import { useRouter, useSearchParams } from "next/navigation"

import { useLinks } from "@/hooks/use-links"
import { linkStatus, relativeTime } from "@/lib/links"
import { truncateMiddle } from "@/lib/utils"
import type { ListLinksParams } from "@/types/api"
import type { LinkStatus } from "@/lib/links"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { CopyButton } from "@/components/app/copy-button"
import { StatusBadge } from "@/components/app/status-badge"
import { LinkRowActions } from "@/components/app/link-row-actions"
import { StatStrip } from "@/components/app/stat-strip"

const PAGE_SIZE = 20

// Status filter options for URL params
const STATUS_FILTERS: { value: string; label: string }[] = [
  { value: "all", label: "All" },
  { value: "active", label: "Active" },
  { value: "expires-soon", label: "Expiring" },
  { value: "expired", label: "Expired" },
]

// Dashboard link table: stat strip, debounced search, status filter in URL
// params, sortable columns, redesigned rows with mono slugs and tabular nums.
export function LinksTable() {
  const router = useRouter()
  const searchParams = useSearchParams()

  // Read filters from URL params
  const initialSearch = searchParams.get("q") ?? ""
  const initialStatus = searchParams.get("status") ?? "all"

  const [search, setSearch] = useState(initialSearch)
  const [debounced, setDebounced] = useState(initialSearch)
  const [statusFilter, setStatusFilter] = useState(initialStatus)
  const [page, setPage] = useState(1)
  const [sort, setSort] = useState<NonNullable<ListLinksParams["sort"]>>("created_at")
  const [order, setOrder] = useState<NonNullable<ListLinksParams["order"]>>("desc")

  // Debounce search input by 300ms and reset to page 1 on a new term.
  useEffect(() => {
    const id = setTimeout(() => {
      setDebounced(search)
      setPage(1)
    }, 300)
    return () => clearTimeout(id)
  }, [search])

  // Sync filters to URL params
  useEffect(() => {
    const params = new URLSearchParams()
    if (debounced) params.set("q", debounced)
    if (statusFilter !== "all") params.set("status", statusFilter)
    const qs = params.toString()
    router.replace(`/dashboard${qs ? `?${qs}` : ""}`, { scroll: false })
  }, [debounced, statusFilter, router])

  const params = useMemo<ListLinksParams>(
    () => ({ page, limit: PAGE_SIZE, sort, order, q: debounced || undefined }),
    [page, sort, order, debounced],
  )

  const { data, isLoading, isError } = useLinks(params)

  function toggleSort(field: NonNullable<ListLinksParams["sort"]>) {
    if (sort === field) {
      setOrder((o) => (o === "asc" ? "desc" : "asc"))
    } else {
      setSort(field)
      setOrder("desc")
    }
  }

  const total = data?.pagination.total ?? 0
  const allRows = data?.data ?? []

  // Client-side status filter (until backend supports it). Because this only
  // filters the already-fetched page, pagination is scoped to the filtered
  // rows on the current page rather than the server's unfiltered total —
  // otherwise the footer counts and page math would contradict what's shown.
  const isFiltered = statusFilter !== "all"
  const rows = isFiltered
    ? allRows.filter((link) => linkStatus(link) === (statusFilter as LinkStatus))
    : allRows

  const displayedTotal = isFiltered ? rows.length : total
  const totalPages = isFiltered ? 1 : Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <div className="space-y-6">
      {/* Stat strip */}
      {!isLoading && !isError && (
        <StatStrip links={allRows} total={total} />
      )}

      {/* Search + status filter */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search by URL or slug…"
          className="max-w-sm font-mono text-sm"
          aria-label="Search links"
        />
        <div className="flex gap-1">
          {STATUS_FILTERS.map((f) => (
            <button
              key={f.value}
              type="button"
              onClick={() => {
                setStatusFilter(f.value)
                setPage(1)
              }}
              className={`rounded-md px-3 py-1.5 text-xs font-medium transition-colors ${
                statusFilter === f.value
                  ? "bg-primary text-primary-foreground"
                  : "text-dim hover:bg-[var(--surface-2)] hover:text-foreground"
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {/* Table */}
      <div className="rounded-lg border border-[var(--border)]">
        <Table>
          <TableHeader>
            <TableRow className="border-b border-[var(--border)]">
              <TableHead className="w-8" />
              <TableHead>Slug</TableHead>
              <TableHead>Destination</TableHead>
              <TableHead className="text-right font-mono">Clicks</TableHead>
              <TableHead>
                <button
                  type="button"
                  onClick={() => toggleSort("created_at")}
                  className="hover:text-foreground"
                >
                  Created {sort === "created_at" ? (order === "asc" ? "↑" : "↓") : ""}
                </button>
              </TableHead>
              <TableHead className="w-12 text-right">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  {Array.from({ length: 6 }).map((__, j) => (
                    <TableCell key={j}>
                      <Skeleton className="h-5 w-full" />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : isError ? (
              <TableRow>
                <TableCell colSpan={6} className="py-10 text-center text-[var(--bad)]">
                  Couldn&apos;t load your links. Please try again.
                </TableCell>
              </TableRow>
            ) : rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="py-16 text-center">
                  {debounced || statusFilter !== "all" ? (
                    <p className="text-dim">No links match your filters.</p>
                  ) : (
                    <div>
                      <p className="font-mono text-2xl text-faint">&#x2234;</p>
                      <p className="mt-2 text-dim">No links yet.</p>
                      <p className="mt-1 text-sm text-faint">
                        Paste a URL on the{" "}
                        <NextLink href="/" className="text-primary hover:underline">home page</NextLink>
                        {" "}to create your first short link.
                      </p>
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ) : (
              rows.map((link) => {
                const status = linkStatus(link)
                const isExpired = status === "expired"

                return (
                  <TableRow
                    key={link.id}
                    className={isExpired ? "opacity-50" : undefined}
                  >
                    {/* Status shape */}
                    <TableCell className="pr-0">
                      <StatusBadge status={status} />
                    </TableCell>

                    {/* Slug */}
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <span className="font-mono text-sm text-primary hover:underline">
                          {link.slug}
                        </span>
                        <CopyButton value={link.short_url} label="Copy short URL" />
                      </div>
                    </TableCell>

                    {/* Destination — middle-truncated, dim */}
                    <TableCell className="max-w-[200px]">
                      <span className="font-mono text-sm text-dim">
                        {truncateMiddle(link.original_url, 45)}
                      </span>
                    </TableCell>

                    {/* Clicks — mono tabular */}
                    <TableCell className="text-right">
                      <span className="font-mono text-sm tabular-nums">
                        {link.click_count !== undefined ? link.click_count : "—"}
                      </span>
                    </TableCell>

                    {/* Created — relative */}
                    <TableCell className="whitespace-nowrap">
                      <span className="font-mono text-xs text-dim">
                        {relativeTime(link.created_at)}
                      </span>
                      {link.expires_at ? (
                        <span className="ml-2 font-mono text-xs text-faint">
                          expires {relativeTime(link.expires_at)}
                        </span>
                      ) : null}
                    </TableCell>

                    {/* Actions */}
                    <TableCell className="text-right">
                      <LinkRowActions link={link} />
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between text-sm text-dim">
        <span className="font-mono tabular-nums">
          {displayedTotal} link{displayedTotal === 1 ? "" : "s"}
          {isFiltered ? " on this page" : ""}
        </span>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={isFiltered || page <= 1}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            Previous
          </Button>
          <span className="font-mono tabular-nums">
            {page}/{totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={isFiltered || page >= totalPages}
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
