-- Índice para query List() (user_id + parent_id NULL + deleted_at NULL)
-- Otimiza a listagem de categorias raiz ordenadas por sequence
CREATE INDEX idx_categories_list
ON categories(user_id, sequence)
WHERE parent_id IS NULL AND deleted_at IS NULL;

-- Índice para filtros user_id + deleted_at (usado em todas queries)
-- Acelera filtragem de categorias ativas de um usuário
CREATE INDEX idx_categories_user_active
ON categories(user_id, deleted_at);

-- Índice para parent_id (usado no LEFT JOIN e CTE recursiva)
-- Otimiza busca de subcategorias e detecção de ciclos
CREATE INDEX idx_categories_parent
ON categories(parent_id)
WHERE deleted_at IS NULL;

-- Índice composto para user + parent (otimiza joins)
-- Acelera queries que filtram por usuário e parent
CREATE INDEX idx_categories_user_parent
ON categories(user_id, parent_id)
WHERE deleted_at IS NULL;
