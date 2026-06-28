"use client"

import { useEffect, useMemo, useState } from "react"

import { useLinks } from "@/hooks/use-links"
import { formatDate, linkStatus } from "@/lib/links"
import type { ListLinksParams } from "@/types/api"
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

const PAGE_SIZE = 20

// The dashboard's link table: debounced search, sortable columns, and
// pagination, all driven through TanStack Query (hooks/use-links).
export function LinksTable() {
  const [search, setSearch] = useState("")
  const [debounced, setDebounced] = useState("")
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
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const rows = data?.data ?? []

  return (
    <div className="space-y-4">
      <Input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search by URL or slug…"
        className="max-w-sm"
        aria-label="Search links"
      />

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Slug</TableHead>
              <TableHead>Destination</TableHead>
              <TableHead>
                <button
                  type="button"
                  onClick={() => toggleSort("created_at")}
                  className="hover:text-foreground"
                >
                  Created {sort === "created_at" ? (order === "asc" ? "↑" : "↓") : ""}
                </button>
              </TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="w-12 text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  {Array.from({ length: 5 }).map((__, j) => (
                    <TableCell key={j}>
                      <Skeleton className="h-5 w-full" />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : isError ? (
              <TableRow>
                <TableCell colSpan={5} className="py-10 text-center text-destructive">
                  Couldn&apos;t load your links. Please try again.
                </TableCell>
              </TableRow>
            ) : rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="py-10 text-center text-muted-foreground">
                  {debounced ? "No links match your search." : "You haven't created any links yet."}
                </TableCell>
              </TableRow>
            ) : (
              rows.map((link) => (
                <TableRow key={link.id}>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <span className="font-mono text-sm text-primary">{link.slug}</span>
                      <CopyButton value={link.short_url} label="Copy short URL" />
                    </div>
                  </TableCell>
                  <TableCell className="max-w-xs truncate text-muted-foreground">
                    {link.original_url}
                  </TableCell>
                  <TableCell className="whitespace-nowrap text-muted-foreground">
                    {formatDate(link.created_at)}
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={linkStatus(link)} />
                  </TableCell>
                  <TableCell className="text-right">
                    <LinkRowActions link={link} />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>
          {total} link{total === 1 ? "" : "s"}
        </span>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            Previous
          </Button>
          <span>
            Page {page} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
