import { NextResponse } from "next/server"

import {
  backendURL,
  clearRefreshCookie,
  getRefreshCookie,
  setRefreshCookie,
} from "@/lib/server/backend"
import type { RefreshResponse } from "@/types/api"

// Exchanges the httpOnly refresh-token cookie for a fresh access token. The
// backend rotates the refresh token, so we store the new one back into the
// cookie. Used on app load (silent session restore) and on 401 retries.
export async function POST() {
  const refreshToken = await getRefreshCookie()
  if (!refreshToken) {
    return NextResponse.json(
      { error: { code: "UNAUTHORIZED", message: "No session.", status: 401 } },
      { status: 401 },
    )
  }

  const res = await fetch(`${backendURL()}/api/v1/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })

  if (!res.ok) {
    // The refresh token is invalid/expired/revoked — drop the stale cookie so
    // the client stops retrying with it.
    await clearRefreshCookie()
    return NextResponse.json(
      { error: { code: "UNAUTHORIZED", message: "Session expired.", status: 401 } },
      { status: 401 },
    )
  }

  const data = (await res.json()) as RefreshResponse
  await setRefreshCookie(data.refresh_token)

  return NextResponse.json({
    access_token: data.access_token,
    expires_in: data.expires_in,
  })
}
