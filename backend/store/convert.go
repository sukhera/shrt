package store

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sukhera/shrt/backend/db"
)

// clampInt32 narrows a non-negative pagination value to int32, saturating at the
// bounds. Limits and offsets are already validated upstream, so this only guards
// against overflow on pathological input.
func clampInt32(n int) int32 {
	switch {
	case n < 0:
		return 0
	case n > math.MaxInt32:
		return math.MaxInt32
	default:
		return int32(n)
	}
}

// toUUID converts an optional UUID string into a pgtype.UUID. A nil pointer
// yields an invalid (NULL) UUID, used for anonymous links. A malformed string
// is reported as ErrNotFound — an unparseable id can never match a row, so
// callers treat it the same as a miss.
func toUUID(s *string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if s == nil || *s == "" {
		return u, nil
	}
	if err := u.Scan(*s); err != nil {
		return u, fmt.Errorf("%w: invalid uuid %q", ErrNotFound, *s)
	}
	return u, nil
}

// mustUUID parses a UUID string known to be valid (it came from the database).
// It panics on failure, signalling a programming error rather than bad input.
func mustUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		panic(fmt.Sprintf("store: mustUUID(%q): %v", s, err))
	}
	return u
}

// uuidToStringPtr returns the canonical string form of a UUID, or nil when the
// UUID is NULL (e.g. an anonymous link's user_id).
func uuidToStringPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := u.String()
	return &s
}

// toTimestamptz converts an optional time into a pgtype.Timestamptz. A nil
// pointer yields a NULL timestamp.
func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// timePtr returns a pointer to the time, or nil when the timestamp is NULL.
func timePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505) — used to translate a concurrent slug insert race
// into ErrAliasTaken.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// The row→detail converters below adapt the (near-identical) sqlc row types into
// the store's LinkDetail. They are intentionally separate functions because sqlc
// emits a distinct struct per query.

func createRowToDetail(r db.CreateLinkRow) *LinkDetail {
	return &LinkDetail{
		ID:          r.ID.String(),
		UserID:      uuidToStringPtr(r.UserID),
		Slug:        r.Slug,
		OriginalURL: r.OriginalUrl,
		IsCustom:    r.IsCustom,
		ExpiresAt:   timePtr(r.ExpiresAt),
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}
}

func updateRowToDetail(r db.UpdateLinkRow) *LinkDetail {
	return &LinkDetail{
		ID:          r.ID.String(),
		UserID:      uuidToStringPtr(r.UserID),
		Slug:        r.Slug,
		OriginalURL: r.OriginalUrl,
		IsCustom:    r.IsCustom,
		ExpiresAt:   timePtr(r.ExpiresAt),
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}
}

func listRowToDetail(r db.GetLinksByUserIDRow) *LinkDetail {
	return &LinkDetail{
		ID:          r.ID.String(),
		UserID:      uuidToStringPtr(r.UserID),
		Slug:        r.Slug,
		OriginalURL: r.OriginalUrl,
		IsCustom:    r.IsCustom,
		ExpiresAt:   timePtr(r.ExpiresAt),
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}
}

func slugUserRowToDetail(r db.GetLinkBySlugAndUserRow) *LinkDetail {
	return &LinkDetail{
		ID:          r.ID.String(),
		UserID:      uuidToStringPtr(r.UserID),
		Slug:        r.Slug,
		OriginalURL: r.OriginalUrl,
		IsCustom:    r.IsCustom,
		ExpiresAt:   timePtr(r.ExpiresAt),
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}
}
