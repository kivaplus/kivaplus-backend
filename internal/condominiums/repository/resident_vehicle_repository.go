package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
)

// ResidentVehiclePostgresRepository implements the ResidentVehicleRepository interface using PostgreSQL
type ResidentVehiclePostgresRepository struct {
	db *database.Connection
}

// NewResidentVehiclePostgresRepository creates a new PostgreSQL resident vehicle repository
func NewResidentVehiclePostgresRepository(db *database.Connection) *ResidentVehiclePostgresRepository {
	return &ResidentVehiclePostgresRepository{
		db: db,
	}
}

// Create creates a new resident-vehicle relationship in the database
func (r *ResidentVehiclePostgresRepository) Create(ctx context.Context, residentVehicle *domain.ResidentVehicle) error {
	query := `
		INSERT INTO resident_vehicle (resident_id, vehicle_id, main_vehicle, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		residentVehicle.ResidentID,
		residentVehicle.VehicleID,
		residentVehicle.MainVehicle,
		residentVehicle.CreatedAt,
	).Scan(&residentVehicle.ID)

	if err != nil {
		return fmt.Errorf("failed to create resident vehicle relationship: %w", err)
	}

	return nil
}

// GetByID retrieves a resident-vehicle relationship by ID
func (r *ResidentVehiclePostgresRepository) GetByID(ctx context.Context, id int64) (*domain.ResidentVehicle, error) {
	query := `
		SELECT id, resident_id, vehicle_id, main_vehicle, created_at
		FROM resident_vehicle
		WHERE id = $1`

	var residentVehicle domain.ResidentVehicle

	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&residentVehicle.ID,
		&residentVehicle.ResidentID,
		&residentVehicle.VehicleID,
		&residentVehicle.MainVehicle,
		&residentVehicle.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident vehicle relationship not found")
		}
		return nil, fmt.Errorf("failed to get resident vehicle relationship: %w", err)
	}

	return &residentVehicle, nil
}

// GetByResidentID retrieves all vehicles for a resident
func (r *ResidentVehiclePostgresRepository) GetByResidentID(ctx context.Context, residentID int64) ([]*domain.ResidentVehicle, error) {
	query := `
		SELECT id, resident_id, vehicle_id, main_vehicle, created_at
		FROM resident_vehicle
		WHERE resident_id = $1
		ORDER BY main_vehicle DESC, created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, residentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query resident vehicles: %w", err)
	}
	defer rows.Close()

	var residentVehicles []*domain.ResidentVehicle
	for rows.Next() {
		var residentVehicle domain.ResidentVehicle

		err := rows.Scan(
			&residentVehicle.ID,
			&residentVehicle.ResidentID,
			&residentVehicle.VehicleID,
			&residentVehicle.MainVehicle,
			&residentVehicle.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resident vehicle: %w", err)
		}

		residentVehicles = append(residentVehicles, &residentVehicle)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate resident vehicles: %w", err)
	}

	return residentVehicles, nil
}

// GetByVehicleID retrieves all residents for a vehicle
func (r *ResidentVehiclePostgresRepository) GetByVehicleID(ctx context.Context, vehicleID int64) ([]*domain.ResidentVehicle, error) {
	query := `
		SELECT id, resident_id, vehicle_id, main_vehicle, created_at
		FROM resident_vehicle
		WHERE vehicle_id = $1
		ORDER BY main_vehicle DESC, created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query vehicle residents: %w", err)
	}
	defer rows.Close()

	var residentVehicles []*domain.ResidentVehicle
	for rows.Next() {
		var residentVehicle domain.ResidentVehicle

		err := rows.Scan(
			&residentVehicle.ID,
			&residentVehicle.ResidentID,
			&residentVehicle.VehicleID,
			&residentVehicle.MainVehicle,
			&residentVehicle.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resident vehicle: %w", err)
		}

		residentVehicles = append(residentVehicles, &residentVehicle)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate resident vehicles: %w", err)
	}

	return residentVehicles, nil
}

// GetMainVehicle retrieves the main vehicle for a resident
func (r *ResidentVehiclePostgresRepository) GetMainVehicle(ctx context.Context, residentID int64) (*domain.ResidentVehicle, error) {
	query := `
		SELECT id, resident_id, vehicle_id, main_vehicle, created_at
		FROM resident_vehicle
		WHERE resident_id = $1 AND main_vehicle = true`

	var residentVehicle domain.ResidentVehicle

	err := r.db.DB.QueryRowContext(ctx, query, residentID).Scan(
		&residentVehicle.ID,
		&residentVehicle.ResidentID,
		&residentVehicle.VehicleID,
		&residentVehicle.MainVehicle,
		&residentVehicle.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("main vehicle not found for resident")
		}
		return nil, fmt.Errorf("failed to get main vehicle: %w", err)
	}

	return &residentVehicle, nil
}

// SetMainVehicle sets a vehicle as the main vehicle for a resident
func (r *ResidentVehiclePostgresRepository) SetMainVehicle(ctx context.Context, residentID, vehicleID int64) error {
	// Start a transaction to ensure atomicity
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, unset all main vehicles for this resident
	unsetQuery := `UPDATE resident_vehicle SET main_vehicle = false WHERE resident_id = $1`
	_, err = tx.ExecContext(ctx, unsetQuery, residentID)
	if err != nil {
		return fmt.Errorf("failed to unset main vehicles: %w", err)
	}

	// Then, set the specified vehicle as main
	setQuery := `UPDATE resident_vehicle SET main_vehicle = true WHERE resident_id = $1 AND vehicle_id = $2`
	result, err := tx.ExecContext(ctx, setQuery, residentID, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to set main vehicle: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("resident vehicle relationship not found")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete deletes a resident-vehicle relationship
func (r *ResidentVehiclePostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM resident_vehicle WHERE id = $1`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete resident vehicle relationship: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("resident vehicle relationship not found")
	}

	return nil
}

// DeleteByResidentAndVehicle deletes a relationship by resident and vehicle IDs
func (r *ResidentVehiclePostgresRepository) DeleteByResidentAndVehicle(ctx context.Context, residentID, vehicleID int64) error {
	query := `DELETE FROM resident_vehicle WHERE resident_id = $1 AND vehicle_id = $2`

	result, err := r.db.DB.ExecContext(ctx, query, residentID, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to delete resident vehicle relationship: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("resident vehicle relationship not found")
	}

	return nil
}

// Exists checks if a resident-vehicle relationship exists
func (r *ResidentVehiclePostgresRepository) Exists(ctx context.Context, residentID, vehicleID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM resident_vehicle WHERE resident_id = $1 AND vehicle_id = $2)`

	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, residentID, vehicleID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check resident vehicle relationship existence: %w", err)
	}

	return exists, nil
}
