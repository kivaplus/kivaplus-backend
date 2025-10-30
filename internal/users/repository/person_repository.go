package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
)

type PersonData struct {
	userID        sql.NullInt64
	addressID     sql.NullInt64
	phoneNumber   sql.NullString
	email         sql.NullString
	occupation    sql.NullString
	observations  sql.NullString
	maritalStatus sql.NullString
	birthdayDate  sql.NullTime
	deletedAt     sql.NullTime
}

// PersonPostgresRepository implements the PersonRepository interface using PostgreSQL
type PersonPostgresRepository struct {
	db *database.Connection
}

// NewPersonPostgresRepository creates a new PostgreSQL person repository
func NewPersonPostgresRepository(db *database.Connection) *PersonPostgresRepository {
	return &PersonPostgresRepository{
		db: db,
	}
}

// Create creates a new person in the database
func (r *PersonPostgresRepository) Create(ctx context.Context, person *domain.Person) error {
	query := `
		INSERT INTO person (name, person_type, phone_number, email, address_id, occupation,
		marital_status, birthday, user_id, observation, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		person.Name,
		person.PersonType,
		person.PhoneNumber,
		person.Email,
		person.AddressID,
		person.Occupation,
		person.MaritalStatus,
		person.BirthdayDate,
		person.UserID,
		person.Observations,
		person.CreatedAt,
		person.UpdatedAt,
	).Scan(&person.ID)

	if err != nil {
		return fmt.Errorf("failed to create person: %w", err)
	}

	return nil
}

// GetByID retrieves a person by ID
func (r *PersonPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Person, error) {
	query := `
		SELECT id, name, person_type, phone_number, email, address_id, occupation, marital_status,
		 birthday, user_id, observations, created_at, updated_at, deleted_at
		FROM person
		WHERE id = $1 AND deleted_at IS NULL`

	person := &domain.Person{}
	var phoneNumber, email, occupation, observations sql.NullString
	var maritalStatus sql.NullString
	var birthdayDate, deletedAt sql.NullTime
	var userID, addressID sql.NullInt64

	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&person.ID,
		&person.Name,
		&person.PersonType,
		&phoneNumber,
		&email,
		&addressID,
		&occupation,
		&maritalStatus,
		&birthdayDate,
		&userID,
		&observations,
		&person.CreatedAt,
		&person.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person not found")
		}
		return nil, fmt.Errorf("failed to get person by ID: %w", err)
	}

	validateFields(person, PersonData{
		userID:        userID,
		addressID:     addressID,
		phoneNumber:   phoneNumber,
		email:         email,
		occupation:    occupation,
		observations:  observations,
		maritalStatus: maritalStatus,
		birthdayDate:  birthdayDate,
		deletedAt:     deletedAt,
	})

	return person, nil
}

// HasCompleteProfile verify if user has complete profile
// A complete profile requires: name, phone_number, and at least one additional field (occupation or marital_status)
func (r *PersonPostgresRepository) HasCompleteProfile(ctx context.Context, userID int64) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM person
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND name IS NOT NULL AND name != ''
		  AND phone_number IS NOT NULL AND phone_number != ''
		  AND (
		    (occupation IS NOT NULL AND occupation != '') OR
		    (marital_status IS NOT NULL AND marital_status != '') OR
		    (birthday IS NOT NULL)
		  )`

	var count int
	err := r.db.DB.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check profile completeness: %w", err)
	}

	return count > 0, nil
}

// GetByUserID retrieves a person by user ID
func (r *PersonPostgresRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Person, error) {
	query := `
		SELECT id, name, person_type, phone_number, email, address_id, occupation, marital_status, birthday, user_id, observations, created_at, updated_at, deleted_at
		FROM person
		WHERE user_id = $1 AND deleted_at IS NULL`

	person := &domain.Person{}
	var phoneNumber, email, occupation, observations sql.NullString
	var maritalStatus sql.NullString
	var birthdayDate, deletedAt sql.NullTime
	var userId, addressID sql.NullInt64

	err := r.db.DB.QueryRowContext(ctx, query, userID).Scan(
		&person.ID,
		&person.Name,
		&person.PersonType,
		&phoneNumber,
		&email,
		&addressID,
		&occupation,
		&maritalStatus,
		&birthdayDate,
		&userId,
		&observations,
		&person.CreatedAt,
		&person.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person not found")
		}
		return nil, fmt.Errorf("failed to get person by user ID: %w", err)
	}

	validateFields(person, PersonData{
		userID:        userId,
		addressID:     addressID,
		phoneNumber:   phoneNumber,
		email:         email,
		occupation:    occupation,
		observations:  observations,
		maritalStatus: maritalStatus,
		birthdayDate:  birthdayDate,
		deletedAt:     deletedAt,
	})

	return person, nil
}

// Update updates an existing person
func (r *PersonPostgresRepository) Update(ctx context.Context, person *domain.Person) error {
	query := `
		UPDATE person
		SET name = $2, person_type = $3, phone_number = $4, email = $5, address_id = $6, occupation = $7, marital_status = $8, birthday = $9, user_id = $10, observations = $11, updated_at = $12
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.DB.ExecContext(
		ctx,
		query,
		person.ID,
		person.Name,
		person.PersonType,
		person.PhoneNumber,
		person.Email,
		person.AddressID,
		person.Occupation,
		person.MaritalStatus,
		person.BirthdayDate,
		person.UserID,
		person.Observations,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update person: %w", err)
	}

	return nil
}

// Delete soft deletes a person
func (r *PersonPostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE person SET deleted_at = $2, updated_at = $2 WHERE id = $1`

	_, err := r.db.DB.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}

	return nil
}

func validateFields(person *domain.Person, data PersonData) {
	// Handle nullable fields
	if data.addressID.Valid {
		person.AddressID = &data.addressID.Int64
	}
	if data.phoneNumber.Valid {
		person.PhoneNumber = data.phoneNumber.String
	}
	if data.email.Valid {
		person.Email = data.email.String
	}
	if data.occupation.Valid {
		person.Occupation = data.occupation.String
	}
	if data.observations.Valid {
		person.Observations = data.observations.String
	}
	if data.maritalStatus.Valid {
		ec := domain.MaritalStatus(data.maritalStatus.String)
		person.MaritalStatus = ec
	}
	if data.birthdayDate.Valid {
		person.BirthdayDate = data.birthdayDate.Time
	}
	if data.userID.Valid {
		person.UserID = &data.userID.Int64
	}
	if data.deletedAt.Valid {
		person.DeletedAt = &data.deletedAt.Time
	}
}
