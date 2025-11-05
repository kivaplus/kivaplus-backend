package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/validator"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// CreateUser implements the create user use case
type CreateUser struct {
	userRepo   ports.UserRepository
	personRepo ports.PersonRepository
	roleRepo   ports.RoleRepository
	logger     logger.Logger
}

// NewCreateUser creates a new CreateUser use case
func NewCreateUser(
	userRepo ports.UserRepository,
	personRepo ports.PersonRepository,
	roleRepo ports.RoleRepository,
	logger logger.Logger,
) *CreateUser {
	return &CreateUser{
		userRepo:   userRepo,
		personRepo: personRepo,
		roleRepo:   roleRepo,
		logger:     logger,
	}
}

// Execute creates a new user atomically (User + Person + Default Role)
func (uc *CreateUser) Execute(ctx context.Context, req ports.CreateUserRequest) (*ports.CreateUserResponse, error) {
	uc.logger.Info("Creating new user", "email", req.Email)

	// Validate input (outside transaction for performance)
	v := validator.New()
	v.ValidateRequired("name", req.Name)
	v.ValidateEmail("email", req.Email)
	v.ValidatePassword("password", req.Password)
	v.ValidatePassword("confirm_password", req.ConfirmPassword)

	if req.Password != req.ConfirmPassword {
		v.AddError("confirm_password", "passwords do not match")
	}

	if v.HasErrors() {
		uc.logger.Error("Validation failed", "errors", v.Errors())
		return nil, fmt.Errorf("validation failed: %v", v.Errors())
	}

	// Check if user already exists (outside transaction for performance)
	exists, err := uc.userRepo.Exists(ctx, req.Email)
	if err != nil {
		uc.logger.Error("Failed to check if user exists", "error", err, "email", req.Email)
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}

	if exists {
		uc.logger.Warn("User already exists", "email", req.Email)
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// 1. Create new user
	user, err := domain.NewUser(req.Email, req.Password)
	if err != nil {
		uc.logger.Error("Failed to create user domain object", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save user to repository
	if err := uc.userRepo.Create(ctx, user); err != nil {
		uc.logger.Error("Failed to save user to repository", "error", err, "email", req.Email)
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// 2. Create person record without address (address_id will be NULL)
	person, err := domain.NewPerson(req.Name, req.Email, user.ID)
	if err != nil {
		uc.logger.Error("Failed to create person domain object", "error", err)
		return nil, fmt.Errorf("failed to create person: %w", err)
	}

	// AddressID is already set to nil in NewPerson function
	uc.logger.Info("Creating person with marital status", "maritalStatus", person.MaritalStatus, "name", person.Name)

	if err := uc.personRepo.Create(ctx, person); err != nil {
		uc.logger.Error("Failed to save person to repository", "error", err, "userID", user.ID, "maritalStatus", person.MaritalStatus)
		return nil, fmt.Errorf("failed to save person: %w", err)
	}

	// Note: No need to update user with person_id anymore
	// The relationship is maintained through person.user_id

	// 3. Assign default role
	defaultRoleID := 3 // Sindico role (default for new users)
	uc.logger.Info("Assigning default role", "userID", user.ID, "roleID", defaultRoleID)
	if err := uc.roleRepo.AddUserRole(ctx, user.ID, defaultRoleID, nil); err != nil {
		uc.logger.Error("Failed to assign default role", "error", err, "userID", user.ID, "roleID", defaultRoleID)
		return nil, fmt.Errorf("failed to assign default role: %w", err)
	}
	uc.logger.Info("Default role assigned successfully", "userID", user.ID, "roleID", defaultRoleID)

	uc.logger.Info("User created successfully", "email", req.Email, "userID", user.ID)

	return &ports.CreateUserResponse{
		User:   user,
		Person: person,
	}, nil
}
