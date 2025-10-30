package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// HTTPHandler handles HTTP requests for user operations
type HTTPHandler struct {
	createUserUC      ports.CreateUserService
	loginUC           ports.LoginService
	refreshTokenUC    ports.RefreshTokenService
	completeProfileUC ports.CompleteProfileService
	logger            logger.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(createUserUC ports.CreateUserService, loginUC ports.LoginService, refreshTokenUC ports.RefreshTokenService, completeProfileUC ports.CompleteProfileService, logger logger.Logger) *HTTPHandler {
	return &HTTPHandler{
		createUserUC:      createUserUC,
		loginUC:           loginUC,
		refreshTokenUC:    refreshTokenUC,
		completeProfileUC: completeProfileUC,
		logger:            logger,
	}
}

// Register handles user registration
func (h *HTTPHandler) Register(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing user registration request")

	var req ports.CreateUserRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Failed to unmarshal registration request", "error", err)
		return h.errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	user, err := h.createUserUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err)

		// Check if it's a validation error or user already exists
		if contains(err.Error(), "validation failed") || contains(err.Error(), "already exists") {
			return h.errorResponse(http.StatusBadRequest, err.Error()), nil
		}

		return h.errorResponse(http.StatusInternalServerError, "Internal server error"), nil
	}

	response := map[string]interface{}{
		"message": "User created successfully",
		"user":    user,
	}

	return h.successResponse(http.StatusCreated, response), nil
}

// Login handles user authentication
func (h *HTTPHandler) Login(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing user login request")

	var req ports.LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Failed to unmarshal login request", "error", err)
		return h.errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	loginResponse, err := h.loginUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error("Failed to authenticate user", "error", err)

		// Check if it's invalid credentials
		if contains(err.Error(), "invalid credentials") || contains(err.Error(), "inactive") {
			return h.errorResponse(http.StatusUnauthorized, "Invalid credentials"), nil
		}

		return h.errorResponse(http.StatusInternalServerError, "Internal server error"), nil
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

	return h.successResponse(http.StatusOK, response), nil
}

// GetProfile handles getting user profile (placeholder)
func (h *HTTPHandler) GetProfile(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing get profile request")

	// TODO: Extract user ID from JWT token in Authorization header
	// TODO: Implement get profile use case

	return h.errorResponse(http.StatusNotImplemented, "Not implemented yet"), nil
}

// UpdateProfile handles updating user profile (placeholder)
func (h *HTTPHandler) UpdateProfile(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing update profile request")

	// TODO: Extract user ID from JWT token in Authorization header
	// TODO: Implement update profile use case

	return h.errorResponse(http.StatusNotImplemented, "Not implemented yet"), nil
}

// Helper methods

func (h *HTTPHandler) successResponse(statusCode int, data interface{}) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(data)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
		},
		Body: string(body),
	}
}

func (h *HTTPHandler) errorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	errorBody := map[string]string{"error": message}
	body, _ := json.Marshal(errorBody)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
		},
		Body: string(body),
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ListUsers handles listing users (admin/sindico only)
func (h *HTTPHandler) ListUsers(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing list users request")

	// TODO: Implement list users use case
	// TODO: Apply filters based on user role (sindico can only see users from their condominios)

	response := map[string]interface{}{
		"message": "List users endpoint - implementation pending",
		"users":   []interface{}{},
	}

	return h.successResponse(http.StatusOK, response), nil
}

// RefreshToken handles token refresh requests
func (h *HTTPHandler) RefreshToken(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
