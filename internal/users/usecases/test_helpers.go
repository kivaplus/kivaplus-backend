package usecases

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
)

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastAccess(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) Exists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

type MockPersonRepository struct {
	mock.Mock
}

func (m *MockPersonRepository) Create(ctx context.Context, person *domain.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *MockPersonRepository) GetByID(ctx context.Context, id int64) (*domain.Person, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Person), args.Error(1)
}

func (m *MockPersonRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Person, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Person), args.Error(1)
}

func (m *MockPersonRepository) Update(ctx context.Context, person *domain.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *MockPersonRepository) HasCompleteProfile(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPersonRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) AddUserRole(ctx context.Context, userID int64, roleID int, condominiumID *int64) error {
	args := m.Called(ctx, userID, roleID, condominiumID)
	return args.Error(0)
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID int64) ([]jwt.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]jwt.Role), args.Error(1)
}

func (m *MockRoleRepository) RemoveUserRole(ctx context.Context, userID int64, roleID int, condominiumID *int64) error {
	args := m.Called(ctx, userID, roleID, condominiumID)
	return args.Error(0)
}

func (m *MockRoleRepository) GetUserRolesInCondominium(ctx context.Context, userID int64, condominiumID int64) ([]jwt.Role, error) {
	args := m.Called(ctx, userID, condominiumID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]jwt.Role), args.Error(1)
}

type MockAddressRepository struct {
	mock.Mock
}

func (m *MockAddressRepository) Create(ctx context.Context, address *domain.Address) (*domain.Address, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Address), args.Error(1)
}

func (m *MockAddressRepository) GetByID(ctx context.Context, id int64) (*domain.Address, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Address), args.Error(1)
}

func (m *MockAddressRepository) Update(ctx context.Context, address *domain.Address) error {
	args := m.Called(ctx, address)
	return args.Error(0)
}

func (m *MockAddressRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Test helper functions
func createTestUser() *domain.User {
	user, _ := domain.NewUser("john@example.com", "SecurePass123!")
	user.ID = 1
	user.Active = true
	return user
}

func createTestRoles() []jwt.Role {
	return []jwt.Role{
		{
			ID:   3,
			Name: "morador",
		},
	}
}

func createTestAddress() *domain.Address {
	address := &domain.Address{
		ID:           1,
		Street:       "Rua das Flores",
		Number:       "123",
		Complement:   "Apto 45",
		Neighborhood: "Centro",
		City:         "São Paulo",
		State:        "SP",
		ZipCode:      "01234-567",
		Country:      "Brasil",
	}
	return address
}

func createTestPerson() *domain.Person {
	addressID := int64(1)
	userID := int64(1)
	person := &domain.Person{
		ID:          1,
		UserID:      &userID,
		Name:        "John Doe",
		Email:       "john@example.com",
		PhoneNumber: "11999999999",
		AddressID:   &addressID,
	}
	return person
}
