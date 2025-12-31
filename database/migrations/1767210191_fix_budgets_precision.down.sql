-- Rollback budgets precision fixes

-- 10. Remove indexes
DROP INDEX IF EXISTS idx_budget_items_budget_id;
DROP INDEX IF EXISTS idx_budgets_user_id;
DROP INDEX IF EXISTS idx_budget_items_deleted_at;
DROP INDEX IF EXISTS idx_budgets_deleted_at;

-- 8. Remove unique constraint
ALTER TABLE budget_items DROP CONSTRAINT IF EXISTS uk_budget_items_budget_category;

-- 7. Restore updated_at defaults in budget_items
ALTER TABLE budget_items ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE budget_items ALTER COLUMN updated_at SET NOT NULL;

-- 6. Restore calculated columns (not recommended, but for rollback)
ALTER TABLE budget_items ADD COLUMN IF NOT EXISTS percentage_used DECIMAL(5, 2) DEFAULT 0;
ALTER TABLE budget_items ADD COLUMN IF NOT EXISTS percentage_total DECIMAL(5, 2) DEFAULT 0;

-- 5. Restore budget_items percentage precision
ALTER TABLE budget_items ALTER COLUMN percentage_goal TYPE DECIMAL(5, 2);

-- 4. Restore updated_at defaults in budgets
ALTER TABLE budgets ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE budgets ALTER COLUMN updated_at SET NOT NULL;

-- 3. Restore budgets percentage precision
ALTER TABLE budgets ALTER COLUMN percentage_used TYPE DECIMAL(5, 2);

-- 2. Remove per-user unique constraint
ALTER TABLE budgets DROP CONSTRAINT IF EXISTS uk_budgets_user_date;

-- 1. Restore global date unique constraint
ALTER TABLE budgets ADD CONSTRAINT uk_budgets_date UNIQUE (date);
