package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
)

// EmployeePostgresRepository implements the EmployeeRepository interface using PostgreSQL
type EmployeePostgresRepository struct {
	db *database.Connection
}

// NewEmployeePostgresRepository creates a new PostgreSQL employee repository
func NewEmployeePostgresRepository(db *database.Connection) *EmployeePostgresRepository {
	return &EmployeePostgresRepository{
		db: db,
	}
}

// Create creates a new employee in the database
func (r *EmployeePostgresRepository) Create(ctx context.Context, employee *domain.Employee) error {
	query := `
		INSERT INTO employee (person_id, condominium_id, occupation, salary, start_date, end_date, status, observations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		employee.PersonID,
		employee.CondominiumID,
		employee.Occupation,
		employee.Salary,
		employee.StartDate,
		employee.EndDate,
		employee.Status,
		employee.Observations,
		employee.CreatedAt,
		employee.UpdatedAt,
	).Scan(&employee.ID)

	if err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}

	return nil
}

// GetByID retrieves an employee by ID
func (r *EmployeePostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Employee, error) {
	query := `
		SELECT id, person_id, condominium_id, occupation, salary, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM employee
		WHERE id = $1 AND deleted_at IS NULL`

	var employee domain.Employee
	var salary sql.NullFloat64
	var endDate sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&employee.ID,
		&employee.PersonID,
		&employee.CondominiumID,
		&employee.Occupation,
		&salary,
		&employee.StartDate,
		&endDate,
		&employee.Status,
		&employee.Observations,
		&employee.CreatedAt,
		&employee.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	// Handle nullable fields
	if salary.Valid {
		employee.Salary = &salary.Float64
	}
	if endDate.Valid {
		employee.EndDate = &endDate.Time
	}
	if deletedAt.Valid {
		employee.DeletedAt = &deletedAt.Time
	}

	return &employee, nil
}

// GetByPersonID retrieves an employee by person ID
func (r *EmployeePostgresRepository) GetByPersonID(ctx context.Context, personID int64) (*domain.Employee, error) {
	query := `
		SELECT id, person_id, condominium_id, occupation, salary, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM employee
		WHERE person_id = $1 AND deleted_at IS NULL`

	var employee domain.Employee
	var salary sql.NullFloat64
	var endDate sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.DB.QueryRowContext(ctx, query, personID).Scan(
		&employee.ID,
		&employee.PersonID,
		&employee.CondominiumID,
		&employee.Occupation,
		&salary,
		&employee.StartDate,
		&endDate,
		&employee.Status,
		&employee.Observations,
		&employee.CreatedAt,
		&employee.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	// Handle nullable fields
	if salary.Valid {
		employee.Salary = &salary.Float64
	}
	if endDate.Valid {
		employee.EndDate = &endDate.Time
	}
	if deletedAt.Valid {
		employee.DeletedAt = &deletedAt.Time
	}

	return &employee, nil
}

// GetByCondominiumID retrieves all employees for a condominium
func (r *EmployeePostgresRepository) GetByCondominiumID(ctx context.Context, condominiumID int64) ([]*domain.Employee, error) {
	query := `
		SELECT id, person_id, condominium_id, occupation, salary, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM employee
		WHERE condominium_id = $1 AND deleted_at IS NULL
		ORDER BY start_date DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, condominiumID)
	if err != nil {
		return nil, fmt.Errorf("failed to query employees: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		var employee domain.Employee
		var salary sql.NullFloat64
		var endDate sql.NullTime
		var deletedAt sql.NullTime

		err := rows.Scan(
			&employee.ID,
			&employee.PersonID,
			&employee.CondominiumID,
			&employee.Occupation,
			&salary,
			&employee.StartDate,
			&endDate,
			&employee.Status,
			&employee.Observations,
			&employee.CreatedAt,
			&employee.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}

		// Handle nullable fields
		if salary.Valid {
			employee.Salary = &salary.Float64
		}
		if endDate.Valid {
			employee.EndDate = &endDate.Time
		}
		if deletedAt.Valid {
			employee.DeletedAt = &deletedAt.Time
		}

		employees = append(employees, &employee)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate employees: %w", err)
	}

	return employees, nil
}

// GetByStatus retrieves employees by status
func (r *EmployeePostgresRepository) GetByStatus(ctx context.Context, condominiumID int64, status domain.EmployeeStatus) ([]*domain.Employee, error) {
	query := `
		SELECT id, person_id, condominium_id, occupation, salary, start_date, end_date, status, observations, created_at, updated_at, deleted_at
		FROM employee
		WHERE condominium_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY start_date DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, condominiumID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query employees by status: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		var employee domain.Employee
		var salary sql.NullFloat64
		var endDate sql.NullTime
		var deletedAt sql.NullTime

		err := rows.Scan(
			&employee.ID,
			&employee.PersonID,
			&employee.CondominiumID,
			&employee.Occupation,
			&salary,
			&employee.StartDate,
			&endDate,
			&employee.Status,
			&employee.Observations,
			&employee.CreatedAt,
			&employee.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}

		// Handle nullable fields
		if salary.Valid {
			employee.Salary = &salary.Float64
		}
		if endDate.Valid {
			employee.EndDate = &endDate.Time
		}
		if deletedAt.Valid {
			employee.DeletedAt = &deletedAt.Time
		}

		employees = append(employees, &employee)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate employees: %w", err)
	}

	return employees, nil
}

// Update updates an existing employee
func (r *EmployeePostgresRepository) Update(ctx context.Context, employee *domain.Employee) error {
	query := `
		UPDATE employee
		SET occupation = $1, salary = $2, start_date = $3, end_date = $4, status = $5, observations = $6, updated_at = $7
		WHERE id = $8 AND deleted_at IS NULL`

	result, err := r.db.DB.ExecContext(
		ctx,
		query,
		employee.Occupation,
		employee.Salary,
		employee.StartDate,
		employee.EndDate,
		employee.Status,
		employee.Observations,
		employee.UpdatedAt,
		employee.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update employee: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("employee not found")
	}

	return nil
}

// Delete soft deletes an employee
func (r *EmployeePostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE employee SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("employee not found")
	}

	return nil
}

// Exists checks if an employee exists for a person in a condominium
func (r *EmployeePostgresRepository) Exists(ctx context.Context, personID, condominiumID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM employee WHERE person_id = $1 AND condominium_id = $2 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.DB.QueryRowContext(ctx, query, personID, condominiumID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check employee existence: %w", err)
	}

	return exists, nil
}
