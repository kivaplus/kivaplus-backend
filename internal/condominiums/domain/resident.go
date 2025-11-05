package domain

import "time"

// Resident represents a resident in a condominium unit
type Resident struct {
	ID        int64      `json:"id"`
	PersonID  int64      `json:"person_id"`
	UnitID    int64      `json:"unit_id"`
	IsOwner   bool       `json:"is_owner"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewResident creates a new resident
func NewResident(personID, unitID int64, isOwner bool) (*Resident, error) {
	now := time.Now()
	return &Resident{
		PersonID:  personID,
		UnitID:    unitID,
		IsOwner:   isOwner,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SetOwnership sets the ownership status
func (r *Resident) SetOwnership(isOwner bool) {
	r.IsOwner = isOwner
	r.UpdatedAt = time.Now()
}

// IsDeleted returns true if the resident is soft deleted
func (r *Resident) IsDeleted() bool {
	return r.DeletedAt != nil
}
