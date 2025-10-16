CREATE TABLE veiculo (
    id BIGSERIAL PRIMARY KEY,
    placa VARCHAR(10) NOT NULL UNIQUE,
    modelo VARCHAR(100) NOT NULL,
    marca VARCHAR(100) NOT NULL,
    cor VARCHAR(50),
    ano INTEGER,
    tipo VARCHAR(20) NOT NULL DEFAULT 'carro' CHECK (tipo IN ('carro', 'moto', 'caminhonete', 'outros')),
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_veiculo_placa ON veiculo(placa);
CREATE INDEX idx_veiculo_tipo ON veiculo(tipo);
CREATE INDEX idx_veiculo_deleted_at ON veiculo(deleted_at);

CREATE TRIGGER update_veiculo_updated_at
    BEFORE UPDATE ON veiculo
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
