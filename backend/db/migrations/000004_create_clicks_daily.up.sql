-- Per-link daily click rollups for sparkline data and aggregate counts.
-- Upserted asynchronously on each redirect — never blocks the 301/302.

CREATE TABLE IF NOT EXISTS clicks_daily (
    link_id  UUID        NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    day      DATE        NOT NULL,
    count    INTEGER     NOT NULL DEFAULT 0,
    PRIMARY KEY (link_id, day)
);

-- Fast per-link lookups for the sparkline (last 30 days).
CREATE INDEX IF NOT EXISTS idx_clicks_daily_link_day
    ON clicks_daily (link_id, day DESC);
