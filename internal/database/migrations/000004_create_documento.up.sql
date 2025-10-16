CREATE TABLE documento (
    id BIGSERIAL PRIMARY KEY,
    pessoa_id BIGINT NOT NULL REFERENCES pessoa(id) ON DELETE CASCADE,
    tipo VARCHAR(10) NOT NULL CHECK (tipo IN ('CPF', 'RG', 'CNH', 'CNPJ')),
    numero VARCHAR(30) NOT NULL,
    orgao_emissor VARCHAR(20),
    data_emissao DATE,
    data_validade DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_pessoa_tipo_documento UNIQUE(pessoa_id, tipo)
);

CREATE INDEX idx_documento_pessoa_id ON documento(pessoa_id);
CREATE INDEX idx_documento_numero ON documento(numero);

CREATE TRIGGER update_documento_updated_at
    BEFORE UPDATE ON documento
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
