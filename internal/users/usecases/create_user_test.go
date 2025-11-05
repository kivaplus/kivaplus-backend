package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// Test helper functions
func setupCreateUserTest() (*CreateUser, *MockUserRepository, *MockPersonRepository, *MockRoleRepository) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockLogger := logger.New()

	useCase := NewCreateUser(mockUserRepo, mockPersonRepo, mockRoleRepo, mockLogger)

	return useCase, mockUserRepo, mockPersonRepo, mockRoleRepo
}

func createValidRequest() ports.CreateUserRequest {
	return ports.CreateUserRequest{
		Name:            "John Doe",
		Email:           "john@example.com",
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
	}
}

// Success scenarios
func TestCreateUser_Success(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupCreateUserTest()
	ctx := context.Background()
	req := createValidRequest()

	// Setup mocks for success scenario
	mockUserRepo.On("Exists", ctx, req.Email).Return(false, nil)
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)
	mockRoleRepo.On("AddUserRole", ctx, mock.AnythingOfType("int64"), 3, (*int64)(nil)).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.User)
	assert.NotNil(t, result.Person)
	assert.Equal(t, req.Email, result.User.Email)
	assert.Equal(t, req.Name, result.Person.Name)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// Validation error scenarios
func TestCreateUser_ValidationErrors(t *testing.T) {
	useCase, _, _, _ := setupCreateUserTest()
	ctx := context.Background()

	tests := []struct {
		name        string
		request     ports.CreateUserRequest
		expectedErr string
	}{
		{
			name: "Empty name",
			request: ports.CreateUserRequest{
				Name:            "",
				Email:           "john@example.com",
				Password:        "SecurePass123!",
				ConfirmPassword: "SecurePass123!",
			},
			expectedErr: "validation failed",
		},
		{
			name: "Invalid email",
			request: ports.CreateUserRequest{
				Name:            "John Doe",
				Email:           "invalid-email",
				Password:        "SecurePass123!",
				ConfirmPassword: "SecurePass123!",
			},
			expectedErr: "validation failed",
		},
		{
			name: "Weak password",
			request: ports.CreateUserRequest{
				Name:            "John Doe",
				Email:           "john@example.com",
				Password:        "123",
				ConfirmPassword: "123",
			},
			expectedErr: "validation failed",
		},
		{
			name: "Password mismatch",
			request: ports.CreateUserRequest{
				Name:            "John Doe",
				Email:           "john@example.com",
				Password:        "SecurePass123!",
				ConfirmPassword: "DifferentPass123!",
			},
			expectedErr: "validation failed",
		},
		{
			name: "Empty password",
			request: ports.CreateUserRequest{
				Name:            "John Doe",
				Email:           "john@example.com",
				Password:        "",
				ConfirmPassword: "",
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

// User already exists scenario
func TestCreateUser_UserAlreadyExists(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupCreateUserTest()
	ctx := context.Background()
	req := createValidRequest()

	// Setup mock to return that user exists
	mockUserRepo.On("Exists", ctx, req.Email).Return(true, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user with email john@example.com already exists")

	mockUserRepo.AssertExpectations(t)
}

// Repository error scenarios
func TestCreateUser_ExistsCheckError(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupCreateUserTest()
	ctx := context.Background()
	req := createValidRequest()

	// Setup mock to return error when checking if user exists
	mockUserRepo.On("Exists", ctx, req.Email).Return(false, errors.New("database error"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check if user exists")

	mockUserRepo.AssertExpectations(t)
}

func TestCreateUser_UserCreateError(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupCreateUserTest()
	ctx := context.Background()
	req := createValidRequest()

	// Setup mocks
	mockUserRepo.On("Exists", ctx, req.Email).Return(false, nil)
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("user creation failed"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save user")

	mockUserRepo.AssertExpectations(t)
}

func TestCreateUser_PersonCreateError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _ := setupCreateUserTest()
	ctx := context.Background()
	req := createValidRequest()

	// Setup mocks
	mockUserRepo.On("Exists", ctx, req.Email).Return(false, nil)
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(errors.New("person creation failed"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save person")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

func TestCreateUser_RoleAssignmentError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupCreateUserTest()
	ctx := context.Background()
	req := createValidRequest()

	// Setup mocks
	mockUserRepo.On("Exists", ctx, req.Email).Return(false, nil)
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)
	mockRoleRepo.On("AddUserRole", ctx, mock.AnythingOfType("int64"), 3, (*int64)(nil)).Return(errors.New("role assignment failed"))

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to assign default role")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// Domain object creation error scenarios
func TestCreateUser_InvalidUserDomain(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupCreateUserTest()
	ctx := context.Background()

	// Create request with invalid data that would cause domain object creation to fail
	req := ports.CreateUserRequest{
		Name:            "John Doe",
		Email:           "", // This should cause domain.NewUser to fail
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
	}

	// Setup mock (won't be called due to validation failure)
	mockUserRepo.On("Exists", ctx, req.Email).Return(false, nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert - should fail at validation level, not domain creation
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

// Constructor test
func TestNewCreateUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockLogger := logger.New()

	useCase := NewCreateUser(mockUserRepo, mockPersonRepo, mockRoleRepo, mockLogger)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockUserRepo, useCase.userRepo)
	assert.Equal(t, mockPersonRepo, useCase.personRepo)
	assert.Equal(t, mockRoleRepo, useCase.roleRepo)
	assert.Equal(t, mockLogger, useCase.logger)
}

// Edge cases
func TestCreateUser_ContextCancellation(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupCreateUserTest()
	req := createValidRequest()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Setup mock to return context cancelled error
	mockUserRepo.On("Exists", ctx, req.Email).Return(false, context.Canceled)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check if user exists")

	mockUserRepo.AssertExpectations(t)
}

// Integration-style test with realistic data
func TestCreateUser_RealisticScenario(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockRoleRepo := setupCreateUserTest()
	ctx := context.Background()

	req := ports.CreateUserRequest{
		Name:            "Maria Silva Santos",
		Email:           "maria.silva@condominio.com.br",
		Password:        "MinhaSenh@Segura123!",
		ConfirmPassword: "MinhaSenh@Segura123!",
	}

	// Setup mocks for realistic scenario
	mockUserRepo.On("Exists", ctx, req.Email).Return(false, nil)
	mockUserRepo.On("Create", ctx, mock.MatchedBy(func(user *domain.User) bool {
		return user.Email == req.Email && user.Active == true
	})).Return(nil)
	mockPersonRepo.On("Create", ctx, mock.MatchedBy(func(person *domain.Person) bool {
		return person.Name == req.Name && person.Email == req.Email
	})).Return(nil)
	mockRoleRepo.On("AddUserRole", ctx, mock.AnythingOfType("int64"), 3, (*int64)(nil)).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, req.Email, result.User.Email)
	assert.Equal(t, req.Name, result.Person.Name)
	assert.True(t, result.User.Active)
	assert.NotEmpty(t, result.User.PasswordHash)

	// Verify all mocks were called with expected parameters
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}
