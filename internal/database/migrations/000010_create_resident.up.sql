CREATE TABLE resident (
    id BIGSERIAL PRIMARY KEY,
    person_id BIGINT NOT NULL UNIQUE REFERENCES person(id) ON DELETE CASCADE,
    unit_id BIGINT NOT NULL REFERENCES unit(id) ON DELETE CASCADE,
    resident_type VARCHAR(20) NOT NULL CHECK (resident_type IN ('proprietario', 'inquilino', 'dependente')),
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo')),
    start_date DATE NOT NULL,
    end_date DATE,
    financial_responsible BOOLEAN NOT NULL DEFAULT FALSE,
    observations TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_end_date CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_resident_person_id ON resident(person_id);
CREATE INDEX idx_resident_unit_id ON resident(unit_id);
CREATE INDEX idx_resident_status ON resident(status);
CREATE INDEX idx_resident_deleted_at ON resident(deleted_at);

CREATE TRIGGER update_resident_updated_at
    BEFORE UPDATE ON resident
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
