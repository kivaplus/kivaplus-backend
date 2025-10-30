package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// Test helper functions for Login use case
func setupLoginTest() (*Login, *MockUserRepository, *MockPersonRepository, *MockRoleRepository) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockLogger := logger.New()

	useCase := NewLogin(mockUserRepo, mockPersonRepo, mockRoleRepo, mockLogger)

	return useCase, mockUserRepo, mockPersonRepo, mockRoleRepo
}

func createValidLoginRequest() ports.LoginRequest {
	return ports.LoginRequest{
		Email:    "john@example.com",
		Password: "SecurePass123!",
	}
}

// Success scenarios
func TestLogin_Success_CompleteProfile(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()
	user := createTestUser()
	roles := createTestRoles()

	// Setup mocks for success scenario with complete profile
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRoleRepo.On("GetUserRoles", ctx, user.ID).Return(roles, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)
	mockUserRepo.On("UpdateLastAccess", ctx, user.ID).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, user, result.User)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestLogin_Success_IncompleteProfile(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()
	user := createTestUser()
	roles := createTestRoles()

	// Setup mocks for success scenario with incomplete profile
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRoleRepo.On("GetUserRoles", ctx, user.ID).Return(roles, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(false, nil)
	mockUserRepo.On("UpdateLastAccess", ctx, user.ID).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, user, result.User)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestLogin_Success_UpdateLastAccessFails(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()
	user := createTestUser()
	roles := createTestRoles()

	// Setup mocks - UpdateLastAccess fails but login should still succeed
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRoleRepo.On("GetUserRoles", ctx, user.ID).Return(roles, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)
	mockUserRepo.On("UpdateLastAccess", ctx, user.ID).Return(errors.New("update failed"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert - should still succeed even if UpdateLastAccess fails
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// Validation error scenarios
func TestLogin_ValidationErrors(t *testing.T) {
	useCase, _, _, _ := setupLoginTest()
	ctx := context.Background()

	tests := []struct {
		name        string
		request     ports.LoginRequest
		expectedErr string
	}{
		{
			name: "Empty email",
			request: ports.LoginRequest{
				Email:    "",
				Password: "SecurePass123!",
			},
			expectedErr: "validation failed",
		},
		{
			name: "Invalid email format",
			request: ports.LoginRequest{
				Email:    "invalid-email",
				Password: "SecurePass123!",
			},
			expectedErr: "validation failed",
		},
		{
			name: "Empty password",
			request: ports.LoginRequest{
				Email:    "john@example.com",
				Password: "",
			},
			expectedErr: "validation failed",
		},
		{
			name: "Both empty",
			request: ports.LoginRequest{
				Email:    "",
				Password: "",
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

// Authentication failure scenarios
func TestLogin_UserNotFound(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()

	// Setup mock to return user not found error
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("user not found"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")

	mockUserRepo.AssertExpectations(t)
}

func TestLogin_InactiveUser(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()

	// Create inactive user
	user := createTestUser()
	user.Active = false

	// Setup mock
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user account is inactive")

	mockUserRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupLoginTest()
	ctx := context.Background()
	user := createTestUser()

	// Create request with wrong password
	req := ports.LoginRequest{
		Email:    "john@example.com",
		Password: "WrongPassword123!",
	}

	// Setup mock
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")

	mockUserRepo.AssertExpectations(t)
}

// Repository error scenarios
func TestLogin_GetRolesError(t *testing.T) {
	useCase, mockUserRepo, _, mockRoleRepo := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()
	user := createTestUser()

	// Setup mocks
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRoleRepo.On("GetUserRoles", ctx, user.ID).Return(nil, errors.New("roles fetch failed"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user roles")

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestLogin_ProfileCheckError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()
	user := createTestUser()
	roles := createTestRoles()

	// Setup mocks
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRoleRepo.On("GetUserRoles", ctx, user.ID).Return(roles, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(false, errors.New("profile check failed"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check profile completeness")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// Constructor test
func TestNewLogin(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockLogger := logger.New()

	useCase := NewLogin(mockUserRepo, mockPersonRepo, mockRoleRepo, mockLogger)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockUserRepo, useCase.userRepo)
	assert.Equal(t, mockPersonRepo, useCase.personRepo)
	assert.Equal(t, mockRoleRepo, useCase.roleRepo)
	assert.NotNil(t, useCase.jwtService)
	assert.Equal(t, mockLogger, useCase.logger)
}

// Edge cases
func TestLogin_ContextCancellation(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupLoginTest()
	req := createValidLoginRequest()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Setup mock to return context cancelled error
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, context.Canceled)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")

	mockUserRepo.AssertExpectations(t)
}

// Integration-style test with realistic data
func TestLogin_RealisticScenario(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupLoginTest()
	ctx := context.Background()

	// Realistic Brazilian user data
	req := ports.LoginRequest{
		Email:    "maria.silva@condominio.com.br",
		Password: "MinhaSenh@Segura123!",
	}

	user, _ := domain.NewUser(req.Email, req.Password)
	user.ID = 42
	user.CreatedAt = time.Now().Add(-30 * 24 * time.Hour) // 30 days ago

	roles := []jwt.Role{
		{
			ID:              2,
			Name:            "sindico",
			CondominiumID:   nil,
			CondominiumName: "",
		},
		{
			ID:              3,
			Name:            "morador",
			CondominiumID:   func() *int64 { id := int64(100); return &id }(),
			CondominiumName: "Edifício Residencial Jardim das Flores",
		},
	}

	// Setup mocks
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRoleRepo.On("GetUserRoles", ctx, user.ID).Return(roles, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)
	mockUserRepo.On("UpdateLastAccess", ctx, user.ID).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, user, result.User)
	assert.Equal(t, req.Email, result.User.Email)
	assert.True(t, result.User.Active)

	// Verify tokens are different
	assert.NotEqual(t, result.AccessToken, result.RefreshToken)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// Test with empty roles
func TestLogin_EmptyRoles(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()
	user := createTestUser()

	// Setup mocks with empty roles
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRoleRepo.On("GetUserRoles", ctx, user.ID).Return([]jwt.Role{}, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)
	mockUserRepo.On("UpdateLastAccess", ctx, user.ID).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert - should still succeed with empty roles
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// Test with multiple roles
func TestLogin_MultipleRoles(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupLoginTest()
	ctx := context.Background()
	req := createValidLoginRequest()
	user := createTestUser()

	// Multiple roles scenario
	roles := []jwt.Role{
		{ID: 1, Name: "super_admin"},
		{ID: 2, Name: "sindico", CondominiumID: func() *int64 { id := int64(100); return &id }()},
		{ID: 3, Name: "morador", CondominiumID: func() *int64 { id := int64(200); return &id }()},
	}

	// Setup mocks
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRoleRepo.On("GetUserRoles", ctx, user.ID).Return(roles, nil)
	mockPersonRepo.On("HasCompleteProfile", ctx, user.ID).Return(true, nil)
	mockUserRepo.On("UpdateLastAccess", ctx, user.ID).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
