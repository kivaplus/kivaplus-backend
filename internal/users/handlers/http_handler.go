package handlers

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kivaplus/kivaplus-backend/internal/shared/cache"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/permissions"
	"github.com/kivaplus/kivaplus-backend/internal/shared/response"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// HTTPHandler handles HTTP requests for user operations
type HTTPHandler struct {
	getProfileUC      ports.GetProfileService
	completeProfileUC ports.CompleteProfileService
	permissionChecker *permissions.EnhancedChecker
	permissionService *permissions.Service
	cache             *cache.RedisService
	logger            logger.Logger
}

// NewHTTPHandler creates a new HTTP handler with enhanced permissions
func NewHTTPHandler(
	getProfileUC ports.GetProfileService,
	completeProfileUC ports.CompleteProfileService,
	permissionChecker *permissions.EnhancedChecker,
	permissionService *permissions.Service,
	cache *cache.RedisService,
	logger logger.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		getProfileUC:      getProfileUC,
		completeProfileUC: completeProfileUC,
		permissionChecker: permissionChecker,
		permissionService: permissionService,
		cache:             cache,
		logger:            logger,
	}
}

// GetProfile handles getting user profile with enhanced validation
func (h *HTTPHandler) GetProfile(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing get profile request")

	// Validate request with permission checking
	validator := h.permissionChecker.NewValidator(ctx, request)
	enhancedClaims, err := validator.RequirePermission("profile", "read", nil)
	if err != nil {
		h.logger.Error("Permission validation failed", "error", err)
		return response.Unauthorized("Access denied"), nil
	}

	userID, err := enhancedClaims.GetUserID()
	if err != nil {
		h.logger.Error("Invalid user ID in token", "error", err)
		return response.Unauthorized("Invalid token"), nil
	}

	h.logger.Info("Validated user access", "userID", userID)

	getProfile, err := h.getProfileUC.Execute(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user profile", "error", err, "userID", userID)
		return response.InternalServerError("Failed to retrieve profile"), nil
	}

	h.logger.Info("Retrieved user profile", "userID", userID)
	return response.OK(getProfile.Profile, "Profile retrieved successfully"), nil
}

// UpdateProfile handles updating user profile with optimized token management
func (h *HTTPHandler) UpdateProfile(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing update profile request")

	// Validate request with permission checking
	validator := h.permissionChecker.NewValidator(ctx, request)
	enhancedClaims, err := validator.RequirePermission("profile", "update", nil)
	if err != nil {
		h.logger.Error("Permission validation failed", "error", err)
		return response.Unauthorized("Access denied"), nil
	}

	userID, err := enhancedClaims.GetUserID()
	if err != nil {
		h.logger.Error("Invalid user ID in token", "error", err)
		return response.Unauthorized("Invalid token"), nil
	}

	// Parse request body
	var req ports.UpdateProfileRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Failed to unmarshal update profile request", "error", err)
		return response.BadRequest("Invalid request body"), nil
	}

	// Execute the complete profile use case
	result, err := h.completeProfileUC.Execute(ctx, userID, req)
	if err != nil {
		h.logger.Error("Failed to update profile", "error", err, "userID", userID)

		// Check if it's a validation error
		if strings.Contains(err.Error(), "validation failed") {
			return response.UnprocessableEntity(err.Error()), nil
		}

		return response.InternalServerError("Failed to update profile"), nil
	}

	// Invalidate user permissions cache after profile update
	if err := h.permissionService.InvalidateUserPermissions(ctx, enhancedClaims.Claims.UserID); err != nil {
		h.logger.Warn("Failed to invalidate user permissions cache", "userID", userID, "error", err)
	}

	// For profile completion, generate new token with updated version
	if result.AccessToken != "" {
		tokens := map[string]string{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		}
		return response.ProfileUpdated(result.Profile, tokens), nil
	}

	return response.ProfileUpdated(result.Profile), nil
}

// Helper methods

// validateRequest provides a fluent interface for request validation
func (h *HTTPHandler) validateRequest(ctx context.Context, request events.APIGatewayProxyRequest) *permissions.RequestValidator {
	return h.permissionChecker.NewValidator(ctx, request)
}

// invalidateUserCache invalidates all cached data for a user
func (h *HTTPHandler) invalidateUserCache(ctx context.Context, userID string) {
	if err := h.permissionService.InvalidateUserPermissions(ctx, userID); err != nil {
		h.logger.Warn("Failed to invalidate user cache", "userID", userID, "error", err)
	}
}

// Old helper methods removed - now using shared response utilities

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ListUsers handles listing users with enhanced permission checking
func (h *HTTPHandler) ListUsers(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing list users request")

	// Validate request with admin permission checking
	validator := h.permissionChecker.NewValidator(ctx, request)
	enhancedClaims, err := validator.RequirePermission("usuarios", "list", nil)
	if err != nil {
		h.logger.Error("Permission validation failed", "error", err)
		return response.InsufficientPermissions("admin access required"), nil
	}

	// Check if user has admin access
	hasAdminAccess := enhancedClaims.HasRole(int(jwt.RoleSuperAdmin)) ||
		enhancedClaims.HasRole(int(jwt.RoleAdmin)) ||
		enhancedClaims.HasRole(int(jwt.RoleSindico))

	if !hasAdminAccess {
		return response.InsufficientPermissions("admin access required"), nil
	}

	// TODO: Implement list users use case with proper filtering
	// TODO: Apply filters based on user role (admin/sindico can only see users from their condominios)

	data := map[string]interface{}{
		"users": []interface{}{},
		"user_permissions": map[string]interface{}{
			"is_super_admin": enhancedClaims.HasRole(int(jwt.RoleSuperAdmin)),
			"is_admin":       enhancedClaims.HasRole(int(jwt.RoleAdmin)),
			"is_sindico":     enhancedClaims.HasRole(int(jwt.RoleSindico)),
			"condominiums":   enhancedClaims.GetCondominiumsWithRole(int(jwt.RoleAdmin)),
		},
	}

	return response.OK(data, "List users endpoint - implementation pending"), nil
}
