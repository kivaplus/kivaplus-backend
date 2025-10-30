CREATE TABLE address (
    id BIGSERIAL PRIMARY KEY,
    postal_code VARCHAR(10) NOT NULL,
    streey VARCHAR(255) NOT NULL,
    num VARCHAR(20) NOT NULL,
    complement VARCHAR(255),
    neighborhood VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(2) NOT NULL,
    country VARCHAR(2) NOT NULL DEFAULT 'BR',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_address_postal_code ON address(postal_code);
CREATE INDEX idx_address_city_state ON address(city, state);

CREATE TRIGGER update_address_updated_at
    BEFORE UPDATE ON address
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
