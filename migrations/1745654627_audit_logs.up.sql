CREATE TABLE IF NOT EXISTS audit_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    affected_account_id UUID REFERENCES accounts(id),
    action              VARCHAR(255) NOT NULL,
    metadata            JSONB,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP
);