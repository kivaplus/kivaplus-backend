package authorization

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
)

// Permission representa uma permissão necessária para acessar uma rota
type Permission struct {
	Resource string `json:"resource"` // ex: "users", "condominium", "unit"
	Action   string `json:"action"`   // ex: "create", "read", "update", "delete"
	Roles    []int  `json:"roles"`    // IDs dos papéis que podem executar esta ação
	Scope    string `json:"scope"`    // "global", "condominio", "own" (próprios dados)
}

// RoutePermission mapeia rotas para permissões necessárias
type RoutePermission struct {
	Method      string     `json:"method"`       // GET, POST, PUT, DELETE
	PathPattern string     `json:"path_pattern"` // Padrão da rota (com regex)
	Permission  Permission `json:"permission"`
}

// AuthorizationService gerencia as permissões do sistema
type AuthorizationService struct {
	routePermissions []RoutePermission
}

// RoutesConfiguration represents the centralized routes configuration
type RoutesConfiguration struct {
	Version string       `json:"version"`
	Groups  []RouteGroup `json:"groups"`
}

// RouteGroup represents a logical grouping of routes
type RouteGroup struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Routes      []RouteConfig `json:"routes"`
}

// RouteConfig represents a single route configuration
type RouteConfig struct {
	Name        string `json:"name"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	PathPattern string `json:"path_pattern"`
	Public      bool   `json:"public"`
	Roles       []int  `json:"roles,omitempty"`
	Resource    string `json:"resource,omitempty"`
	Action      string `json:"action,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Lambda      string `json:"lambda"`
	Description string `json:"description,omitempty"`
}

// GetAllRoutes returns all routes from all groups
func (rc *RoutesConfiguration) GetAllRoutes() []RouteConfig {
	var allRoutes []RouteConfig
	for _, group := range rc.Groups {
		allRoutes = append(allRoutes, group.Routes...)
	}
	return allRoutes
}

// NewAuthorizationService cria um novo serviço de autorização
func NewAuthorizationService() *AuthorizationService {
	return &AuthorizationService{
		routePermissions: getRoutePermissionsFromConfig(),
	}
}

// CheckPermission verifica se o usuário tem permissão para acessar uma rota
func (a *AuthorizationService) CheckPermission(method, path string, claims *jwt.Claims) (bool, error) {
	// Encontra a permissão necessária para a rota
	routePermission, err := a.findRoutePermission(method, path)
	if err != nil {
		return false, fmt.Errorf("route permission not found: %w", err)
	}

	// Super admin tem acesso a tudo
	if claims.IsSuperAdmin() {
		return true, nil
	}

	// Verifica se o usuário tem algum dos papéis necessários
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

	// Verifica escopo específico
	switch routePermission.Permission.Scope {
	case "global":
		// Acesso global - já verificado pelos papéis
		return true, nil
	case "condominium":
		// Precisa verificar se tem acesso ao condomínio específico
		return a.checkCondominiumAccess(path, claims)
	case "own":
		// Precisa verificar se está acessando próprios dados
		return a.checkOwnDataAccess(path, claims)
	default:
		return true, nil
	}
}

// findRoutePermission encontra a permissão para uma rota específica
func (a *AuthorizationService) findRoutePermission(method, path string) (*RoutePermission, error) {
	for _, rp := range a.routePermissions {
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

// checkCondominioAccess verifica acesso a condomínio específico
func (a *AuthorizationService) checkCondominiumAccess(path string, claims *jwt.Claims) (bool, error) {
	// Extrai o ID do condomínio da URL
	condominiumID, err := extractCondominiumIDFromPath(path)
	if err != nil {
		return false, fmt.Errorf("failed to extract condominio ID: %w", err)
	}

	return claims.CanAccessCondominium(condominiumID), nil
}

// checkOwnDataAccess verifica acesso aos próprios dados
func (a *AuthorizationService) checkOwnDataAccess(path string, claims *jwt.Claims) (bool, error) {
	// Extrai o ID do usuário da URL
	userID, err := extractUserIDFromPath(path)
	if err != nil {
		return false, fmt.Errorf("failed to extract user ID: %w", err)
	}

	return claims.UserID == userID, nil
}

// extractCondominioIDFromPath extrai o ID do condomínio da URL
func extractCondominiumIDFromPath(path string) (int64, error) {
	// Padrões possíveis: /condominios/{id}, /condominios/{id}/unidades, etc.
	re := regexp.MustCompile(`/condominios/(\d+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return 0, fmt.Errorf("condominio ID not found in path: %s", path)
	}

	id, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid condominio ID: %s", matches[1])
	}

	return id, nil
}

// extractUserIDFromPath extrai o ID do usuário da URL
func extractUserIDFromPath(path string) (string, error) {
	// Padrões possíveis: /usuarios/{id}, /profile, etc.
	if strings.Contains(path, "/profile") {
		// Para rotas de perfil, não há ID na URL, deve usar o do token
		return "", fmt.Errorf("user ID should come from token for profile routes")
	}

	re := regexp.MustCompile(`/usuarios/(\w+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", fmt.Errorf("user ID not found in path: %s", path)
	}

	return matches[1], nil
}

// getRoutePermissionsFromConfig loads route permissions from centralized configuration
func getRoutePermissionsFromConfig() []RoutePermission {
	// Try to load from centralized configuration first
	config, err := loadRoutesConfiguration()
	if err != nil {
		fmt.Printf("⚠️ Failed to load centralized routes configuration: %v\n", err)
		fmt.Println("🔄 Falling back to hardcoded permissions...")
		return getHardcodedRoutePermissions()
	}

	fmt.Printf("✅ Loaded route permissions from centralized configuration: %d routes\n", len(config.GetAllRoutes()))

	// Convert centralized configuration to authorization format
	return convertConfigToPermissions(config)
}

// loadRoutesConfiguration loads the centralized routes configuration
func loadRoutesConfiguration() (*RoutesConfiguration, error) {
	// Try multiple possible paths
	possiblePaths := []string{
		"configs/routes.json",
		"../configs/routes.json",
		"../../configs/routes.json",
	}

	for _, path := range possiblePaths {
		if config, err := loadConfigFromPath(path); err == nil {
			return config, nil
		}
	}

	return nil, fmt.Errorf("routes configuration not found in any of the expected paths")
}

// loadConfigFromPath loads configuration from a specific path
func loadConfigFromPath(path string) (*RoutesConfiguration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config RoutesConfiguration
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// convertConfigToPermissions converts centralized config to authorization permissions
func convertConfigToPermissions(config *RoutesConfiguration) []RoutePermission {
	var permissions []RoutePermission

	for _, group := range config.Groups {
		for _, route := range group.Routes {
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
	}

	return permissions
}

// getHardcodedRoutePermissions is the fallback hardcoded permissions (kept for safety)
func getHardcodedRoutePermissions() []RoutePermission {
	factory := NewRoutePermissionFactory()

	// ===== ROTAS PÚBLICAS =====
	factory.
		AddPublicRoute("POST", "^/register$", "public", "register").
		AddPublicRoute("POST", "^/login$", "public", "login")

	// ===== USUÁRIOS =====
	factory.
		AddRoute("GET", "^/usuarios$").
		WithResource("usuarios").WithAction("list").
		WithRoles(AdminAndSindico()...).GlobalScope().Build().
		AddRoute("GET", "^/usuarios/\\w+$").
		WithResource("usuarios").WithAction("read").
		WithRoles(AdminAndSindico()...).OwnScope().Build().
		AddRoute("PUT", "^/usuarios/\\w+$").
		WithResource("usuarios").WithAction("update").
		WithRoles(AdminSindicoMorador()...).OwnScope().Build()

	// ===== PERFIL =====
	factory.
		AddRoute("GET", "^/profile$").
		WithResource("profile").WithAction("read").
		WithRoles(AllRoles()...).OwnScope().Build().
		AddRoute("PUT", "^/profile$").
		WithResource("profile").WithAction("update").
		WithRoles(AllRoles()...).OwnScope().Build()

	// Basic fallback routes - add more as needed
	return factory.Build()
}
