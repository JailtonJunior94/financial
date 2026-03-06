-- 1. Remove indexes dependent on parent_id
DROP INDEX IF EXISTS idx_categories_parent;
DROP INDEX IF EXISTS idx_categories_user_parent;
DROP INDEX IF EXISTS idx_categories_list;

-- 2. Remove FK and parent_id column
ALTER TABLE categories DROP CONSTRAINT IF EXISTS fk_categories_parent;
ALTER TABLE categories DROP COLUMN IF EXISTS parent_id;

-- 3. Recreate listing index without parent_id
CREATE INDEX idx_categories_list ON categories(user_id, sequence)
    WHERE deleted_at IS NULL;

-- 4. Create subcategories table
CREATE TABLE subcategories (
    id          UUID         NOT NULL,
    category_id UUID         NOT NULL,
    user_id     UUID         NOT NULL,
    name        VARCHAR(255) NOT NULL,
    sequence    INT          NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,

    CONSTRAINT pk_subcategories PRIMARY KEY (id),
    CONSTRAINT fk_subcategories_categories FOREIGN KEY (category_id)
        REFERENCES categories(id) ON DELETE CASCADE,
    CONSTRAINT fk_subcategories_users FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);

-- 5. Indexes for subcategories
CREATE INDEX idx_subcategories_category_list
    ON subcategories(category_id, sequence)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_subcategories_user_active
    ON subcategories(user_id, deleted_at);

CREATE INDEX idx_subcategories_cursor
    ON subcategories(category_id, sequence, id)
    WHERE deleted_at IS NULL;
