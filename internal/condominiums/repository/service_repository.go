package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
)

// ServicePostgresRepository implements the ServiceRepository interface using PostgreSQL
type ServicePostgresRepository struct {
	db *database.Connection
}

// NewServicePostgresRepository creates a new PostgreSQL service repository
func NewServicePostgresRepository(db *database.Connection) *ServicePostgresRepository {
	return &ServicePostgresRepository{
		db: db,
	}
}

// Create creates a new service in the database
func (r *ServicePostgresRepository) Create(ctx context.Context, service *domain.Service) error {
	query := `
		INSERT INTO service (outsourcing_contract_id, name, description, value, frequency, observations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		service.OutsourcingContractID,
		service.Name,
		service.Description,
		service.Value,
		service.Frequency,
		service.Observations,
		service.CreatedAt,
		service.UpdatedAt,
	).Scan(&service.ID)

	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	return nil
}

// GetByID retrieves a service by ID
func (r *ServicePostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Service, error) {
	query := `
		SELECT id, outsourcing_contract_id, name, description, value, frequency, observations, created_at, updated_at
		FROM service
		WHERE id = $1`

	var service domain.Service
	var description sql.NullString

	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&service.ID,
		&service.OutsourcingContractID,
		&service.Name,
		&description,
		&service.Value,
		&service.Frequency,
		&service.Observations,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		service.Description = description.String
	}

	return &service, nil
}

// GetByContractID retrieves all services for an outsourcing contract
func (r *ServicePostgresRepository) GetByContractID(ctx context.Context, contractID int64) ([]*domain.Service, error) {
	query := `
		SELECT id, outsourcing_contract_id, name, description, value, frequency, observations, created_at, updated_at
		FROM service
		WHERE outsourcing_contract_id = $1
		ORDER BY name`

	rows, err := r.db.DB.QueryContext(ctx, query, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var services []*domain.Service
	for rows.Next() {
		var service domain.Service
		var description sql.NullString

		err := rows.Scan(
			&service.ID,
			&service.OutsourcingContractID,
			&service.Name,
			&description,
			&service.Value,
			&service.Frequency,
			&service.Observations,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			service.Description = description.String
		}

		services = append(services, &service)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate services: %w", err)
	}

	return services, nil
}

// GetByFrequency retrieves services by frequency
func (r *ServicePostgresRepository) GetByFrequency(ctx context.Context, contractID int64, frequency domain.ServiceFrequency) ([]*domain.Service, error) {
	query := `
		SELECT id, outsourcing_contract_id, name, description, value, frequency, observations, created_at, updated_at
		FROM service
		WHERE outsourcing_contract_id = $1 AND frequency = $2
		ORDER BY name`

	rows, err := r.db.DB.QueryContext(ctx, query, contractID, frequency)
	if err != nil {
		return nil, fmt.Errorf("failed to query services by frequency: %w", err)
	}
	defer rows.Close()

	var services []*domain.Service
	for rows.Next() {
		var service domain.Service
		var description sql.NullString

		err := rows.Scan(
			&service.ID,
			&service.OutsourcingContractID,
			&service.Name,
			&description,
			&service.Value,
			&service.Frequency,
			&service.Observations,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			service.Description = description.String
		}

		services = append(services, &service)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate services: %w", err)
	}

	return services, nil
}

// Update updates an existing service
func (r *ServicePostgresRepository) Update(ctx context.Context, service *domain.Service) error {
	query := `
		UPDATE service
		SET name = $1, description = $2, value = $3, frequency = $4, observations = $5, updated_at = $6
		WHERE id = $7`

	result, err := r.db.DB.ExecContext(
		ctx,
		query,
		service.Name,
		service.Description,
		service.Value,
		service.Frequency,
		service.Observations,
		service.UpdatedAt,
		service.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service not found")
	}

	return nil
}

// Delete deletes a service
func (r *ServicePostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM service WHERE id = $1`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service not found")
	}

	return nil
}
