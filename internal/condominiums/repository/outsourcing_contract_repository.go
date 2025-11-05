package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
)

// OutsourcingContractPostgresRepository implements the OutsourcingContractRepository interface using PostgreSQL
type OutsourcingContractPostgresRepository struct {
	db *database.Connection
}

// NewOutsourcingContractPostgresRepository creates a new PostgreSQL outsourcing contract repository
func NewOutsourcingContractPostgresRepository(db *database.Connection) *OutsourcingContractPostgresRepository {
	return &OutsourcingContractPostgresRepository{
		db: db,
	}
}

// Create creates a new outsourcing contract in the database
func (r *OutsourcingContractPostgresRepository) Create(ctx context.Context, contract *domain.OutsourcingContract) error {
	query := `
		INSERT INTO outsourcing_contract (condominium_id, outsourcing_company_id, contract_number, monthly_amount, due_day, start_date, end_date, status, observations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		contract.CondominiumID,
		contract.OutsourcingCompanyID,
		contract.ContractNumber,
		contract.MonthlyAmount,
		contract.DueDay,
		contract.StartDate,
		contract.EndDate,
		contract.Status,
		contract.Observations,
		contract.CreatedAt,
		contract.UpdatedAt,
	).Scan(&contract.ID)

	if err != nil {
		return fmt.Errorf("failed to create outsourcing contract: %w", err)
	}

	return nil
}

// GetByID retrieves an outsourcing contract by ID
func (r *OutsourcingContractPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.OutsourcingContract, error) {
	query := `
		SELECT id, condominium_id, outsourcing_company_id, contract_number, monthly_amount, due_day, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM outsourcing_contract
		WHERE id = $1 AND deleted_at IS NULL`

	return r.scanContract(ctx, query, id)
}

// GetByContractNumber retrieves an outsourcing contract by contract number
func (r *OutsourcingContractPostgresRepository) GetByContractNumber(ctx context.Context, contractNumber string) (*domain.OutsourcingContract, error) {
	query := `
		SELECT id, condominium_id, outsourcing_company_id, contract_number, monthly_amount, due_day, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM outsourcing_contract
		WHERE contract_number = $1 AND deleted_at IS NULL`

	return r.scanContract(ctx, query, contractNumber)
}

// GetByCondominiumID retrieves all contracts for a condominium
func (r *OutsourcingContractPostgresRepository) GetByCondominiumID(ctx context.Context, condominiumID int64) ([]*domain.OutsourcingContract, error) {
	query := `
		SELECT id, condominium_id, outsourcing_company_id, contract_number, monthly_amount, due_day, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM outsourcing_contract
		WHERE condominium_id = $1 AND deleted_at IS NULL
		ORDER BY start_date DESC`

	return r.scanContracts(ctx, query, condominiumID)
}

// GetByCompanyID retrieves all contracts for an outsourcing company
func (r *OutsourcingContractPostgresRepository) GetByCompanyID(ctx context.Context, companyID int64) ([]*domain.OutsourcingContract, error) {
	query := `
		SELECT id, condominium_id, outsourcing_company_id, contract_number, monthly_amount, due_day, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM outsourcing_contract
		WHERE outsourcing_company_id = $1 AND deleted_at IS NULL
		ORDER BY start_date DESC`

	return r.scanContracts(ctx, query, companyID)
}

// GetByStatus retrieves contracts by status
func (r *OutsourcingContractPostgresRepository) GetByStatus(ctx context.Context, status domain.ContractStatus) ([]*domain.OutsourcingContract, error) {
	query := `
		SELECT id, condominium_id, outsourcing_company_id, contract_number, monthly_amount, due_day, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM outsourcing_contract
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY start_date DESC`

	return r.scanContracts(ctx, query, status)
}

// Update updates an existing outsourcing contract
func (r *OutsourcingContractPostgresRepository) Update(ctx context.Context, contract *domain.OutsourcingContract) error {
	query := `
		UPDATE outsourcing_contract
		SET contract_number = $1, monthly_amount = $2, due_day = $3, start_date = $4, end_date = $5, status = $6, observations = $7, updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL`

	result, err := r.db.DB.ExecContext(
		ctx,
		query,
		contract.ContractNumber,
		contract.MonthlyAmount,
		contract.DueDay,
		contract.StartDate,
		contract.EndDate,
		contract.Status,
		contract.Observations,
		contract.UpdatedAt,
		contract.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update outsourcing contract: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("outsourcing contract not found")
	}

	return nil
}

// Delete soft deletes an outsourcing contract
func (r *OutsourcingContractPostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE outsourcing_contract SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete outsourcing contract: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("outsourcing contract not found")
	}

	return nil
}

// Exists checks if a contract exists by contract number
func (r *OutsourcingContractPostgresRepository) Exists(ctx context.Context, contractNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM outsourcing_contract WHERE contract_number = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, contractNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check outsourcing contract existence: %w", err)
	}

	return exists, nil
}

// Helper methods

func (r *OutsourcingContractPostgresRepository) scanContract(ctx context.Context, query string, args ...interface{}) (*domain.OutsourcingContract, error) {
	var contract domain.OutsourcingContract
	var endDate sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, args...).Scan(
		&contract.ID,
		&contract.CondominiumID,
		&contract.OutsourcingCompanyID,
		&contract.ContractNumber,
		&contract.MonthlyAmount,
		&contract.DueDay,
		&contract.StartDate,
		&endDate,
		&contract.Status,
		&contract.Observations,
		&contract.CreatedAt,
		&contract.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("outsourcing contract not found")
		}
		return nil, fmt.Errorf("failed to get outsourcing contract: %w", err)
	}

	// Handle nullable fields
	if endDate.Valid {
		contract.EndDate = &endDate.Time
	}
	if deletedAt.Valid {
		contract.DeletedAt = &deletedAt.Time
	}

	return &contract, nil
}

func (r *OutsourcingContractPostgresRepository) scanContracts(ctx context.Context, query string, args ...interface{}) ([]*domain.OutsourcingContract, error) {
	rows, err := r.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query outsourcing contracts: %w", err)
	}
	defer rows.Close()

	var contracts []*domain.OutsourcingContract
	for rows.Next() {
		var contract domain.OutsourcingContract
		var endDate sql.NullTime
		var deletedAt sql.NullTime

		err := rows.Scan(
			&contract.ID,
			&contract.CondominiumID,
			&contract.OutsourcingCompanyID,
			&contract.ContractNumber,
			&contract.MonthlyAmount,
			&contract.DueDay,
			&contract.StartDate,
			&endDate,
			&contract.Status,
			&contract.Observations,
			&contract.CreatedAt,
			&contract.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outsourcing contract: %w", err)
		}

		// Handle nullable fields
		if endDate.Valid {
			contract.EndDate = &endDate.Time
		}
		if deletedAt.Valid {
			contract.DeletedAt = &deletedAt.Time
		}

		contracts = append(contracts, &contract)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate outsourcing contracts: %w", err)
	}

	return contracts, nil
}
