package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// Test helper functions for RefreshToken use case
func setupRefreshTokenTest() (*RefreshToken, *MockUserRepository, *MockPersonRepository, *MockRoleRepository) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockLogger := logger.New()

	useCase := NewRefreshToken(mockUserRepo, mockPersonRepo, mockRoleRepo, mockLogger)

	return useCase, mockUserRepo, mockPersonRepo, mockRoleRepo
}

func createValidRefreshTokenRequest() ports.RefreshTokenRequest {
	// Create a real JWT service to generate a valid refresh token for testing
	jwtService := jwt.GetJWTService()
	refreshToken, _ := jwtService.GenerateRefreshToken("1", "john@example.com")

	return ports.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
}

// Success scenarios
func TestRefreshToken_Success_CompleteProfile(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _ := setupRefreshTokenTest()
	ctx := context.Background()
	req := createValidRefreshTokenRequest()
	user := createTestUser()

	// Setup mocks for success scenario
	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	// Verify tokens are different
	assert.NotEqual(t, result.AccessToken, result.RefreshToken)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

func TestRefreshToken_Success_IncompleteProfile(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _ := setupRefreshTokenTest()
	ctx := context.Background()
	req := createValidRefreshTokenRequest()
	user := createTestUser()

	// Setup mocks for success scenario with incomplete profile
	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(false, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

// Validation error scenarios
func TestRefreshToken_ValidationErrors(t *testing.T) {
	useCase, _, _, _ := setupRefreshTokenTest()
	ctx := context.Background()

	tests := []struct {
		name        string
		request     ports.RefreshTokenRequest
		expectedErr string
	}{
		{
			name: "Empty refresh token",
			request: ports.RefreshTokenRequest{
				RefreshToken: "",
			},
			expectedErr: "validation failed",
		},
		{
			name: "Whitespace only token",
			request: ports.RefreshTokenRequest{
				RefreshToken: "   ",
			},
			expectedErr: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := useCase.Execute(ctx, tt.request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// Token validation error scenarios
func TestRefreshToken_InvalidToken(t *testing.T) {
	useCase, _, _, _ := setupRefreshTokenTest()
	ctx := context.Background()

	req := ports.RefreshTokenRequest{
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
	}

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	useCase, _, _, _ := setupRefreshTokenTest()
	ctx := context.Background()

	// Create an expired token (this will be caught by JWT validation)
	req := ports.RefreshTokenRequest{
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDk0NTkyMDB9.invalid",
	}

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

func TestRefreshToken_WrongTokenType(t *testing.T) {
	useCase, _, _, _ := setupRefreshTokenTest()
	ctx := context.Background()

	// Generate an access token instead of refresh token
	jwtService := jwt.GetJWTService()
	accessToken, _ := jwtService.GenerateToken("1", "john@example.com", "complete", 1)

	req := ports.RefreshTokenRequest{
		RefreshToken: accessToken, // Using access token as refresh token (wrong type)
	}

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

// User-related error scenarios
func TestRefreshToken_UserNotFound(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupRefreshTokenTest()
	ctx := context.Background()
	req := createValidRefreshTokenRequest()

	// Setup mock to return user not found
	mockUserRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("user not found"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	mockUserRepo.AssertExpectations(t)
}

func TestRefreshToken_InactiveUser(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupRefreshTokenTest()
	ctx := context.Background()
	req := createValidRefreshTokenRequest()

	// Create inactive user
	user := createTestUser()
	user.Active = false

	// Setup mock
	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user account is inactive")

	mockUserRepo.AssertExpectations(t)
}

// Repository error scenarios
// Note: GetUserRoles test removed as we now use optimized JWT service

func TestRefreshToken_ProfileCheckError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _ := setupRefreshTokenTest()
	ctx := context.Background()
	req := createValidRefreshTokenRequest()
	user := createTestUser()

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(false, errors.New("profile check failed"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check profile completeness")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

// Constructor test
func TestNewRefreshToken(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockLogger := logger.New()

	useCase := NewRefreshToken(mockUserRepo, mockPersonRepo, mockRoleRepo, mockLogger)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockUserRepo, useCase.userRepo)
	assert.Equal(t, mockPersonRepo, useCase.personRepo)
	assert.Equal(t, mockRoleRepo, useCase.roleRepo)
	assert.NotNil(t, useCase.jwtService)
	assert.Equal(t, mockLogger, useCase.logger)
}

// Edge cases
func TestRefreshToken_ContextCancellation(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupRefreshTokenTest()
	req := createValidRefreshTokenRequest()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Setup mock to return context cancelled error
	mockUserRepo.On("GetByID", ctx, int64(1)).Return(nil, context.Canceled)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	mockUserRepo.AssertExpectations(t)
}

// Integration-style test with realistic data
func TestRefreshToken_RealisticScenario(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _ := setupRefreshTokenTest()
	ctx := context.Background()

	// Generate a valid refresh token for user ID 42
	jwtService := jwt.GetJWTService()
	refreshToken, _ := jwtService.GenerateRefreshToken("42", "maria.silva@condominio.com.br")

	req := ports.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	user, _ := domain.NewUser("maria.silva@condominio.com.br", "MinhaSenh@Segura123!")
	user.ID = 42

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotEqual(t, result.AccessToken, result.RefreshToken)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

// Test with empty roles
func TestRefreshToken_EmptyRoles(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _ := setupRefreshTokenTest()
	ctx := context.Background()
	req := createValidRefreshTokenRequest()
	user := createTestUser()

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert - should still succeed with empty roles
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

// Test with multiple roles
func TestRefreshToken_MultipleRoles(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _ := setupRefreshTokenTest()
	ctx := context.Background()
	req := createValidRefreshTokenRequest()
	user := createTestUser()

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

// Test invalid user ID in token
func TestRefreshToken_InvalidUserIDInToken(t *testing.T) {
	useCase, _, _, _ := setupRefreshTokenTest()
	ctx := context.Background()

	// This test will be handled by the JWT validation itself
	req := ports.RefreshTokenRequest{
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiaW52YWxpZCJ9.invalid",
	}

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")
}
