"use client"

import { toast } from "sonner"

import { ApiError } from "@/lib/api"
import { useDeleteLink } from "@/hooks/use-links"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import type { Link } from "@/types/api"

interface DeleteLinkDialogProps {
  link: Link
  open: boolean
  onOpenChange: (open: boolean) => void
}

// Destructive confirmation for deleting a link. Uses AlertDialog per the design
// system rule that destructive actions always confirm.
export function DeleteLinkDialog({ link, open, onOpenChange }: DeleteLinkDialogProps) {
  const remove = useDeleteLink()

  async function onConfirm() {
    try {
      await remove.mutateAsync(link.slug)
      toast.success("Link deleted")
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Couldn't delete the link")
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete this link?</AlertDialogTitle>
          <AlertDialogDescription>
            <span className="font-mono text-primary">{link.slug}</span> will stop redirecting
            immediately. This cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault()
              void onConfirm()
            }}
            disabled={remove.isPending}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {remove.isPending ? "Deleting…" : "Delete"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
