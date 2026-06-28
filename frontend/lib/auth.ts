// Access-token storage. The access token lives in a module-scoped variable only
// — never localStorage or a cookie reachable by JS (CLAUDE.md). It is lost on a
// hard refresh and recovered via a silent refresh against the httpOnly refresh
// cookie. The refresh token itself is never held here; it lives in an httpOnly
// cookie set by the Next.js /api route layer.

let accessToken: string | null = null

/** Returns the current in-memory access token, or null if unauthenticated. */
export function getAccessToken(): string | null {
  return accessToken
}

/** Stores the access token in memory. Pass null to clear it (logout). */
export function setAccessToken(token: string | null): void {
  accessToken = token
}
