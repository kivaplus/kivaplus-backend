-- Rollback: Make address_id NOT NULL again
-- Note: This will fail if there are records with NULL address_id

ALTER TABLE person ALTER COLUMN address_id SET NOT NULL;
