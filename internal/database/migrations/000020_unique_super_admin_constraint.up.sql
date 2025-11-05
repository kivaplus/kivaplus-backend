-- Migration to ensure only one super_admin can exist
-- This creates a unique partial index that allows only one super_admin role

-- Create a unique partial index to ensure only one super_admin exists
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_unique_super_admin
ON users_role (role_id)
WHERE role_id = 1 AND condominium_id IS NULL;

-- Add a comment explaining the constraint
COMMENT ON INDEX idx_unique_super_admin IS 'Ensures only one super_admin (role_id=1) can exist in the system';
