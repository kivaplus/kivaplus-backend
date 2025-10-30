package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/handlers"
	"github.com/kivaplus/kivaplus-backend/internal/users/repository"
	"github.com/kivaplus/kivaplus-backend/internal/users/usecases"
)

var (
	userHandler *handlers.HTTPHandler
	log         logger.Logger
)

func init() {
	log = logger.New()

	// Initialize database connection
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		panic(err)
	}

	// Initialize repositories
	userRepo := repository.NewUserPostgresRepository(db)
	personRepo := repository.NewPersonPostgresRepository(db)
	roleRepo := repository.NewRolePostgresRepository(db)
	addressRepo := repository.NewAddressPostgresRepository(db)

	// Initialize use cases
	createUserUC := usecases.NewCreateUser(userRepo, personRepo, roleRepo, log)
	loginUC := usecases.NewLogin(userRepo, personRepo, roleRepo, log)
	refreshTokenUC := usecases.NewRefreshToken(userRepo, personRepo, roleRepo, log)
	completeProfileUC := usecases.NewCompleteProfile(userRepo, personRepo, addressRepo, log)

	// Initialize handler
	userHandler = handlers.NewHTTPHandler(createUserUC, loginUC, refreshTokenUC, completeProfileUC, log)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Info("Processing request", "path", request.Path, "method", request.HTTPMethod)

	// Rotas públicas (não precisam de autenticação)
	switch request.Path {
	case "/register":
		if request.HTTPMethod == "POST" {
			return userHandler.Register(ctx, request)
		}
	case "/login":
		if request.HTTPMethod == "POST" {
			return userHandler.Login(ctx, request)
		}
	}

	// Rotas privadas (precisam de autenticação - já validadas pelo Authorizer)
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

	// Verificar se precisa completar perfil
	if requiresProfileCompletion(request) {
		return profileCompletionRequiredResponse(), nil
	}

	errorBody, _ := json.Marshal(map[string]string{"error": "Route not found"})
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Headers: map[string]string{
			"Content-Type": "application/json",
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
