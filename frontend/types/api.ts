// Shared API types for the shrt backend. These mirror the contract in
// IMPLEMENTATION-PLAN.md § 4 — keep them in sync with the Go handlers.

/** A shortened link, as returned by the links endpoints. */
export interface Link {
  id: string
  slug: string
  short_url: string
  original_url: string
  is_custom: boolean
  expires_at: string | null
  created_at: string
  updated_at: string
  /** Total click count — present once the click rollup backend (R5) is deployed. */
  click_count?: number
}

/** Pagination metadata returned by GET /links. */
export interface Pagination {
  page: number
  limit: number
  total: number
}

/** GET /links response body. */
export interface ListLinksResponse {
  data: Link[]
  pagination: Pagination
}

/** Query parameters accepted by GET /links. */
export interface ListLinksParams {
  page?: number
  limit?: number
  sort?: "created_at" | "expires_at"
  order?: "asc" | "desc"
  q?: string
}

/** POST /links request body. */
export interface CreateLinkInput {
  url: string
  alias?: string
  expires_at?: string | null
}

/** PATCH /links/:slug request body — any subset of fields. */
export interface UpdateLinkInput {
  url?: string
  alias?: string
  expires_at?: string | null
}

/** Public user shape returned by auth endpoints. */
export interface User {
  id: string
  email: string
}

/** POST /auth/register response body. */
export interface RegisterResponse {
  user: User
  access_token: string
  refresh_token: string
}

/** POST /auth/login response body. */
export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

/** POST /auth/refresh response body. */
export interface RefreshResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

/** Credentials shared by register and login. */
export interface Credentials {
  email: string
  password: string
}

/** The error envelope returned for all API errors (contract § 4.2). */
export interface ApiErrorBody {
  error: {
    code: string
    message: string
    status: number
  }
}
