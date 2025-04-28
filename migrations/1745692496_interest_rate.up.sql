CREATE TABLE IF NOT EXISTS interest_rates (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rate                  BIGINT NOT NULL,
    calculation_frequency VARCHAR(50) NOT NULL,
    created_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at            TIMESTAMP
);

-- Create an account to book interest transactions
INSERT INTO users (id, email, password, first_name, last_name)
    VALUES (
        '00000000-1111-1111-1111-000000000000',
        'interestaccount@payterbank.app',
        gen_random_uuid(),
        'Interest',
        'Account'
);

INSERT INTO accounts (id, user_id, account_number, status, account_type, currency)
    VALUES (
        '00000000-1111-1111-1111-000000000000',
        '00000000-1111-1111-1111-000000000000',
        '00001111',
        'ACTIVE',
        'EXTERNAL',
        'GBP'
);