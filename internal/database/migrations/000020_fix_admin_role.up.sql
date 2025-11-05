-- Fix the typo in admin role name and ensure proper role structure
-- Update existing 'adminstador' to 'admin' and adjust role IDs

-- First, update any existing users with the old admin role
UPDATE users_role SET role_id = 2 WHERE role_id = (SELECT id FROM role WHERE name = 'adminstador');

-- Delete the old misspelled role
DELETE FROM role WHERE name = 'adminstador';

-- Insert the correct admin role with ID 2
INSERT INTO role (id, name, description, permissions) VALUES
(2, 'admin', 'Administrador do condomínio - gestão completa do condomínio', '{
    "condominio": {"read": true, "update": true},
    "unidade": {"create": true, "read": true, "update": true, "delete": true},
    "morador": {"create": true, "read": true, "update": true, "delete": true},
    "funcionario": {"create": true, "read": true, "update": true, "delete": true},
    "veiculo": {"create": true, "read": true, "update": true, "delete": true},
    "contrato": {"create": true, "read": true, "update": true, "delete": true}
}')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    permissions = EXCLUDED.permissions;

-- Update sindico role to ID 3 if needed
UPDATE role SET id = 3 WHERE name = 'sindico' AND id != 3;
UPDATE users_role SET role_id = 3 WHERE role_id = (SELECT id FROM role WHERE name = 'sindico' AND id != 3);

-- Update morador role to ID 4 if needed
UPDATE role SET id = 4 WHERE name = 'morador' AND id != 4;
UPDATE users_role SET role_id = 4 WHERE role_id = (SELECT id FROM role WHERE name = 'morador' AND id != 4);

-- Update funcionario role to ID 5 if needed
UPDATE role SET id = 5 WHERE name = 'funcionario' AND id != 5;
UPDATE users_role SET role_id = 5 WHERE role_id = (SELECT id FROM role WHERE name = 'funcionario' AND id != 5);

-- Reset the sequence to ensure future inserts use correct IDs
SELECT setval('role_id_seq', 5, true);
