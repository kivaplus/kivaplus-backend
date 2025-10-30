package ports

import (
	"context"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int64) (*domain.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *domain.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id int64) error

	// Exists checks if a user exists by email
	Exists(ctx context.Context, email string) (bool, error)

	// UpdateLastAccess updates the user's last access time
	UpdateLastAccess(ctx context.Context, id int64) error
}

// PersonRepository defines the interface for person data operations
type PersonRepository interface {
	// Create creates a new person
	Create(ctx context.Context, pessoa *domain.Person) error

	// GetByID retrieves a person by ID
	GetByID(ctx context.Context, id int64) (*domain.Person, error)

	// GetByUserID retrieves a person by user ID
	GetByUserID(ctx context.Context, usuarioID int64) (*domain.Person, error)

	// Update updates an existing person
	Update(ctx context.Context, pessoa *domain.Person) error

	// Delete soft deletes a person
	Delete(ctx context.Context, id int64) error

	// HasCompleteProfile verifica se o usuário tem um perfil completo
	HasCompleteProfile(ctx context.Context, userID int64) (bool, error)
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	// AddUserRole adds a role to a user
	AddUserRole(ctx context.Context, userID int64, roleID int, condominiumID *int64) error

	// GetUserRoles retrieves all roles for a user
	GetUserRoles(ctx context.Context, userID int64) ([]jwt.Role, error)

	// RemoveUserRole removes a role from a user
	RemoveUserRole(ctx context.Context, userID int64, roleID int, condominiumID *int64) error

	// GetUserRolesInCondominium retrieves user roles in a specific condominium
	GetUserRolesInCondominium(ctx context.Context, userID int64, condominiumID int64) ([]jwt.Role, error)
}

type AddressRepository interface {
	// Create creates a new address
	Create(ctx context.Context, address *domain.Address) (*domain.Address, error)

	// GetByID retrieves an address by ID
	GetByID(ctx context.Context, id int64) (*domain.Address, error)

	// Update updates an existing address
	Update(ctx context.Context, address *domain.Address) error

	// Delete soft deletes an address
	Delete(ctx context.Context, id int64) error
}
