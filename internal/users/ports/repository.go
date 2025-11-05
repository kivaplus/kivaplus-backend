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
	Create(ctx context.Context, person *domain.Person) error

	// GetByID retrieves a person by ID
	GetByID(ctx context.Context, id int64) (*domain.Person, error)

	// GetByUserID retrieves a person by user ID
	GetByUserID(ctx context.Context, userID int64) (*domain.Person, error)

	// Update updates an existing person
	Update(ctx context.Context, person *domain.Person) error

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

	// RemoveUserRoleWithoutCondominium removes a role from a user that has no condominium_id (NULL)
	RemoveUserRoleWithoutCondominium(ctx context.Context, userID int64, roleID int) error

	// GetUserRolesInCondominium retrieves user roles in a specific condominium
	GetUserRolesInCondominium(ctx context.Context, userID int64, condominiumID int64) ([]jwt.Role, error)

	// GetUsersByRole retrieves all users with a specific role (for security validation)
	GetUsersByRole(ctx context.Context, roleID int) ([]int64, error)
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

// PersonDocumentRepository defines the interface for person document data operations
type PersonDocumentRepository interface {
	// Create creates a new person document
	Create(ctx context.Context, document *domain.PersonDocument) error

	// GetByPersonID retrieves all documents for a person
	GetByPersonID(ctx context.Context, personID int64) ([]*domain.PersonDocument, error)

	// GetByPersonIDAndType retrieves a specific document type for a person
	GetByPersonIDAndType(ctx context.Context, personID int64, docType domain.DocumentType) (*domain.PersonDocument, error)

	// Update updates an existing person document
	Update(ctx context.Context, document *domain.PersonDocument) error

	// Delete deletes a person document
	Delete(ctx context.Context, id int64) error

	// ExistsByNumberAndType checks if a document with the same number and type already exists
	ExistsByNumberAndType(ctx context.Context, number string, docType domain.DocumentType) (bool, error)

	// ExistsByNumberAndTypeForDifferentPerson checks if a document exists for a different person
	ExistsByNumberAndTypeForDifferentPerson(ctx context.Context, number string, docType domain.DocumentType, excludePersonID int64) (bool, error)
}
