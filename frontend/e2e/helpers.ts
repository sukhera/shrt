import { type APIRequestContext, expect } from "@playwright/test"

// Direct backend base URL for seeding data and asserting redirect behaviour that
// the frontend doesn't surface (e.g. raw 410 on expired links).
export const BACKEND_URL = process.env.E2E_BACKEND_URL ?? "http://localhost:8080"

/** A unique email per test run so registration never collides. */
export function uniqueEmail(prefix = "e2e"): string {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1e6)}@example.com`
}

export const TEST_PASSWORD = "supersecret123"

/** Registers a user straight against the backend and returns the access token. */
export async function registerViaApi(
  request: APIRequestContext,
  email: string,
  password = TEST_PASSWORD,
): Promise<string> {
  const res = await request.post(`${BACKEND_URL}/api/v1/auth/register`, {
    data: { email, password },
  })
  expect(res.ok(), `register ${email}`).toBeTruthy()
  const body = await res.json()
  return body.access_token as string
}

/** Creates a link via the backend API, optionally as an authenticated user. */
export async function createLinkViaApi(
  request: APIRequestContext,
  url: string,
  opts: { token?: string; alias?: string; expiresAt?: string } = {},
): Promise<{ slug: string; short_url: string }> {
  const res = await request.post(`${BACKEND_URL}/api/v1/links`, {
    headers: opts.token ? { Authorization: `Bearer ${opts.token}` } : {},
    data: {
      url,
      ...(opts.alias ? { alias: opts.alias } : {}),
      ...(opts.expiresAt ? { expires_at: opts.expiresAt } : {}),
    },
  })
  expect(res.ok(), `create link ${url}`).toBeTruthy()
  return res.json()
}
