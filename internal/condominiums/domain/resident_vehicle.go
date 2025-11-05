package domain

import (
	"fmt"
	"time"
)

// ResidentVehicle represents the relationship between a resident and their vehicles
type ResidentVehicle struct {
	ID          int64     `json:"id"`
	ResidentID  int64     `json:"resident_id"`
	VehicleID   int64     `json:"vehicle_id"`
	MainVehicle bool      `json:"main_vehicle"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewResidentVehicle creates a new resident-vehicle relationship
func NewResidentVehicle(residentID, vehicleID int64, mainVehicle bool) (*ResidentVehicle, error) {
	if residentID <= 0 {
		return nil, fmt.Errorf("resident ID must be greater than 0")
	}
	if vehicleID <= 0 {
		return nil, fmt.Errorf("vehicle ID must be greater than 0")
	}

	return &ResidentVehicle{
		ResidentID:  residentID,
		VehicleID:   vehicleID,
		MainVehicle: mainVehicle,
		CreatedAt:   time.Now(),
	}, nil
}

// SetAsMainVehicle sets this vehicle as the main vehicle for the resident
func (rv *ResidentVehicle) SetAsMainVehicle() {
	rv.MainVehicle = true
}

// UnsetAsMainVehicle removes this vehicle as the main vehicle for the resident
func (rv *ResidentVehicle) UnsetAsMainVehicle() {
	rv.MainVehicle = false
}

// IsMainVehicle returns true if this is the resident's main vehicle
func (rv *ResidentVehicle) IsMainVehicle() bool {
	return rv.MainVehicle
}
