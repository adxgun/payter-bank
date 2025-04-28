-- name: SaveUser :one
INSERT INTO users(
    email, password, first_name, last_name, user_type
) VALUES ($1, $2, $3, $4, $5) RETURNING id, email;

-- name: GetUserByEmail :one
SELECT id, email, password, first_name, last_name, user_type, created_at, updated_at
    FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT id, email, password, first_name, last_name, user_type, created_at, updated_at
    FROM users WHERE id = $1 LIMIT 1;

-- name: GetProfileByUserID :one
SELECT
    accounts.id AS account_id,
    users.id AS user_id,
    users.email AS email,
    users.first_name AS first_name,
    users.last_name AS last_name,
    accounts.account_type AS account_type,
    users.user_type AS user_type,
    users.created_at AS registered_at
FROM users
         JOIN accounts ON users.id = accounts.user_id
WHERE users.id = $1 LIMIT 1;

-- name: SaveAccount :one
INSERT INTO accounts(
    user_id, account_number, status, account_type, currency
) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetAccountByID :one
SELECT id, user_id, account_number, status, account_type, currency, created_at, updated_at
    FROM accounts WHERE id = $1 LIMIT 1;

-- name: UpdateAccountStatus :exec
UPDATE accounts
    SET status = $1, updated_at = CURRENT_TIMESTAMP
    WHERE id = $2;

-- name: GetAllActiveAccounts :many
SELECT
    users.id as user_id,
    accounts.id as account_id,
    account_number,
    status,
    account_type,
    currency
    FROM accounts
    JOIN users ON users.id = accounts.user_id
WHERE users.user_type='CUSTOMER'
    AND accounts.status = 'ACTIVE'
    AND accounts.account_type = 'CURRENT';