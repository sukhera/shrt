// Typed API client. This is the only place the browser calls fetch (CLAUDE.md).
//
// Two destinations:
//   - Same-origin Next.js routes under /api/* for auth (they manage the httpOnly
//     refresh cookie server-side).
//   - The Go backend under NEXT_PUBLIC_API_URL for links, with the in-memory
//     access token attached as a Bearer header.
//
// On a 401 from the backend the client attempts a single silent refresh and
// retries once, so an expired access token is transparently renewed.

import { getAccessToken, setAccessToken } from "@/lib/auth"
import type {
  ApiErrorBody,
  CreateLinkInput,
  Credentials,
  Link,
  ListLinksParams,
  ListLinksResponse,
  RegisterResponse,
  UpdateLinkInput,
  User,
} from "@/types/api"

/** A typed error carrying the backend's error envelope fields. */
export class ApiError extends Error {
  readonly code: string
  readonly status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.name = "ApiError"
    this.code = code
    this.status = status
  }
}

function backendBase(): string {
  const url = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"
  return url.replace(/\/$/, "")
}

/** Parses an error response body into an ApiError, with sensible fallbacks. */
async function toApiError(res: Response): Promise<ApiError> {
  try {
    const body = (await res.json()) as ApiErrorBody
    if (body.error) {
      return new ApiError(body.error.code, body.error.message, body.error.status)
    }
  } catch {
    // fall through to a generic error
  }
  return new ApiError("INTERNAL", "Something went wrong. Please try again.", res.status)
}

/** Decodes a successful response, tolerating empty (204) bodies. */
async function decode<T>(res: Response): Promise<T> {
  if (res.status === 204) {
    return undefined as T
  }
  return (await res.json()) as T
}

/**
 * Attempts a silent token refresh via the Next.js /api/refresh route. On success
 * the new access token is stored in memory and returned; on failure the access
 * token is cleared and null is returned.
 */
async function refreshAccessToken(): Promise<string | null> {
  const res = await fetch("/api/refresh", { method: "POST" })
  if (!res.ok) {
    setAccessToken(null)
    return null
  }
  const data = (await res.json()) as { access_token: string }
  setAccessToken(data.access_token)
  return data.access_token
}

/**
 * Calls a backend endpoint with the access token attached. On 401 it refreshes
 * once and retries; a second 401 surfaces as an ApiError.
 */
async function backendFetch<T>(
  path: string,
  init: RequestInit = {},
  retry = true,
): Promise<T> {
  const token = getAccessToken()
  const headers = new Headers(init.headers)
  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }

  const res = await fetch(`${backendBase()}${path}`, { ...init, headers })

  if (res.status === 401 && retry) {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      return backendFetch<T>(path, init, false)
    }
  }

  if (!res.ok) {
    throw await toApiError(res)
  }
  return decode<T>(res)
}

/** Calls a same-origin Next.js /api route (auth proxy). */
async function localFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json")
  }
  const token = getAccessToken()
  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }
  const res = await fetch(path, { ...init, headers })
  if (!res.ok) {
    throw await toApiError(res)
  }
  return decode<T>(res)
}

// ── Auth ────────────────────────────────────────────────────────────────────

export const auth = {
  async register(creds: Credentials): Promise<{ user: User; access_token: string }> {
    return localFetch("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(creds),
    })
  },

  async login(creds: Credentials): Promise<{ access_token: string; expires_in: number }> {
    return localFetch("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(creds),
    })
  },

  async logout(): Promise<void> {
    await localFetch<void>("/api/auth/logout", { method: "POST" })
    setAccessToken(null)
  },

  /** Restores a session on app load using the httpOnly refresh cookie. */
  async refresh(): Promise<string | null> {
    return refreshAccessToken()
  },
}

// ── Links ─────────────────────────────────────────────────────────────────

export const links = {
  async create(input: CreateLinkInput): Promise<Link> {
    return backendFetch<Link>("/api/v1/links", {
      method: "POST",
      body: JSON.stringify(input),
    })
  },

  async list(params: ListLinksParams = {}): Promise<ListLinksResponse> {
    const search = new URLSearchParams()
    if (params.page) search.set("page", String(params.page))
    if (params.limit) search.set("limit", String(params.limit))
    if (params.sort) search.set("sort", params.sort)
    if (params.order) search.set("order", params.order)
    if (params.q) search.set("q", params.q)
    const qs = search.toString()
    return backendFetch<ListLinksResponse>(`/api/v1/links${qs ? `?${qs}` : ""}`)
  },

  async get(slug: string): Promise<Link> {
    return backendFetch<Link>(`/api/v1/links/${encodeURIComponent(slug)}`)
  },

  async update(slug: string, input: UpdateLinkInput): Promise<Link> {
    return backendFetch<Link>(`/api/v1/links/${encodeURIComponent(slug)}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    })
  },

  async remove(slug: string): Promise<void> {
    return backendFetch<void>(`/api/v1/links/${encodeURIComponent(slug)}`, {
      method: "DELETE",
    })
  },
}
