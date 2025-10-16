CREATE TABLE papel (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(50) NOT NULL UNIQUE,
    descricao TEXT,
    permissoes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO papel (nome, descricao, permissoes) VALUES
('super_admin', 'Administrador do sistema - acesso total', '{"all": true}'),
('sindico', 'Síndico do condomínio - gestão completa do condomínio', '{
    "condominio": {"read": true, "update": true},
    "unidade": {"create": true, "read": true, "update": true, "delete": true},
    "morador": {"create": true, "read": true, "update": true, "delete": true},
    "funcionario": {"create": true, "read": true, "update": true, "delete": true},
    "veiculo": {"create": true, "read": true, "update": true, "delete": true},
    "contrato": {"create": true, "read": true, "update": true, "delete": true}
}'),
('morador', 'Morador - acesso restrito aos próprios dados', '{
    "unidade": {"read": true},
    "morador": {"read": true},
    "veiculo": {"create": true, "read": true, "update": true}
}'),
('funcionario', 'Funcionário do condomínio', '{
    "morador": {"read": true},
    "veiculo": {"read": true}
}');
