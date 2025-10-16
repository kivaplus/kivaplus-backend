CREATE TABLE pessoa (
    id BIGSERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    tipo_pessoa VARCHAR(2) NOT NULL CHECK (tipo_pessoa IN ('PF', 'PJ')),
    telefone VARCHAR(20),
    email VARCHAR(255),
    endereco_id BIGINT NOT NULL REFERENCES endereco(id) ON DELETE RESTRICT,
    profissao VARCHAR(100),
    estado_civil VARCHAR(20) CHECK (estado_civil IN ('solteiro', 'casado', 'divorciado', 'viuvo', 'uniao_estavel', 'outro')),
    data_nascimento DATE,
    usuario_id BIGINT,
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_pessoa_nome ON pessoa(nome);
CREATE INDEX idx_pessoa_email ON pessoa(email);
CREATE INDEX idx_pessoa_tipo ON pessoa(tipo_pessoa);
CREATE INDEX idx_pessoa_deleted_at ON pessoa(deleted_at);

CREATE TRIGGER update_pessoa_updated_at
    BEFORE UPDATE ON pessoa
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
