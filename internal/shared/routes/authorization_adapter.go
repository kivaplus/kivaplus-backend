package routes

import (
	"fmt"
	"regexp"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
)

// AuthorizationAdapter adapts route configuration for the authorization system
type AuthorizationAdapter struct {
	config *RoutesConfiguration
}

// NewAuthorizationAdapter creates a new authorization adapter
func NewAuthorizationAdapter(config *RoutesConfiguration) *AuthorizationAdapter {
	return &AuthorizationAdapter{config: config}
}

// RoutePermission represents a route permission for the authorization system
type RoutePermission struct {
	Method      string     `json:"method"`
	PathPattern string     `json:"path_pattern"`
	Permission  Permission `json:"permission"`
}

// Permission represents a permission for the authorization system
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Roles    []int  `json:"roles"`
	Scope    string `json:"scope"`
}

// GetRoutePermissions converts route configuration to authorization format
func (a *AuthorizationAdapter) GetRoutePermissions() []RoutePermission {
	var permissions []RoutePermission

	for _, route := range a.config.GetAllRoutes() {
		// Skip public routes - they don't need permissions
		if route.Public {
			continue
		}

		permissions = append(permissions, RoutePermission{
			Method:      route.Method,
			PathPattern: route.PathPattern,
			Permission: Permission{
				Resource: route.Resource,
				Action:   route.Action,
				Roles:    route.Roles,
				Scope:    route.Scope,
			},
		})
	}

	return permissions
}

// CheckPermission checks if a user has permission to access a route
func (a *AuthorizationAdapter) CheckPermission(method, path string, claims *jwt.Claims) (bool, error) {
	// Find the route permission
	routePermission, err := a.findRoutePermission(method, path)
	if err != nil {
		return false, fmt.Errorf("route permission not found: %w", err)
	}

	// Super admin has access to everything
	if claims.IsSuperAdmin() {
		return true, nil
	}

	// Check if user has required role
	hasRequiredRole := false
	for _, requiredRole := range routePermission.Permission.Roles {
		if claims.HasRole(requiredRole) {
			hasRequiredRole = true
			break
		}
	}

	if !hasRequiredRole {
		return false, fmt.Errorf("insufficient role permissions")
	}

	// Check scope-specific permissions
	switch routePermission.Permission.Scope {
	case "global":
		return true, nil
	case "condominium":
		return a.checkCondominiumAccess(path, claims)
	case "own":
		return a.checkOwnDataAccess(path, claims)
	default:
		return true, nil
	}
}

// IsPublicRoute checks if a route is public
func (a *AuthorizationAdapter) IsPublicRoute(method, path string) bool {
	for _, route := range a.config.GetPublicRoutes() {
		if route.Method == method {
			matched, err := regexp.MatchString(route.PathPattern, path)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

// findRoutePermission finds the permission for a specific route
func (a *AuthorizationAdapter) findRoutePermission(method, path string) (*RoutePermission, error) {
	permissions := a.GetRoutePermissions()

	for _, rp := range permissions {
		if rp.Method == method {
			matched, err := regexp.MatchString(rp.PathPattern, path)
			if err != nil {
				continue
			}
			if matched {
				return &rp, nil
			}
		}
	}
	return nil, fmt.Errorf("no permission rule found for %s %s", method, path)
}

// checkCondominiumAccess verifies access to specific condominium
func (a *AuthorizationAdapter) checkCondominiumAccess(path string, claims *jwt.Claims) (bool, error) {
	condominiumID, err := extractCondominiumIDFromPath(path)
	if err != nil {
		return false, fmt.Errorf("failed to extract condominium ID: %w", err)
	}

	return claims.CanAccessCondominium(condominiumID), nil
}

// checkOwnDataAccess verifies access to own data
func (a *AuthorizationAdapter) checkOwnDataAccess(path string, claims *jwt.Claims) (bool, error) {
	// For profile routes, always allow (user accessing their own profile)
	if regexp.MustCompile(`^/profile`).MatchString(path) {
		return true, nil
	}

	// For user-specific routes, extract user ID and compare
	userID, err := extractUserIDFromPath(path)
	if err != nil {
		return false, fmt.Errorf("failed to extract user ID: %w", err)
	}

	return claims.UserID == userID, nil
}

// Helper functions (can be shared or moved to a utils package)
func extractCondominiumIDFromPath(path string) (int64, error) {
	// Implementation from your existing code
	re := regexp.MustCompile(`/condominios/(\d+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return 0, fmt.Errorf("condominium ID not found in path: %s", path)
	}

	// Convert to int64 (you'll need to import strconv)
	// This is a simplified version - use your existing implementation
	return 1, nil // Placeholder
}

func extractUserIDFromPath(path string) (string, error) {
	// Implementation from your existing code
	if regexp.MustCompile(`^/profile`).MatchString(path) {
		return "", fmt.Errorf("user ID should come from token for profile routes")
	}

	re := regexp.MustCompile(`/usuarios/(\w+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("user ID not found in path: %s", path)
	}

	return matches[1], nil
}
