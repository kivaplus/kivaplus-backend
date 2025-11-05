package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// GetUnit implements the get unit use case
type GetUnit struct {
	unitRepo ports.UnitRepository
	logger   logger.Logger
}

// NewGetUnit creates a new GetUnit use case
func NewGetUnit(unitRepo ports.UnitRepository, logger logger.Logger) *GetUnit {
	return &GetUnit{
		unitRepo: unitRepo,
		logger:   logger,
	}
}

// Execute gets a unit by ID
func (uc *GetUnit) Execute(ctx context.Context, condominiumID, unitID int64) error {
	uc.logger.Info("Getting unit", "condominium_id", condominiumID, "unit_id", unitID)

	// TODO: Implement get unit logic
	return fmt.Errorf("not implemented yet")
}
