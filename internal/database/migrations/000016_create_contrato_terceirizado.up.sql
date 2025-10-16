CREATE TABLE contrato_terceirizado (
    id BIGSERIAL PRIMARY KEY,
    condominio_id BIGINT NOT NULL REFERENCES condominio(id) ON DELETE CASCADE,
    empresa_id BIGINT NOT NULL REFERENCES empresa_terceirizada(id) ON DELETE RESTRICT,
    numero_contrato VARCHAR(100) NOT NULL UNIQUE,
    valor_mensal NUMERIC(12, 2) NOT NULL,
    dia_vencimento INTEGER NOT NULL CHECK (dia_vencimento >= 1 AND dia_vencimento <= 31),
    data_inicio DATE NOT NULL,
    data_fim DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'suspenso', 'cancelado', 'finalizado')),
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_data_fim CHECK (data_fim IS NULL OR data_fim >= data_inicio)
);

CREATE INDEX idx_contrato_terceirizado_condominio_id ON contrato_terceirizado(condominio_id);
CREATE INDEX idx_contrato_terceirizado_empresa_id ON contrato_terceirizado(empresa_id);
CREATE INDEX idx_contrato_terceirizado_status ON contrato_terceirizado(status);
CREATE INDEX idx_contrato_terceirizado_numero ON contrato_terceirizado(numero_contrato);
CREATE INDEX idx_contrato_terceirizado_deleted_at ON contrato_terceirizado(deleted_at);

CREATE TRIGGER update_contrato_terceirizado_updated_at
    BEFORE UPDATE ON contrato_terceirizado
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
