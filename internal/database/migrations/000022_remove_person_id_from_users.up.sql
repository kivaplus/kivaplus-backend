-- Remove redundant person_id from users table
-- The relationship is maintained through person.user_id

-- First, ensure all person records have the correct user_id
-- (in case there are any inconsistencies)
UPDATE person
SET user_id = u.id
FROM users u
WHERE u.person_id = person.id
AND person.user_id IS NULL;

-- Remove the foreign key constraint from person table
ALTER TABLE person DROP CONSTRAINT IF EXISTS fk_person_users;

-- Drop the index on users.person_id
DROP INDEX IF EXISTS idx_users_person_id;

-- Remove the person_id column from users table
ALTER TABLE users DROP COLUMN IF EXISTS person_id;

-- Recreate the foreign key constraint on person.user_id with proper naming
ALTER TABLE person
    ADD CONSTRAINT fk_person_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Ensure the index on person.user_id exists
CREATE INDEX IF NOT EXISTS idx_person_user_id ON person(user_id);

-- Add a unique constraint to ensure 1:1 relationship
ALTER TABLE person
    ADD CONSTRAINT uk_person_user_id UNIQUE (user_id);
