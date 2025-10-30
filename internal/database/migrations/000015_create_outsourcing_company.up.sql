CREATE TABLE outsourcing_company (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    document_number VARCHAR(18) NOT NULL UNIQUE,
    legal_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    email VARCHAR(255),
    address_id BIGINT NOT NULL REFERENCES address(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_outsourcing_company_document_number ON outsourcing_company(document_number);
CREATE INDEX idx_outsourcing_company_name ON outsourcing_company(name);
CREATE INDEX idx_outsourcing_company_deleted_at ON outsourcing_company(deleted_at);

CREATE TRIGGER update_outsourcing_company_updated_at
    BEFORE UPDATE ON outsourcing_company
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
