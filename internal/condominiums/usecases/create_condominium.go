package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/validator"
	addressDomain "github.com/kivaplus/kivaplus-backend/internal/users/domain"
	userPorts "github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// CreateCondominium implements the create condominium use case
type CreateCondominium struct {
	condominiumRepo ports.CondominiumRepository
	personRepo      userPorts.PersonRepository
	addressRepo     userPorts.AddressRepository
	roleRepo        userPorts.RoleRepository
	userRepo        userPorts.UserRepository
	jwtService      *jwt.Service
	logger          logger.Logger
}

// NewCreateCondominium creates a new CreateCondominium use case
func NewCreateCondominium(
	condominiumRepo ports.CondominiumRepository,
	personRepo userPorts.PersonRepository,
	addressRepo userPorts.AddressRepository,
	roleRepo userPorts.RoleRepository,
	userRepo userPorts.UserRepository,
	jwtService *jwt.Service,
	logger logger.Logger,
) *CreateCondominium {
	return &CreateCondominium{
		condominiumRepo: condominiumRepo,
		personRepo:      personRepo,
		addressRepo:     addressRepo,
		roleRepo:        roleRepo,
		userRepo:        userRepo,
		jwtService:      jwtService,
		logger:          logger,
	}
}

// Execute creates a new condominium and assigns the user as manager and admin
func (uc *CreateCondominium) Execute(ctx context.Context, userID int64, req ports.CreateCondominiumRequest) (*ports.CreateCondominiumResponse, error) {
	uc.logger.Info("Creating new condominium", "user_id", userID, "name", req.Name)

	// Validate input
	v := validator.New()
	v.ValidateRequired("name", req.Name)
	v.ValidateMinLength("name", req.Name, 2)
	v.ValidateMaxLength("name", req.Name, 255)
	v.ValidateRequired("document_number", req.DocumentNumber)
	v.ValidateDocument("document_number", req.DocumentNumber, "CNPJ") // Condominiums should have CNPJ

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
		uc.logger.Error("Condominium validation failed", "errors", v.Errors())
		return nil, fmt.Errorf("validation failed: %v", v.Errors())
	}

	// Check if condominium already exists
	exists, err := uc.condominiumRepo.Exists(ctx, req.DocumentNumber)
	if err != nil {
		uc.logger.Error("Failed to check if condominium exists", "error", err)
		return nil, fmt.Errorf("failed to check condominium existence: %w", err)
	}

	if exists {
		uc.logger.Warn("Condominium already exists", "document_number", req.DocumentNumber)
		return nil, fmt.Errorf("condominium with document number %s already exists", req.DocumentNumber)
	}

	// Get the person who will be the manager
	person, err := uc.personRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get person for user", "error", err, "user_id", userID)
		return nil, fmt.Errorf("user profile not found: %w", err)
	}

	// Create address for the condominium
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

	// Create the condominium
	condominium, err := domain.NewCondominium(req.Name, req.DocumentNumber, createdAddress.ID, person.ID)
	if err != nil {
		uc.logger.Error("Failed to create condominium domain object", "error", err)
		return nil, fmt.Errorf("failed to create condominium: %w", err)
	}

	// Set optional fields
	if req.PhoneNumber != "" || req.Email != "" {
		condominium.SetContactInfo(req.PhoneNumber, req.Email)
	}
	if req.Observations != "" {
		condominium.SetObservations(req.Observations)
	}

	// Set the user as administrator as well
	condominium.SetAdministrator(person.ID)

	if err := uc.condominiumRepo.Create(ctx, condominium); err != nil {
		uc.logger.Error("Failed to create condominium", "error", err)
		return nil, fmt.Errorf("failed to create condominium: %w", err)
	}

	// Use goroutines for role management operations
	type roleResult struct {
		operation string
		err       error
	}

	roleResultChan := make(chan roleResult, 3)

	// Remove old sindico role concurrently
	go func() {
		err := uc.roleRepo.RemoveUserRoleWithoutCondominium(ctx, userID, 3)
		if err != nil {
			uc.logger.Warn("Failed to remove old sindico role", "error", err, "user_id", userID)
		}
		roleResultChan <- roleResult{"remove_old_sindico", err}
	}()

	// Assign sindico role concurrently
	go func() {
		err := uc.roleRepo.AddUserRole(ctx, userID, 3, &condominium.ID)
		roleResultChan <- roleResult{"add_sindico", err}
	}()

	// Assign admin role concurrently
	go func() {
		err := uc.roleRepo.AddUserRole(ctx, userID, 2, &condominium.ID)
		roleResultChan <- roleResult{"add_admin", err}
	}()

	// Collect results from role operations
	var sindicoErr, adminErr error
	for i := 0; i < 3; i++ {
		result := <-roleResultChan
		switch result.operation {
		case "add_sindico":
			if result.err != nil {
				uc.logger.Error("Failed to assign sindico role", "error", result.err, "user_id", userID, "condominium_id", condominium.ID)
				sindicoErr = result.err
			}
		case "add_admin":
			if result.err != nil {
				uc.logger.Error("Failed to assign admin role", "error", result.err, "user_id", userID, "condominium_id", condominium.ID)
				adminErr = result.err
			}
		}
	}

	// Check for critical role assignment errors
	if sindicoErr != nil {
		return nil, fmt.Errorf("failed to assign sindico role: %w", sindicoErr)
	}
	if adminErr != nil {
		return nil, fmt.Errorf("failed to assign admin role: %w", adminErr)
	}

	uc.logger.Info("Condominium created successfully", "condominium_id", condominium.ID, "user_id", userID)

	// Format the response
	condominiumData := &ports.CondominiumData{
		ID:             condominium.ID,
		Name:           condominium.Name,
		DocumentNumber: condominium.DocumentNumber,
		PhoneNumber:    condominium.PhoneNumber,
		Email:          condominium.Email,
		Observations:   condominium.Observations,
		Address: &ports.AddressData{
			ZipCode:      createdAddress.ZipCode,
			Street:       createdAddress.Street,
			Number:       createdAddress.Number,
			Complement:   createdAddress.Complement,
			Neighborhood: createdAddress.Neighborhood,
			City:         createdAddress.City,
			State:        createdAddress.State,
			Country:      createdAddress.Country,
		},
		Manager: &ports.PersonData{
			ID:          person.ID,
			Name:        person.Name,
			Email:       person.Email,
			PhoneNumber: person.PhoneNumber,
		},
		Administrator: &ports.PersonData{
			ID:          person.ID,
			Name:        person.Name,
			Email:       person.Email,
			PhoneNumber: person.PhoneNumber,
		},
		UserRole: "sindico", // The creator becomes the sindico
	}

	// With the new dynamic permission system, we don't need to generate new tokens
	// The existing token will work with the new roles automatically
	uc.logger.Info("Condominium created with dynamic permissions", "condominium_id", condominium.ID, "user_id", userID)

	return &ports.CreateCondominiumResponse{
		Condominium: condominiumData,
		Message:     "Condominium created successfully.",
		// No new tokens needed - existing token works with new permissions
	}, nil
}
