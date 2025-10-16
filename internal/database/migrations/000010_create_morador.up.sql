CREATE TABLE morador (
    id BIGSERIAL PRIMARY KEY,
    pessoa_id BIGINT NOT NULL UNIQUE REFERENCES pessoa(id) ON DELETE CASCADE,
    unidade_id BIGINT NOT NULL REFERENCES unidade(id) ON DELETE CASCADE,
    tipo VARCHAR(20) NOT NULL CHECK (tipo IN ('proprietario', 'inquilino', 'dependente')),
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo')),
    data_entrada DATE NOT NULL,
    data_saida DATE,
    responsavel_financeiro BOOLEAN NOT NULL DEFAULT FALSE,
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_data_saida CHECK (data_saida IS NULL OR data_saida >= data_entrada)
);

CREATE INDEX idx_morador_pessoa_id ON morador(pessoa_id);
CREATE INDEX idx_morador_unidade_id ON morador(unidade_id);
CREATE INDEX idx_morador_status ON morador(status);
CREATE INDEX idx_morador_deleted_at ON morador(deleted_at);

CREATE TRIGGER update_morador_updated_at
    BEFORE UPDATE ON morador
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
