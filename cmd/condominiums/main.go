package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/handlers"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/repository"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/usecases"
	"github.com/kivaplus/kivaplus-backend/internal/shared/cache"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/permissions"
	userRepository "github.com/kivaplus/kivaplus-backend/internal/users/repository"
)

var handler *handlers.HTTPHandler

func init() {
	// Initialize logger
	appLogger := logger.New()

	// Initialize database connection
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Redis cache
	redisService := cache.NewRedisService()
	if err := redisService.Ping(); err != nil {
		appLogger.Warn("Redis connection failed, continuing without cache", "error", err)
		// Continue without Redis for development/fallback
	}

	// Initialize JWT service (singleton)
	jwtService := jwt.GetJWTService()

	// Initialize repositories
	userRepo := userRepository.NewUserPostgresRepository(db)
	personRepo := userRepository.NewPersonPostgresRepository(db)
	addressRepo := userRepository.NewAddressPostgresRepository(db)
	roleRepo := userRepository.NewRolePostgresRepository(db)
	enhancedRoleRepo := userRepository.NewEnhancedRoleRepository(db, appLogger)
	condominiumRepo := repository.NewCondominiumPostgresRepository(db)

	// Initialize permission service
	permissionService := permissions.NewPermissionService(redisService, enhancedRoleRepo, appLogger)

	// Initialize enhanced permission checker
	permissionChecker := permissions.NewEnhancedChecker(jwtService, permissionService, redisService, appLogger)

	// Initialize use cases
	createCondominiumUseCase := usecases.NewCreateCondominium(condominiumRepo, personRepo, addressRepo, roleRepo, userRepo, jwtService, appLogger)
	listCondominiumsUseCase := usecases.NewListCondominiums(condominiumRepo, personRepo, addressRepo, roleRepo, appLogger)
	getCondominiumUseCase := usecases.NewGetCondominium(condominiumRepo, personRepo, addressRepo, roleRepo, appLogger)

	// Initialize handler with enhanced permissions
	handler = handlers.NewHTTPHandler(
		createCondominiumUseCase,
		listCondominiumsUseCase,
		getCondominiumUseCase,
		permissionChecker,
		permissionService,
		redisService,
		appLogger,
	)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	// Route requests based on HTTP method and path
	switch request.HTTPMethod {
	case "POST":
		if request.Resource == "/condominios" {
			return handler.CreateCondominium(ctx, request)
		}
	case "GET":
		if request.Resource == "/condominios" {
			return handler.ListCondominiums(ctx, request)
		}
		if request.Resource == "/condominios/{id}" {
			return handler.GetCondominium(ctx, request)
		}
	}

	// Route not found
	return events.APIGatewayProxyResponse{
		StatusCode: 404,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Origin, Content-Type, Authorization",
		},
		Body: `{"error":"not_found","message":"Route not found"}`,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
