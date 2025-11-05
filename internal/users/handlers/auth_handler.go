package handlers

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/response"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	createUserUC   ports.CreateUserService
	loginUC        ports.LoginService
	refreshTokenUC ports.RefreshTokenService
	logger         logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	createUserUC ports.CreateUserService,
	loginUC ports.LoginService,
	refreshTokenUC ports.RefreshTokenService,
	logger logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		createUserUC:   createUserUC,
		loginUC:        loginUC,
		refreshTokenUC: refreshTokenUC,
		logger:         logger,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing register request")

	var req ports.CreateUserRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Failed to parse register request", "error", err)
		return response.BadRequest("Invalid request format"), nil
	}

	user, err := h.createUserUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err)
		return response.BadRequest(err.Error()), nil
	}

	// Return the user data directly with a single message
	return response.Created(user, "User created successfully"), nil
}

// Login handles user authentication
func (h *AuthHandler) Login(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing login request")

	var req ports.LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Failed to parse login request", "error", err)
		return response.BadRequest("Invalid request format"), nil
	}

	loginResponse, err := h.loginUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error("Failed to authenticate user", "error", err)
		return response.InvalidCredentials(), nil
	}

	user := map[string]interface{}{
		"id":     loginResponse.User.ID,
		"email":  loginResponse.User.Email,
		"active": loginResponse.User.Active,
	}

	tokens := map[string]string{
		"access_token":  loginResponse.AccessToken,
		"refresh_token": loginResponse.RefreshToken,
	}

	return response.LoginSuccess(user, tokens), nil
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing refresh token request")

	var req ports.RefreshTokenRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Failed to parse refresh token request", "error", err)
		return response.BadRequest("Invalid request format"), nil
	}

	refreshResponse, err := h.refreshTokenUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error("Failed to refresh token", "error", err)
		return response.InvalidRefreshToken(), nil
	}

	tokens := map[string]string{
		"access_token":  refreshResponse.AccessToken,
		"refresh_token": refreshResponse.RefreshToken,
	}

	return response.OK(tokens, "Token refreshed successfully"), nil
}

// Health handles health check requests
func (h *AuthHandler) Health(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing health check request")

	healthData := map[string]interface{}{
		"status":    "healthy",
		"service":   "kivaplus-backend",
		"timestamp": "2025-11-04T15:40:00Z",
		"version":   "1.0.0",
	}

	return response.OK(healthData, "Service is healthy"), nil
}
