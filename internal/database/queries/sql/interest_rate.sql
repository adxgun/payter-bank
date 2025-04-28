-- name: SaveInterestRate :one
INSERT INTO interest_rates(
    rate, calculation_frequency
) VALUES ($1, $2) RETURNING *;

-- name: UpdateRate :exec
UPDATE interest_rates
    SET rate = $1, updated_at = CURRENT_TIMESTAMP
    WHERE id = $2 RETURNING *;

-- name: UpdateCalculationFrequency :exec
UPDATE interest_rates
    SET calculation_frequency = $1, updated_at = CURRENT_TIMESTAMP
    WHERE id = $2 RETURNING *;

-- name: GetInterestRates :many
SELECT * FROM interest_rates ORDER BY created_at DESC LIMIT 1;