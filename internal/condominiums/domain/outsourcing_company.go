package domain

import (
	"fmt"
	"time"
)

// OutsourcingCompany represents a company that provides outsourced services
type OutsourcingCompany struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	DocumentNumber string     `json:"document_number"`
	LegalName      string     `json:"legal_name"`
	PhoneNumber    string     `json:"phone_number,omitempty"`
	Email          string     `json:"email,omitempty"`
	AddressID      int64      `json:"address_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewOutsourcingCompany creates a new outsourcing company
func NewOutsourcingCompany(name, documentNumber, legalName string, addressID int64) (*OutsourcingCompany, error) {
	if name == "" {
		return nil, fmt.Errorf("company name is required")
	}
	if documentNumber == "" {
		return nil, fmt.Errorf("document number is required")
	}
	if legalName == "" {
		return nil, fmt.Errorf("legal name is required")
	}
	if addressID <= 0 {
		return nil, fmt.Errorf("address ID must be greater than 0")
	}

	now := time.Now()
	return &OutsourcingCompany{
		Name:           name,
		DocumentNumber: documentNumber,
		LegalName:      legalName,
		AddressID:      addressID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// SetContactInfo sets the company contact information
func (oc *OutsourcingCompany) SetContactInfo(phoneNumber, email string) {
	oc.PhoneNumber = phoneNumber
	oc.Email = email
	oc.UpdatedAt = time.Now()
}

// UpdateName updates the company name
func (oc *OutsourcingCompany) UpdateName(name string) error {
	if name == "" {
		return fmt.Errorf("company name cannot be empty")
	}
	oc.Name = name
	oc.UpdatedAt = time.Now()
	return nil
}

// UpdateLegalName updates the legal name
func (oc *OutsourcingCompany) UpdateLegalName(legalName string) error {
	if legalName == "" {
		return fmt.Errorf("legal name cannot be empty")
	}
	oc.LegalName = legalName
	oc.UpdatedAt = time.Now()
	return nil
}

// IsDeleted returns true if the company is soft deleted
func (oc *OutsourcingCompany) IsDeleted() bool {
	return oc.DeletedAt != nil
}
