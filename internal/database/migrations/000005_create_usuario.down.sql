ALTER TABLE pessoa DROP CONSTRAINT IF EXISTS fk_pessoa_usuario;
DROP TRIGGER IF EXISTS update_usuario_updated_at ON usuario;
DROP TABLE IF EXISTS usuario CASCADE;
