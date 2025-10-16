CREATE TABLE funcionario (
    id BIGSERIAL PRIMARY KEY,
    pessoa_id BIGINT NOT NULL REFERENCES pessoa(id) ON DELETE CASCADE,
    condominio_id BIGINT NOT NULL REFERENCES condominio(id) ON DELETE CASCADE,
    cargo VARCHAR(100) NOT NULL,
    salario NUMERIC(12, 2),
    data_admissao DATE NOT NULL,
    data_demissao DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo', 'afastado', 'ferias')),
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT uk_pessoa_condominio_funcionario UNIQUE(pessoa_id, condominio_id),
    CONSTRAINT chk_data_demissao CHECK (data_demissao IS NULL OR data_demissao >= data_admissao)
);

CREATE INDEX idx_funcionario_pessoa_id ON funcionario(pessoa_id);
CREATE INDEX idx_funcionario_condominio_id ON funcionario(condominio_id);
CREATE INDEX idx_funcionario_status ON funcionario(status);
CREATE INDEX idx_funcionario_deleted_at ON funcionario(deleted_at);

CREATE TRIGGER update_funcionario_updated_at
    BEFORE UPDATE ON funcionario
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
