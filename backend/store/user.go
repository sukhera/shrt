package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sukhera/shrt/backend/db"
)

// bcryptCost is the work factor for password hashing. The PRD requires ≥ 12
// (S-05).
const bcryptCost = 12

// minPasswordLen is the minimum accepted password length (API contract).
const minPasswordLen = 8

// User is the public view of a user account — never includes the password hash.
type User struct {
	ID        string
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AuthResult bundles a user with a freshly issued token pair, returned by
// register and login.
type AuthResult struct {
	User   User
	Tokens TokenPair
}

// Register creates a new user with a bcrypt-hashed password and issues an
// initial token pair. It returns ErrEmailTaken if the email already exists and
// ErrInvalidCredentials if the password is too short.
func (s *Store) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	email = normalizeEmail(email)
	if err := validateCredentials(email, password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	row, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	userID := row.ID.String()
	tokens, err := s.issueTokenPair(ctx, userID, row.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User: User{
			ID:        userID,
			Email:     row.Email,
			Role:      row.Role,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Tokens: *tokens,
	}, nil
}

// Login verifies an email/password pair and issues a token pair. Unknown email
// and wrong password both return ErrInvalidCredentials (no account enumeration).
func (s *Store) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = normalizeEmail(email)

	row, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Hash a throwaway password so the response time does not reveal
			// whether the email exists (timing-attack mitigation).
			_ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(password))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	userID := row.ID.String()
	tokens, err := s.issueTokenPair(ctx, userID, row.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User: User{
			ID:        userID,
			Email:     row.Email,
			Role:      row.Role,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Tokens: *tokens,
	}, nil
}

// dummyBcryptHash is a valid bcrypt hash of a random string, used only to spend
// comparable CPU time on the unknown-email path so login timing is uniform.
const dummyBcryptHash = "$2a$12$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// validateCredentials enforces the minimal email/password rules shared by
// register. Email format is checked loosely (presence of "@" with a non-empty
// local and domain part); the database UNIQUE constraint is the source of truth
// for uniqueness.
func validateCredentials(email, password string) error {
	at := strings.IndexByte(email, '@')
	if at <= 0 || at == len(email)-1 || strings.Contains(email[at+1:], "@") {
		return fmt.Errorf("%w: email is not valid", ErrInvalidCredentials)
	}
	if len(password) < minPasswordLen {
		return fmt.Errorf("%w: password must be at least %d characters", ErrInvalidCredentials, minPasswordLen)
	}
	return nil
}

// normalizeEmail lowercases and trims an email so lookups and uniqueness are
// case-insensitive.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
