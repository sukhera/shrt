-- name: UpsertClickDaily :exec
-- Idempotent daily click increment. Called async on each redirect.
INSERT INTO clicks_daily (link_id, day, count)
VALUES ($1, CURRENT_DATE, 1)
ON CONFLICT (link_id, day) DO UPDATE SET count = clicks_daily.count + 1;

-- name: GetClickStats :many
-- Last N days of click data for a link, ordered chronologically.
-- Returns rows only for days with clicks; the caller fills zero-days.
SELECT day, count
FROM clicks_daily
WHERE link_id = $1
  AND day >= CURRENT_DATE - sqlc.arg('days_back')::int
ORDER BY day ASC;

-- name: GetTotalClicks :one
-- Lifetime click count for a single link.
SELECT COALESCE(SUM(count), 0)::bigint AS total
FROM clicks_daily
WHERE link_id = $1;

-- name: GetTotalClicksByUser :many
-- Total clicks per link for a user's links. Used to populate click_count
-- in the dashboard list without N+1 queries.
SELECT cd.link_id, COALESCE(SUM(cd.count), 0)::bigint AS total
FROM clicks_daily cd
JOIN links l ON l.id = cd.link_id
WHERE l.user_id = $1 AND l.deleted_at IS NULL
GROUP BY cd.link_id;
