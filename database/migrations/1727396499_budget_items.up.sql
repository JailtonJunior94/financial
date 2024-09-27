CREATE TABLE budget_items (
    id CHAR(36) NOT NULL,
    budget_id CHAR(36) NOT NULL,
    category_id CHAR(36) NOT NULL,
    percentage_goal DECIMAL(5, 2) DEFAULT 0,
    amount_goal DECIMAL(10, 2) DEFAULT 0,
    amount_used DECIMAL(10, 2) DEFAULT 0,
    percentage_used DECIMAL(5, 2) DEFAULT 0,
    percentage_total DECIMAL(5, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (budget_id) REFERENCES budgets(id),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);