package authorization

import (
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/routes"
)

// AuthorizationServiceV2 is the new version using centralized route configuration
type AuthorizationServiceV2 struct {
	routeAdapter *routes.AuthorizationAdapter
}

// NewAuthorizationServiceV2 creates a new authorization service using centralized routes
func NewAuthorizationServiceV2(configPath string) (*AuthorizationServiceV2, error) {
	// Load routes configuration
	config, err := routes.LoadRoutesConfiguration(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load routes configuration: %w", err)
	}

	// Create authorization adapter
	adapter := routes.NewAuthorizationAdapter(config)

	return &AuthorizationServiceV2{
		routeAdapter: adapter,
	}, nil
}

// NewAuthorizationServiceV2WithConfig creates a new authorization service with provided config
func NewAuthorizationServiceV2WithConfig(config *routes.RoutesConfiguration) *AuthorizationServiceV2 {
	adapter := routes.NewAuthorizationAdapter(config)
	return &AuthorizationServiceV2{
		routeAdapter: adapter,
	}
}

// CheckPermission verifies if the user has permission to access a route
func (a *AuthorizationServiceV2) CheckPermission(method, path string, claims *jwt.Claims) (bool, error) {
	return a.routeAdapter.CheckPermission(method, path, claims)
}

// IsPublicRoute checks if a route is public (doesn't require authentication)
func (a *AuthorizationServiceV2) IsPublicRoute(method, path string) bool {
	return a.routeAdapter.IsPublicRoute(method, path)
}

// GetRoutePermissions returns all route permissions (for debugging/inspection)
func (a *AuthorizationServiceV2) GetRoutePermissions() []routes.RoutePermission {
	return a.routeAdapter.GetRoutePermissions()
}

// ValidateConfiguration validates the current route configuration
func (a *AuthorizationServiceV2) ValidateConfiguration() error {
	permissions := a.routeAdapter.GetRoutePermissions()
	if len(permissions) == 0 {
		return fmt.Errorf("no route permissions configured")
	}

	// Add more validation logic as needed
	return nil
}
