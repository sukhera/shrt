// Server-only helpers for the Next.js route handlers that proxy auth to the Go
// backend. These run on the server, so they may read non-public env vars and
// handle the refresh token, which must never reach the browser as JS-readable
// state.

import { cookies } from "next/headers"

/** Name of the httpOnly cookie that holds the backend refresh token. */
export const REFRESH_COOKIE = "shrt_refresh"

/** Lifetime of the refresh cookie, matching the backend refresh TTL (30 days). */
const REFRESH_MAX_AGE = 60 * 60 * 24 * 30

/**
 * Base URL of the Go backend. Prefers the server-only API_URL and falls back to
 * the public one so a single var works in local dev.
 */
export function backendURL(): string {
  const url = process.env.API_URL ?? process.env.NEXT_PUBLIC_API_URL
  if (!url) {
    throw new Error("API_URL / NEXT_PUBLIC_API_URL is not set")
  }
  return url.replace(/\/$/, "")
}

/** Sets the httpOnly refresh-token cookie. */
export async function setRefreshCookie(token: string): Promise<void> {
  const jar = await cookies()
  jar.set(REFRESH_COOKIE, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: REFRESH_MAX_AGE,
  })
}

/** Reads the refresh token from the httpOnly cookie, or null if absent. */
export async function getRefreshCookie(): Promise<string | null> {
  const jar = await cookies()
  return jar.get(REFRESH_COOKIE)?.value ?? null
}

/** Clears the refresh-token cookie (logout). */
export async function clearRefreshCookie(): Promise<void> {
  const jar = await cookies()
  jar.delete(REFRESH_COOKIE)
}
