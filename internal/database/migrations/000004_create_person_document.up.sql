CREATE TABLE person_document (
    id BIGSERIAL PRIMARY KEY,
    person_document_id BIGINT NOT NULL REFERENCES person(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('CPF', 'RG', 'CNH', 'CNPJ')),
    num VARCHAR(30) NOT NULL,
    issuing_authority VARCHAR(20),
    issue_date DATE,
    expiration_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_person_type_document_person UNIQUE(person_document_id, type)
);

CREATE INDEX idx_person_document_person_id ON person_document(person_document_id);
CREATE INDEX idx_person_document_num ON person_document(num);

CREATE TRIGGER update_person_document_updated_at
    BEFORE UPDATE ON person_document
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
