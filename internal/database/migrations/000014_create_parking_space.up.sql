CREATE TABLE parking_space (
    id BIGSERIAL PRIMARY KEY,
    unit_id BIGINT NOT NULL REFERENCES unit(id) ON DELETE CASCADE,
    num VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('coberta', 'descoberta')),
    available BOOLEAN NOT NULL DEFAULT TRUE,
    vehicle_id BIGINT REFERENCES vehicle(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_unit_num_parking UNIQUE(unit_id, num)
);

CREATE INDEX idx_parking_space_unit_id ON parking_space(unit_id);
CREATE INDEX idx_parking_space_vehicle_id ON parking_space(vehicle_id);
CREATE INDEX idx_parking_space_available ON parking_space(available);

CREATE TRIGGER update_parking_space_updated_at
    BEFORE UPDATE ON parking_space
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
