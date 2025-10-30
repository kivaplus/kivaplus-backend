package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RouteConfig represents a single route configuration
type RouteConfig struct {
	// Route identification
	Name        string `json:"name" yaml:"name"`
	Method      string `json:"method" yaml:"method"`
	Path        string `json:"path" yaml:"path"`
	PathPattern string `json:"path_pattern" yaml:"path_pattern"` // Regex pattern for matching

	// Authorization
	Public   bool   `json:"public" yaml:"public"`
	Roles    []int  `json:"roles" yaml:"roles"`
	Resource string `json:"resource" yaml:"resource"`
	Action   string `json:"action" yaml:"action"`
	Scope    string `json:"scope" yaml:"scope"` // "global", "condominium", "own"

	// Infrastructure
	Lambda      string `json:"lambda" yaml:"lambda"` // Which Lambda handles this route
	Description string `json:"description" yaml:"description"`
}

// RouteGroup represents a logical grouping of routes
type RouteGroup struct {
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description" yaml:"description"`
	Routes      []RouteConfig `json:"routes" yaml:"routes"`
}

// RoutesConfiguration holds all route configurations
type RoutesConfiguration struct {
	Version string       `json:"version" yaml:"version"`
	Groups  []RouteGroup `json:"groups" yaml:"groups"`
}

// GetAllRoutes returns all routes from all groups
func (rc *RoutesConfiguration) GetAllRoutes() []RouteConfig {
	var allRoutes []RouteConfig
	for _, group := range rc.Groups {
		allRoutes = append(allRoutes, group.Routes...)
	}
	return allRoutes
}

// GetPublicRoutes returns only public routes
func (rc *RoutesConfiguration) GetPublicRoutes() []RouteConfig {
	var publicRoutes []RouteConfig
	for _, route := range rc.GetAllRoutes() {
		if route.Public {
			publicRoutes = append(publicRoutes, route)
		}
	}
	return publicRoutes
}

// GetProtectedRoutes returns only protected routes
func (rc *RoutesConfiguration) GetProtectedRoutes() []RouteConfig {
	var protectedRoutes []RouteConfig
	for _, route := range rc.GetAllRoutes() {
		if !route.Public {
			protectedRoutes = append(protectedRoutes, route)
		}
	}
	return protectedRoutes
}

// GetRoutesByLambda returns routes handled by a specific Lambda
func (rc *RoutesConfiguration) GetRoutesByLambda(lambdaName string) []RouteConfig {
	var routes []RouteConfig
	for _, route := range rc.GetAllRoutes() {
		if route.Lambda == lambdaName {
			routes = append(routes, route)
		}
	}
	return routes
}

// FindRoute finds a route by method and path pattern
func (rc *RoutesConfiguration) FindRoute(method, path string) (*RouteConfig, error) {
	for _, route := range rc.GetAllRoutes() {
		if route.Method == method && route.Path == path {
			return &route, nil
		}
	}
	return nil, fmt.Errorf("route not found: %s %s", method, path)
}

// LoadRoutesConfiguration loads routes from a JSON file
func LoadRoutesConfiguration(configPath string) (*RoutesConfiguration, error) {
	// Default to embedded configuration if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultRoutesConfiguration(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read routes config: %w", err)
	}

	var config RoutesConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse routes config: %w", err)
	}

	return &config, nil
}

// SaveRoutesConfiguration saves routes to a JSON file
func SaveRoutesConfiguration(config *RoutesConfiguration, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal routes config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write routes config: %w", err)
	}

	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	return "configs/routes.json"
}
