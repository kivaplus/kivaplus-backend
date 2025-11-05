package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// Mock services
type MockCreateUserService struct {
	mock.Mock
}

func (m *MockCreateUserService) Execute(ctx context.Context, req ports.CreateUserRequest) (*ports.CreateUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.CreateUserResponse), args.Error(1)
}

type MockLoginService struct {
	mock.Mock
}

func (m *MockLoginService) Execute(ctx context.Context, req ports.LoginRequest) (*ports.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.LoginResponse), args.Error(1)
}

type MockRefreshTokenService struct {
	mock.Mock
}

func (m *MockRefreshTokenService) Execute(ctx context.Context, req ports.RefreshTokenRequest) (*ports.RefreshTokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.RefreshTokenResponse), args.Error(1)
}

type MockCompleteProfileService struct {
	mock.Mock
}

func (m *MockCompleteProfileService) Execute(ctx context.Context, userID int64, req ports.UpdateProfileRequest) (*ports.CompleteProfileResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.CompleteProfileResponse), args.Error(1)
}

// TestDependencies holds all the mocked dependencies for testing
type TestDependencies struct {
	CreateUserUC   *MockCreateUserService
	LoginUC        *MockLoginService
	RefreshTokenUC *MockRefreshTokenService
	Handler        *AuthHandler
}

// setupTestHandler creates a new HTTPHandler with all mocked dependencies
func setupTestHandler() *TestDependencies {
	mockCreateUserUC := new(MockCreateUserService)
	mockLoginUC := new(MockLoginService)
	mockRefreshTokenUC := new(MockRefreshTokenService)
	mockLogger := logger.New()

	authHandler := NewAuthHandler(mockCreateUserUC, mockLoginUC, mockRefreshTokenUC, mockLogger)

	return &TestDependencies{
		CreateUserUC:   mockCreateUserUC,
		LoginUC:        mockLoginUC,
		RefreshTokenUC: mockRefreshTokenUC,
		Handler:        authHandler,
	}
}

// Helper functions for creating test data
func createTestUser() *domain.User {
	return &domain.User{
		ID:     1,
		Email:  "test@example.com",
		Active: true,
	}
}

func createTestPerson(userID int64) *domain.Person {
	return &domain.Person{
		ID:     1,
		Name:   "Test User",
		UserID: &userID,
	}
}

func createAPIGatewayRequest(body string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Body: body,
	}
}

func TestHTTPHandler_Register_Success(t *testing.T) {
	// Arrange
	deps := setupTestHandler()
	user := createTestUser()
	person := createTestPerson(user.ID)

	expectedResponse := &ports.CreateUserResponse{
		User:   user,
		Person: person,
	}

	deps.CreateUserUC.On("Execute", mock.Anything, mock.MatchedBy(func(req ports.CreateUserRequest) bool {
		return req.Email == "test@example.com" && req.Name == "Test User"
	})).Return(expectedResponse, nil)

	request := createAPIGatewayRequest(`{
		"name": "Test User",
		"email": "test@example.com",
		"password": "password123",
		"confirm_password": "password123"
	}`)

	// Act
	response, err := deps.Handler.Register(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 201, response.StatusCode)
	assert.Contains(t, response.Body, "User created successfully")

	deps.CreateUserUC.AssertExpectations(t)
}

func TestHTTPHandler_Register_ValidationError(t *testing.T) {
	// Arrange
	deps := setupTestHandler()

	deps.CreateUserUC.On("Execute", mock.Anything, mock.Anything).Return(nil,
		fmt.Errorf("validation failed: email is required"))

	request := createAPIGatewayRequest(`{
		"name": "Test User",
		"email": "",
		"password": "password123",
		"confirm_password": "password123"
	}`)

	// Act
	response, err := deps.Handler.Register(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 400, response.StatusCode)
	assert.Contains(t, response.Body, "validation failed")

	deps.CreateUserUC.AssertExpectations(t)
}

func TestHTTPHandler_Login_Success(t *testing.T) {
	// Arrange
	deps := setupTestHandler()
	user := createTestUser()

	expectedResponse := &ports.LoginResponse{
		AccessToken:  "jwt-access-token-here",
		RefreshToken: "jwt-refresh-token-here",
		User:         user,
	}

	deps.LoginUC.On("Execute", mock.Anything, mock.MatchedBy(func(req ports.LoginRequest) bool {
		return req.Email == "test@example.com" && req.Password == "password123"
	})).Return(expectedResponse, nil)

	request := createAPIGatewayRequest(`{
		"email": "test@example.com",
		"password": "password123"
	}`)

	// Act
	response, err := deps.Handler.Login(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)

	var responseBody map[string]any
	err = json.Unmarshal([]byte(response.Body), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "jwt-access-token-here", responseBody["access_token"])
	assert.Equal(t, "jwt-refresh-token-here", responseBody["refresh_token"])

	deps.LoginUC.AssertExpectations(t)
}

func TestHTTPHandler_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	deps := setupTestHandler()

	deps.LoginUC.On("Execute", mock.Anything, mock.Anything).Return(nil,
		fmt.Errorf("invalid credentials"))

	request := createAPIGatewayRequest(`{
		"email": "test@example.com",
		"password": "wrongpassword"
	}`)

	// Act
	response, err := deps.Handler.Login(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 401, response.StatusCode)
	assert.Contains(t, response.Body, "Invalid credentials")

	deps.LoginUC.AssertExpectations(t)
}

func TestHTTPHandler_RefreshToken_Success(t *testing.T) {
	// Arrange
	deps := setupTestHandler()

	expectedResponse := &ports.RefreshTokenResponse{
		AccessToken:  "new-jwt-access-token",
		RefreshToken: "new-jwt-refresh-token",
	}

	deps.RefreshTokenUC.On("Execute", mock.Anything, mock.MatchedBy(func(req ports.RefreshTokenRequest) bool {
		return req.RefreshToken == "valid-refresh-token"
	})).Return(expectedResponse, nil)

	request := createAPIGatewayRequest(`{
		"refresh_token": "valid-refresh-token"
	}`)

	// Act
	response, err := deps.Handler.RefreshToken(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)

	var responseBody map[string]any
	err = json.Unmarshal([]byte(response.Body), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "new-jwt-access-token", responseBody["access_token"])
	assert.Equal(t, "new-jwt-refresh-token", responseBody["refresh_token"])

	deps.RefreshTokenUC.AssertExpectations(t)
}

func TestHTTPHandler_RefreshToken_InvalidToken(t *testing.T) {
	// Arrange
	deps := setupTestHandler()

	deps.RefreshTokenUC.On("Execute", mock.Anything, mock.Anything).Return(nil,
		fmt.Errorf("invalid refresh token"))

	request := createAPIGatewayRequest(`{
		"refresh_token": "invalid-refresh-token"
	}`)

	// Act
	response, err := deps.Handler.RefreshToken(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 401, response.StatusCode)
	assert.Contains(t, response.Body, "Invalid refresh token")

	deps.RefreshTokenUC.AssertExpectations(t)
}

func TestHTTPHandler_Register_InvalidJSON(t *testing.T) {
	// Arrange
	deps := setupTestHandler()

	request := createAPIGatewayRequest(`{"invalid": json}`) // Invalid JSON

	// Act
	response, err := deps.Handler.Register(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 400, response.StatusCode)
	assert.Contains(t, response.Body, "Invalid request format")
}
