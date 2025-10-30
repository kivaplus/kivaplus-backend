CREATE TABLE employee (
    id BIGSERIAL PRIMARY KEY,
    person_id BIGINT NOT NULL REFERENCES person(id) ON DELETE CASCADE,
    condominium_id BIGINT NOT NULL REFERENCES condominium(id) ON DELETE CASCADE,
    occupation VARCHAR(100) NOT NULL,
    salary NUMERIC(12, 2),
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo', 'afastado', 'ferias')),
    observations TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT uk_person_condominium_employee UNIQUE(person_id, condominium_id),
    CONSTRAINT chk_end_date CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_employee_person_id ON employee(person_id);
CREATE INDEX idx_employee_condominium_id ON employee(condominium_id);
CREATE INDEX idx_employee_status ON employee(status);
CREATE INDEX idx_employee_deleted_at ON employee(deleted_at);

CREATE TRIGGER update_employee_updated_at
    BEFORE UPDATE ON employee
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
