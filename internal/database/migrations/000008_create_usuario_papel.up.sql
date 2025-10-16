CREATE TABLE usuario_papel (
    id BIGSERIAL PRIMARY KEY,
    usuario_id BIGINT NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    papel_id INTEGER NOT NULL REFERENCES papel(id) ON DELETE CASCADE,
    condominio_id BIGINT REFERENCES condominio(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_usuario_papel_condominio UNIQUE(usuario_id, papel_id, condominio_id)
);

CREATE INDEX idx_usuario_papel_usuario ON usuario_papel(usuario_id);
CREATE INDEX idx_usuario_papel_papel ON usuario_papel(papel_id);
CREATE INDEX idx_usuario_papel_condominio ON usuario_papel(condominio_id);
