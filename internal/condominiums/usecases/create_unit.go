package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// CreateUnit implements the create unit use case
type CreateUnit struct {
	unitRepo ports.UnitRepository
	logger   logger.Logger
}

// NewCreateUnit creates a new CreateUnit use case
func NewCreateUnit(unitRepo ports.UnitRepository, logger logger.Logger) *CreateUnit {
	return &CreateUnit{
		unitRepo: unitRepo,
		logger:   logger,
	}
}

// Execute creates a new unit
func (uc *CreateUnit) Execute(ctx context.Context, condominiumID int64) error {
	uc.logger.Info("Creating unit", "condominium_id", condominiumID)

	// TODO: Implement create unit logic
	return fmt.Errorf("not implemented yet")
}
