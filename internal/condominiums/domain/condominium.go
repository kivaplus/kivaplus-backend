package domain

import "time"

// Condominium represents a condominium in the system
type Condominium struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	DocumentNumber  string     `json:"document_number"`
	PhoneNumber     string     `json:"phone_number,omitempty"`
	Email           string     `json:"email,omitempty"`
	AddressID       int64      `json:"address_id"`
	ManagerID       int64      `json:"manager_id"`
	AdministratorID *int64     `json:"administrator_id,omitempty"`
	Observations    string     `json:"observations,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// NewCondominium creates a new condominium
func NewCondominium(name, documentNumber string, addressID, managerID int64) (*Condominium, error) {
	now := time.Now()
	return &Condominium{
		Name:           name,
		DocumentNumber: documentNumber,
		AddressID:      addressID,
		ManagerID:      managerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// SetAdministrator sets the administrator for the condominium
func (c *Condominium) SetAdministrator(administratorID int64) {
	c.AdministratorID = &administratorID
	c.UpdatedAt = time.Now()
}

// SetContactInfo sets the contact information
func (c *Condominium) SetContactInfo(phoneNumber, email string) {
	c.PhoneNumber = phoneNumber
	c.Email = email
	c.UpdatedAt = time.Now()
}

// SetObservations sets the observations
func (c *Condominium) SetObservations(observations string) {
	c.Observations = observations
	c.UpdatedAt = time.Now()
}

// IsDeleted returns true if the condominium is soft deleted
func (c *Condominium) IsDeleted() bool {
	return c.DeletedAt != nil
}
