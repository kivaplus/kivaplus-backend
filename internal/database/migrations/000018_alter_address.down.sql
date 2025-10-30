-- Rollback: Rename columns back to original names
-- zip_code -> postal_code
-- street -> streey

-- Drop the index on zip_code since we're renaming the column back
DROP INDEX IF EXISTS idx_address_zip_code;

-- Rename the columns back to original names
ALTER TABLE address RENAME COLUMN zip_code TO postal_code;
ALTER TABLE address RENAME COLUMN street TO streey;

-- Recreate the original index
CREATE INDEX idx_address_postal_code ON address(postal_code);
