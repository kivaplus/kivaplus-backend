package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
)

// UserPostgresRepository implements the UserRepository interface using PostgreSQL
type UserPostgresRepository struct {
	db *database.Connection
}

// NewUserPostgresRepository creates a new PostgreSQL user repository
func NewUserPostgresRepository(db *database.Connection) *UserPostgresRepository {
	return &UserPostgresRepository{
		db: db,
	}
}

// Create creates a new user in the database
func (r *UserPostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by email
func (r *UserPostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, active, last_access, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	user := &domain.User{}
	var lastAccess, deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Active,
		&lastAccess,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Handle nullable fields
	if lastAccess.Valid {
		user.LastAccess = &lastAccess.Time
	}
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, active, last_access, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	user := &domain.User{}
	var lastAccess, deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Active,
		&lastAccess,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Handle nullable fields
	if lastAccess.Valid {
		user.LastAccess = &lastAccess.Time
	}
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return user, nil
}

// Update updates an existing user
func (r *UserPostgresRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, active = $4, last_access = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.DB.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Active,
		user.LastAccess,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete soft deletes a user
func (r *UserPostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE users SET deleted_at = $2, updated_at = $2 WHERE id = $1`

	_, err := r.db.DB.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Exists checks if a user exists by email
func (r *UserPostgresRepository) Exists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

// UpdateLastAccess updates the user's last access time
func (r *UserPostgresRepository) UpdateLastAccess(ctx context.Context, id int64) error {
	query := `UPDATE users SET last_access = $2, updated_at = $2 WHERE id = $1`

	_, err := r.db.DB.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last access: %w", err)
	}

	return nil
}
