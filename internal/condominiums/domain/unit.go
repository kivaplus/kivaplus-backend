package domain

import "time"

// Unit represents a unit within a condominium
type Unit struct {
	ID            int64      `json:"id"`
	CondominiumID int64      `json:"condominium_id"`
	Number        string     `json:"number"`
	Building      string     `json:"building,omitempty"`
	Floor         int        `json:"floor,omitempty"`
	FloorAreaM2   float64    `json:"floor_area_m2,omitempty"`
	PropertyID    int64      `json:"property_id"`
	HasGarage     bool       `json:"has_garage"`
	Observations  string     `json:"observations,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// NewUnit creates a new unit
func NewUnit(condominiumID int64, number string, propertyID int64) (*Unit, error) {
	now := time.Now()
	return &Unit{
		CondominiumID: condominiumID,
		Number:        number,
		PropertyID:    propertyID,
		HasGarage:     false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// SetBuilding sets the building information
func (u *Unit) SetBuilding(building string) {
	u.Building = building
	u.UpdatedAt = time.Now()
}

// SetFloor sets the floor information
func (u *Unit) SetFloor(floor int) {
	u.Floor = floor
	u.UpdatedAt = time.Now()
}

// SetFloorArea sets the floor area
func (u *Unit) SetFloorArea(areaM2 float64) {
	u.FloorAreaM2 = areaM2
	u.UpdatedAt = time.Now()
}

// SetGarage sets garage availability
func (u *Unit) SetGarage(hasGarage bool) {
	u.HasGarage = hasGarage
	u.UpdatedAt = time.Now()
}

// SetObservations sets observations
func (u *Unit) SetObservations(observations string) {
	u.Observations = observations
	u.UpdatedAt = time.Now()
}

// IsDeleted returns true if the unit is soft deleted
func (u *Unit) IsDeleted() bool {
	return u.DeletedAt != nil
}
