-- Rollback role permissions to Portuguese resource names

-- Rollback super_admin
UPDATE role SET permissions = '{"all": true}' WHERE name = 'super_admin';

-- Rollback administrator to adminstador with Portuguese resources
UPDATE role SET
    name = 'adminstador',
    permissions = '{
        "condominio": {"read": true, "update": true},
        "unidade": {"create": true, "read": true, "update": true, "delete": true},
        "morador": {"create": true, "read": true, "update": true, "delete": true},
        "funcionario": {"create": true, "read": true, "update": true, "delete": true},
        "veiculo": {"create": true, "read": true, "update": true, "delete": true},
        "contrato": {"create": true, "read": true, "update": true, "delete": true}
    }'
WHERE name = 'administrator';

-- Rollback sindico to Portuguese resources
UPDATE role SET permissions = '{
    "condominio": {"read": true, "update": true},
    "unidade": {"create": true, "read": true, "update": true, "delete": true},
    "morador": {"create": true, "read": true, "update": true, "delete": true},
    "funcionario": {"create": true, "read": true, "update": true, "delete": true},
    "veiculo": {"create": true, "read": true, "update": true, "delete": true},
    "contrato": {"create": true, "read": true, "update": true, "delete": true}
}' WHERE name = 'sindico';

-- Rollback morador to Portuguese resources
UPDATE role SET permissions = '{
    "unidade": {"read": true},
    "morador": {"read": true},
    "veiculo": {"create": true, "read": true, "update": true}
}' WHERE name = 'morador';

-- Rollback funcionario to Portuguese resources
UPDATE role SET permissions = '{
    "morador": {"read": true},
    "veiculo": {"read": true}
}' WHERE name = 'funcionario';
