-- Rollback: Restore address constraints
-- Note: This will fail if there are records with NULL address_id or if referenced addresses don't exist

-- Make address_id NOT NULL again
ALTER TABLE person ALTER COLUMN address_id SET NOT NULL;

-- Restore the foreign key constraint
ALTER TABLE person ADD CONSTRAINT person_address_id_fkey
    FOREIGN KEY (address_id) REFERENCES address(id) ON DELETE RESTRICT;
