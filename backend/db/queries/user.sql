-- name: CreateUser :one
-- Inserts a new user. The caller hashes the password before calling.
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id, email, role, created_at, updated_at;

-- name: GetUserByEmail :one
-- Fetches a user (including password_hash) by email for login.
SELECT id, email, password_hash, role, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
-- Fetches a user by id. Used after refresh-token validation to issue a new
-- access token with the current role.
SELECT id, email, role, created_at, updated_at
FROM users
WHERE id = $1;
