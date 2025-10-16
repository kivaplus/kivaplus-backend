CREATE TABLE servico (
    id BIGSERIAL PRIMARY KEY,
    contrato_id BIGINT NOT NULL REFERENCES contrato_terceirizado(id) ON DELETE CASCADE,
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    valor NUMERIC(12, 2) NOT NULL,
    frequencia VARCHAR(20) NOT NULL CHECK (frequencia IN ('unico', 'diario', 'semanal', 'mensal', 'anual')),
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_servico_contrato_id ON servico(contrato_id);
CREATE INDEX idx_servico_nome ON servico(nome);

CREATE TRIGGER update_servico_updated_at
    BEFORE UPDATE ON servico
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
