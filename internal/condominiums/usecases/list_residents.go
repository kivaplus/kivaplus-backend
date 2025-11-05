package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// ListResidents implements the list residents use case
type ListResidents struct {
	residentRepo ports.ResidentRepository
	logger       logger.Logger
}

// NewListResidents creates a new ListResidents use case
func NewListResidents(residentRepo ports.ResidentRepository, logger logger.Logger) *ListResidents {
	return &ListResidents{
		residentRepo: residentRepo,
		logger:       logger,
	}
}

// Execute lists all residents for a condominium
func (uc *ListResidents) Execute(ctx context.Context, condominiumID int64) error {
	uc.logger.Info("Listing residents", "condominium_id", condominiumID)

	// TODO: Implement list residents logic
	return fmt.Errorf("not implemented yet")
}
