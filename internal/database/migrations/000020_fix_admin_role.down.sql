-- Rollback the admin role fix
-- This is a complex rollback since we're fixing data integrity

-- Revert admin role back to the misspelled version
UPDATE role SET name = 'adminstador', description = 'Administrador do condomínio - gestão completa do condomínio'
WHERE id = 2 AND name = 'admin';

-- Note: This rollback doesn't fully restore the original state since we're fixing a typo
-- In production, you might want to keep the fix and not allow rollback
