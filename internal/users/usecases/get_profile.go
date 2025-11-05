package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

type GetProfile struct {
	userRepo           ports.UserRepository
	personRepo         ports.PersonRepository
	addressRepo        ports.AddressRepository
	personDocumentRepo ports.PersonDocumentRepository
	logger             logger.Logger
}

// NewGetProfile creates a new GetProfile use case
func NewGetProfile(userRepo ports.UserRepository, personRepo ports.PersonRepository, addressRepo ports.AddressRepository, personDocumentRepo ports.PersonDocumentRepository, logger logger.Logger) *GetProfile {
	return &GetProfile{
		userRepo:           userRepo,
		personRepo:         personRepo,
		addressRepo:        addressRepo,
		personDocumentRepo: personDocumentRepo,
		logger:             logger,
	}
}

// Execute gets the user profile by retrieving user, person, and address information
func (uc *GetProfile) Execute(ctx context.Context, userID int64) (*ports.GetProfileResponse, error) {
	// Validate user ID
	if userID <= 0 {
		uc.logger.Error("Invalid user ID", "user_id", userID)
		return nil, fmt.Errorf("invalid user ID: must be greater than 0")
	}

	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user by ID", "error", err, "user_id", userID)
		return nil, fmt.Errorf("user not found")
	}

	// Check if person already exists for this user
	person, err := uc.personRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("Debug: GetByUserID error in get_profile", "error", err, "error_string", err.Error(), "userID", userID)
		if err.Error() == "person not found" || err.Error() == "sql: no rows in result set" {
			uc.logger.Error("Person not found for user", "user_id", userID)
			return nil, fmt.Errorf("profile not found")
		}
		uc.logger.Error("Failed to check existing person", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}

	// Additional safety check for nil person
	if person == nil {
		uc.logger.Error("Person is nil after successful query", "user_id", userID)
		return nil, fmt.Errorf("profile not found")
	}

	var address *domain.Address
	if person.AddressID != nil {
		address, err = uc.addressRepo.GetByID(ctx, *person.AddressID)
		if err != nil {
			uc.logger.Error("address not found", "error", err, "address_id", *person.AddressID)
			return nil, fmt.Errorf("address not found: %w", err)
		}
	}

	// Get document information concurrently (try to get CPF and CNPJ in parallel)
	var document string
	var documentType domain.DocumentType

	type docResult struct {
		doc     *domain.PersonDocument
		err     error
		docType domain.DocumentType
	}

	// Use goroutines to fetch both document types concurrently
	cpfChan := make(chan docResult, 1)
	cnpjChan := make(chan docResult, 1)

	go func() {
		doc, err := uc.personDocumentRepo.GetByPersonIDAndType(ctx, person.ID, domain.CPF)
		cpfChan <- docResult{doc, err, domain.CPF}
	}()

	go func() {
		doc, err := uc.personDocumentRepo.GetByPersonIDAndType(ctx, person.ID, domain.CNPJ)
		cnpjChan <- docResult{doc, err, domain.CNPJ}
	}()

	// Collect results - prioritize CPF over CNPJ
	cpfResult := <-cpfChan
	cnpjResult := <-cnpjChan

	if cpfResult.err == nil && cpfResult.doc != nil {
		document = cpfResult.doc.Number
		documentType = cpfResult.doc.Type
	} else if cnpjResult.err == nil && cnpjResult.doc != nil {
		document = cnpjResult.doc.Number
		documentType = cnpjResult.doc.Type
	}

	// Format the response
	profileData := &ports.ProfileData{
		Name:          person.Name,
		Email:         person.Email,
		Document:      document,
		DocumentType:  documentType,
		PhoneNumber:   person.PhoneNumber,
		Occupation:    person.Occupation,
		MaritalStatus: person.MaritalStatus,
	}

	if person.BirthdayDate != nil && !person.BirthdayDate.IsZero() {
		profileData.BirthdayDate = person.BirthdayDate.Format("2006-01-02")
	}

	if address != nil {
		profileData.Address = &ports.AddressData{
			ZipCode:      address.ZipCode,
			Street:       address.Street,
			Number:       address.Number,
			Complement:   address.Complement,
			Neighborhood: address.Neighborhood,
			City:         address.City,
			State:        address.State,
			Country:      address.Country,
		}
	}

	return &ports.GetProfileResponse{
		Profile: profileData,
		Message: "User Profile",
	}, nil
}
