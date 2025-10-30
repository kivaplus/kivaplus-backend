CREATE TABLE resident_vehicle (
    id BIGSERIAL PRIMARY KEY,
    resident_id BIGINT NOT NULL REFERENCES resident(id) ON DELETE CASCADE,
    vehicle_id BIGINT NOT NULL REFERENCES vehicle(id) ON DELETE CASCADE,
    main_vehicle BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_resident_vehicle UNIQUE(resident_id, vehicle_id)
);

CREATE INDEX idx_resident_vehicle_resident ON resident_vehicle(resident_id);
CREATE INDEX idx_resident_vehicle_vehicle ON resident_vehicle(vehicle_id);
