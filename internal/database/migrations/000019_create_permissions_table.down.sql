-- Rollback permissions table creation

-- Remove the permission_type column from role table
ALTER TABLE role DROP COLUMN IF EXISTS permission_type;

-- Drop the permissions table (CASCADE will handle foreign key constraints)
DROP TABLE IF EXISTS permissions CASCADE;
