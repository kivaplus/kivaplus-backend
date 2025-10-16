CREATE TABLE morador_veiculo (
    id BIGSERIAL PRIMARY KEY,
    morador_id BIGINT NOT NULL REFERENCES morador(id) ON DELETE CASCADE,
    veiculo_id BIGINT NOT NULL REFERENCES veiculo(id) ON DELETE CASCADE,
    principal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_morador_veiculo UNIQUE(morador_id, veiculo_id)
);

CREATE INDEX idx_morador_veiculo_morador ON morador_veiculo(morador_id);
CREATE INDEX idx_morador_veiculo_veiculo ON morador_veiculo(veiculo_id);
