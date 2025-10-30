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

// Test helper functions for CompleteProfile use case
func setupCompleteProfileTest() (*CompleteProfile, *MockUserRepository, *MockPersonRepository, *MockAddressRepository) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockAddressRepo := new(MockAddressRepository)
	mockLogger := logger.New()

	useCase := NewCompleteProfile(mockUserRepo, mockPersonRepo, mockAddressRepo, mockLogger)

	return useCase, mockUserRepo, mockPersonRepo, mockAddressRepo
}

func createValidUpdateProfileRequest() ports.UpdateProfileRequest {
	email := "john@example.com"
	phoneNumber := "+55 11 99999-9999"
	occupation := "Software Engineer"
	return ports.UpdateProfileRequest{
		Name:        "John Doe Silva",
		PersonType:  domain.IndividualEntity,
		PhoneNumber: phoneNumber,
		Email:       email,
		Occupation:  occupation,
		Address: ports.CreateAddressRequest{
			ZipCode:      "01234-567",
			Street:       "Rua das Flores",
			Number:       "123",
			Complement:   "Apto 45",
			Neighborhood: "Centro",
			City:         "São Paulo",
			State:        "SP",
			Country:      "Brasil",
		},
	}
}

// Success scenarios
func TestCompleteProfile_Success_NewProfile(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()
	address := createTestAddress()

	// Setup mocks for new profile creation
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("pessoa not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user, result.User)
	assert.NotNil(t, result.Person)
	assert.Equal(t, "Profile complete successfully.", result.Message)
	assert.Equal(t, req.Name, result.Person.Name)
	assert.Equal(t, req.PhoneNumber, result.Person.PhoneNumber)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}

func TestCompleteProfile_Success_UpdateExistingProfile(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()
	address := createTestAddress()

	// Create existing person
	existingPerson, _ := domain.NewPerson("Old Name", "old@example.com", userID)
	existingPerson.ID = 1

	// Setup mocks for updating existing profile
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(existingPerson, nil)
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Update", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user, result.User)
	assert.NotNil(t, result.Person)
	assert.Equal(t, "Profile complete successfully.", result.Message)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}

// Validation error scenarios
func TestCompleteProfile_ValidationErrors(t *testing.T) {
	useCase, _, _, _ := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)

	tests := []struct {
		name        string
		request     ports.UpdateProfileRequest
		expectedErr string
	}{
		{
			name: "Empty name",
			request: func() ports.UpdateProfileRequest {
				req := createValidUpdateProfileRequest()
				req.Name = ""
				return req
			}(),
			expectedErr: "validation failed",
		},
		{
			name: "Name too short",
			request: func() ports.UpdateProfileRequest {
				req := createValidUpdateProfileRequest()
				req.Name = "A"
				return req
			}(),
			expectedErr: "validation failed",
		},
		{
			name: "Name too long",
			request: func() ports.UpdateProfileRequest {
				req := createValidUpdateProfileRequest()
				req.Name = string(make([]byte, 256)) // 256 characters
				return req
			}(),
			expectedErr: "validation failed",
		},
		{
			name: "Invalid email",
			request: func() ports.UpdateProfileRequest {
				req := createValidUpdateProfileRequest()
				invalidEmail := "invalid-email"
				req.Email = invalidEmail
				return req
			}(),
			expectedErr: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := useCase.Execute(ctx, userID, tt.request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// User-related error scenarios
func TestCompleteProfile_UserNotFound(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()

	// Setup mock to return user not found
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	mockUserRepo.AssertExpectations(t)
}

// Repository error scenarios
func TestCompleteProfile_PersonCheckError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _ := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("database error"))

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check existing profile")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

func TestCompleteProfile_AddressCreateError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("pessoa not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(nil, errors.New("address creation failed"))

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create address")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}

func TestCompleteProfile_PersonCreateError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()
	address := createTestAddress()

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("pessoa not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(errors.New("person creation failed"))

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create profile")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}

func TestCompleteProfile_PersonUpdateError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()
	address := createTestAddress()

	// Create existing person
	existingPerson, _ := domain.NewPerson("Old Name", "old@example.com", userID)
	existingPerson.ID = 1

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(existingPerson, nil)
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Update", ctx, mock.AnythingOfType("*domain.Person")).Return(errors.New("person update failed"))

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update profile")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}

// Constructor test
func TestNewCompleteProfile(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockAddressRepo := new(MockAddressRepository)
	mockLogger := logger.New()

	useCase := NewCompleteProfile(mockUserRepo, mockPersonRepo, mockAddressRepo, mockLogger)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockUserRepo, useCase.userRepo)
	assert.Equal(t, mockPersonRepo, useCase.personRepo)
	assert.Equal(t, mockAddressRepo, useCase.addressRepo)
	assert.Equal(t, mockLogger, useCase.logger)
}

// Edge cases
func TestCompleteProfile_ContextCancellation(t *testing.T) {
	useCase, mockUserRepo, _, _ := setupCompleteProfileTest()
	userID := int64(1)
	req := createValidUpdateProfileRequest()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Setup mock to return context cancelled error
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, context.Canceled)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")

	mockUserRepo.AssertExpectations(t)
}

// Integration-style test with realistic data
func TestCompleteProfile_RealisticScenario(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(42)

	// Realistic Brazilian user data
	email := "maria.silva@condominio.com.br"
	req := ports.UpdateProfileRequest{
		Name:        "Maria Silva Santos",
		PhoneNumber: "+55 11 98765-4321",
		Email:       email,
		Occupation:  "Síndica Profissional",
		Address: ports.CreateAddressRequest{
			ZipCode:      "04567-890",
			Street:       "Avenida Paulista",
			Number:       "1000",
			Complement:   "Conjunto 1501",
			Neighborhood: "Bela Vista",
			City:         "São Paulo",
			State:        "SP",
			Country:      "Brasil",
		},
	}

	user, _ := domain.NewUser(email, "MinhaSenh@Segura123!")
	user.ID = userID

	address, _ := domain.NewAddress(
		req.Address.ZipCode,
		req.Address.Street,
		req.Address.Number,
		req.Address.Complement,
		req.Address.Neighborhood,
		req.Address.City,
		req.Address.State,
		req.Address.Country,
	)
	address.ID = 100

	// Setup mocks for new profile creation
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("pessoa not found"))
	mockAddressRepo.On("Create", ctx, mock.MatchedBy(func(addr *domain.Address) bool {
		return addr.ZipCode == req.Address.ZipCode && addr.Street == req.Address.Street
	})).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.MatchedBy(func(person *domain.Person) bool {
		return person.Name == req.Name && person.PhoneNumber == req.PhoneNumber
	})).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, user, result.User)
	assert.Equal(t, req.Name, result.Person.Name)
	assert.Equal(t, req.PhoneNumber, result.Person.PhoneNumber)
	assert.Equal(t, req.Email, result.Person.Email)
	assert.Equal(t, req.Occupation, result.Person.Occupation)
	assert.Equal(t, &address.ID, result.Person.AddressID)
	assert.Equal(t, "Profile complete successfully.", result.Message)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}

// Test with optional fields
func TestCompleteProfile_OptionalFields(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	user := createTestUser()
	address := createTestAddress()

	// Request with minimal required fields
	req := ports.UpdateProfileRequest{
		Name: "John Doe",
		Address: ports.CreateAddressRequest{
			ZipCode:      "01234-567",
			Street:       "Rua das Flores",
			Number:       "123",
			Neighborhood: "Centro",
			City:         "São Paulo",
			State:        "SP",
			Country:      "Brasil",
		},
		// Optional fields not provided: PhoneNumber, Email, Occupation, Complement
	}

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("pessoa not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Person.Name)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}

// Test with empty email (should be allowed)
func TestCompleteProfile_EmptyEmail(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	user := createTestUser()
	address := createTestAddress()

	req := createValidUpdateProfileRequest()
	emptyEmail := ""
	req.Email = emptyEmail

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("pessoa not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert - should succeed with empty email
	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}
