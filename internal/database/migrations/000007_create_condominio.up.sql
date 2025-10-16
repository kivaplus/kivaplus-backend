CREATE TABLE condominio (
    id BIGSERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    cnpj VARCHAR(18) NOT NULL UNIQUE,
    telefone VARCHAR(20),
    email VARCHAR(255),
    endereco_id BIGINT NOT NULL REFERENCES endereco(id) ON DELETE RESTRICT,
    sindico_id BIGINT NOT NULL REFERENCES pessoa(id) ON DELETE RESTRICT,
    administrador_id BIGINT REFERENCES pessoa(id) ON DELETE SET NULL,
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_condominio_cnpj ON condominio(cnpj);
CREATE INDEX idx_condominio_nome ON condominio(nome);
CREATE INDEX idx_condominio_sindico_id ON condominio(sindico_id);
CREATE INDEX idx_condominio_deleted_at ON condominio(deleted_at);

CREATE TRIGGER update_condominio_updated_at
    BEFORE UPDATE ON condominio
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
