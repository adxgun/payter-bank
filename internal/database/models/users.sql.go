// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: users.sql

package models

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const getAccountByCurrency = `-- name: GetAccountByCurrency :one
SELECT id, user_id, account_number, account_type, status, currency, created_at, updated_at, deleted_at, balance FROM accounts WHERE currency = $1 AND user_id = $2 LIMIT 1
`

type GetAccountByCurrencyParams struct {
	Currency Currency  `json:"currency"`
	UserID   uuid.UUID `json:"user_id"`
}

func (q *Queries) GetAccountByCurrency(ctx context.Context, arg GetAccountByCurrencyParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccountByCurrency, arg.Currency, arg.UserID)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.AccountNumber,
		&i.AccountType,
		&i.Status,
		&i.Currency,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Balance,
	)
	return i, err
}

const getAccountByID = `-- name: GetAccountByID :one
SELECT id, user_id, account_number, status, account_type, currency, created_at, updated_at
    FROM accounts WHERE id = $1 LIMIT 1
`

type GetAccountByIDRow struct {
	ID            uuid.UUID    `json:"id"`
	UserID        uuid.UUID    `json:"user_id"`
	AccountNumber string       `json:"account_number"`
	Status        Status       `json:"status"`
	AccountType   AccountType  `json:"account_type"`
	Currency      Currency     `json:"currency"`
	CreatedAt     sql.NullTime `json:"created_at"`
	UpdatedAt     sql.NullTime `json:"updated_at"`
}

func (q *Queries) GetAccountByID(ctx context.Context, id uuid.UUID) (GetAccountByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getAccountByID, id)
	var i GetAccountByIDRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.AccountNumber,
		&i.Status,
		&i.AccountType,
		&i.Currency,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAccountDetailsByID = `-- name: GetAccountDetailsByID :one
SELECT
    users.id as user_id,
    accounts.id as account_id,
    users.first_name as first_name,
    users.last_name as last_name,
    account_number,
    status,
    account_type,
    currency,
    accounts.balance,
    accounts.created_at
FROM accounts
         JOIN users ON users.id = accounts.user_id
WHERE accounts.id = $1
`

type GetAccountDetailsByIDRow struct {
	UserID        uuid.UUID     `json:"user_id"`
	AccountID     uuid.UUID     `json:"account_id"`
	FirstName     string        `json:"first_name"`
	LastName      string        `json:"last_name"`
	AccountNumber string        `json:"account_number"`
	Status        Status        `json:"status"`
	AccountType   AccountType   `json:"account_type"`
	Currency      Currency      `json:"currency"`
	Balance       sql.NullInt64 `json:"balance"`
	CreatedAt     sql.NullTime  `json:"created_at"`
}

func (q *Queries) GetAccountDetailsByID(ctx context.Context, id uuid.UUID) (GetAccountDetailsByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getAccountDetailsByID, id)
	var i GetAccountDetailsByIDRow
	err := row.Scan(
		&i.UserID,
		&i.AccountID,
		&i.FirstName,
		&i.LastName,
		&i.AccountNumber,
		&i.Status,
		&i.AccountType,
		&i.Currency,
		&i.Balance,
		&i.CreatedAt,
	)
	return i, err
}

const getAccountStats = `-- name: GetAccountStats :one
SELECT
    (SELECT COUNT(*) FROM users) as total_users,
    COUNT(*) AS total,
    COUNT(CASE WHEN status = 'CLOSED' THEN 1 END) AS closed,
    COUNT(CASE WHEN status = 'SUSPENDED' THEN 1 END) AS suspended
FROM accounts
`

type GetAccountStatsRow struct {
	TotalUsers int64 `json:"total_users"`
	Total      int64 `json:"total"`
	Closed     int64 `json:"closed"`
	Suspended  int64 `json:"suspended"`
}

func (q *Queries) GetAccountStats(ctx context.Context) (GetAccountStatsRow, error) {
	row := q.db.QueryRowContext(ctx, getAccountStats)
	var i GetAccountStatsRow
	err := row.Scan(
		&i.TotalUsers,
		&i.Total,
		&i.Closed,
		&i.Suspended,
	)
	return i, err
}

const getAllActiveAccounts = `-- name: GetAllActiveAccounts :many
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
    AND accounts.account_type = 'CURRENT'
`

type GetAllActiveAccountsRow struct {
	UserID        uuid.UUID   `json:"user_id"`
	AccountID     uuid.UUID   `json:"account_id"`
	AccountNumber string      `json:"account_number"`
	Status        Status      `json:"status"`
	AccountType   AccountType `json:"account_type"`
	Currency      Currency    `json:"currency"`
}

func (q *Queries) GetAllActiveAccounts(ctx context.Context) ([]GetAllActiveAccountsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllActiveAccounts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllActiveAccountsRow
	for rows.Next() {
		var i GetAllActiveAccountsRow
		if err := rows.Scan(
			&i.UserID,
			&i.AccountID,
			&i.AccountNumber,
			&i.Status,
			&i.AccountType,
			&i.Currency,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllCurrentAccounts = `-- name: GetAllCurrentAccounts :many
SELECT
    u.first_name,
    u.last_name,
    u.id AS user_id,
    a.account_number,
    a.id AS account_id,
    a.balance,
    a.account_type,
    a.status,
    a.currency,
    a.created_at
FROM accounts a
         JOIN users u ON u.id = a.user_id
WHERE a.account_type = 'CURRENT'
ORDER BY a.created_at DESC
`

type GetAllCurrentAccountsRow struct {
	FirstName     string        `json:"first_name"`
	LastName      string        `json:"last_name"`
	UserID        uuid.UUID     `json:"user_id"`
	AccountNumber string        `json:"account_number"`
	AccountID     uuid.UUID     `json:"account_id"`
	Balance       sql.NullInt64 `json:"balance"`
	AccountType   AccountType   `json:"account_type"`
	Status        Status        `json:"status"`
	Currency      Currency      `json:"currency"`
	CreatedAt     sql.NullTime  `json:"created_at"`
}

func (q *Queries) GetAllCurrentAccounts(ctx context.Context) ([]GetAllCurrentAccountsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllCurrentAccounts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllCurrentAccountsRow
	for rows.Next() {
		var i GetAllCurrentAccountsRow
		if err := rows.Scan(
			&i.FirstName,
			&i.LastName,
			&i.UserID,
			&i.AccountNumber,
			&i.AccountID,
			&i.Balance,
			&i.AccountType,
			&i.Status,
			&i.Currency,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProfileByUserID = `-- name: GetProfileByUserID :one
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
WHERE users.id = $1 LIMIT 1
`

type GetProfileByUserIDRow struct {
	AccountID    uuid.UUID    `json:"account_id"`
	UserID       uuid.UUID    `json:"user_id"`
	Email        string       `json:"email"`
	FirstName    string       `json:"first_name"`
	LastName     string       `json:"last_name"`
	AccountType  AccountType  `json:"account_type"`
	UserType     UserType     `json:"user_type"`
	RegisteredAt sql.NullTime `json:"registered_at"`
}

func (q *Queries) GetProfileByUserID(ctx context.Context, id uuid.UUID) (GetProfileByUserIDRow, error) {
	row := q.db.QueryRowContext(ctx, getProfileByUserID, id)
	var i GetProfileByUserIDRow
	err := row.Scan(
		&i.AccountID,
		&i.UserID,
		&i.Email,
		&i.FirstName,
		&i.LastName,
		&i.AccountType,
		&i.UserType,
		&i.RegisteredAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, password, first_name, last_name, user_type, created_at, updated_at
    FROM users WHERE email = $1 LIMIT 1
`

type GetUserByEmailRow struct {
	ID        uuid.UUID    `json:"id"`
	Email     string       `json:"email"`
	Password  string       `json:"password"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	UserType  UserType     `json:"user_type"`
	CreatedAt sql.NullTime `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i GetUserByEmailRow
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Password,
		&i.FirstName,
		&i.LastName,
		&i.UserType,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, password, first_name, last_name, user_type, created_at, updated_at
    FROM users WHERE id = $1 LIMIT 1
`

type GetUserByIDRow struct {
	ID        uuid.UUID    `json:"id"`
	Email     string       `json:"email"`
	Password  string       `json:"password"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	UserType  UserType     `json:"user_type"`
	CreatedAt sql.NullTime `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (GetUserByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i GetUserByIDRow
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Password,
		&i.FirstName,
		&i.LastName,
		&i.UserType,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const saveAccount = `-- name: SaveAccount :one
INSERT INTO accounts(
    user_id, account_number, status, account_type, currency
) VALUES ($1, $2, $3, $4, $5) RETURNING id, user_id, account_number, account_type, status, currency, created_at, updated_at, deleted_at, balance
`

type SaveAccountParams struct {
	UserID        uuid.UUID   `json:"user_id"`
	AccountNumber string      `json:"account_number"`
	Status        Status      `json:"status"`
	AccountType   AccountType `json:"account_type"`
	Currency      Currency    `json:"currency"`
}

func (q *Queries) SaveAccount(ctx context.Context, arg SaveAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, saveAccount,
		arg.UserID,
		arg.AccountNumber,
		arg.Status,
		arg.AccountType,
		arg.Currency,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.AccountNumber,
		&i.AccountType,
		&i.Status,
		&i.Currency,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Balance,
	)
	return i, err
}

const saveUser = `-- name: SaveUser :one
INSERT INTO users(
    email, password, first_name, last_name, user_type
) VALUES ($1, $2, $3, $4, $5) RETURNING id, email
`

type SaveUserParams struct {
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	UserType  UserType `json:"user_type"`
}

type SaveUserRow struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (q *Queries) SaveUser(ctx context.Context, arg SaveUserParams) (SaveUserRow, error) {
	row := q.db.QueryRowContext(ctx, saveUser,
		arg.Email,
		arg.Password,
		arg.FirstName,
		arg.LastName,
		arg.UserType,
	)
	var i SaveUserRow
	err := row.Scan(&i.ID, &i.Email)
	return i, err
}

const updateAccountStatus = `-- name: UpdateAccountStatus :exec
UPDATE accounts
    SET status = $1, updated_at = CURRENT_TIMESTAMP
    WHERE id = $2
`

type UpdateAccountStatusParams struct {
	Status Status    `json:"status"`
	ID     uuid.UUID `json:"id"`
}

func (q *Queries) UpdateAccountStatus(ctx context.Context, arg UpdateAccountStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateAccountStatus, arg.Status, arg.ID)
	return err
}
