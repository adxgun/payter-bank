CREATE TABLE IF NOT EXISTS transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_account_id     UUID NOT NULL REFERENCES accounts(id),
    to_account_id       UUID NOT NULL REFERENCES accounts(id),
    amount              BIGINT NOT NULL,
    reference_number    VARCHAR(255) NOT NULL,
    description         VARCHAR(255),
    status              VARCHAR(50) NOT NULL,
    currency            VARCHAR(3) NOT NULL,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP
);