CREATE TABLE budget_items (
    id UUID NOT NULL,
    budget_id UUID NOT NULL,
    category_id UUID NOT NULL,
    percentage_goal DECIMAL(5, 2) DEFAULT 0,
    amount_goal DECIMAL(10, 2) DEFAULT 0,
    amount_used DECIMAL(10, 2) DEFAULT 0,
    percentage_used DECIMAL(5, 2) DEFAULT 0,
    percentage_total DECIMAL(5, 2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk_budget_items PRIMARY KEY (id),
    CONSTRAINT fk_budget_items_budgets FOREIGN KEY (budget_id) REFERENCES budgets(id),
    CONSTRAINT fk_budget_items_categories FOREIGN KEY (category_id) REFERENCES categories(id)
);