package domain

import "time"

// VehicleType represents the type of vehicle
type VehicleType string

const (
	Car   VehicleType = "carro"
	Bike  VehicleType = "moto"
	Truck VehicleType = "caminhonete"
	Other VehicleType = "outros"
)

// Vehicle represents a vehicle
type Vehicle struct {
	ID           int64       `json:"id"`
	PlateNumber  string      `json:"plate_number"`
	Model        string      `json:"model"`
	Make         string      `json:"make"`
	Color        string      `json:"color,omitempty"`
	Year         int         `json:"year,omitempty"`
	VehicleType  VehicleType `json:"vehicle_type"`
	Observations string      `json:"observations,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	DeletedAt    *time.Time  `json:"deleted_at,omitempty"`
}

// NewVehicle creates a new vehicle
func NewVehicle(plateNumber, model, make string) (*Vehicle, error) {
	now := time.Now()
	return &Vehicle{
		PlateNumber: plateNumber,
		Model:       model,
		Make:        make,
		VehicleType: Car, // Default to car
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// SetColor sets the vehicle color
func (v *Vehicle) SetColor(color string) {
	v.Color = color
	v.UpdatedAt = time.Now()
}

// SetYear sets the vehicle year
func (v *Vehicle) SetYear(year int) {
	v.Year = year
	v.UpdatedAt = time.Now()
}

// SetVehicleType sets the vehicle type
func (v *Vehicle) SetVehicleType(vehicleType VehicleType) {
	v.VehicleType = vehicleType
	v.UpdatedAt = time.Now()
}

// SetObservations sets observations
func (v *Vehicle) SetObservations(observations string) {
	v.Observations = observations
	v.UpdatedAt = time.Now()
}

// IsDeleted returns true if the vehicle is soft deleted
func (v *Vehicle) IsDeleted() bool {
	return v.DeletedAt != nil
}
