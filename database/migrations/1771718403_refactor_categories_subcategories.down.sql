DROP TABLE IF EXISTS subcategories;

DROP INDEX IF EXISTS idx_categories_list;

ALTER TABLE categories ADD COLUMN IF NOT EXISTS parent_id UUID;
ALTER TABLE categories ADD CONSTRAINT fk_categories_parent
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE RESTRICT;

CREATE INDEX idx_categories_list ON categories(user_id, sequence)
    WHERE parent_id IS NULL AND deleted_at IS NULL;
CREATE INDEX idx_categories_parent ON categories(parent_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_user_parent ON categories(user_id, parent_id)
    WHERE deleted_at IS NULL;
