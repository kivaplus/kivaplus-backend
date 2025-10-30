package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
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
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"Invalid request format"}`,
		}, nil
	}

	user, err := h.createUserUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: fmt.Sprintf(`{"error":"%s"}`, err.Error()),
		}, nil
	}

	response := map[string]interface{}{
		"message": "User created successfully",
		"user": map[string]interface{}{
			"id":     user.User.ID,
			"email":  user.User.Email,
			"active": user.User.Active,
		},
	}

	responseBody, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

// Login handles user authentication
func (h *AuthHandler) Login(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing login request")

	var req ports.LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Failed to parse login request", "error", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"Invalid request format"}`,
		}, nil
	}

	loginResponse, err := h.loginUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error("Failed to authenticate user", "error", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"Invalid credentials"}`,
		}, nil
	}

	response := map[string]interface{}{
		"access_token":  loginResponse.AccessToken,
		"refresh_token": loginResponse.RefreshToken,
		"user": map[string]interface{}{
			"id":     loginResponse.User.ID,
			"email":  loginResponse.User.Email,
			"active": loginResponse.User.Active,
		},
	}

	responseBody, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing refresh token request")

	var req ports.RefreshTokenRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Failed to parse refresh token request", "error", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"Invalid request format"}`,
		}, nil
	}

	refreshResponse, err := h.refreshTokenUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error("Failed to refresh token", "error", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"Invalid refresh token"}`,
		}, nil
	}

	response := map[string]interface{}{
		"access_token":  refreshResponse.AccessToken,
		"refresh_token": refreshResponse.RefreshToken,
	}

	responseBody, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}
