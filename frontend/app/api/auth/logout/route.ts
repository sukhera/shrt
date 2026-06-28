import { NextRequest, NextResponse } from "next/server"

import {
  backendURL,
  clearRefreshCookie,
  getRefreshCookie,
} from "@/lib/server/backend"

// Revokes the refresh token on the backend and clears the httpOnly cookie. The
// backend's logout endpoint is auth-guarded by the access token (forwarded from
// the client's Authorization header); the refresh token to revoke comes from the
// httpOnly cookie. It always clears the cookie so the user is logged out locally
// even if the backend call fails.
export async function POST(req: NextRequest) {
  const refreshToken = await getRefreshCookie()
  const authHeader = req.headers.get("Authorization")

  if (refreshToken && authHeader) {
    try {
      await fetch(`${backendURL()}/api/v1/auth/logout`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: authHeader,
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })
    } catch {
      // Ignore backend errors — the local cookie clear below is what matters.
    }
  }

  await clearRefreshCookie()
  return new NextResponse(null, { status: 204 })
}
