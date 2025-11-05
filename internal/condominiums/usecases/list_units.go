package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// ListUnits implements the list units use case
type ListUnits struct {
	unitRepo ports.UnitRepository
	logger   logger.Logger
}

// NewListUnits creates a new ListUnits use case
func NewListUnits(unitRepo ports.UnitRepository, logger logger.Logger) *ListUnits {
	return &ListUnits{
		unitRepo: unitRepo,
		logger:   logger,
	}
}

// Execute lists all units for a condominium
func (uc *ListUnits) Execute(ctx context.Context, condominiumID int64) error {
	uc.logger.Info("Listing units", "condominium_id", condominiumID)

	// TODO: Implement list units logic
	return fmt.Errorf("not implemented yet")
}
