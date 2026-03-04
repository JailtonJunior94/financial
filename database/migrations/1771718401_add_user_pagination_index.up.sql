-- Migration: Add user cursor-based pagination index
-- Description: Índice composto para suportar keyset pagination de usuários ordenados por (name, id)
-- Performance: Keyset pagination O(1) excluindo soft-deleted records

CREATE INDEX IF NOT EXISTS idx_users_name_id
ON users(name, id)
WHERE deleted_at IS NULL;
