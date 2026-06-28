package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sukhera/shrt/backend/db"
)

// aliasPattern constrains custom aliases to URL-safe characters. Length bounds
// are checked separately so we can return a clear message.
var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

const (
	aliasMinLen = 3
	aliasMaxLen = 64
)

// LinkDetail is the full API view of a link, returned by the CRUD endpoints.
// Unlike the redirect-path Link, it carries identity and audit fields.
type LinkDetail struct {
	ID          string
	UserID      *string
	Slug        string
	OriginalURL string
	IsCustom    bool
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateLinkInput carries the validated inputs for creating a link. UserID is
// nil for anonymous shortens. Alias is an optional custom slug.
type CreateLinkInput struct {
	URL       string
	Alias     string
	ExpiresAt *time.Time
	UserID    *string
}

// ListLinksInput parameterises a paginated, optionally searched/sorted list of a
// user's links.
type ListLinksInput struct {
	UserID string
	Page   int
	Limit  int
	Sort   string // "created_at" | "expires_at"
	Order  string // "asc" | "desc"
	Search string
}

// UpdateLinkInput carries the fields a PATCH may change. A nil pointer leaves the
// field unchanged; ClearExpiry distinguishes "set expiry to null" from "leave
// expiry alone" since ExpiresAt being nil already means the latter.
type UpdateLinkInput struct {
	URL         *string
	Alias       *string
	ExpiresAt   *time.Time
	ClearExpiry bool
}

// CreateLink validates and persists a new link. It checks the URL, runs Safe
// Browsing, resolves the slug (custom alias or generated), and inserts the row.
func (s *Store) CreateLink(ctx context.Context, in CreateLinkInput) (*LinkDetail, error) {
	if err := validateURL(in.URL); err != nil {
		return nil, err
	}
	if err := s.checkSafeBrowsing(ctx, in.URL); err != nil {
		return nil, err
	}

	slug, isCustom, err := s.resolveSlug(ctx, in.Alias)
	if err != nil {
		return nil, err
	}

	userID, err := toUUID(in.UserID)
	if err != nil {
		return nil, err
	}

	row, err := s.queries.CreateLink(ctx, db.CreateLinkParams{
		UserID:      userID,
		Slug:        slug,
		OriginalUrl: in.URL,
		IsCustom:    isCustom,
		ExpiresAt:   toTimestamptz(in.ExpiresAt),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAliasTaken
		}
		return nil, fmt.Errorf("create link: %w", err)
	}
	return createRowToDetail(row), nil
}

// resolveSlug returns the slug to use and whether it is a custom alias. An empty
// alias triggers random generation; a non-empty alias is validated and checked
// for collisions.
func (s *Store) resolveSlug(ctx context.Context, alias string) (string, bool, error) {
	if alias == "" {
		slug, err := s.generateSlug(ctx)
		if err != nil {
			return "", false, err
		}
		return slug, false, nil
	}

	if err := validateAlias(alias); err != nil {
		return "", false, err
	}
	exists, err := s.queries.SlugExists(ctx, alias)
	if err != nil {
		return "", false, fmt.Errorf("check alias existence: %w", err)
	}
	if exists {
		return "", false, ErrAliasTaken
	}
	return alias, true, nil
}

// ListByUser returns a page of the user's links plus the total matching count.
func (s *Store) ListByUser(ctx context.Context, in ListLinksInput) ([]LinkDetail, int, error) {
	userID, err := toUUID(&in.UserID)
	if err != nil {
		return nil, 0, err
	}

	search := pgtype.Text{}
	if in.Search != "" {
		search = pgtype.Text{String: in.Search, Valid: true}
	}

	rows, err := s.queries.GetLinksByUserID(ctx, db.GetLinksByUserIDParams{
		UserID:       userID,
		Search:       search,
		SortExpires:  in.Sort == "expires_at",
		SortAsc:      in.Order == "asc",
		ResultLimit:  clampInt32(in.Limit),
		ResultOffset: clampInt32((in.Page - 1) * in.Limit),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list links: %w", err)
	}

	total, err := s.queries.CountLinksByUserID(ctx, db.CountLinksByUserIDParams{
		UserID: userID,
		Search: search,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count links: %w", err)
	}

	details := make([]LinkDetail, len(rows))
	for i, row := range rows {
		details[i] = *listRowToDetail(row)
	}
	return details, int(total), nil
}

// GetDetailBySlug returns a single link owned by userID. It returns ErrNotFound
// when no active link with that slug belongs to the user.
func (s *Store) GetDetailBySlug(ctx context.Context, userID, slug string) (*LinkDetail, error) {
	uid, err := toUUID(&userID)
	if err != nil {
		return nil, err
	}
	row, err := s.queries.GetLinkBySlugAndUser(ctx, db.GetLinkBySlugAndUserParams{Slug: slug, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get link detail: %w", err)
	}
	return slugUserRowToDetail(row), nil
}

// UpdateLink applies a partial update to a user's link, enforcing ownership via
// the slug+user lookup. A changed destination is re-validated and re-checked
// against Safe Browsing; a changed alias is validated for collisions. The Redis
// cache entry for the old (and, if renamed, new) slug is invalidated.
func (s *Store) UpdateLink(ctx context.Context, userID, slug string, in UpdateLinkInput) (*LinkDetail, error) {
	current, err := s.GetDetailBySlug(ctx, userID, slug)
	if err != nil {
		return nil, err
	}

	params := db.UpdateLinkParams{ID: mustUUID(current.ID)}

	if in.URL != nil {
		if err := validateURL(*in.URL); err != nil {
			return nil, err
		}
		if err := s.checkSafeBrowsing(ctx, *in.URL); err != nil {
			return nil, err
		}
		params.OriginalUrl = pgtype.Text{String: *in.URL, Valid: true}
	}

	newSlug := current.Slug
	if in.Alias != nil && *in.Alias != current.Slug {
		if err := validateAlias(*in.Alias); err != nil {
			return nil, err
		}
		exists, err := s.queries.SlugExists(ctx, *in.Alias)
		if err != nil {
			return nil, fmt.Errorf("check alias existence: %w", err)
		}
		if exists {
			return nil, ErrAliasTaken
		}
		params.Slug = pgtype.Text{String: *in.Alias, Valid: true}
		newSlug = *in.Alias
	}

	// expires_at is set verbatim by the query (COALESCE is not used for it), so
	// we must always pass the intended value: the new expiry, an explicit clear,
	// or the existing value when the field is untouched.
	switch {
	case in.ExpiresAt != nil:
		params.ExpiresAt = toTimestamptz(in.ExpiresAt)
	case in.ClearExpiry:
		params.ExpiresAt = pgtype.Timestamptz{}
	default:
		params.ExpiresAt = toTimestamptz(current.ExpiresAt)
	}

	row, err := s.queries.UpdateLink(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAliasTaken
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update link: %w", err)
	}

	s.invalidateSlugCache(ctx, current.Slug)
	if newSlug != current.Slug {
		s.invalidateSlugCache(ctx, newSlug)
	}
	return updateRowToDetail(row), nil
}

// DeleteLink soft-deletes a user's link and invalidates its cache entry. It
// returns ErrNotFound when the user owns no active link with that slug.
func (s *Store) DeleteLink(ctx context.Context, userID, slug string) error {
	current, err := s.GetDetailBySlug(ctx, userID, slug)
	if err != nil {
		return err
	}
	if err := s.queries.SoftDeleteLink(ctx, mustUUID(current.ID)); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	s.invalidateSlugCache(ctx, slug)
	return nil
}

// invalidateSlugCache removes the cached entry for a slug. Failures are logged
// and ignored — a stale cache entry self-heals at its TTL.
func (s *Store) invalidateSlugCache(ctx context.Context, slug string) {
	if err := s.rdb.Del(ctx, slugCacheKey(slug)).Err(); err != nil {
		slog.Warn("redis del failed", "slug", slug, "err", err)
	}
}

// validateURL rejects malformed URLs and non-HTTP(S) schemes.
func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https", ErrInvalidURL)
	}
	if u.Host == "" {
		return fmt.Errorf("%w: missing host", ErrInvalidURL)
	}
	return nil
}

// validateAlias enforces the custom-alias character set and length bounds.
func validateAlias(alias string) error {
	if len(alias) < aliasMinLen || len(alias) > aliasMaxLen {
		return fmt.Errorf("%w: alias must be %d–%d characters", ErrInvalidURL, aliasMinLen, aliasMaxLen)
	}
	if !aliasPattern.MatchString(alias) {
		return fmt.Errorf("%w: alias may contain only letters, digits, hyphen, underscore", ErrInvalidURL)
	}
	return nil
}
