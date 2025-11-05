package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	userPorts "github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// GetCondominium implements the get condominium use case
type GetCondominium struct {
	condominiumRepo ports.CondominiumRepository
	personRepo      userPorts.PersonRepository
	addressRepo     userPorts.AddressRepository
	roleRepo        userPorts.RoleRepository
	logger          logger.Logger
}

// NewGetCondominium creates a new GetCondominium use case
func NewGetCondominium(
	condominiumRepo ports.CondominiumRepository,
	personRepo userPorts.PersonRepository,
	addressRepo userPorts.AddressRepository,
	roleRepo userPorts.RoleRepository,
	logger logger.Logger,
) *GetCondominium {
	return &GetCondominium{
		condominiumRepo: condominiumRepo,
		personRepo:      personRepo,
		addressRepo:     addressRepo,
		roleRepo:        roleRepo,
		logger:          logger,
	}
}

// Execute gets a specific condominium by ID
func (uc *GetCondominium) Execute(ctx context.Context, userID int64, condominiumID int64) (*ports.GetCondominiumResponse, error) {
	uc.logger.Info("Getting condominium", "user_id", userID, "condominium_id", condominiumID)

	// Get the condominium
	condominium, err := uc.condominiumRepo.GetByID(ctx, condominiumID)
	if err != nil {
		uc.logger.Error("Failed to get condominium", "error", err, "condominium_id", condominiumID)
		return nil, fmt.Errorf("condominium not found")
	}

	// Get address
	address, err := uc.addressRepo.GetByID(ctx, condominium.AddressID)
	if err != nil {
		uc.logger.Error("Failed to get address", "error", err, "address_id", condominium.AddressID)
		return nil, fmt.Errorf("failed to get condominium address: %w", err)
	}

	// Get manager
	manager, err := uc.personRepo.GetByID(ctx, condominium.ManagerID)
	if err != nil {
		uc.logger.Error("Failed to get manager", "error", err, "manager_id", condominium.ManagerID)
		return nil, fmt.Errorf("failed to get condominium manager: %w", err)
	}

	// Get administrator (optional)
	var administrator *ports.PersonData
	if condominium.AdministratorID != nil {
		admin, err := uc.personRepo.GetByID(ctx, *condominium.AdministratorID)
		if err != nil {
			uc.logger.Warn("Failed to get administrator", "error", err, "administrator_id", *condominium.AdministratorID)
		} else {
			administrator = &ports.PersonData{
				ID:          admin.ID,
				Name:        admin.Name,
				Email:       admin.Email,
				PhoneNumber: admin.PhoneNumber,
			}
		}
	}

	// Get user's role in this condominium
	userRoles, err := uc.roleRepo.GetUserRolesInCondominium(ctx, userID, condominiumID)
	if err != nil {
		uc.logger.Error("Failed to get user roles", "error", err, "user_id", userID, "condominium_id", condominiumID)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	userRole := "unknown"
	if len(userRoles) > 0 {
		// Use the highest priority role (assuming lower ID = higher priority)
		userRole = userRoles[0].Name
		for _, role := range userRoles {
			if role.ID < userRoles[0].ID {
				userRole = role.Name
			}
		}
	}

	condominiumData := &ports.CondominiumData{
		ID:             condominium.ID,
		Name:           condominium.Name,
		DocumentNumber: condominium.DocumentNumber,
		PhoneNumber:    condominium.PhoneNumber,
		Email:          condominium.Email,
		Observations:   condominium.Observations,
		Address: &ports.AddressData{
			ZipCode:      address.ZipCode,
			Street:       address.Street,
			Number:       address.Number,
			Complement:   address.Complement,
			Neighborhood: address.Neighborhood,
			City:         address.City,
			State:        address.State,
			Country:      address.Country,
		},
		Manager: &ports.PersonData{
			ID:          manager.ID,
			Name:        manager.Name,
			Email:       manager.Email,
			PhoneNumber: manager.PhoneNumber,
		},
		Administrator: administrator,
		UserRole:      userRole,
	}

	return &ports.GetCondominiumResponse{
		Condominium: condominiumData,
		Message:     "Condominium retrieved successfully.",
	}, nil
}
