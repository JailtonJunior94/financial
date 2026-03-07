CREATE TABLE transactions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id),
    category_id          UUID NOT NULL REFERENCES categories(id),
    subcategory_id       UUID REFERENCES subcategories(id),
    card_id              UUID REFERENCES cards(id),
    invoice_id           UUID REFERENCES invoices(id),
    installment_group_id UUID,
    description          VARCHAR(255) NOT NULL,
    amount               DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    payment_method       VARCHAR(10) NOT NULL
        CHECK (payment_method IN ('pix','boleto','ted','debit','credit')),
    transaction_date     DATE NOT NULL,
    installment_number   SMALLINT CHECK (installment_number >= 1),
    installment_total    SMALLINT CHECK (installment_total >= 1),
    status               VARCHAR(10) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','cancelled')),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ,
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX idx_transactions_user_id
    ON transactions(user_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_transactions_invoice_id
    ON transactions(invoice_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_transactions_installment_group
    ON transactions(installment_group_id)
    WHERE installment_group_id IS NOT NULL;

CREATE INDEX idx_transactions_user_date
    ON transactions(user_id, transaction_date DESC) WHERE deleted_at IS NULL;
