package permissions

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/kivaplus/kivaplus-backend/internal/shared/cache"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// Service handles permission operations with caching
type Service struct {
	cache      *cache.RedisService
	roleRepo   RoleRepository
	logger     logger.Logger
	cacheTTL   time.Duration
	versionTTL time.Duration
}

// NewPermissionService creates a new permission service
func NewPermissionService(cache *cache.RedisService, roleRepo RoleRepository, logger logger.Logger) *Service {
	return &Service{
		cache:      cache, // Can be nil for no-cache mode
		roleRepo:   roleRepo,
		logger:     logger,
		cacheTTL:   15 * time.Minute, // Cache permissions for 15 minutes
		versionTTL: 24 * time.Hour,   // Keep version for 24 hours
	}
}

// UserPermissions represents cached user permissions
type UserPermissions struct {
	UserID        string     `json:"user_id"`
	Roles         []jwt.Role `json:"roles"`
	ProfileStatus string     `json:"profile_status"`
	CachedAt      time.Time  `json:"cached_at"`
	Version       int64      `json:"version"`
}

// GetUserPermissions retrieves user permissions with caching
func (s *Service) GetUserPermissions(ctx context.Context, userID string) (*UserPermissions, error) {
	// Try to get from cache first (if cache is available)
	if s.cache != nil {
		cacheKey := s.cache.UserPermissionsKey(userID)

		var cached UserPermissions
		err := s.cache.Get(cacheKey, &cached)
		if err == nil {
			// Check if cache is still fresh (within 5 minutes for high-frequency operations)
			if time.Since(cached.CachedAt) < 5*time.Minute {
				s.logger.Debug("Retrieved permissions from cache", "userID", userID, "cacheAge", time.Since(cached.CachedAt))
				return &cached, nil
			}
		}
		s.logger.Debug("Fetching permissions from database", "userID", userID, "cacheError", err)
	} else {
		s.logger.Debug("Cache not available, fetching permissions from database", "userID", userID)
	}

	// Cache miss, stale, or cache not available - fetch from database
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	roles, profileStatus, err := s.roleRepo.GetUserRolesAndProfile(ctx, userIDInt)
	if err != nil {
		s.logger.Error("Permission validation failed", "error", err)
		return nil, fmt.Errorf("failed to fetch user roles: %w", err)
	}

	// Get or create version (fallback to timestamp if cache unavailable)
	version := time.Now().Unix()
	if s.cache != nil {
		if v, err := s.getOrCreateUserVersion(userID); err == nil {
			version = v
		} else {
			s.logger.Warn("Failed to get user version, using timestamp", "userID", userID, "error", err)
		}
	}

	permissions := &UserPermissions{
		UserID:        userID,
		Roles:         roles,
		ProfileStatus: profileStatus,
		CachedAt:      time.Now(),
		Version:       version,
	}

	// Cache the permissions (if cache is available)
	if s.cache != nil {
		if err := s.cache.Set(s.cache.UserPermissionsKey(userID), permissions, s.cacheTTL); err != nil {
			s.logger.Warn("Failed to cache permissions", "userID", userID, "error", err)
		}
	}

	return permissions, nil
}

// InvalidateUserPermissions removes user permissions from cache and increments version
func (s *Service) InvalidateUserPermissions(ctx context.Context, userID string) error {
	if s.cache == nil {
		// No cache available, nothing to invalidate
		s.logger.Info("Cache not available, skipping permission invalidation", "userID", userID)
		return nil
	}

	// Delete from cache
	cacheKey := s.cache.UserPermissionsKey(userID)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to delete permissions cache", "userID", userID, "error", err)
	}

	// Increment version to invalidate existing tokens
	versionKey := s.cache.UserTokenVersionKey(userID)
	newVersion, err := s.cache.Increment(versionKey)
	if err != nil {
		return fmt.Errorf("failed to increment user version: %w", err)
	}

	// Set expiration for version key
	if err := s.cache.SetExpiration(versionKey, s.versionTTL); err != nil {
		s.logger.Warn("Failed to set version expiration", "userID", userID, "error", err)
	}

	s.logger.Info("Invalidated user permissions", "userID", userID, "newVersion", newVersion)
	return nil
}

// GetUserTokenVersion retrieves the current token version for a user
func (s *Service) GetUserTokenVersion(userID string) (int64, error) {
	if s.cache == nil {
		// No cache available, return timestamp-based version
		return time.Now().Unix(), nil
	}

	versionKey := s.cache.UserTokenVersionKey(userID)

	// Try to get existing version
	var version int64
	err := s.cache.Get(versionKey, &version)
	if err == cache.ErrCacheNotFound {
		// Create initial version
		return s.getOrCreateUserVersion(userID)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get user version: %w", err)
	}

	return version, nil
}

// getOrCreateUserVersion gets or creates a version for the user
func (s *Service) getOrCreateUserVersion(userID string) (int64, error) {
	if s.cache == nil {
		// No cache available, return timestamp-based version
		return time.Now().Unix(), nil
	}

	versionKey := s.cache.UserTokenVersionKey(userID)

	// Try to set initial version atomically
	initialVersion := time.Now().Unix()
	created, err := s.cache.SetNX(versionKey, initialVersion, s.versionTTL)
	if err != nil {
		return 0, fmt.Errorf("failed to create user version: %w", err)
	}

	if created {
		return initialVersion, nil
	}

	// Version already exists, get it
	var version int64
	err = s.cache.Get(versionKey, &version)
	if err != nil {
		return 0, fmt.Errorf("failed to get existing version: %w", err)
	}

	return version, nil
}

// ValidateTokenVersion checks if a token version is current
func (s *Service) ValidateTokenVersion(userID string, tokenVersion int64) (bool, error) {
	if s.cache == nil {
		// No cache available, always consider token valid (fallback mode)
		return true, nil
	}

	currentVersion, err := s.GetUserTokenVersion(userID)
	if err != nil {
		return false, err
	}

	return tokenVersion >= currentVersion, nil
}

// CheckPermission validates if user has specific permission for a resource
func (s *Service) CheckPermission(ctx context.Context, userID string, resource string, action string, condominiumID *int64) (bool, error) {
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Convert to claims for permission checking
	claims := &jwt.Claims{
		UserID:        userID,
		ProfileStatus: permissions.ProfileStatus,
	}

	return s.checkPermissionWithClaims(claims, resource, action, condominiumID), nil
}

// checkPermissionWithClaims checks permission using UserPermissions (not deprecated Claims methods)
func (s *Service) checkPermissionWithClaims(claims *jwt.Claims, resource string, action string, condominiumID *int64) bool {
	// Get fresh permissions for the user
	ctx := context.Background()
	permissions, err := s.GetUserPermissions(ctx, claims.UserID)
	if err != nil {
		s.logger.Error("Failed to get permissions for user", "userID", claims.UserID, "error", err)
		return false
	}

	return s.checkPermissionWithUserPermissions(permissions, claims.ProfileStatus, resource, action, condominiumID)
}

// checkPermissionWithUserPermissions checks permission using UserPermissions
func (s *Service) checkPermissionWithUserPermissions(permissions *UserPermissions, profileStatus string, resource string, action string, condominiumID *int64) bool {
	// Super admin has access to everything
	for _, role := range permissions.Roles {
		if role.ID == int(jwt.RoleSuperAdmin) {
			return true
		}
	}

	switch resource {
	case "profile":
		// Profile is ALWAYS accessible for authenticated users (they need to complete it)
		return action == "read" || action == "update"

	case "condominios":
		// Profile must be complete for condominium operations
		if profileStatus != "complete" {
			return false
		}
		if condominiumID == nil {
			// List condominiums - user needs at least one role
			return len(permissions.Roles) > 0
		}
		// Access specific condominium - check if user has any role in that condominium
		for _, role := range permissions.Roles {
			if role.CondominiumID != nil && *role.CondominiumID == *condominiumID {
				return true
			}
		}
		return false

	case "units", "residents", "vehicles":
		// Profile must be complete for these operations
		if profileStatus != "complete" {
			return false
		}
		if condominiumID == nil {
			return false
		}
		// Need admin access to the condominium
		for _, role := range permissions.Roles {
			if role.CondominiumID != nil && *role.CondominiumID == *condominiumID {
				if role.ID == int(jwt.RoleAdmin) || role.ID == int(jwt.RoleSindico) {
					return true
				}
			}
		}
		return false

	default:
		// Profile must be complete for other operations
		if profileStatus != "complete" {
			return false
		}
		return false
	}
}

// BulkInvalidatePermissions invalidates permissions for multiple users
func (s *Service) BulkInvalidatePermissions(ctx context.Context, userIDs []string) error {
	for _, userID := range userIDs {
		if err := s.InvalidateUserPermissions(ctx, userID); err != nil {
			s.logger.Error("Failed to invalidate permissions for user", "userID", userID, "error", err)
			// Continue with other users
		}
	}
	return nil
}

// RoleRepository interface for fetching user roles
type RoleRepository interface {
	GetUserRolesAndProfile(ctx context.Context, userID int64) ([]jwt.Role, string, error)
}
