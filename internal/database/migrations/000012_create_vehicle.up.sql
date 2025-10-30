CREATE TABLE vehicle (
    id BIGSERIAL PRIMARY KEY,
    plate_number VARCHAR(10) NOT NULL UNIQUE,
    model VARCHAR(100) NOT NULL,
    make VARCHAR(100) NOT NULL,
    color VARCHAR(50),
    year INTEGER,
    vehicle_type VARCHAR(20) NOT NULL DEFAULT 'carro' CHECK (vehicle_type IN ('carro', 'moto', 'caminhonete', 'outros')),
    observations TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_vehicle_plate_number ON vehicle(plate_number);
CREATE INDEX idx_vehicle_vehicle_type ON vehicle(vehicle_type);
CREATE INDEX idx_vehicle_deleted_at ON vehicle(deleted_at);

CREATE TRIGGER update_vehicle_updated_at
    BEFORE UPDATE ON vehicle
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
