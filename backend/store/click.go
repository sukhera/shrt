package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sukhera/shrt/backend/db"
)

// ClickDay is a single day's click count for the stats API.
type ClickDay struct {
	Day   time.Time `json:"day"`
	Count int       `json:"count"`
}

// LinkStats is the response shape for GET /api/v1/links/{id}/stats.
type LinkStats struct {
	LinkID string     `json:"link_id"`
	Total  int64      `json:"total"`
	Daily  []ClickDay `json:"daily"`
}

// RecordClick fires an async upsert to increment today's click count for a
// link. It never blocks the caller — errors are logged and dropped. The linkID
// is the UUID string from the redirect-path Link struct.
func (s *Store) RecordClick(ctx context.Context, linkID string) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		uid := mustUUID(linkID)
		if err := s.queries.UpsertClickDaily(bgCtx, uid); err != nil {
			slog.Warn("click upsert failed", "link_id", linkID, "err", err)
		}
	}()
}

// GetLinkStats returns 30-day click data for a link. Zero-days are filled in
// so the sparkline always has 30 data points.
func (s *Store) GetLinkStats(ctx context.Context, linkID string) (*LinkStats, error) {
	uid := mustUUID(linkID)

	total, err := s.queries.GetTotalClicks(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get total clicks: %w", err)
	}

	rows, err := s.queries.GetClickStats(ctx, db.GetClickStatsParams{
		LinkID:   uid,
		DaysBack: 30,
	})
	if err != nil {
		return nil, fmt.Errorf("get click stats: %w", err)
	}

	daily := fillZeroDays(rows, 30)

	return &LinkStats{
		LinkID: linkID,
		Total:  total,
		Daily:  daily,
	}, nil
}

// GetClickCountsByUser returns a map of link_id → total clicks for all of a
// user's active links. Used by the list endpoint to populate click_count
// without N+1 queries.
func (s *Store) GetClickCountsByUser(ctx context.Context, userID string) (map[string]int64, error) {
	uid, err := toUUID(&userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.GetTotalClicksByUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get click counts by user: %w", err)
	}

	counts := make(map[string]int64, len(rows))
	for _, r := range rows {
		counts[uuidString(r.LinkID)] = r.Total
	}
	return counts, nil
}

// fillZeroDays produces a slice of exactly `days` entries ending at today,
// filling in 0 for days with no clicks.
func fillZeroDays(rows []db.GetClickStatsRow, days int) []ClickDay {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	start := today.AddDate(0, 0, -(days - 1))

	// Index rows by day for O(1) lookup
	byDay := make(map[string]int32, len(rows))
	for _, r := range rows {
		key := r.Day.Format("2006-01-02")
		byDay[key] = r.Count
	}

	result := make([]ClickDay, days)
	for i := range days {
		d := start.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		result[i] = ClickDay{Day: d, Count: int(byDay[key])}
	}
	return result
}

// uuidString extracts the string from a pgtype.UUID. Panics if invalid.
func uuidString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return u.String()
}
