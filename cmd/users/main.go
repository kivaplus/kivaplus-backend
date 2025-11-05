package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kivaplus/kivaplus-backend/internal/shared/cache"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/permissions"
	"github.com/kivaplus/kivaplus-backend/internal/users/handlers"
	"github.com/kivaplus/kivaplus-backend/internal/users/repository"
	"github.com/kivaplus/kivaplus-backend/internal/users/usecases"
)

var (
	userHandler       *handlers.HTTPHandler
	authHandler       *handlers.AuthHandler
	permissionChecker *permissions.EnhancedChecker
	log               logger.Logger
)

func init() {
	log = logger.New()

	// Initialize database connection
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		panic(err)
	}

	// Initialize Redis cache
	redisService := cache.NewRedisService()
	if err := redisService.Ping(); err != nil {
		log.Warn("Redis connection failed, continuing without cache", "error", err)
		redisService = nil
	}

	// Initialize repositories
	userRepo := repository.NewUserPostgresRepository(db)
	personRepo := repository.NewPersonPostgresRepository(db)
	addressRepo := repository.NewAddressPostgresRepository(db)
	personDocRepo := repository.NewPersonDocumentPostgresRepository(db)
	roleRepo := repository.NewRolePostgresRepository(db)
	enhancedRoleRepo := repository.NewEnhancedRoleRepository(db, log)

	// Initialize JWT service (singleton)
	jwtService := jwt.GetJWTService()

	// Initialize permission service
	var permissionService *permissions.Service
	if redisService != nil {
		permissionService = permissions.NewPermissionService(redisService, enhancedRoleRepo, log)
		permissionChecker = permissions.NewEnhancedChecker(jwtService, permissionService, redisService, log)
		log.Info("Enhanced permission system initialized with Redis cache")
	} else {
		permissionService = permissions.NewPermissionService(nil, enhancedRoleRepo, log)
		permissionChecker = permissions.NewEnhancedChecker(jwtService, permissionService, nil, log)
		log.Info("Enhanced permission system initialized without cache")
	}

	// Initialize use cases
	createUserUC := usecases.NewCreateUser(userRepo, personRepo, roleRepo, log)
	loginUC := usecases.NewLogin(userRepo, personRepo, roleRepo, log)
	refreshTokenUC := usecases.NewRefreshToken(userRepo, personRepo, roleRepo, log)
	getProfileUC := usecases.NewGetProfile(userRepo, personRepo, addressRepo, personDocRepo, log)
	completeProfileUC := usecases.NewCompleteProfile(userRepo, personRepo, addressRepo, personDocRepo, roleRepo, jwtService, log)

	// Initialize handlers with enhanced permissions
	authHandler = handlers.NewAuthHandler(createUserUC, loginUC, refreshTokenUC, log)
	userHandler = handlers.NewHTTPHandler(
		getProfileUC,
		completeProfileUC,
		permissionChecker,
		permissionService,
		redisService,
		log,
	)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Info("Processing request", "path", request.Path, "method", request.HTTPMethod)

	// Handle CORS preflight requests
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Origin, Content-Type, Authorization",
			},
		}, nil
	}

	// Public routes (no authentication required)
	switch request.Path {
	case "/register":
		if request.HTTPMethod == "POST" {
			return authHandler.Register(ctx, request)
		}
	case "/login":
		if request.HTTPMethod == "POST" {
			return authHandler.Login(ctx, request)
		}
	case "/token/refresh":
		if request.HTTPMethod == "POST" {
			return authHandler.RefreshToken(ctx, request)
		}
	}

	// Protected routes (authentication required - validated by Authorizer)
	switch request.Path {
	case "/profile":
		if request.HTTPMethod == "GET" {
			return userHandler.GetProfile(ctx, request)
		}
		if request.HTTPMethod == "PUT" {
			return userHandler.UpdateProfile(ctx, request)
		}
	case "/usuarios":
		if request.HTTPMethod == "GET" {
			return userHandler.ListUsers(ctx, request)
		}
	}

	// Check if profile completion is required
	if requiresProfileCompletion(request) {
		return profileCompletionRequiredResponse(), nil
	}

	// Route not found
	errorBody, _ := json.Marshal(map[string]string{
		"error":   "not_found",
		"message": "Route not found",
	})
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Origin, Content-Type, Authorization",
		},
		Body: string(errorBody),
	}, nil
}

// requiresProfileCompletion verifica se o usuário precisa completar o perfil
func requiresProfileCompletion(request events.APIGatewayProxyRequest) bool {
	// Verifica se o contexto do authorizer indica que precisa completar perfil
	if authContext, exists := request.RequestContext.Authorizer["requiresProfileCompletion"]; exists {
		if requires, ok := authContext.(bool); ok && requires {
			return true
		}
	}
	return false
}

// profileCompletionRequiredResponse retorna resposta indicando que precisa completar perfil
func profileCompletionRequiredResponse() events.APIGatewayProxyResponse {
	response := map[string]interface{}{
		"error":   "profile_incomplete",
		"message": "Please complete your profile before accessing this resource",
		"action":  "redirect_to_profile_completion",
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusPreconditionRequired, // 428
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

func main() {
	lambda.Start(handler)
}
