package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// Test helper functions for CompleteProfile use case
func setupCompleteProfileTest() (*CompleteProfile, *MockUserRepository, *MockPersonRepository, *MockAddressRepository, *MockPersonDocumentRepository) {
	mockUserRepo := new(MockUserRepository)
	mockPersonRepo := new(MockPersonRepository)
	mockAddressRepo := new(MockAddressRepository)
	mockPersonDocumentRepo := new(MockPersonDocumentRepository)
	mockRoleRepo := new(MockRoleRepository)
	jwtService := jwt.GetJWTService()
	mockLogger := logger.New()

	useCase := NewCompleteProfile(mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo, mockRoleRepo, jwtService, mockLogger)

	return useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo
}

func createValidUpdateProfileRequest() ports.UpdateProfileRequest {
	email := "john@example.com"
	phoneNumber := "+55 11 99999-9999"
	occupation := "Software Engineer"
	return ports.UpdateProfileRequest{
		Name:          "John Doe Silva",
		Document:      "12345678901",
		DocumentType:  domain.CPF,
		PersonType:    domain.IndividualEntity,
		PhoneNumber:   phoneNumber,
		Email:         email,
		Occupation:    occupation,
		MaritalStatus: domain.Single,
		BirthdayDate:  "1990-01-01",
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
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()
	address := createTestAddress()

	// Setup mocks for new profile creation
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Document validation mocks
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, int64(-1)).Return(false, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))

	// Profile creation mocks
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)

	// Document creation mocks
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(nil, errors.New("document not found")).Once()
	mockPersonDocumentRepo.On("Create", ctx, mock.AnythingOfType("*domain.PersonDocument")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(&domain.PersonDocument{
		Number: req.Document,
		Type:   req.DocumentType,
	}, nil).Once()

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Profile)
	assert.Equal(t, "Profile completed successfully.", result.Message)
	assert.Equal(t, req.Name, result.Profile.Name)
	assert.Equal(t, req.PhoneNumber, result.Profile.PhoneNumber)
	assert.Equal(t, req.Document, result.Profile.Document)
	assert.Equal(t, req.DocumentType, result.Profile.DocumentType)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
	mockPersonDocumentRepo.AssertExpectations(t)
}

func TestCompleteProfile_Success_UpdateExistingProfile(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
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

	// Document validation mocks
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, existingPerson.ID).Return(false, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(existingPerson, nil)

	// Profile update mocks
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Update", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)

	// Document update mocks
	existingDoc := &domain.PersonDocument{
		ID:       1,
		PersonID: existingPerson.ID,
		Number:   "98765432100",
		Type:     req.DocumentType,
	}
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, existingPerson.ID, req.DocumentType).Return(existingDoc, nil)
	mockPersonDocumentRepo.On("Update", ctx, mock.AnythingOfType("*domain.PersonDocument")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, existingPerson.ID, req.DocumentType).Return(&domain.PersonDocument{
		Number: req.Document,
		Type:   req.DocumentType,
	}, nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Profile)
	assert.Equal(t, "Profile completed successfully.", result.Message)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
	mockPersonDocumentRepo.AssertExpectations(t)
}

// Validation error scenarios
func TestCompleteProfile_ValidationErrors(t *testing.T) {
	useCase, mockUserRepo, _, _, _ := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	user := createTestUser()

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
			// Mock user exists for validation tests
			mockUserRepo.On("GetByID", ctx, userID).Return(user, nil).Once()

			result, err := useCase.Execute(ctx, userID, tt.request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// User-related error scenarios
func TestCompleteProfile_UserNotFound(t *testing.T) {
	useCase, mockUserRepo, _, _, _ := setupCompleteProfileTest()
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
	useCase, mockUserRepo, mockPersonRepo, _, _ := setupCompleteProfileTest()
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
	assert.Contains(t, err.Error(), "failed to validate document ownership")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
}

func TestCompleteProfile_AddressCreateError(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, int64(-1)).Return(false, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))
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
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()
	address := createTestAddress()

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, int64(-1)).Return(false, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))
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
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
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
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, existingPerson.ID).Return(false, nil)
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
	mockPersonDocumentRepo := new(MockPersonDocumentRepository)
	mockRoleRepo := new(MockRoleRepository)
	jwtService := jwt.GetJWTService()
	mockLogger := logger.New()

	useCase := NewCompleteProfile(mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo, mockRoleRepo, jwtService, mockLogger)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockUserRepo, useCase.userRepo)
	assert.Equal(t, mockPersonRepo, useCase.personRepo)
	assert.Equal(t, mockAddressRepo, useCase.addressRepo)
	assert.Equal(t, mockPersonDocumentRepo, useCase.personDocumentRepo)
	assert.Equal(t, mockLogger, useCase.logger)
}

// Edge cases
func TestCompleteProfile_ContextCancellation(t *testing.T) {
	useCase, mockUserRepo, _, _, _ := setupCompleteProfileTest()
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
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(42)

	// Realistic Brazilian user data
	email := "maria.silva@condominio.com.br"
	req := ports.UpdateProfileRequest{
		Name:          "Maria Silva Santos",
		Document:      "12345678901",
		DocumentType:  domain.CPF,
		PhoneNumber:   "+55 11 98765-4321",
		Email:         email,
		Occupation:    "Síndica Profissional",
		MaritalStatus: domain.Married,
		BirthdayDate:  "1985-03-15",
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
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, int64(-1)).Return(false, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))
	mockAddressRepo.On("Create", ctx, mock.MatchedBy(func(addr *domain.Address) bool {
		return addr.ZipCode == req.Address.ZipCode && addr.Street == req.Address.Street
	})).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.MatchedBy(func(person *domain.Person) bool {
		return person.Name == req.Name && person.PhoneNumber == req.PhoneNumber
	})).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(nil, errors.New("document not found")).Once()
	mockPersonDocumentRepo.On("Create", ctx, mock.AnythingOfType("*domain.PersonDocument")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(&domain.PersonDocument{
		Number: req.Document,
		Type:   req.DocumentType,
	}, nil).Once()

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotNil(t, result.Profile)
	assert.Equal(t, req.Name, result.Profile.Name)
	assert.Equal(t, req.PhoneNumber, result.Profile.PhoneNumber)
	assert.Equal(t, req.Email, result.Profile.Email)
	assert.Equal(t, req.Occupation, result.Profile.Occupation)
	assert.Equal(t, req.Document, result.Profile.Document)
	assert.Equal(t, req.DocumentType, result.Profile.DocumentType)
	assert.Equal(t, req.MaritalStatus, result.Profile.MaritalStatus)
	assert.Equal(t, req.BirthdayDate, result.Profile.BirthdayDate)
	assert.Equal(t, "Profile completed successfully.", result.Message)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
	mockPersonDocumentRepo.AssertExpectations(t)
}

// Test with optional fields
func TestCompleteProfile_OptionalFields(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, _ := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	user := createTestUser()
	address := createTestAddress()

	// Request with minimal required fields (no document)
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
		// Optional fields not provided: Document, PhoneNumber, Email, Occupation, Complement
	}

	// Setup mocks (no document validation needed since no document provided)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Profile)
	assert.Equal(t, req.Name, result.Profile.Name)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
}

// Test with empty email (should be allowed)
func TestCompleteProfile_EmptyEmail(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	user := createTestUser()
	address := createTestAddress()

	req := createValidUpdateProfileRequest()
	emptyEmail := ""
	req.Email = emptyEmail

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, int64(-1)).Return(false, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(nil, errors.New("document not found")).Once()
	mockPersonDocumentRepo.On("Create", ctx, mock.AnythingOfType("*domain.PersonDocument")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(&domain.PersonDocument{
		Number: req.Document,
		Type:   req.DocumentType,
	}, nil).Once()

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert - should succeed with empty email
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Profile)

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
	mockPersonDocumentRepo.AssertExpectations(t)
}

// Document validation tests
func TestCompleteProfile_DocumentValidation_InvalidCPF(t *testing.T) {
	useCase, mockUserRepo, _, _, _ := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	user := createTestUser()

	req := createValidUpdateProfileRequest()
	req.Document = "123456789" // Invalid CPF (only 9 digits)

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")

	mockUserRepo.AssertExpectations(t)
}

func TestCompleteProfile_DocumentValidation_InvalidCNPJ(t *testing.T) {
	useCase, mockUserRepo, _, _, _ := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	user := createTestUser()

	req := createValidUpdateProfileRequest()
	req.Document = "123456789012" // Invalid CNPJ (only 12 digits)
	req.DocumentType = domain.CNPJ

	// Setup mocks
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")

	mockUserRepo.AssertExpectations(t)
}

func TestCompleteProfile_DocumentValidation_DuplicateDocument(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, _, mockPersonDocumentRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	user := createTestUser()

	// Setup mocks - document already exists for another user
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, int64(-1)).Return(true, nil)

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")

	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockPersonDocumentRepo.AssertExpectations(t)
}

func TestCompleteProfile_DocumentValidation_ValidCNPJ(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	req.Document = "12345678901234" // Valid CNPJ (14 digits)
	req.DocumentType = domain.CNPJ

	user := createTestUser()
	address := createTestAddress()

	// Setup mocks for new profile creation with CNPJ
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, int64(-1)).Return(false, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(nil, errors.New("document not found")).Once()
	mockPersonDocumentRepo.On("Create", ctx, mock.AnythingOfType("*domain.PersonDocument")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(&domain.PersonDocument{
		Number: req.Document,
		Type:   req.DocumentType,
	}, nil).Once()

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Profile)
	assert.Equal(t, req.Document, result.Profile.Document)
	assert.Equal(t, req.DocumentType, result.Profile.DocumentType)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
	mockPersonDocumentRepo.AssertExpectations(t)
}

func TestCompleteProfile_DocumentValidation_FormattedCPF(t *testing.T) {
	useCase, mockUserRepo, mockPersonRepo, mockAddressRepo, mockPersonDocumentRepo := setupCompleteProfileTest()
	ctx := context.Background()
	userID := int64(1)
	req := createValidUpdateProfileRequest()
	req.Document = "123.456.789-01" // Formatted CPF

	user := createTestUser()
	address := createTestAddress()

	// Setup mocks for new profile creation with formatted CPF
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockPersonDocumentRepo.On("ExistsByNumberAndTypeForDifferentPerson", ctx, req.Document, req.DocumentType, int64(-1)).Return(false, nil)
	mockPersonRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("person not found"))
	mockAddressRepo.On("Create", ctx, mock.AnythingOfType("*domain.Address")).Return(address, nil)
	mockPersonRepo.On("Create", ctx, mock.AnythingOfType("*domain.Person")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(nil, errors.New("document not found")).Once()
	mockPersonDocumentRepo.On("Create", ctx, mock.AnythingOfType("*domain.PersonDocument")).Return(nil)
	mockPersonDocumentRepo.On("GetByPersonIDAndType", ctx, mock.AnythingOfType("int64"), req.DocumentType).Return(&domain.PersonDocument{
		Number: req.Document,
		Type:   req.DocumentType,
	}, nil).Once()

	// Execute
	result, err := useCase.Execute(ctx, userID, req)

	// Assert - should succeed with formatted CPF
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Profile)
	assert.Equal(t, req.Document, result.Profile.Document)

	// Verify all mocks were called
	mockUserRepo.AssertExpectations(t)
	mockPersonRepo.AssertExpectations(t)
	mockAddressRepo.AssertExpectations(t)
	mockPersonDocumentRepo.AssertExpectations(t)
}
