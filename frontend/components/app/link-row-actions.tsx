"use client"

import { MoreHorizontal, Pencil, Trash2 } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { EditLinkDialog } from "@/components/app/edit-link-dialog"
import { DeleteLinkDialog } from "@/components/app/delete-link-dialog"
import type { Link } from "@/types/api"

// Per-row actions menu opening the edit dialog or the delete confirmation.
export function LinkRowActions({ link }: { link: Link }) {
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" aria-label={`Actions for ${link.slug}`}>
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onSelect={() => setEditOpen(true)}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={() => setDeleteOpen(true)}
            className="text-destructive focus:text-destructive"
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <EditLinkDialog link={link} open={editOpen} onOpenChange={setEditOpen} />
      <DeleteLinkDialog link={link} open={deleteOpen} onOpenChange={setDeleteOpen} />
    </>
  )
}
