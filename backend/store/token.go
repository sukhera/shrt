package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sukhera/shrt/backend/db"
)

// AccessClaims are the custom claims embedded in an access token. RegisteredClaims
// supplies the standard sub/exp/iat fields.
type AccessClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair is the result of authenticating: a short-lived access token and a
// long-lived opaque refresh token (returned in plaintext to the client once;
// only its hash is stored).
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // access-token lifetime in seconds
}

// createAccessToken signs an RS256 access token for the user, embedding the role
// claim used by admin authorization (S-07).
func (s *Store) createAccessToken(userID, role string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.cfg.JWTAccessTTL)
	claims := AccessClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.jwtPrivate)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

// ParseAccessToken validates an access token's signature and expiry and returns
// its claims. Any failure is normalised to ErrInvalidToken so handlers respond
// 401 without leaking the specific cause.
func (s *Store) ParseAccessToken(raw string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtPublic, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return claims, nil
}

// newRefreshToken generates a cryptographically random opaque refresh token and
// returns it alongside its SHA-256 hash. Only the hash is persisted (S-10).
func newRefreshToken() (token, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("read random bytes: %w", err)
	}
	token = hex.EncodeToString(buf)
	hash = hashToken(token)
	return token, hash, nil
}

// hashToken returns the hex-encoded SHA-256 of a refresh token. Lookups hash the
// presented token and compare against the stored hash.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// issueTokenPair creates an access token plus a stored refresh token for a user.
// Used by register, login, and refresh.
func (s *Store) issueTokenPair(ctx context.Context, userID, role string) (*TokenPair, error) {
	access, _, err := s.createAccessToken(userID, role)
	if err != nil {
		return nil, err
	}

	refresh, hash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	uid, err := toUUID(&userID)
	if err != nil {
		return nil, err
	}
	if _, err := s.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    uid,
		TokenHash: hash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(s.cfg.JWTRefreshTTL), Valid: true},
	}); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.cfg.JWTAccessTTL.Seconds()),
	}, nil
}

// IssueAccessToken signs a standalone access token for a user and role, without
// creating a refresh token. It is used by tests and by callers that already hold
// a validated session and only need a fresh access token.
func (s *Store) IssueAccessToken(userID, role string) (string, error) {
	token, _, err := s.createAccessToken(userID, role)
	return token, err
}

// RefreshToken validates a refresh token, then issues a new token pair and
// revokes the old refresh token (rotation). An unknown, expired, or revoked
// token yields ErrInvalidToken.
func (s *Store) RefreshToken(ctx context.Context, presented string) (*TokenPair, error) {
	hash := hashToken(presented)
	row, err := s.queries.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("lookup refresh token: %w", err)
	}

	if row.RevokedAt.Valid || !row.ExpiresAt.Valid || !row.ExpiresAt.Time.After(time.Now()) {
		return nil, ErrInvalidToken
	}

	user, err := s.queries.GetUserByID(ctx, row.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("load user for refresh: %w", err)
	}

	// Rotate: revoke the presented token before issuing a replacement so a leaked
	// refresh token cannot be reused after a legitimate refresh.
	if err := s.queries.RevokeRefreshToken(ctx, hash); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	return s.issueTokenPair(ctx, row.UserID.String(), user.Role)
}

// Logout revokes a refresh token. Revoking an unknown or already-revoked token
// is a no-op (the user is logged out either way), so it never errors on that.
func (s *Store) Logout(ctx context.Context, presented string) error {
	if err := s.queries.RevokeRefreshToken(ctx, hashToken(presented)); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}
