CREATE TYPE user_type AS ENUM (
    'EXTERNAL',
    'ADMIN',
    'CUSTOMER'
);

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    first_name  VARCHAR(255) NOT NULL,
    last_name   VARCHAR(255) NOT NULL,
    user_type   user_type NOT NULL DEFAULT 'CUSTOMER',
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP
);

CREATE TYPE account_type AS ENUM (
    'EXTERNAL',
    'CURRENT'
);

CREATE TYPE status as ENUM (
    'PENDING',
    'ACTIVE',
    'SUSPENDED',
    'CLOSED'
);

CREATE TYPE currency as ENUM (
    'GBP',
    'EUR',
    'JPY'
);

CREATE TABLE IF NOT EXISTS accounts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id),
    account_number VARCHAR(255) NOT NULL UNIQUE,
    account_type   account_type NOT NULL DEFAULT 'CURRENT',
    status         status NOT NULL DEFAULT 'PENDING',
    currency       currency NOT NULL DEFAULT 'GBP',
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMP
);

-- create users and accounts to represent transactions from and to external accounts.
INSERT INTO users (id, email, password, first_name, last_name, user_type)
    VALUES ('00000000-0000-0000-0000-000000000000', 'external@payterbank.com', gen_random_uuid(), 'External', 'Account', 'CUSTOMER');

INSERT INTO accounts (id, user_id, account_number, account_type)
    VALUES ('00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', '00000000', 'EXTERNAL');
