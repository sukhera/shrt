import { NextRequest, NextResponse } from "next/server"

import { backendURL, setRefreshCookie } from "@/lib/server/backend"
import type { ApiErrorBody, LoginResponse } from "@/types/api"

// Proxies login to the Go backend, stores the returned refresh token in an
// httpOnly cookie, and returns only the access token (+ expiry) to the browser.
export async function POST(req: NextRequest) {
  const body = await req.text()

  const res = await fetch(`${backendURL()}/api/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
  })

  if (!res.ok) {
    const err = (await res.json()) as ApiErrorBody
    return NextResponse.json(err, { status: res.status })
  }

  const data = (await res.json()) as LoginResponse
  await setRefreshCookie(data.refresh_token)

  return NextResponse.json({
    access_token: data.access_token,
    expires_in: data.expires_in,
  })
}
