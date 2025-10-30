CREATE TABLE outsourcing_contract (
    id BIGSERIAL PRIMARY KEY,
    condominium_id BIGINT NOT NULL REFERENCES condominium(id) ON DELETE CASCADE,
    outsourcing_company_id BIGINT NOT NULL REFERENCES outsourcing_company(id) ON DELETE RESTRICT,
    contract_number VARCHAR(100) NOT NULL UNIQUE,
    monthly_amount NUMERIC(12, 2) NOT NULL,
    due_day INTEGER NOT NULL CHECK (due_day >= 1 AND due_day <= 31),
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'suspenso', 'cancelado', 'finalizado')),
    observations TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_end_date CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_outsourcing_contract_condominium_id ON outsourcing_contract(condominium_id);
CREATE INDEX idx_outsourcing_contract_outsourcing_company_id ON outsourcing_contract(outsourcing_company_id);
CREATE INDEX idx_outsourcing_contract_status ON outsourcing_contract(status);
CREATE INDEX idx_outsourcing_contract_contract_number ON outsourcing_contract(contract_number);
CREATE INDEX idx_outsourcing_contract_deleted_at ON outsourcing_contract(deleted_at);

CREATE TRIGGER update_outsourcing_contract_updated_at
    BEFORE UPDATE ON outsourcing_contract
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
