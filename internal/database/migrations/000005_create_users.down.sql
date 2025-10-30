ALTER TABLE person DROP CONSTRAINT IF EXISTS fk_person_users;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TABLE IF EXISTS users CASCADE;
