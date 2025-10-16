CREATE TABLE endereco (
    id BIGSERIAL PRIMARY KEY,
    cep VARCHAR(10) NOT NULL,
    logradouro VARCHAR(255) NOT NULL,
    numero VARCHAR(20) NOT NULL,
    complemento VARCHAR(255),
    bairro VARCHAR(100) NOT NULL,
    cidade VARCHAR(100) NOT NULL,
    estado VARCHAR(2) NOT NULL,
    pais VARCHAR(2) NOT NULL DEFAULT 'BR',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_endereco_cep ON endereco(cep);
CREATE INDEX idx_endereco_cidade_estado ON endereco(cidade, estado);

CREATE TRIGGER update_endereco_updated_at
    BEFORE UPDATE ON endereco
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
