-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: SaveUser :one
INSERT INTO users (email, password, external_auth_id, created_at, updated_at)
    VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    RETURNING id, email, external_auth_id, password, created_at, updated_at;

-- name: SaveAccount :one
INSERT INTO accounts (user_id, first_name, last_name, account_type, created_at, updated_at)
    VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    RETURNING id, user_id, first_name, last_name, account_type, created_at, updated_at;

-- name: GetAccountByUserID :one
SELECT * FROM accounts WHERE user_id = $1 LIMIT 1;

-- name: GetProfileByUserID :one
SELECT
    accounts.id AS account_id,
    users.id AS user_id,
    users.email AS email,
    accounts.first_name AS first_name,
    accounts.last_name AS last_name,
    accounts.account_type AS account_type,
    users.created_at AS registered_at
    FROM users
    JOIN accounts ON users.id = accounts.user_id
    WHERE users.id = $1 LIMIT 1;

-- name: UserWithAuthExists :one
SELECT
    accounts.id AS account_id,
    users.id AS user_id
    FROM users
    JOIN accounts ON users.id = accounts.user_id
    WHERE users.external_auth_id = $1 LIMIT 1;

-- name: GetAccountByID :one
SELECT * FROM accounts WHERE id = $1 LIMIT 1;