-- name: CreateRefreshToken :one
-- Stores a refresh token by its SHA-256 hash (never the plaintext token).
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at;

-- name: GetRefreshTokenByHash :one
-- Looks up a refresh token by its hash. The caller checks expiry and revocation.
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
-- Marks a refresh token revoked (idempotent: only sets revoked_at if still active).
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: DeleteExpiredTokens :exec
-- Housekeeping: removes tokens past their expiry. Intended for a periodic job.
DELETE FROM refresh_tokens
WHERE expires_at < NOW();
