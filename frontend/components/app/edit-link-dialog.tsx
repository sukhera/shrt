"use client"

import { useState } from "react"
import { toast } from "sonner"

import { ApiError } from "@/lib/api"
import { useUpdateLink } from "@/hooks/use-links"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { Link, UpdateLinkInput } from "@/types/api"

interface EditLinkDialogProps {
  link: Link
  open: boolean
  onOpenChange: (open: boolean) => void
}

// Converts an ISO timestamp to the value a datetime-local input expects
// ("YYYY-MM-DDTHH:mm") in local time.
function toLocalInput(iso: string | null): string {
  if (!iso) return ""
  const d = new Date(iso)
  const tzOffset = d.getTimezoneOffset() * 60_000
  return new Date(d.getTime() - tzOffset).toISOString().slice(0, 16)
}

// Edit dialog pre-filled with a link's current destination, alias, and expiry.
// Submitting issues a PATCH; TanStack Query invalidation refreshes the table.
export function EditLinkDialog({ link, open, onOpenChange }: EditLinkDialogProps) {
  const update = useUpdateLink()
  const [url, setUrl] = useState(link.original_url)
  const [alias, setAlias] = useState(link.slug)
  const [expiresAt, setExpiresAt] = useState(toLocalInput(link.expires_at))

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()

    const input: UpdateLinkInput = {}
    if (url.trim() !== link.original_url) input.url = url.trim()
    if (alias.trim() !== link.slug) input.alias = alias.trim()
    // Expiry: empty clears it (null), a value sets it, unchanged is omitted.
    const nextExpiry = expiresAt ? new Date(expiresAt).toISOString() : null
    if (nextExpiry !== link.expires_at) input.expires_at = nextExpiry

    try {
      await update.mutateAsync({ slug: link.slug, input })
      toast.success("Link updated")
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Couldn't update the link")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit link</DialogTitle>
          <DialogDescription>Update the destination, alias, or expiry.</DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-url">Destination URL</Label>
            <Input
              id="edit-url"
              type="url"
              required
              value={url}
              onChange={(e) => setUrl(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-alias">Alias</Label>
            <Input id="edit-alias" value={alias} onChange={(e) => setAlias(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-expires">Expires at</Label>
            <Input
              id="edit-expires"
              type="datetime-local"
              value={expiresAt}
              onChange={(e) => setExpiresAt(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={update.isPending}>
              {update.isPending ? "Saving…" : "Save changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
