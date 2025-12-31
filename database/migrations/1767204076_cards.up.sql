CREATE TABLE cards (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    due_day INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk_cards PRIMARY KEY (id),
    CONSTRAINT fk_cards_users FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT chk_due_day CHECK (due_day >= 1 AND due_day <= 31)
);

CREATE INDEX idx_cards_user_id ON cards(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_cards_deleted_at ON cards(deleted_at);
