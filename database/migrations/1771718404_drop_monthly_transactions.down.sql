CREATE TABLE IF NOT EXISTS monthly_transactions (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS transaction_items (
    id                     UUID PRIMARY KEY,
    monthly_transaction_id UUID NOT NULL REFERENCES monthly_transactions(id),
    created_at             TIMESTAMPTZ NOT NULL
);
