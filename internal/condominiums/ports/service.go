package ports

import (
	"context"
)

// CreateCondominiumRequest represents the request to create a condominium
type CreateCondominiumRequest struct {
	Name           string                          `json:"name"`
	DocumentNumber string                          `json:"document_number"`
	PhoneNumber    string                          `json:"phone_number,omitempty"`
	Email          string                          `json:"email,omitempty"`
	Observations   string                          `json:"observations,omitempty"`
	Address        CreateCondominiumAddressRequest `json:"address"`
}

// CreateCondominiumAddressRequest represents the address data for condominium creation
type CreateCondominiumAddressRequest struct {
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
}

// CreateCondominiumResponse represents the response from creating a condominium
type CreateCondominiumResponse struct {
	Condominium  *CondominiumData `json:"condominium"`
	Message      string           `json:"message"`
	AccessToken  string           `json:"access_token,omitempty"`
	RefreshToken string           `json:"refresh_token,omitempty"`
}

// CondominiumData represents the formatted condominium data for responses
type CondominiumData struct {
	ID             int64        `json:"id"`
	Name           string       `json:"name"`
	DocumentNumber string       `json:"document_number"`
	PhoneNumber    string       `json:"phone_number,omitempty"`
	Email          string       `json:"email,omitempty"`
	Observations   string       `json:"observations,omitempty"`
	Address        *AddressData `json:"address"`
	Manager        *PersonData  `json:"manager"`
	Administrator  *PersonData  `json:"administrator,omitempty"`
	UserRole       string       `json:"user_role"` // Role of the current user in this condominium
}

// AddressData represents the formatted address data
type AddressData struct {
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
}

// PersonData represents the formatted person data
type PersonData struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// ListCondominiumsResponse represents the response from listing condominiums
type ListCondominiumsResponse struct {
	Condominiums []*CondominiumData `json:"condominiums"`
	Message      string             `json:"message"`
}

// GetCondominiumResponse represents the response from getting a single condominium
type GetCondominiumResponse struct {
	Condominium *CondominiumData `json:"condominium"`
	Message     string           `json:"message"`
}

// CreateCondominiumService defines the interface for creating condominiums
type CreateCondominiumService interface {
	Execute(ctx context.Context, userID int64, req CreateCondominiumRequest) (*CreateCondominiumResponse, error)
}

// ListCondominiumsService defines the interface for listing user's condominiums
type ListCondominiumsService interface {
	Execute(ctx context.Context, userID int64) (*ListCondominiumsResponse, error)
}

// GetCondominiumService defines the interface for getting a single condominium
type GetCondominiumService interface {
	Execute(ctx context.Context, userID int64, condominiumID int64) (*GetCondominiumResponse, error)
}
