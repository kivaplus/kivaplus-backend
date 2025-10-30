CREATE TABLE unit (
    id BIGSERIAL PRIMARY KEY,
    condominium_id BIGINT NOT NULL REFERENCES condominium(id) ON DELETE CASCADE,
    num VARCHAR(20) NOT NULL,
    building VARCHAR(10),
    floor INTEGER,
    floor_area_m2 NUMERIC(10, 2),
    property_id BIGINT NOT NULL REFERENCES person(id) ON DELETE RESTRICT,
    has_garage BOOLEAN NOT NULL DEFAULT FALSE,
    observations TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT uk_condominium_building_num UNIQUE(condominium_id, building, num)
);

CREATE INDEX idx_unit_condominium_id ON unit(condominium_id);
CREATE INDEX idx_unit_property_id ON unit(property_id);
CREATE INDEX idx_unit_num ON unit(num);
CREATE INDEX idx_unit_deleted_at ON unit(deleted_at);

CREATE TRIGGER update_unit_updated_at
    BEFORE UPDATE ON unit
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
