CREATE TABLE users_role (
    id BIGSERIAL PRIMARY KEY,
    users_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES role(id) ON DELETE CASCADE,
    condominium_id BIGINT REFERENCES condominium(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_users_role_condominium UNIQUE(users_id, role_id, condominium_id)
);

CREATE INDEX idx_users_role_user ON users_role(users_id);
CREATE INDEX idx_users_role_role ON users_role(role_id);
CREATE INDEX idx_users_role_condominium ON users_role(condominium_id);
