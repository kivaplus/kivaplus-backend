-- Remove address constraints and set address_id to NULL
-- This migration removes the foreign key constraint and sets all address_id values to NULL

-- First, drop the foreign key constraint
ALTER TABLE person DROP CONSTRAINT IF EXISTS person_address_id_fkey;

-- Set all existing address_id values to NULL
UPDATE person SET address_id = NULL WHERE address_id IS NOT NULL;

-- Ensure the column is nullable (should already be done by migration 000019)
ALTER TABLE person ALTER COLUMN address_id DROP NOT NULL;
