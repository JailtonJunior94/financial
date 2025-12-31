-- Create invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    card_id UUID NOT NULL,
    reference_month DATE NOT NULL,
    due_date DATE NOT NULL,
    total_amount BIGINT NOT NULL DEFAULT 0, -- Stored as cents
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign keys
    CONSTRAINT fk_invoices_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_invoices_card FOREIGN KEY (card_id) REFERENCES cards(id),

    -- Business constraints
    CONSTRAINT uk_invoices_user_card_month
        UNIQUE (user_id, card_id, reference_month)
        WHERE deleted_at IS NULL,

    -- Ensure total_amount is non-negative
    CONSTRAINT chk_invoices_total_amount CHECK (total_amount >= 0)
);

-- Indexes for performance
CREATE INDEX idx_invoices_user_id ON invoices(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_card_id ON invoices(card_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_reference_month ON invoices(reference_month) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_user_reference_month ON invoices(user_id, reference_month) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_due_date ON invoices(due_date) WHERE deleted_at IS NULL;

-- Create invoice_items table
CREATE TABLE IF NOT EXISTS invoice_items (
    id UUID PRIMARY KEY,
    invoice_id UUID NOT NULL,
    category_id UUID NOT NULL,
    purchase_date DATE NOT NULL,
    description VARCHAR(255) NOT NULL,
    total_amount BIGINT NOT NULL, -- Original purchase total (in cents)
    installment_number INT NOT NULL,
    installment_total INT NOT NULL,
    installment_amount BIGINT NOT NULL, -- This installment's amount (in cents)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign keys
    CONSTRAINT fk_invoice_items_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id),
    CONSTRAINT fk_invoice_items_category FOREIGN KEY (category_id) REFERENCES categories(id),

    -- Business constraints
    CONSTRAINT chk_invoice_items_installment_number
        CHECK (installment_number > 0 AND installment_number <= installment_total),
    CONSTRAINT chk_invoice_items_installment_total
        CHECK (installment_total > 0),
    CONSTRAINT chk_invoice_items_amounts
        CHECK (total_amount > 0 AND installment_amount > 0)
);

-- Indexes for performance
CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoice_items_category_id ON invoice_items(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoice_items_purchase_date ON invoice_items(purchase_date) WHERE deleted_at IS NULL;

-- Composite index for finding items by purchase origin (for updates/deletes)
CREATE INDEX idx_invoice_items_purchase_origin
    ON invoice_items(purchase_date, category_id, description)
    WHERE deleted_at IS NULL;

-- Composite index for finding items by invoice and sorting
CREATE INDEX idx_invoice_items_invoice_purchase
    ON invoice_items(invoice_id, purchase_date, installment_number)
    WHERE deleted_at IS NULL;
