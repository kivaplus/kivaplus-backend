package domain

import (
	"time"
)

// PersonType represents the type of person (individual or legal entity)
type PersonType string

const (
	IndividualEntity PersonType = "PF"
	LegalEntity      PersonType = "PJ"
)

// MariageStatis represents marital status
type MaritalStatus string

const (
	Solteiro     MaritalStatus = "solteiro"
	Casado       MaritalStatus = "casado"
	Divorciado   MaritalStatus = "divorciado"
	Viuvo        MaritalStatus = "viuvo"
	UniaoEstavel MaritalStatus = "uniao_estavel"
	Outro        MaritalStatus = "outro"
)

// Person represents a person (individual or legal entity)
type Person struct {
	ID            int64         `json:"id"`
	Name          string        `json:"name"`
	PersonType    PersonType    `json:"person_type"`
	PhoneNumber   string        `json:"phone_number,omitempty"`
	Email         string        `json:"email,omitempty"`
	AddressID     *int64        `json:"address_id,omitempty"`
	Occupation    string        `json:"occupation,omitempty"`
	MaritalStatus MaritalStatus `json:"marital_status,omitempty"`
	BirthdayDate  time.Time     `json:"birthday_date,omitempty"`
	UserID        *int64        `json:"user_id,omitempty"`
	Observations  string        `json:"observations,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	DeletedAt     *time.Time    `json:"deleted_at,omitempty"`
}

// NewPerson creates a new person
func NewPerson(name string, email string, userID int64) (*Person, error) {
	now := time.Now()
	return &Person{
		Name:       name,
		PersonType: IndividualEntity, // Default to 'PF' (individual person)
		Email:      email,
		AddressID:  nil, // Will be set later when address is created
		UserID:     &userID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// UpdateProfile updates the person's profile information
func (p *Person) UpdateProfile(name string, phoneNumber, email, occupation string, addressID *int64) {
	p.Name = name
	p.PhoneNumber = phoneNumber
	p.Email = email
	p.Occupation = occupation
	p.AddressID = addressID
	p.UpdatedAt = time.Now()
}

// SetMaritalStatus sets the marital status
func (p *Person) SetMaritalStatus(maritalStatus MaritalStatus) {
	p.MaritalStatus = maritalStatus
	p.UpdatedAt = time.Now()
}

// SetBirthdayDate sets the birth date
func (p *Person) SetBirthdayDate(birthdayDate time.Time) {
	p.BirthdayDate = birthdayDate
	p.UpdatedAt = time.Now()
}

// LinkToUsuario links this person to a user
func (p *Person) LinkToUser(userID int64) {
	p.UserID = &userID
	p.UpdatedAt = time.Now()
}

// IsDeleted returns true if the person is soft deleted
func (p *Person) IsDeleted() bool {
	return p.DeletedAt != nil
}
