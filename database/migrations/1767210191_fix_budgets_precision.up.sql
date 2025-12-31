-- Fix budgets table for financial precision and DDD compliance

-- 1. Remove incorrect unique constraint on date (should be per user)
ALTER TABLE budgets DROP CONSTRAINT IF EXISTS uk_budgets_date;

-- 2. Add correct unique constraint (one budget per user per month)
ALTER TABLE budgets ADD CONSTRAINT uk_budgets_user_date UNIQUE (user_id, date);

-- 3. Fix percentage precision (scale 3 for 100.000%)
ALTER TABLE budgets ALTER COLUMN percentage_used TYPE DECIMAL(6, 3);

-- 4. Fix updated_at to be nullable without default (only set on actual updates)
ALTER TABLE budgets ALTER COLUMN updated_at DROP DEFAULT;
ALTER TABLE budgets ALTER COLUMN updated_at DROP NOT NULL;

-- 5. Fix budget_items percentages precision
ALTER TABLE budget_items ALTER COLUMN percentage_goal TYPE DECIMAL(6, 3);

-- 6. Remove calculated columns from database (should be computed in domain)
ALTER TABLE budget_items DROP COLUMN IF EXISTS percentage_used;
ALTER TABLE budget_items DROP COLUMN IF EXISTS percentage_total;

-- 7. Fix updated_at in budget_items
ALTER TABLE budget_items ALTER COLUMN updated_at DROP DEFAULT;
ALTER TABLE budget_items ALTER COLUMN updated_at DROP NOT NULL;

-- 8. Add unique constraint to prevent duplicate categories in same budget
ALTER TABLE budget_items ADD CONSTRAINT uk_budget_items_budget_category
    UNIQUE (budget_id, category_id) WHERE deleted_at IS NULL;

-- 9. Add index for soft delete queries
CREATE INDEX IF NOT EXISTS idx_budgets_deleted_at ON budgets(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_budget_items_deleted_at ON budget_items(deleted_at) WHERE deleted_at IS NULL;

-- 10. Add index for user queries
CREATE INDEX IF NOT EXISTS idx_budgets_user_id ON budgets(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_budget_items_budget_id ON budget_items(budget_id) WHERE deleted_at IS NULL;
