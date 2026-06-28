-- name: GetLinkBySlug :one
-- Active (non-deleted) link by slug. Used by the redirect hot path.
SELECT id, user_id, slug, original_url, is_custom, expires_at, created_at, updated_at
FROM links
WHERE slug = $1 AND deleted_at IS NULL;

-- name: GetLinksByUserID :many
-- Paginated list of a user's active links. An optional case-insensitive search
-- (sqlc.narg('search')) filters by slug or destination URL. Sorting is driven by
-- two boolean flags so the four (column × direction) combinations resolve in SQL
-- without dynamic string building: sort_expires picks expires_at over created_at,
-- and sort_asc picks ascending over descending.
SELECT id, user_id, slug, original_url, is_custom, expires_at, created_at, updated_at
FROM links
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR slug ILIKE '%' || sqlc.narg('search') || '%'
    OR original_url ILIKE '%' || sqlc.narg('search') || '%'
  )
ORDER BY
  CASE WHEN sqlc.arg('sort_expires')::bool AND sqlc.arg('sort_asc')::bool THEN expires_at END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_expires')::bool AND NOT sqlc.arg('sort_asc')::bool THEN expires_at END DESC NULLS LAST,
  CASE WHEN NOT sqlc.arg('sort_expires')::bool AND sqlc.arg('sort_asc')::bool THEN created_at END ASC,
  CASE WHEN NOT sqlc.arg('sort_expires')::bool AND NOT sqlc.arg('sort_asc')::bool THEN created_at END DESC
LIMIT sqlc.arg('result_limit') OFFSET sqlc.arg('result_offset');

-- name: CountLinksByUserID :one
-- Total active links for a user, honoring the same optional search filter as the
-- list query so pagination totals stay consistent.
SELECT COUNT(*)
FROM links
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR slug ILIKE '%' || sqlc.narg('search') || '%'
    OR original_url ILIKE '%' || sqlc.narg('search') || '%'
  );

-- name: GetLinkBySlugAndUser :one
-- A user's single active link by slug. Used for detail, update, and delete so
-- ownership is enforced in the query itself.
SELECT id, user_id, slug, original_url, is_custom, expires_at, created_at, updated_at
FROM links
WHERE slug = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: SlugExists :one
-- Reports whether an active link already uses this slug. Used by slug generation
-- (collision retry) and custom-alias validation.
SELECT EXISTS (
  SELECT 1 FROM links WHERE slug = $1 AND deleted_at IS NULL
);

-- name: CreateLink :one
INSERT INTO links (user_id, slug, original_url, is_custom, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, slug, original_url, is_custom, expires_at, created_at, updated_at;

-- name: UpdateLink :one
-- Updates mutable fields of an active link. NULL arguments leave a field unchanged
-- (handled via COALESCE) except expires_at, which is set verbatim so it can be cleared.
UPDATE links
SET original_url = COALESCE(sqlc.narg('original_url'), original_url),
    slug         = COALESCE(sqlc.narg('slug'), slug),
    expires_at   = sqlc.narg('expires_at'),
    updated_at   = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, user_id, slug, original_url, is_custom, expires_at, created_at, updated_at;

-- name: SoftDeleteLink :exec
UPDATE links
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
