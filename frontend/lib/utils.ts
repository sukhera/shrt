import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Middle-truncate a URL for display. Keeps the domain and the tail of the path,
 * replacing the middle with "…" when the string exceeds `max` characters.
 */
export function truncateMiddle(str: string, max: number = 40): string {
  if (str.length <= max) return str
  const keep = Math.floor((max - 1) / 2)
  return str.slice(0, keep) + "…" + str.slice(str.length - keep)
}
