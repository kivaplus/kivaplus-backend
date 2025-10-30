CREATE TABLE person (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    person_type VARCHAR(2) NOT NULL CHECK (person_type IN ('PF', 'PJ')),
    phone_number VARCHAR(20),
    email VARCHAR(255),
    address_id BIGINT NOT NULL REFERENCES address(id) ON DELETE RESTRICT,
    occupation VARCHAR(100),
    marital_status VARCHAR(20) CHECK (marital_status IN ('solteiro', 'casado', 'divorciado', 'viuvo', 'uniao_estavel', 'outro')),
    birthday DATE,
    user_id BIGINT,
    observation TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_person_name ON person(name);
CREATE INDEX idx_person_email ON person(email);
CREATE INDEX idx_person_type ON person(person_type);
CREATE INDEX idx_person_deleted_at ON person(deleted_at);

CREATE TRIGGER update_person_updated_at
    BEFORE UPDATE ON person
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
