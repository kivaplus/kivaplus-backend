package repository

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
)

// NewPersonPostgresRepository creates a new PostgreSQL person repository
func NewAddressPostgresRepository(db *database.Connection) *AddressPostgresRepository {
	return &AddressPostgresRepository{
		db: db,
	}
}

func (r *AddressPostgresRepository) Create(ctx context.Context, address *domain.Address) (*domain.Address, error) {
	query := `
		INSERT INTO address (zip_code, street, num, complement, neighborhood, city,
		state, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err := r.db.DB.QueryRowContext(
		ctx,
		query,
		address.ZipCode,
		address.Street,
		address.Number,
		address.Complement,
		address.Neighborhood,
		address.City,
		address.State,
		address.Country,
		address.CreatedAt,
		address.UpdatedAt,
	).Scan(&address.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create person: %w", err)
	}

	return address, nil
}

// PersonPostgresRepository implements the PersonRepository interface using PostgreSQL
type AddressPostgresRepository struct {
	db *database.Connection
}

// Delete implements ports.AddressRepository.
func (r *AddressPostgresRepository) Delete(ctx context.Context, id int64) error {
	panic("unimplemented")
}

// GetByID implements ports.AddressRepository.
func (r *AddressPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Address, error) {
	query := `
		SELECT id, zip_code, street, num, complement, neighborhood, city, state, country, created_at, updated_at
		FROM address
		WHERE id = $1`

	address := &domain.Address{}
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&address.ID,
		&address.ZipCode,
		&address.Street,
		&address.Number,
		&address.Complement,
		&address.Neighborhood,
		&address.City,
		&address.State,
		&address.Country,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get address by ID: %w", err)
	}

	return address, nil
}

// Update implements ports.AddressRepository.
func (r *AddressPostgresRepository) Update(ctx context.Context, address *domain.Address) error {
	panic("unimplemented")
}
