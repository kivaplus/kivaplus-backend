-- Rollback migration for unique super_admin constraint

-- Drop the unique index
DROP INDEX IF EXISTS idx_unique_super_admin;
