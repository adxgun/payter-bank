-- name: GetAccountBalance :one
SELECT
    a.id AS account_id,
    a.account_number AS account_number,
    a.currency AS currency,
    a.account_type AS account_type,
    (
        COALESCE(SUM(CASE WHEN t.to_account_id = a.id THEN t.amount ELSE 0 END), 0) -
        COALESCE(SUM(CASE WHEN t.from_account_id = a.id THEN t.amount ELSE 0 END), 0)
    ) AS balance
FROM accounts a
    LEFT JOIN
        transactions t ON a.id = t.from_account_id OR a.id = t.to_account_id
    WHERE a.id = $1
GROUP BY
    a.id, a.account_number, a.currency LIMIT 1;

-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = $1;

-- name: GetTransactionsByAccountID :many
SELECT
    t.id AS transaction_id,
    t.from_account_id AS from_account_id,
    t.to_account_id AS to_account_id,
    t.amount AS amount,
    t.reference_number AS reference_number,
    t.description AS description,
    t.status AS status,
    t.currency AS currency,
    t.created_at AS created_at,
    t.updated_at AS updated_at
FROM
    transactions t
WHERE
    t.from_account_id = $1 OR t.to_account_id = $1
ORDER BY
    t.created_at DESC;

-- name: SaveTransaction :one
INSERT INTO transactions(
    from_account_id, to_account_id, amount, reference_number, description, status, currency
) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;