package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kivaplus/kivaplus-backend/internal/shared/authorization"
	"github.com/kivaplus/kivaplus-backend/internal/shared/cache"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// AuthorizerResponse represents the Lambda authorizer response
type AuthorizerResponse struct {
	PrincipalID    string                 `json:"principalId"`
	PolicyDocument PolicyDocument         `json:"policyDocument"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

type PolicyDocument struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

type Statement struct {
	Action   string `json:"Action"`
	Effect   string `json:"Effect"`
	Resource string `json:"Resource"`
}

var (
	jwtService       *jwt.Service
	authzService     *authorization.AuthorizationService
	permissionsCache *cache.PermissionsCache
	log              logger.Logger
)

func init() {
	log = logger.New()
	jwtService = jwt.NewJWTService()

	// Use the new centralized authorization service
	authzService = authorization.NewAuthorizationService()
	log.Info("Authorization service initialized with centralized routes configuration")

	// Initialize database connection for permissions cache
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Error("Failed to connect to database for permissions cache", "error", err)
		// Continue without cache - will use centralized configuration
	} else {
		permissionsCache = cache.NewPermissionsCache(db, log)
	}
}

func handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (AuthorizerResponse, error) {
	log.Info("Processing authorization request",
		"methodArn", event.MethodArn,
		"authorizationToken", maskToken(event.AuthorizationToken))

	// Extract method and path from methodArn
	method, path, err := parseMethodArn(event.MethodArn)
	if err != nil {
		log.Error("Failed to parse method ARN", "error", err, "methodArn", event.MethodArn)
		return AuthorizerResponse{}, fmt.Errorf("invalid method ARN: %v", err)
	}

	log.Info("Parsed request", "method", method, "path", path)

	// Check if it's a public route (login, register)
	if isPublicRoute(method, path) {
		log.Info("Public route accessed", "method", method, "path", path)
		return generatePolicy("public", "Allow", event.MethodArn, nil), nil
	}

	// Extract and validate token
	token := event.AuthorizationToken
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if token == "" {
		log.Warn("Missing authorization token")
		return AuthorizerResponse{}, fmt.Errorf("missing authorization token")
	}

	// Validate JWT token
	claims, err := jwtService.ValidateAccessToken(token)
	if err != nil {
		log.Error("Token validation failed", "error", err)
		return AuthorizerResponse{}, fmt.Errorf("unauthorized: %v", err)
	}

	log.Info("Token validated successfully",
		"userID", claims.UserID,
		"email", claims.Email,
		"rolesCount", len(claims.Roles),
		"profileStatus", claims.ProfileStatus)

	// Check if profile is incomplete and redirect to complete profile
	if claims.ProfileStatus == "incomplete" && !isProfileCompletionRoute(method, path) {
		log.Info("User has incomplete profile, should complete profile first", "userID", claims.UserID)

		// Allow access but add context indicating profile completion needed
		policy := generatePolicy(claims.UserID, "Allow", event.MethodArn, claims)
		policy.Context["requiresProfileCompletion"] = true
		return policy, nil
	}

	// Check permissions for the route using cached permissions
	hasPermission, err := checkPermissionWithCache(ctx, method, path, claims)
	if err != nil {
		log.Error("Permission check failed", "error", err, "userID", claims.UserID, "method", method, "path", path)
		return AuthorizerResponse{}, fmt.Errorf("permission denied: %v", err)
	}

	if !hasPermission {
		log.Warn("Access denied", "userID", claims.UserID, "method", method, "path", path)
		return AuthorizerResponse{}, fmt.Errorf("access denied")
	}

	log.Info("Access granted", "userID", claims.UserID, "method", method, "path", path)

	// Generate policy with user context
	policy := generatePolicy(claims.UserID, "Allow", event.MethodArn, claims)
	return policy, nil
}

// parseMethodArn extrai o método HTTP e o path do methodArn
func parseMethodArn(methodArn string) (string, string, error) {
	// Format: arn:aws:execute-api:region:account:api-id/stage/METHOD/resource-path
	parts := strings.Split(methodArn, "/")
	if len(parts) < 4 {
		return "", "", fmt.Errorf("invalid methodArn format: %s", methodArn)
	}

	method := parts[2]
	path := "/" + strings.Join(parts[3:], "/")

	return method, path, nil
}

// isPublicRoute verifica se a rota é pública (não precisa de autenticação)
func isPublicRoute(method, path string) bool {
	publicRoutes := map[string][]string{
		"POST": {"/register", "/login"},
		"GET":  {"/health", "/version"},
	}

	if routes, exists := publicRoutes[method]; exists {
		for _, route := range routes {
			if path == route {
				return true
			}
		}
	}

	return false
}

// generatePolicy gera a política de acesso
func generatePolicy(principalID, effect, resource string, claims *jwt.Claims) AuthorizerResponse {
	context := map[string]interface{}{
		"userID": principalID,
	}

	// Adiciona informações do usuário ao contexto se disponível
	if claims != nil {
		context["email"] = claims.Email
		context["tokenType"] = claims.TokenType

		// Serializa os roles para o contexto
		if rolesJSON, err := json.Marshal(claims.Roles); err == nil {
			context["roles"] = string(rolesJSON)
		}

		// Adiciona flags de conveniência
		context["isSuperAdmin"] = claims.IsSuperAdmin()

		// Adiciona lista de condomínios onde é síndico
		sindicos := claims.GetCondominiumsWhereSindico()
		if sindicosJSON, err := json.Marshal(sindicos); err == nil {
			context["sindicoCondominios"] = string(sindicosJSON)
		}

		// Adiciona lista de condomínios onde é morador
		moradores := claims.GetCondominiumsWhereMorador()
		if moradoresJSON, err := json.Marshal(moradores); err == nil {
			context["moradorCondominios"] = string(moradoresJSON)
		}
	}

	return AuthorizerResponse{
		PrincipalID: principalID,
		PolicyDocument: PolicyDocument{
			Version: "2012-10-17",
			Statement: []Statement{
				{
					Action:   "execute-api:Invoke",
					Effect:   effect,
					Resource: resource,
				},
			},
		},
		Context: context,
	}
}

// isProfileCompletionRoute verifica se a rota é para completar perfil
func isProfileCompletionRoute(method, path string) bool {
	profileRoutes := map[string][]string{
		"GET": {"/profile"},
		"PUT": {"/profile"},
	}

	if routes, exists := profileRoutes[method]; exists {
		for _, route := range routes {
			if path == route {
				return true
			}
		}
	}

	return false
}

// checkPermissionWithCache verifica permissões usando o cache
func checkPermissionWithCache(ctx context.Context, method, path string, claims *jwt.Claims) (bool, error) {
	// Se não temos cache, usa o sistema padrão
	if permissionsCache == nil {
		return authzService.CheckPermission(method, path, claims)
	}

	// Super admin sempre tem acesso
	if claims.IsSuperAdmin() {
		return true, nil
	}

	// Extrai resource e action do path e method
	resource, action := extractResourceAndAction(method, path)
	if resource == "" {
		// Se não conseguir extrair, usa o sistema padrão
		return authzService.CheckPermission(method, path, claims)
	}

	// Verifica se algum dos roles do usuário tem a permissão necessária
	for _, role := range claims.Roles {
		hasPermission, err := permissionsCache.HasPermission(ctx, role.ID, resource, action)
		if err != nil {
			log.Error("Failed to check permission in cache", "error", err, "roleID", role.ID, "resource", resource, "action", action)
			continue
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// extractResourceAndAction extrai o recurso e ação do método e path
func extractResourceAndAction(method, path string) (string, string) {
	// Mapeia métodos HTTP para ações
	methodToAction := map[string]string{
		"GET":    "read",
		"POST":   "create",
		"PUT":    "update",
		"DELETE": "delete",
	}

	action, exists := methodToAction[method]
	if !exists {
		return "", ""
	}

	// Extrai o recurso do path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) == 0 {
		return "", ""
	}

	// Mapeia paths para recursos
	pathToResource := map[string]string{
		"usuarios":     "users",
		"condominios":  "condominio",
		"unidades":     "unidade",
		"moradores":    "morador",
		"funcionarios": "funcionario",
		"veiculos":     "veiculo",
		"contratos":    "contrato",
		"profile":      "users", // Profile é considerado como usuário
	}

	resource := pathParts[0]
	if mappedResource, exists := pathToResource[resource]; exists {
		return mappedResource, action
	}

	// Se não encontrar mapeamento, usa o próprio path como recurso
	return resource, action
}

// maskToken mascara o token para logs (mostra apenas os primeiros e últimos caracteres)
func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:5] + "..." + token[len(token)-5:]
}

func main() {
	lambda.Start(handler)
}
