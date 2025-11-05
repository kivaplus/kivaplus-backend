package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// ListParkingSpaces implements the list parking spaces use case
type ListParkingSpaces struct {
	parkingSpaceRepo ports.ParkingSpaceRepository
	logger           logger.Logger
}

// NewListParkingSpaces creates a new ListParkingSpaces use case
func NewListParkingSpaces(parkingSpaceRepo ports.ParkingSpaceRepository, logger logger.Logger) *ListParkingSpaces {
	return &ListParkingSpaces{
		parkingSpaceRepo: parkingSpaceRepo,
		logger:           logger,
	}
}

// Execute lists all parking spaces for a condominium
func (uc *ListParkingSpaces) Execute(ctx context.Context, condominiumID int64) error {
	uc.logger.Info("Listing parking spaces", "condominium_id", condominiumID)

	// TODO: Implement list parking spaces logic
	return fmt.Errorf("not implemented yet")
}
