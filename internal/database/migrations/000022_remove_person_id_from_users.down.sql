-- Rollback: Add person_id back to users table

-- Remove the unique constraint
ALTER TABLE person DROP CONSTRAINT IF EXISTS uk_person_user_id;

-- Remove the foreign key constraint
ALTER TABLE person DROP CONSTRAINT IF EXISTS fk_person_user_id;

-- Add person_id column back to users
ALTER TABLE users ADD COLUMN person_id BIGINT;

-- Update users.person_id based on person.user_id
UPDATE users
SET person_id = p.id
FROM person p
WHERE p.user_id = users.id;

-- Add foreign key constraint
ALTER TABLE users
    ADD CONSTRAINT fk_users_person_id
    FOREIGN KEY (person_id) REFERENCES person(id) ON DELETE SET NULL;

-- Create index
CREATE INDEX idx_users_person_id ON users(person_id);

-- Recreate the old foreign key constraint on person
ALTER TABLE person
    ADD CONSTRAINT fk_person_users
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
