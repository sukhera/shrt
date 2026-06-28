import { NextRequest, NextResponse } from "next/server"

import { backendURL, setRefreshCookie } from "@/lib/server/backend"
import type { ApiErrorBody, RegisterResponse } from "@/types/api"

// Proxies registration to the Go backend, stores the refresh token in an
// httpOnly cookie, and returns the user + access token to the browser.
export async function POST(req: NextRequest) {
  const body = await req.text()

  const res = await fetch(`${backendURL()}/api/v1/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
  })

  if (!res.ok) {
    const err = (await res.json()) as ApiErrorBody
    return NextResponse.json(err, { status: res.status })
  }

  const data = (await res.json()) as RegisterResponse
  await setRefreshCookie(data.refresh_token)

  return NextResponse.json({
    user: data.user,
    access_token: data.access_token,
  })
}
