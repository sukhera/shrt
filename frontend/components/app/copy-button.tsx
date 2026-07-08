"use client"

import { Check, Copy } from "lucide-react"
import { useState } from "react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"

interface CopyButtonProps {
  value: string
  label?: string
}

// Icon button that copies a value to the clipboard, shows a brief check, and
// fires a Sonner toast. Used for short URLs and slugs.
export function CopyButton({ value, label = "Copy" }: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      toast.success("Link copied")
      setTimeout(() => setCopied(false), 1500)
    } catch {
      toast.error("Couldn't copy to clipboard")
    }
  }

  return (
    <Button variant="ghost" size="icon" aria-label={label} onClick={copy}>
      {copied ? (
        <Check className="h-4 w-4 text-ok" />
      ) : (
        <Copy className="h-4 w-4" />
      )}
    </Button>
  )
}
