CREATE TABLE budgets (
    id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    date DATE NOT NULL UNIQUE,
    amount_goal DECIMAL(10, 2) NOT NULL,
    amount_used DECIMAL(10, 2) DEFAULT 0,
    percentage_used DECIMAL(5, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE KEY uk_budgets_date (date)
);
