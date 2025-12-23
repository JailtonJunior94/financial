-- Remove índices de performance para categorias
DROP INDEX IF EXISTS idx_categories_list;
DROP INDEX IF EXISTS idx_categories_user_active;
DROP INDEX IF EXISTS idx_categories_parent;
DROP INDEX IF EXISTS idx_categories_user_parent;
