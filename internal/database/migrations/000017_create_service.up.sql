CREATE TABLE service (
    id BIGSERIAL PRIMARY KEY,
    outsourcing_contract_id BIGINT NOT NULL REFERENCES outsourcing_contract(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    value NUMERIC(12, 2) NOT NULL,
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('unico', 'diario', 'semanal', 'mensal', 'anual')),
    observations TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_service_contract_id ON service(outsourcing_contract_id);
CREATE INDEX idx_service_name ON service(name);

CREATE TRIGGER update_service_updated_at
    BEFORE UPDATE ON service
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
