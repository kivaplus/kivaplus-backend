-- Make address_id nullable in person table
-- This allows users to register without providing an address initially

ALTER TABLE person ALTER COLUMN address_id DROP NOT NULL;
