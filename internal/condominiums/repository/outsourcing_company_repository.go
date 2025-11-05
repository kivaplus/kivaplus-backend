package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
)

// OutsourcingCompanyPostgresRepository implements the OutsourcingCompanyRepository interface using PostgreSQL
type OutsourcingCompanyPostgresRepository struct {
	db *database.Connection
}

// NewOutsourcingCompanyPostgresRepository creates a new PostgreSQL outsourcing company repository
func NewOutsourcingCompanyPostgresRepository(db *database.Connection) *OutsourcingCompanyPostgresRepository {
	return &OutsourcingCompanyPostgresRepository{
		db: db,
	}
}

// Create creates a new outsourcing company in the database
func (r *OutsourcingCompanyPostgresRepository) Create(ctx context.Context, company *domain.OutsourcingCompany) error {
	query := `
		INSERT INTO outsourcing_company (name, document_number, legal_name, phone_number, email, address_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		company.Name,
		company.DocumentNumber,
		company.LegalName,
		company.PhoneNumber,
		company.Email,
		company.AddressID,
		company.CreatedAt,
		company.UpdatedAt,
	).Scan(&company.ID)

	if err != nil {
		return fmt.Errorf("failed to create outsourcing company: %w", err)
	}

	return nil
}

// GetByID retrieves an outsourcing company by ID
func (r *OutsourcingCompanyPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.OutsourcingCompany, error) {
	query := `
		SELECT id, name, document_number, legal_name, phone_number, email, address_id, created_at, updated_at, deleted_at
		FROM outsourcing_company
		WHERE id = $1 AND deleted_at IS NULL`

	var company domain.OutsourcingCompany
	var phoneNumber, email sql.NullString
	var deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&company.ID,
		&company.Name,
		&company.DocumentNumber,
		&company.LegalName,
		&phoneNumber,
		&email,
		&company.AddressID,
		&company.CreatedAt,
		&company.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("outsourcing company not found")
		}
		return nil, fmt.Errorf("failed to get outsourcing company: %w", err)
	}

	// Handle nullable fields
	if phoneNumber.Valid {
		company.PhoneNumber = phoneNumber.String
	}
	if email.Valid {
		company.Email = email.String
	}
	if deletedAt.Valid {
		company.DeletedAt = &deletedAt.Time
	}

	return &company, nil
}

// GetByDocumentNumber retrieves an outsourcing company by document number
func (r *OutsourcingCompanyPostgresRepository) GetByDocumentNumber(ctx context.Context, documentNumber string) (*domain.OutsourcingCompany, error) {
	query := `
		SELECT id, name, document_number, legal_name, phone_number, email, address_id, created_at, updated_at, deleted_at
		FROM outsourcing_company
		WHERE document_number = $1 AND deleted_at IS NULL`

	var company domain.OutsourcingCompany
	var phoneNumber, email sql.NullString
	var deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, documentNumber).Scan(
		&company.ID,
		&company.Name,
		&company.DocumentNumber,
		&company.LegalName,
		&phoneNumber,
		&email,
		&company.AddressID,
		&company.CreatedAt,
		&company.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("outsourcing company not found")
		}
		return nil, fmt.Errorf("failed to get outsourcing company: %w", err)
	}

	// Handle nullable fields
	if phoneNumber.Valid {
		company.PhoneNumber = phoneNumber.String
	}
	if email.Valid {
		company.Email = email.String
	}
	if deletedAt.Valid {
		company.DeletedAt = &deletedAt.Time
	}

	return &company, nil
}

// GetAll retrieves all outsourcing companies
func (r *OutsourcingCompanyPostgresRepository) GetAll(ctx context.Context) ([]*domain.OutsourcingCompany, error) {
	query := `
		SELECT id, name, document_number, legal_name, phone_number, email, address_id, created_at, updated_at, deleted_at
		FROM outsourcing_company
		WHERE deleted_at IS NULL
		ORDER BY name`

	rows, err := r.db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query outsourcing companies: %w", err)
	}
	defer rows.Close()

	var companies []*domain.OutsourcingCompany
	for rows.Next() {
		var company domain.OutsourcingCompany
		var phoneNumber, email sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.DocumentNumber,
			&company.LegalName,
			&phoneNumber,
			&email,
			&company.AddressID,
			&company.CreatedAt,
			&company.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outsourcing company: %w", err)
		}

		// Handle nullable fields
		if phoneNumber.Valid {
			company.PhoneNumber = phoneNumber.String
		}
		if email.Valid {
			company.Email = email.String
		}
		if deletedAt.Valid {
			company.DeletedAt = &deletedAt.Time
		}

		companies = append(companies, &company)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate outsourcing companies: %w", err)
	}

	return companies, nil
}

// Update updates an existing outsourcing company
func (r *OutsourcingCompanyPostgresRepository) Update(ctx context.Context, company *domain.OutsourcingCompany) error {
	query := `
		UPDATE outsourcing_company
		SET name = $1, document_number = $2, legal_name = $3, phone_number = $4, email = $5, address_id = $6, updated_at = $7
		WHERE id = $8 AND deleted_at IS NULL`

	result, err := r.db.DB.ExecContext(
		ctx,
		query,
		company.Name,
		company.DocumentNumber,
		company.LegalName,
		company.PhoneNumber,
		company.Email,
		company.AddressID,
		company.UpdatedAt,
		company.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update outsourcing company: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("outsourcing company not found")
	}

	return nil
}

// Delete soft deletes an outsourcing company
func (r *OutsourcingCompanyPostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE outsourcing_company SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete outsourcing company: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("outsourcing company not found")
	}

	return nil
}

// Exists checks if an outsourcing company exists by document number
func (r *OutsourcingCompanyPostgresRepository) Exists(ctx context.Context, documentNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM outsourcing_company WHERE document_number = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, documentNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check outsourcing company existence: %w", err)
	}

	return exists, nil
}
