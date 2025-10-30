-- Rename columns in address table
-- postal_code -> zip_code
-- streey -> street

-- First, drop the existing index on postal_code since we're renaming the column
DROP INDEX IF EXISTS idx_address_postal_code;

-- Rename the columns
ALTER TABLE address RENAME COLUMN postal_code TO zip_code;
ALTER TABLE address RENAME COLUMN streey TO street;

-- Recreate the index with the new column name
CREATE INDEX idx_address_zip_code ON address(zip_code);
