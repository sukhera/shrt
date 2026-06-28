CREATE TABLE IF NOT EXISTS links (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        REFERENCES users(id) ON DELETE SET NULL,
    slug         TEXT        NOT NULL,
    original_url TEXT        NOT NULL,
    is_custom    BOOLEAN     NOT NULL DEFAULT FALSE,
    expires_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Redirect hot path: fast slug lookup; UNIQUE only among active (non-deleted)
-- links so a slug can be recycled after its link is soft-deleted.
CREATE UNIQUE INDEX IF NOT EXISTS idx_links_slug_active
    ON links (slug)
    WHERE deleted_at IS NULL;

-- Dashboard list queries by owner
CREATE INDEX IF NOT EXISTS idx_links_user_id
    ON links (user_id)
    WHERE deleted_at IS NULL;
