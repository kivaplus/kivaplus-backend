CREATE TABLE empresa_terceirizada (
    id BIGSERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    cnpj VARCHAR(18) NOT NULL UNIQUE,
    razao_social VARCHAR(255) NOT NULL,
    telefone VARCHAR(20),
    email VARCHAR(255),
    endereco_id BIGINT NOT NULL REFERENCES endereco(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_empresa_terceirizada_cnpj ON empresa_terceirizada(cnpj);
CREATE INDEX idx_empresa_terceirizada_nome ON empresa_terceirizada(nome);
CREATE INDEX idx_empresa_terceirizada_deleted_at ON empresa_terceirizada(deleted_at);

CREATE TRIGGER update_empresa_terceirizada_updated_at
    BEFORE UPDATE ON empresa_terceirizada
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
