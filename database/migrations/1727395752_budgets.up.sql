CREATE TABLE budgets (
    id UUID NOT NULL,
    user_id UUID NOT NULL,
    date DATE NOT NULL UNIQUE,
    amount_goal DECIMAL(10, 2) NOT NULL,
    amount_used DECIMAL(10, 2) DEFAULT 0,
    percentage_used DECIMAL(5, 2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT pk_budgets PRIMARY KEY (id),
    CONSTRAINT fk_budgets_users FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT uk_budgets_date UNIQUE (date)
);
