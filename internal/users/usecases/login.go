package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/validator"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// Login implements the user login use case
type Login struct {
	userRepo   ports.UserRepository
	roleRepo   ports.RoleRepository
	personRepo ports.PersonRepository
	jwtService *jwt.Service
	logger     logger.Logger
}

// NewLogin creates a new Login use case
func NewLogin(userRepo ports.UserRepository, personRepo ports.PersonRepository, roleRepo ports.RoleRepository, logger logger.Logger) *Login {
	return &Login{
		userRepo:   userRepo,
		personRepo: personRepo,
		roleRepo:   roleRepo,
		jwtService: jwt.NewJWTService(),
		logger:     logger,
	}
}

// Execute authenticates a user and returns a JWT token
func (uc *Login) Execute(ctx context.Context, req ports.LoginRequest) (*ports.LoginResponse, error) {
	uc.logger.Info("User login attempt", "email", req.Email)

	// Validate input
	v := validator.New()
	v.ValidateEmail("email", req.Email)
	v.ValidateRequired("password", req.Password)

	if v.HasErrors() {
		uc.logger.Error("Login validation failed", "errors", v.Errors())
		return nil, fmt.Errorf("validation failed: %v", v.Errors())
	}

	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		uc.logger.Error("Failed to get user by email", "error", err, "email", req.Email)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive() {
		uc.logger.Warn("Login attempt for inactive user", "email", req.Email)
		return nil, fmt.Errorf("user account is inactive")
	}

	// Validate password
	if !user.ValidatePassword(req.Password) {
		uc.logger.Warn("Invalid password for user", "email", req.Email)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Get user roles
	roles, err := uc.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		uc.logger.Error("Failed to get user roles", "error", err, "userID", user.ID)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	uc.logger.Info("User roles loaded", "userID", user.ID, "rolesCount", len(roles))

	// Check if profile is complete
	hasCompleteProfile, err := uc.personRepo.HasCompleteProfile(ctx, user.ID)
	if err != nil {
		uc.logger.Error("Failed to check profile completeness", "error", err, "userID", user.ID)
		return nil, fmt.Errorf("failed to check profile completeness: %w", err)
	}

	profileStatus := "incomplete"
	if hasCompleteProfile {
		profileStatus = "complete"
	}

	uc.logger.Info("Profile status checked", "userID", user.ID, "profileStatus", profileStatus)

	// Generate JWT tokens (access and refresh)
	userIDStr := fmt.Sprintf("%d", user.ID)
	accessToken, err := uc.jwtService.GenerateToken(userIDStr, user.Email, roles, profileStatus)
	if err != nil {
		uc.logger.Error("Failed to generate access token", "error", err, "email", req.Email)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.jwtService.GenerateRefreshToken(userIDStr, user.Email)
	if err != nil {
		uc.logger.Error("Failed to generate refresh token", "error", err, "email", req.Email)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last access
	user.UpdateLastAccess()
	if err := uc.userRepo.UpdateLastAccess(ctx, user.ID); err != nil {
		uc.logger.Warn("Failed to update last access", "error", err, "email", req.Email)
		// Don't fail the login for this
	}

	uc.logger.Info("User logged in successfully", "email", req.Email, "id", user.ID)

	return &ports.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}
