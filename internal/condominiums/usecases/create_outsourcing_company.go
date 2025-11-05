package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/validator"
	addressDomain "github.com/kivaplus/kivaplus-backend/internal/users/domain"
	userPorts "github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// CreateOutsourcingCompany implements the create outsourcing company use case
type CreateOutsourcingCompany struct {
	companyRepo ports.OutsourcingCompanyRepository
	addressRepo userPorts.AddressRepository
	logger      logger.Logger
}

// NewCreateOutsourcingCompany creates a new CreateOutsourcingCompany use case
func NewCreateOutsourcingCompany(
	companyRepo ports.OutsourcingCompanyRepository,
	addressRepo userPorts.AddressRepository,
	logger logger.Logger,
) *CreateOutsourcingCompany {
	return &CreateOutsourcingCompany{
		companyRepo: companyRepo,
		addressRepo: addressRepo,
		logger:      logger,
	}
}

// CreateOutsourcingCompanyRequest represents the request to create an outsourcing company
type CreateOutsourcingCompanyRequest struct {
	Name           string                      `json:"name"`
	DocumentNumber string                      `json:"document_number"`
	LegalName      string                      `json:"legal_name"`
	PhoneNumber    string                      `json:"phone_number,omitempty"`
	Email          string                      `json:"email,omitempty"`
	Address        CreateCompanyAddressRequest `json:"address"`
}

// CreateCompanyAddressRequest represents the address data for creating a company
type CreateCompanyAddressRequest struct {
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
}

// CreateOutsourcingCompanyResponse represents the response after creating an outsourcing company
type CreateOutsourcingCompanyResponse struct {
	Company *OutsourcingCompanyData `json:"company"`
	Message string                  `json:"message"`
}

// OutsourcingCompanyData represents outsourcing company data in responses
type OutsourcingCompanyData struct {
	ID             int64        `json:"id"`
	Name           string       `json:"name"`
	DocumentNumber string       `json:"document_number"`
	LegalName      string       `json:"legal_name"`
	PhoneNumber    string       `json:"phone_number,omitempty"`
	Email          string       `json:"email,omitempty"`
	Address        *AddressData `json:"address,omitempty"`
}

// AddressData represents address data in responses
type AddressData struct {
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
}

// Execute creates a new outsourcing company
func (uc *CreateOutsourcingCompany) Execute(ctx context.Context, req CreateOutsourcingCompanyRequest) (*CreateOutsourcingCompanyResponse, error) {
	uc.logger.Info("Creating new outsourcing company", "name", req.Name, "document_number", req.DocumentNumber)

	// Validate input
	v := validator.New()
	v.ValidateRequired("name", req.Name)
	v.ValidateMinLength("name", req.Name, 2)
	v.ValidateMaxLength("name", req.Name, 255)
	v.ValidateRequired("document_number", req.DocumentNumber)
	v.ValidateDocument("document_number", req.DocumentNumber, "CNPJ") // Companies should have CNPJ
	v.ValidateRequired("legal_name", req.LegalName)
	v.ValidateMinLength("legal_name", req.LegalName, 2)
	v.ValidateMaxLength("legal_name", req.LegalName, 255)

	if req.Email != "" {
		v.ValidateEmail("email", req.Email)
	}

	// Validate address
	v.ValidateRequired("address.zip_code", req.Address.ZipCode)
	v.ValidateRequired("address.street", req.Address.Street)
	v.ValidateRequired("address.number", req.Address.Number)
	v.ValidateRequired("address.neighborhood", req.Address.Neighborhood)
	v.ValidateRequired("address.city", req.Address.City)
	v.ValidateRequired("address.state", req.Address.State)
	v.ValidateRequired("address.country", req.Address.Country)

	if v.HasErrors() {
		uc.logger.Error("Outsourcing company validation failed", "errors", v.Errors())
		return nil, fmt.Errorf("validation failed: %v", v.Errors())
	}

	// Check if company already exists
	exists, err := uc.companyRepo.Exists(ctx, req.DocumentNumber)
	if err != nil {
		uc.logger.Error("Failed to check company existence", "error", err)
		return nil, fmt.Errorf("failed to check company existence: %w", err)
	}

	if exists {
		uc.logger.Warn("Outsourcing company already exists", "document_number", req.DocumentNumber)
		return nil, fmt.Errorf("outsourcing company with document number %s already exists", req.DocumentNumber)
	}

	// Create address for the company
	address, err := addressDomain.NewAddress(
		req.Address.ZipCode,
		req.Address.Street,
		req.Address.Number,
		req.Address.Complement,
		req.Address.Neighborhood,
		req.Address.City,
		req.Address.State,
		req.Address.Country,
	)
	if err != nil {
		uc.logger.Error("Failed to create address domain object", "error", err)
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	createdAddress, err := uc.addressRepo.Create(ctx, address)
	if err != nil {
		uc.logger.Error("Failed to create address", "error", err)
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	// Create the outsourcing company
	company, err := domain.NewOutsourcingCompany(req.Name, req.DocumentNumber, req.LegalName, createdAddress.ID)
	if err != nil {
		uc.logger.Error("Failed to create outsourcing company domain object", "error", err)
		return nil, fmt.Errorf("failed to create outsourcing company: %w", err)
	}

	// Set optional contact information
	if req.PhoneNumber != "" || req.Email != "" {
		company.SetContactInfo(req.PhoneNumber, req.Email)
	}

	// Save to database
	if err := uc.companyRepo.Create(ctx, company); err != nil {
		uc.logger.Error("Failed to create outsourcing company", "error", err)
		return nil, fmt.Errorf("failed to create outsourcing company: %w", err)
	}

	uc.logger.Info("Outsourcing company created successfully", "company_id", company.ID, "name", req.Name)

	// Format response
	companyData := &OutsourcingCompanyData{
		ID:             company.ID,
		Name:           company.Name,
		DocumentNumber: company.DocumentNumber,
		LegalName:      company.LegalName,
		PhoneNumber:    company.PhoneNumber,
		Email:          company.Email,
		Address: &AddressData{
			ZipCode:      createdAddress.ZipCode,
			Street:       createdAddress.Street,
			Number:       createdAddress.Number,
			Complement:   createdAddress.Complement,
			Neighborhood: createdAddress.Neighborhood,
			City:         createdAddress.City,
			State:        createdAddress.State,
			Country:      createdAddress.Country,
		},
	}

	return &CreateOutsourcingCompanyResponse{
		Company: companyData,
		Message: "Outsourcing company created successfully.",
	}, nil
}
