package usecases

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/validator"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// RefreshToken implements the token refresh use case
type RefreshToken struct {
	userRepo   ports.UserRepository
	roleRepo   ports.RoleRepository
	personRepo ports.PersonRepository
	jwtService *jwt.Service
	logger     logger.Logger
}

// NewRefreshToken creates a new RefreshToken use case
func NewRefreshToken(userRepo ports.UserRepository, personRepo ports.PersonRepository, roleRepo ports.RoleRepository, logger logger.Logger) *RefreshToken {
	return &RefreshToken{
		userRepo:   userRepo,
		personRepo: personRepo,
		roleRepo:   roleRepo,
		jwtService: jwt.NewJWTService(),
		logger:     logger,
	}
}

// Execute refreshes an access token using a valid refresh token
func (uc *RefreshToken) Execute(ctx context.Context, req ports.RefreshTokenRequest) (*ports.RefreshTokenResponse, error) {
	uc.logger.Info("Token refresh attempt")

	// Validate input
	v := validator.New()
	v.ValidateRequired("refresh_token", req.RefreshToken)

	if v.HasErrors() {
		uc.logger.Error("Refresh token validation failed", "errors", v.Errors())
		return nil, fmt.Errorf("validation failed: %v", v.Errors())
	}

	// Validate refresh token
	claims, err := uc.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		uc.logger.Error("Invalid refresh token", "error", err)
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user by ID from token
	userIDStr := claims.UserID
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		uc.logger.Error("Invalid user ID in token", "error", err, "userID", userIDStr)
		return nil, fmt.Errorf("invalid user ID in token")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user by ID", "error", err, "userID", userID)
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is still active
	if !user.IsActive() {
		uc.logger.Warn("Token refresh attempt for inactive user", "userID", userID)
		return nil, fmt.Errorf("user account is inactive")
	}

	// Get fresh user roles (in case they changed)
	roles, err := uc.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		uc.logger.Error("Failed to get user roles", "error", err, "userID", user.ID)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check current profile status
	hasCompleteProfile, err := uc.personRepo.HasCompleteProfile(ctx, user.ID)
	if err != nil {
		uc.logger.Error("Failed to check profile completeness", "error", err, "userID", user.ID)
		return nil, fmt.Errorf("failed to check profile completeness: %w", err)
	}

	profileStatus := "incomplete"
	if hasCompleteProfile {
		profileStatus = "complete"
	}

	// Generate new access token with fresh data
	userIDStr = fmt.Sprintf("%d", user.ID)
	newAccessToken, err := uc.jwtService.GenerateToken(userIDStr, user.Email, roles, profileStatus)
	if err != nil {
		uc.logger.Error("Failed to generate new access token", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	// Optionally generate new refresh token (for token rotation)
	newRefreshToken, err := uc.jwtService.GenerateRefreshToken(userIDStr, user.Email)
	if err != nil {
		uc.logger.Error("Failed to generate new refresh token", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	uc.logger.Info("Token refreshed successfully", "userID", userID)

	return &ports.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
