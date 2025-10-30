CREATE TABLE condominium (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    document_number VARCHAR(18) NOT NULL UNIQUE,
    phone_number VARCHAR(20),
    email VARCHAR(255),
    address_id BIGINT NOT NULL REFERENCES address(id) ON DELETE RESTRICT,
    manager_id BIGINT NOT NULL REFERENCES person(id) ON DELETE RESTRICT,
    administrator_id BIGINT REFERENCES person(id) ON DELETE SET NULL,
    observations TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_condominium_document_number ON condominium(document_number);
CREATE INDEX idx_condominium_name ON condominium(name);
CREATE INDEX idx_condominium_manager_id ON condominium(manager_id);
CREATE INDEX idx_condominium_deleted_at ON condominium(deleted_at);

CREATE TRIGGER update_condominium_updated_at
    BEFORE UPDATE ON condominium
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
