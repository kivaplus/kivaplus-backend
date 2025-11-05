-- Update role permissions to use English resource names to match domain models

-- Update super_admin (keep "all": true as it bypasses resource checks)
UPDATE role SET permissions = '{"all": true}' WHERE name = 'super_admin';

-- Update adminstador (fix typo and use English resources)
UPDATE role SET
    name = 'administrator',
    permissions = '{
        "condominiums": {"read": true, "update": true},
        "units": {"create": true, "read": true, "update": true, "delete": true},
        "residents": {"create": true, "read": true, "update": true, "delete": true},
        "employees": {"create": true, "read": true, "update": true, "delete": true},
        "vehicles": {"create": true, "read": true, "update": true, "delete": true},
        "contracts": {"create": true, "read": true, "update": true, "delete": true},
        "companies": {"create": true, "read": true, "update": true, "delete": true},
        "services": {"create": true, "read": true, "update": true, "delete": true}
    }'
WHERE name = 'adminstador';

-- Update sindico to use English resources
UPDATE role SET permissions = '{
    "condominiums": {"read": true, "update": true},
    "units": {"create": true, "read": true, "update": true, "delete": true},
    "residents": {"create": true, "read": true, "update": true, "delete": true},
    "employees": {"create": true, "read": true, "update": true, "delete": true},
    "vehicles": {"create": true, "read": true, "update": true, "delete": true},
    "contracts": {"create": true, "read": true, "update": true, "delete": true},
    "companies": {"read": true},
    "services": {"read": true}
}' WHERE name = 'sindico';

-- Update morador to use English resources
UPDATE role SET permissions = '{
    "units": {"read": true},
    "residents": {"read": true},
    "vehicles": {"create": true, "read": true, "update": true}
}' WHERE name = 'morador';

-- Update funcionario to use English resources
UPDATE role SET permissions = '{
    "residents": {"read": true},
    "vehicles": {"read": true}
}' WHERE name = 'funcionario';
