CREATE TABLE unidade (
    id BIGSERIAL PRIMARY KEY,
    condominio_id BIGINT NOT NULL REFERENCES condominio(id) ON DELETE CASCADE,
    numero VARCHAR(20) NOT NULL,
    bloco VARCHAR(10),
    andar INTEGER,
    area_m2 NUMERIC(10, 2),
    proprietario_id BIGINT NOT NULL REFERENCES pessoa(id) ON DELETE RESTRICT,
    tem_garagem BOOLEAN NOT NULL DEFAULT FALSE,
    observacoes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT uk_condominio_bloco_numero UNIQUE(condominio_id, bloco, numero)
);

CREATE INDEX idx_unidade_condominio_id ON unidade(condominio_id);
CREATE INDEX idx_unidade_proprietario_id ON unidade(proprietario_id);
CREATE INDEX idx_unidade_numero ON unidade(numero);
CREATE INDEX idx_unidade_deleted_at ON unidade(deleted_at);

CREATE TRIGGER update_unidade_updated_at
    BEFORE UPDATE ON unidade
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
