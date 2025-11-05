package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/handlers"
	"github.com/kivaplus/kivaplus-backend/internal/users/repository"
	"github.com/kivaplus/kivaplus-backend/internal/users/usecases"
)

var (
	authHandler *handlers.AuthHandler
)

func init() {
	// Initialize logger
	logger := logger.New()

	// Initialize database connection
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserPostgresRepository(db)
	personRepo := repository.NewPersonPostgresRepository(db)
	roleRepo := repository.NewRolePostgresRepository(db)

	// Initialize use cases
	createUserUC := usecases.NewCreateUser(userRepo, personRepo, roleRepo, logger)
	loginUC := usecases.NewLogin(userRepo, personRepo, roleRepo, logger)
	refreshTokenUC := usecases.NewRefreshToken(userRepo, personRepo, roleRepo, logger)

	// Initialize auth handler
	authHandler = handlers.NewAuthHandler(createUserUC, loginUC, refreshTokenUC, logger)

	log.Println("✅ Auth lambda initialized successfully")
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Route based on path and method
	switch {
	case request.HTTPMethod == "POST" && request.Path == "/register":
		return authHandler.Register(ctx, request)
	case request.HTTPMethod == "POST" && request.Path == "/login":
		return authHandler.Login(ctx, request)
	case request.HTTPMethod == "POST" && request.Path == "/token/refresh":
		return authHandler.RefreshToken(ctx, request)
	case request.HTTPMethod == "GET" && request.Path == "/health":
		return authHandler.Health(ctx, request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"Route not found"}`,
		}, nil
	}
}
