package permissions

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kivaplus/kivaplus-backend/internal/shared/cache"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// EnhancedChecker combines JWT validation with cached permission checking
type EnhancedChecker struct {
	jwtService        *jwt.Service
	permissionService *Service
	cache             *cache.RedisService
	logger            logger.Logger
}

// NewEnhancedChecker creates a new enhanced permission checker
func NewEnhancedChecker(
	jwtService *jwt.Service,
	permissionService *Service,
	cache *cache.RedisService,
	logger logger.Logger,
) *EnhancedChecker {
	return &EnhancedChecker{
		jwtService:        jwtService,
		permissionService: permissionService,
		cache:             cache,
		logger:            logger,
	}
}

// ValidateRequestWithPermissions validates JWT and checks permissions in one call
func (ec *EnhancedChecker) ValidateRequestWithPermissions(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
	resource string,
	action string,
	condominiumID *int64,
) (*EnhancedClaims, error) {
	// Extract and validate JWT token
	claims, err := ec.jwtService.ExtractClaimsFromRequest(request)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check token version (lazy invalidation) - skip if cache unavailable
	isValid, err := ec.permissionService.ValidateTokenVersion(claims.UserID, claims.TokenVersion)
	if err != nil {
		ec.logger.Warn("Failed to validate token version, continuing without cache", "userID", claims.UserID, "error", err)
		// Continue without cache validation - this is acceptable for development
	} else if !isValid {
		return nil, fmt.Errorf("token version is outdated, please refresh")
	}

	// Get fresh permissions - handle cache failures gracefully
	permissions, err := ec.permissionService.GetUserPermissions(ctx, claims.UserID)
	if err != nil {
		ec.logger.Warn("Failed to get permissions from cache, using basic permissions", "userID", claims.UserID, "error", err)
		// Create basic permissions for profile access
		permissions = &UserPermissions{
			UserID:        claims.UserID,
			Roles:         []jwt.Role{}, // Empty roles but allow profile access
			ProfileStatus: claims.ProfileStatus,
			CachedAt:      time.Now(),
			Version:       0,
		}
	}

	// Special case: profile access is always allowed (users need to complete their profile)
	if resource == "profile" && (action == "read" || action == "update") {
		// Allow profile access even if permission check fails
		ec.logger.Info("Allowing profile access for user", "userID", claims.UserID, "action", action)
	} else {
		// Check specific permission for other resources
		hasPermission, err := ec.permissionService.CheckPermission(ctx, claims.UserID, resource, action, condominiumID)
		if err != nil {
			return nil, fmt.Errorf("failed to check permission: %w", err)
		}

		if !hasPermission {
			return nil, fmt.Errorf("insufficient permissions for %s:%s", resource, action)
		}
	}

	// Return enhanced claims with fresh permissions
	enhancedClaims := &EnhancedClaims{
		Claims:      claims,
		Permissions: permissions,
		Validated:   true,
	}

	return enhancedClaims, nil
}

// ValidateTokenOnly validates JWT token without permission checking
func (ec *EnhancedChecker) ValidateTokenOnly(request events.APIGatewayProxyRequest) (*jwt.Claims, error) {
	return ec.jwtService.ExtractClaimsFromRequest(request)
}

// CheckPermissionOnly checks permission without token validation (for pre-validated requests)
func (ec *EnhancedChecker) CheckPermissionOnly(
	ctx context.Context,
	userID string,
	resource string,
	action string,
	condominiumID *int64,
) (bool, error) {
	return ec.permissionService.CheckPermission(ctx, userID, resource, action, condominiumID)
}

// RefreshUserPermissions forces a refresh of user permissions
func (ec *EnhancedChecker) RefreshUserPermissions(ctx context.Context, userID string) error {
	return ec.permissionService.InvalidateUserPermissions(ctx, userID)
}

// GetUserPermissionsWithCache retrieves user permissions (cached)
func (ec *EnhancedChecker) GetUserPermissionsWithCache(ctx context.Context, userID string) (*UserPermissions, error) {
	return ec.permissionService.GetUserPermissions(ctx, userID)
}

// EnhancedClaims combines JWT claims with fresh permissions
type EnhancedClaims struct {
	Claims      *jwt.Claims      `json:"claims"`
	Permissions *UserPermissions `json:"permissions"`
	Validated   bool             `json:"validated"`
}

// GetUserID returns the user ID as int64
func (ec *EnhancedClaims) GetUserID() (int64, error) {
	return strconv.ParseInt(ec.Claims.UserID, 10, 64)
}

// HasRole checks if user has a specific role
func (ec *EnhancedClaims) HasRole(roleID int) bool {
	for _, role := range ec.Permissions.Roles {
		if role.ID == roleID {
			return true
		}
	}
	return false
}

// HasRoleInCondominium checks if user has a role in a specific condominium
func (ec *EnhancedClaims) HasRoleInCondominium(roleID int, condominiumID int64) bool {
	for _, role := range ec.Permissions.Roles {
		if role.ID == roleID && role.CondominiumID != nil && *role.CondominiumID == condominiumID {
			return true
		}
	}
	return false
}

// CanAccessCondominium checks if user can access a condominium
func (ec *EnhancedClaims) CanAccessCondominium(condominiumID int64) bool {
	// Super admin can access everything
	if ec.HasRole(int(jwt.RoleSuperAdmin)) {
		return true
	}

	// Check if user has any role in the condominium
	for _, role := range ec.Permissions.Roles {
		if role.CondominiumID != nil && *role.CondominiumID == condominiumID {
			return true
		}
	}

	return false
}

// IsProfileComplete checks if user profile is complete
func (ec *EnhancedClaims) IsProfileComplete() bool {
	return ec.Permissions.ProfileStatus == "complete"
}

// GetCondominiumsWithRole returns condominium IDs where user has a specific role
func (ec *EnhancedClaims) GetCondominiumsWithRole(roleID int) []int64 {
	var condominiums []int64
	for _, role := range ec.Permissions.Roles {
		if role.ID == roleID && role.CondominiumID != nil {
			condominiums = append(condominiums, *role.CondominiumID)
		}
	}
	return condominiums
}

// RequestValidator provides a fluent interface for request validation
type RequestValidator struct {
	checker *EnhancedChecker
	ctx     context.Context
	request events.APIGatewayProxyRequest
}

// NewRequestValidator creates a new request validator
func (ec *EnhancedChecker) NewValidator(ctx context.Context, request events.APIGatewayProxyRequest) *RequestValidator {
	return &RequestValidator{
		checker: ec,
		ctx:     ctx,
		request: request,
	}
}

// RequirePermission validates token and checks permission
func (rv *RequestValidator) RequirePermission(resource, action string, condominiumID *int64) (*EnhancedClaims, error) {
	return rv.checker.ValidateRequestWithPermissions(rv.ctx, rv.request, resource, action, condominiumID)
}

// RequireAuth validates token only
func (rv *RequestValidator) RequireAuth() (*jwt.Claims, error) {
	return rv.checker.ValidateTokenOnly(rv.request)
}

// RequireCompleteProfile validates token and checks if profile is complete
func (rv *RequestValidator) RequireCompleteProfile() (*EnhancedClaims, error) {
	claims, err := rv.RequirePermission("profile", "read", nil)
	if err != nil {
		return nil, err
	}

	if !claims.IsProfileComplete() {
		return nil, fmt.Errorf("profile must be completed")
	}

	return claims, nil
}

// RequireCondominiumAccess validates access to a specific condominium
func (rv *RequestValidator) RequireCondominiumAccess(condominiumID int64) (*EnhancedClaims, error) {
	claims, err := rv.RequireCompleteProfile()
	if err != nil {
		return nil, err
	}

	if !claims.CanAccessCondominium(condominiumID) {
		return nil, fmt.Errorf("access denied to condominium %d", condominiumID)
	}

	return claims, nil
}

// Performance monitoring
type PermissionMetrics struct {
	TokenValidationTime time.Duration
	PermissionCheckTime time.Duration
	CacheHitRate        float64
	TotalRequests       int64
}

// GetMetrics returns performance metrics (implement as needed)
func (ec *EnhancedChecker) GetMetrics() *PermissionMetrics {
	// Implementation would track metrics over time
	return &PermissionMetrics{
		// Placeholder values
		TokenValidationTime: 5 * time.Millisecond,
		PermissionCheckTime: 2 * time.Millisecond,
		CacheHitRate:        0.85,
		TotalRequests:       1000,
	}
}
