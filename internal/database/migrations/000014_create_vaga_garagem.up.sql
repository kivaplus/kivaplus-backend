CREATE TABLE vaga_garagem (
    id BIGSERIAL PRIMARY KEY,
    unidade_id BIGINT NOT NULL REFERENCES unidade(id) ON DELETE CASCADE,
    numero VARCHAR(20) NOT NULL,
    tipo VARCHAR(20) NOT NULL CHECK (tipo IN ('coberta', 'descoberta')),
    disponivel BOOLEAN NOT NULL DEFAULT TRUE,
    veiculo_id BIGINT REFERENCES veiculo(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_unidade_numero_vaga UNIQUE(unidade_id, numero)
);

CREATE INDEX idx_vaga_garagem_unidade_id ON vaga_garagem(unidade_id);
CREATE INDEX idx_vaga_garagem_veiculo_id ON vaga_garagem(veiculo_id);
CREATE INDEX idx_vaga_garagem_disponivel ON vaga_garagem(disponivel);

CREATE TRIGGER update_vaga_garagem_updated_at
    BEFORE UPDATE ON vaga_garagem
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
