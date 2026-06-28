-- name: GetLinkBySlug :one
-- Active (non-deleted) link by slug. Used by the redirect hot path.
SELECT id, user_id, slug, original_url, is_custom, expires_at, created_at, updated_at
FROM links
WHERE slug = $1 AND deleted_at IS NULL;

-- name: GetLinksByUserID :many
-- Paginated list of a user's active links, newest first.
SELECT id, user_id, slug, original_url, is_custom, expires_at, created_at, updated_at
FROM links
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

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
