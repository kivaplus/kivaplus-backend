package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// ListVehicles implements the list vehicles use case
type ListVehicles struct {
	vehicleRepo ports.VehicleRepository
	logger      logger.Logger
}

// NewListVehicles creates a new ListVehicles use case
func NewListVehicles(vehicleRepo ports.VehicleRepository, logger logger.Logger) *ListVehicles {
	return &ListVehicles{
		vehicleRepo: vehicleRepo,
		logger:      logger,
	}
}

// Execute lists all vehicles for a condominium
func (uc *ListVehicles) Execute(ctx context.Context, condominiumID int64) error {
	uc.logger.Info("Listing vehicles", "condominium_id", condominiumID)

	// TODO: Implement list vehicles logic
	return fmt.Errorf("not implemented yet")
}
