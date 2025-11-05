package domain

import "time"

// ParkingSpaceType represents the type of parking space
type ParkingSpaceType string

const (
	Covered   ParkingSpaceType = "coberta"
	Uncovered ParkingSpaceType = "descoberta"
)

// ParkingSpace represents a parking space
type ParkingSpace struct {
	ID        int64            `json:"id"`
	UnitID    int64            `json:"unit_id"`
	Number    string           `json:"number"`
	Type      ParkingSpaceType `json:"type"`
	Available bool             `json:"available"`
	VehicleID *int64           `json:"vehicle_id,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// NewParkingSpace creates a new parking space
func NewParkingSpace(unitID int64, number string, spaceType ParkingSpaceType) (*ParkingSpace, error) {
	now := time.Now()
	return &ParkingSpace{
		UnitID:    unitID,
		Number:    number,
		Type:      spaceType,
		Available: true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// AssignVehicle assigns a vehicle to the parking space
func (ps *ParkingSpace) AssignVehicle(vehicleID int64) {
	ps.VehicleID = &vehicleID
	ps.Available = false
	ps.UpdatedAt = time.Now()
}

// RemoveVehicle removes the vehicle from the parking space
func (ps *ParkingSpace) RemoveVehicle() {
	ps.VehicleID = nil
	ps.Available = true
	ps.UpdatedAt = time.Now()
}

// IsOccupied returns true if the parking space is occupied
func (ps *ParkingSpace) IsOccupied() bool {
	return ps.VehicleID != nil
}
