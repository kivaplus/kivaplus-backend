package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
)

// CondominiumPostgresRepository implements the CondominiumRepository interface using PostgreSQL
type CondominiumPostgresRepository struct {
	db *database.Connection
}

// NewCondominiumPostgresRepository creates a new PostgreSQL condominium repository
func NewCondominiumPostgresRepository(db *database.Connection) *CondominiumPostgresRepository {
	return &CondominiumPostgresRepository{
		db: db,
	}
}

// Create creates a new condominium in the database
func (r *CondominiumPostgresRepository) Create(ctx context.Context, condominium *domain.Condominium) error {
	query := `
		INSERT INTO condominium (name, document_number, phone_number, email, address_id, manager_id, administrator_id, observations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		condominium.Name,
		condominium.DocumentNumber,
		condominium.PhoneNumber,
		condominium.Email,
		condominium.AddressID,
		condominium.ManagerID,
		condominium.AdministratorID,
		condominium.Observations,
		condominium.CreatedAt,
		condominium.UpdatedAt,
	).Scan(&condominium.ID)

	if err != nil {
		return fmt.Errorf("failed to create condominium: %w", err)
	}

	return nil
}

// GetByID retrieves a condominium by ID
func (r *CondominiumPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Condominium, error) {
	query := `
		SELECT id, name, document_number, phone_number, email, address_id, manager_id, administrator_id, observations, created_at, updated_at, deleted_at
		FROM condominium
		WHERE id = $1 AND deleted_at IS NULL`

	var condominium domain.Condominium
	var phoneNumber, email, observations sql.NullString
	var administratorID sql.NullInt64
	var deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&condominium.ID,
		&condominium.Name,
		&condominium.DocumentNumber,
		&phoneNumber,
		&email,
		&condominium.AddressID,
		&condominium.ManagerID,
		&administratorID,
		&observations,
		&condominium.CreatedAt,
		&condominium.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("condominium not found")
		}
		return nil, fmt.Errorf("failed to get condominium: %w", err)
	}

	// Handle nullable fields
	if phoneNumber.Valid {
		condominium.PhoneNumber = phoneNumber.String
	}
	if email.Valid {
		condominium.Email = email.String
	}
	if observations.Valid {
		condominium.Observations = observations.String
	}
	if administratorID.Valid {
		condominium.AdministratorID = &administratorID.Int64
	}
	if deletedAt.Valid {
		condominium.DeletedAt = &deletedAt.Time
	}

	return &condominium, nil
}

// GetByDocumentNumber retrieves a condominium by document number
func (r *CondominiumPostgresRepository) GetByDocumentNumber(ctx context.Context, documentNumber string) (*domain.Condominium, error) {
	query := `
		SELECT id, name, document_number, phone_number, email, address_id, manager_id, administrator_id, observations, created_at, updated_at, deleted_at
		FROM condominium
		WHERE document_number = $1 AND deleted_at IS NULL`

	var condominium domain.Condominium
	var phoneNumber, email, observations sql.NullString
	var administratorID sql.NullInt64
	var deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, documentNumber).Scan(
		&condominium.ID,
		&condominium.Name,
		&condominium.DocumentNumber,
		&phoneNumber,
		&email,
		&condominium.AddressID,
		&condominium.ManagerID,
		&administratorID,
		&observations,
		&condominium.CreatedAt,
		&condominium.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("condominium not found")
		}
		return nil, fmt.Errorf("failed to get condominium: %w", err)
	}

	// Handle nullable fields
	if phoneNumber.Valid {
		condominium.PhoneNumber = phoneNumber.String
	}
	if email.Valid {
		condominium.Email = email.String
	}
	if observations.Valid {
		condominium.Observations = observations.String
	}
	if administratorID.Valid {
		condominium.AdministratorID = &administratorID.Int64
	}
	if deletedAt.Valid {
		condominium.DeletedAt = &deletedAt.Time
	}

	return &condominium, nil
}

// GetByUserID retrieves all condominiums where the user has any role
func (r *CondominiumPostgresRepository) GetByUserID(ctx context.Context, userID int64) ([]*domain.Condominium, error) {
	query := `
		SELECT DISTINCT c.id, c.name, c.document_number, c.phone_number, c.email, c.address_id, c.manager_id, c.administrator_id, c.observations, c.created_at, c.updated_at, c.deleted_at
		FROM condominium c
		INNER JOIN users_role ur ON c.id = ur.condominium_id
		WHERE ur.users_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.name`

	rows, err := r.db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query condominiums: %w", err)
	}
	defer rows.Close()

	var condominiums []*domain.Condominium
	for rows.Next() {
		var condominium domain.Condominium
		var phoneNumber, email, observations sql.NullString
		var administratorID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&condominium.ID,
			&condominium.Name,
			&condominium.DocumentNumber,
			&phoneNumber,
			&email,
			&condominium.AddressID,
			&condominium.ManagerID,
			&administratorID,
			&observations,
			&condominium.CreatedAt,
			&condominium.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan condominium: %w", err)
		}

		// Handle nullable fields
		if phoneNumber.Valid {
			condominium.PhoneNumber = phoneNumber.String
		}
		if email.Valid {
			condominium.Email = email.String
		}
		if observations.Valid {
			condominium.Observations = observations.String
		}
		if administratorID.Valid {
			condominium.AdministratorID = &administratorID.Int64
		}
		if deletedAt.Valid {
			condominium.DeletedAt = &deletedAt.Time
		}

		condominiums = append(condominiums, &condominium)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate condominiums: %w", err)
	}

	return condominiums, nil
}

// GetAll retrieves all condominiums (for super admin)
func (r *CondominiumPostgresRepository) GetAll(ctx context.Context) ([]*domain.Condominium, error) {
	query := `
		SELECT id, name, document_number, phone_number, email, address_id, manager_id, administrator_id, observations, created_at, updated_at, deleted_at
		FROM condominium
		WHERE deleted_at IS NULL
		ORDER BY name`

	rows, err := r.db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all condominiums: %w", err)
	}
	defer rows.Close()

	var condominiums []*domain.Condominium
	for rows.Next() {
		var condominium domain.Condominium
		var phoneNumber, email, observations sql.NullString
		var administratorID sql.NullInt64
		var deletedAt sql.NullTime

		err := rows.Scan(
			&condominium.ID,
			&condominium.Name,
			&condominium.DocumentNumber,
			&phoneNumber,
			&email,
			&condominium.AddressID,
			&condominium.ManagerID,
			&administratorID,
			&observations,
			&condominium.CreatedAt,
			&condominium.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan condominium: %w", err)
		}

		// Handle nullable fields
		if phoneNumber.Valid {
			condominium.PhoneNumber = phoneNumber.String
		}
		if email.Valid {
			condominium.Email = email.String
		}
		if observations.Valid {
			condominium.Observations = observations.String
		}
		if administratorID.Valid {
			condominium.AdministratorID = &administratorID.Int64
		}
		if deletedAt.Valid {
			condominium.DeletedAt = &deletedAt.Time
		}

		condominiums = append(condominiums, &condominium)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate all condominiums: %w", err)
	}

	return condominiums, nil
}

// Update updates an existing condominium
func (r *CondominiumPostgresRepository) Update(ctx context.Context, condominium *domain.Condominium) error {
	query := `
		UPDATE condominium
		SET name = $1, document_number = $2, phone_number = $3, email = $4, address_id = $5, manager_id = $6, administrator_id = $7, observations = $8, updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL`

	result, err := r.db.DB.ExecContext(
		ctx,
		query,
		condominium.Name,
		condominium.DocumentNumber,
		condominium.PhoneNumber,
		condominium.Email,
		condominium.AddressID,
		condominium.ManagerID,
		condominium.AdministratorID,
		condominium.Observations,
		condominium.UpdatedAt,
		condominium.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update condominium: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("condominium not found")
	}

	return nil
}

// Delete soft deletes a condominium
func (r *CondominiumPostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE condominium SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete condominium: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("condominium not found")
	}

	return nil
}

// Exists checks if a condominium exists by document number
func (r *CondominiumPostgresRepository) Exists(ctx context.Context, documentNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM condominium WHERE document_number = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, documentNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check condominium existence: %w", err)
	}

	return exists, nil
}
