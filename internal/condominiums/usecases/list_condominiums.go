package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	userDomain "github.com/kivaplus/kivaplus-backend/internal/users/domain"
	userPorts "github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// ListCondominiums implements the list condominiums use case
type ListCondominiums struct {
	condominiumRepo ports.CondominiumRepository
	personRepo      userPorts.PersonRepository
	addressRepo     userPorts.AddressRepository
	roleRepo        userPorts.RoleRepository
	logger          logger.Logger
}

// NewListCondominiums creates a new ListCondominiums use case
func NewListCondominiums(
	condominiumRepo ports.CondominiumRepository,
	personRepo userPorts.PersonRepository,
	addressRepo userPorts.AddressRepository,
	roleRepo userPorts.RoleRepository,
	logger logger.Logger,
) *ListCondominiums {
	return &ListCondominiums{
		condominiumRepo: condominiumRepo,
		personRepo:      personRepo,
		addressRepo:     addressRepo,
		roleRepo:        roleRepo,
		logger:          logger,
	}
}

// Execute lists all condominiums where the user has any role
func (uc *ListCondominiums) Execute(ctx context.Context, userID int64) (*ports.ListCondominiumsResponse, error) {
	uc.logger.Info("Listing condominiums for user", "user_id", userID)

	// Get user roles to determine if user is super admin
	userRoles, err := uc.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user roles", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check if user is super admin with security validation
	isSuperAdmin := false
	superAdminCount := 0
	for _, role := range userRoles {
		if role.ID == 1 { // RoleSuperAdmin = 1
			isSuperAdmin = true
			superAdminCount++
		}
	}

	// Security check: If user claims to be super admin, verify uniqueness
	if isSuperAdmin {
		// Verify that this is the only super admin in the system
		allSuperAdmins, err := uc.roleRepo.GetUsersByRole(ctx, 1) // Get all users with super_admin role
		if err != nil {
			uc.logger.Error("Failed to verify super admin uniqueness", "error", err, "user_id", userID)
			return nil, fmt.Errorf("security validation failed: %w", err)
		}

		// Security violation: Multiple super admins detected
		if len(allSuperAdmins) > 1 {
			uc.logger.Error("SECURITY ALERT: Multiple super admins detected",
				"user_id", userID,
				"super_admin_count", len(allSuperAdmins),
				"super_admin_users", allSuperAdmins)
			return nil, fmt.Errorf("security violation: multiple super admins detected")
		}

		// Verify that the current user is the legitimate super admin
		if len(allSuperAdmins) == 1 && allSuperAdmins[0] != userID {
			uc.logger.Error("SECURITY ALERT: Unauthorized super admin access attempt",
				"attempting_user_id", userID,
				"legitimate_super_admin_id", allSuperAdmins[0])
			return nil, fmt.Errorf("security violation: unauthorized super admin access")
		}

		uc.logger.Info("Super admin access validated", "user_id", userID)
	}

	var condominiums []*domain.Condominium
	if isSuperAdmin {
		// Super admin sees all condominiums
		uc.logger.Info("User is super admin, listing all condominiums", "user_id", userID)
		condominiums, err = uc.condominiumRepo.GetAll(ctx)
		if err != nil {
			uc.logger.Error("Failed to get all condominiums for super admin", "error", err, "user_id", userID)
			return nil, fmt.Errorf("failed to get all condominiums: %w", err)
		}
	} else {
		// Regular user sees only condominiums where they have roles
		condominiums, err = uc.condominiumRepo.GetByUserID(ctx, userID)
		if err != nil {
			uc.logger.Error("Failed to get condominiums for user", "error", err, "user_id", userID)
			return nil, fmt.Errorf("failed to get condominiums: %w", err)
		}
	}

	// Create a map of condominium ID to user role for quick lookup
	roleMap := make(map[int64]string)
	for _, role := range userRoles {
		if role.CondominiumID != nil {
			roleMap[*role.CondominiumID] = role.Name
		}
	}

	// Use goroutines to fetch condominium data in parallel
	type condominiumResult struct {
		data *ports.CondominiumData
		err  error
	}

	resultChan := make(chan condominiumResult, len(condominiums))

	// Launch goroutines for parallel processing
	for _, condominium := range condominiums {
		go func(condo *domain.Condominium) {
			defer func() {
				if r := recover(); r != nil {
					uc.logger.Error("Panic in condominium processing goroutine", "panic", r, "condominium_id", condo.ID)
					resultChan <- condominiumResult{nil, fmt.Errorf("panic processing condominium %d", condo.ID)}
				}
			}()

			// Get address, manager, and administrator concurrently
			type fetchResult struct {
				address       *userDomain.Address
				manager       *userDomain.Person
				administrator *userDomain.Person
				addressErr    error
				managerErr    error
				adminErr      error
			}

			fetchChan := make(chan fetchResult, 1)

			go func() {
				var result fetchResult

				// Use goroutines for concurrent database calls
				addressChan := make(chan struct {
					addr *userDomain.Address
					err  error
				}, 1)
				managerChan := make(chan struct {
					person *userDomain.Person
					err    error
				}, 1)
				adminChan := make(chan struct {
					person *userDomain.Person
					err    error
				}, 1)

				// Fetch address
				go func() {
					addr, err := uc.addressRepo.GetByID(ctx, condo.AddressID)
					addressChan <- struct {
						addr *userDomain.Address
						err  error
					}{addr, err}
				}()

				// Fetch manager
				go func() {
					manager, err := uc.personRepo.GetByID(ctx, condo.ManagerID)
					managerChan <- struct {
						person *userDomain.Person
						err    error
					}{manager, err}
				}()

				// Fetch administrator (if exists)
				go func() {
					if condo.AdministratorID != nil {
						admin, err := uc.personRepo.GetByID(ctx, *condo.AdministratorID)
						adminChan <- struct {
							person *userDomain.Person
							err    error
						}{admin, err}
					} else {
						adminChan <- struct {
							person *userDomain.Person
							err    error
						}{nil, nil}
					}
				}()

				// Collect results
				addressResult := <-addressChan
				managerResult := <-managerChan
				adminResult := <-adminChan

				result.address = addressResult.addr
				result.addressErr = addressResult.err
				result.manager = managerResult.person
				result.managerErr = managerResult.err
				result.administrator = adminResult.person
				result.adminErr = adminResult.err

				fetchChan <- result
			}()

			fetchRes := <-fetchChan

			// Check for critical errors (address and manager are required)
			if fetchRes.addressErr != nil {
				uc.logger.Error("Failed to get address", "error", fetchRes.addressErr, "address_id", condo.AddressID)
				resultChan <- condominiumResult{nil, fetchRes.addressErr}
				return
			}

			if fetchRes.managerErr != nil {
				uc.logger.Error("Failed to get manager", "error", fetchRes.managerErr, "manager_id", condo.ManagerID)
				resultChan <- condominiumResult{nil, fetchRes.managerErr}
				return
			}

			// Administrator is optional, just log warning if failed
			var administrator *ports.PersonData
			if condo.AdministratorID != nil {
				if fetchRes.adminErr != nil {
					uc.logger.Warn("Failed to get administrator", "error", fetchRes.adminErr, "administrator_id", *condo.AdministratorID)
				} else if fetchRes.administrator != nil {
					administrator = &ports.PersonData{
						ID:          fetchRes.administrator.ID,
						Name:        fetchRes.administrator.Name,
						Email:       fetchRes.administrator.Email,
						PhoneNumber: fetchRes.administrator.PhoneNumber,
					}
				}
			}

			// Get user's role in this condominium
			userRole := roleMap[condo.ID]
			if userRole == "" {
				if isSuperAdmin {
					userRole = "super_admin"
				} else {
					userRole = "unknown"
				}
			}

			condominiumData := &ports.CondominiumData{
				ID:             condo.ID,
				Name:           condo.Name,
				DocumentNumber: condo.DocumentNumber,
				PhoneNumber:    condo.PhoneNumber,
				Email:          condo.Email,
				Observations:   condo.Observations,
				Address: &ports.AddressData{
					ZipCode:      fetchRes.address.ZipCode,
					Street:       fetchRes.address.Street,
					Number:       fetchRes.address.Number,
					Complement:   fetchRes.address.Complement,
					Neighborhood: fetchRes.address.Neighborhood,
					City:         fetchRes.address.City,
					State:        fetchRes.address.State,
					Country:      fetchRes.address.Country,
				},
				Manager: &ports.PersonData{
					ID:          fetchRes.manager.ID,
					Name:        fetchRes.manager.Name,
					Email:       fetchRes.manager.Email,
					PhoneNumber: fetchRes.manager.PhoneNumber,
				},
				Administrator: administrator,
				UserRole:      userRole,
			}

			resultChan <- condominiumResult{condominiumData, nil}
		}(condominium)
	}

	// Collect results from all goroutines
	var condominiumsData []*ports.CondominiumData
	for i := 0; i < len(condominiums); i++ {
		result := <-resultChan
		if result.err != nil {
			uc.logger.Error("Failed to process condominium", "error", result.err)
			continue // Skip failed condominiums
		}
		if result.data != nil {
			condominiumsData = append(condominiumsData, result.data)
		}
	}

	message := "Condominiums retrieved successfully."
	if len(condominiumsData) == 0 {
		message = "No condominiums found. Create your first condominium to get started."
	}

	return &ports.ListCondominiumsResponse{
		Condominiums: condominiumsData,
		Message:      message,
	}, nil
}
