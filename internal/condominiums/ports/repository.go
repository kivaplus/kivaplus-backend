package ports

import (
	"context"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
)

// CondominiumRepository defines the interface for condominium data operations
type CondominiumRepository interface {
	// Create creates a new condominium
	Create(ctx context.Context, condominium *domain.Condominium) error

	// GetByID retrieves a condominium by ID
	GetByID(ctx context.Context, id int64) (*domain.Condominium, error)

	// GetByDocumentNumber retrieves a condominium by document number
	GetByDocumentNumber(ctx context.Context, documentNumber string) (*domain.Condominium, error)

	// GetByUserID retrieves all condominiums where the user has any role
	GetByUserID(ctx context.Context, userID int64) ([]*domain.Condominium, error)

	// GetAll retrieves all condominiums (for super admin)
	GetAll(ctx context.Context) ([]*domain.Condominium, error)

	// Update updates an existing condominium
	Update(ctx context.Context, condominium *domain.Condominium) error

	// Delete soft deletes a condominium
	Delete(ctx context.Context, id int64) error

	// Exists checks if a condominium exists by document number
	Exists(ctx context.Context, documentNumber string) (bool, error)
}

// UnitRepository defines the interface for unit data operations
type UnitRepository interface {
	// Create creates a new unit
	Create(ctx context.Context, unit *domain.Unit) error

	// GetByID retrieves a unit by ID
	GetByID(ctx context.Context, id int64) (*domain.Unit, error)

	// GetByCondominiumID retrieves all units for a condominium
	GetByCondominiumID(ctx context.Context, condominiumID int64) ([]*domain.Unit, error)

	// Update updates an existing unit
	Update(ctx context.Context, unit *domain.Unit) error

	// Delete soft deletes a unit
	Delete(ctx context.Context, id int64) error
}

// VehicleRepository defines the interface for vehicle data operations
type VehicleRepository interface {
	// Create creates a new vehicle
	Create(ctx context.Context, vehicle *domain.Vehicle) error

	// GetByID retrieves a vehicle by ID
	GetByID(ctx context.Context, id int64) (*domain.Vehicle, error)

	// GetByPlateNumber retrieves a vehicle by plate number
	GetByPlateNumber(ctx context.Context, plateNumber string) (*domain.Vehicle, error)

	// GetByCondominiumID retrieves all vehicles for a condominium
	GetByCondominiumID(ctx context.Context, condominiumID int64) ([]*domain.Vehicle, error)

	// Update updates an existing vehicle
	Update(ctx context.Context, vehicle *domain.Vehicle) error

	// Delete soft deletes a vehicle
	Delete(ctx context.Context, id int64) error

	// Exists checks if a vehicle exists by plate number
	Exists(ctx context.Context, plateNumber string) (bool, error)
}

// ParkingSpaceRepository defines the interface for parking space data operations
type ParkingSpaceRepository interface {
	// Create creates a new parking space
	Create(ctx context.Context, parkingSpace *domain.ParkingSpace) error

	// GetByID retrieves a parking space by ID
	GetByID(ctx context.Context, id int64) (*domain.ParkingSpace, error)

	// GetByUnitID retrieves all parking spaces for a unit
	GetByUnitID(ctx context.Context, unitID int64) ([]*domain.ParkingSpace, error)

	// GetByCondominiumID retrieves all parking spaces for a condominium
	GetByCondominiumID(ctx context.Context, condominiumID int64) ([]*domain.ParkingSpace, error)

	// Update updates an existing parking space
	Update(ctx context.Context, parkingSpace *domain.ParkingSpace) error

	// Delete deletes a parking space
	Delete(ctx context.Context, id int64) error
}

// ResidentRepository defines the interface for resident data operations
type ResidentRepository interface {
	// Create creates a new resident
	Create(ctx context.Context, resident *domain.Resident) error

	// GetByID retrieves a resident by ID
	GetByID(ctx context.Context, id int64) (*domain.Resident, error)

	// GetByUnitID retrieves all residents for a unit
	GetByUnitID(ctx context.Context, unitID int64) ([]*domain.Resident, error)

	// GetByCondominiumID retrieves all residents for a condominium
	GetByCondominiumID(ctx context.Context, condominiumID int64) ([]*domain.Resident, error)

	// Update updates an existing resident
	Update(ctx context.Context, resident *domain.Resident) error

	// Delete soft deletes a resident
	Delete(ctx context.Context, id int64) error
}

// EmployeeRepository defines the interface for employee data operations
type EmployeeRepository interface {
	// Create creates a new employee
	Create(ctx context.Context, employee *domain.Employee) error

	// GetByID retrieves an employee by ID
	GetByID(ctx context.Context, id int64) (*domain.Employee, error)

	// GetByPersonID retrieves an employee by person ID
	GetByPersonID(ctx context.Context, personID int64) (*domain.Employee, error)

	// GetByCondominiumID retrieves all employees for a condominium
	GetByCondominiumID(ctx context.Context, condominiumID int64) ([]*domain.Employee, error)

	// GetByStatus retrieves employees by status
	GetByStatus(ctx context.Context, condominiumID int64, status domain.EmployeeStatus) ([]*domain.Employee, error)

	// Update updates an existing employee
	Update(ctx context.Context, employee *domain.Employee) error

	// Delete soft deletes an employee
	Delete(ctx context.Context, id int64) error

	// Exists checks if an employee exists for a person in a condominium
	Exists(ctx context.Context, personID, condominiumID int64) (bool, error)
}

// OutsourcingCompanyRepository defines the interface for outsourcing company data operations
type OutsourcingCompanyRepository interface {
	// Create creates a new outsourcing company
	Create(ctx context.Context, company *domain.OutsourcingCompany) error

	// GetByID retrieves an outsourcing company by ID
	GetByID(ctx context.Context, id int64) (*domain.OutsourcingCompany, error)

	// GetByDocumentNumber retrieves an outsourcing company by document number
	GetByDocumentNumber(ctx context.Context, documentNumber string) (*domain.OutsourcingCompany, error)

	// GetAll retrieves all outsourcing companies
	GetAll(ctx context.Context) ([]*domain.OutsourcingCompany, error)

	// Update updates an existing outsourcing company
	Update(ctx context.Context, company *domain.OutsourcingCompany) error

	// Delete soft deletes an outsourcing company
	Delete(ctx context.Context, id int64) error

	// Exists checks if an outsourcing company exists by document number
	Exists(ctx context.Context, documentNumber string) (bool, error)
}

// OutsourcingContractRepository defines the interface for outsourcing contract data operations
type OutsourcingContractRepository interface {
	// Create creates a new outsourcing contract
	Create(ctx context.Context, contract *domain.OutsourcingContract) error

	// GetByID retrieves an outsourcing contract by ID
	GetByID(ctx context.Context, id int64) (*domain.OutsourcingContract, error)

	// GetByContractNumber retrieves an outsourcing contract by contract number
	GetByContractNumber(ctx context.Context, contractNumber string) (*domain.OutsourcingContract, error)

	// GetByCondominiumID retrieves all contracts for a condominium
	GetByCondominiumID(ctx context.Context, condominiumID int64) ([]*domain.OutsourcingContract, error)

	// GetByCompanyID retrieves all contracts for an outsourcing company
	GetByCompanyID(ctx context.Context, companyID int64) ([]*domain.OutsourcingContract, error)

	// GetByStatus retrieves contracts by status
	GetByStatus(ctx context.Context, status domain.ContractStatus) ([]*domain.OutsourcingContract, error)

	// Update updates an existing outsourcing contract
	Update(ctx context.Context, contract *domain.OutsourcingContract) error

	// Delete soft deletes an outsourcing contract
	Delete(ctx context.Context, id int64) error

	// Exists checks if a contract exists by contract number
	Exists(ctx context.Context, contractNumber string) (bool, error)
}

// ServiceRepository defines the interface for service data operations
type ServiceRepository interface {
	// Create creates a new service
	Create(ctx context.Context, service *domain.Service) error

	// GetByID retrieves a service by ID
	GetByID(ctx context.Context, id int64) (*domain.Service, error)

	// GetByContractID retrieves all services for an outsourcing contract
	GetByContractID(ctx context.Context, contractID int64) ([]*domain.Service, error)

	// GetByFrequency retrieves services by frequency
	GetByFrequency(ctx context.Context, contractID int64, frequency domain.ServiceFrequency) ([]*domain.Service, error)

	// Update updates an existing service
	Update(ctx context.Context, service *domain.Service) error

	// Delete deletes a service
	Delete(ctx context.Context, id int64) error
}

// ResidentVehicleRepository defines the interface for resident-vehicle relationship data operations
type ResidentVehicleRepository interface {
	// Create creates a new resident-vehicle relationship
	Create(ctx context.Context, residentVehicle *domain.ResidentVehicle) error

	// GetByID retrieves a resident-vehicle relationship by ID
	GetByID(ctx context.Context, id int64) (*domain.ResidentVehicle, error)

	// GetByResidentID retrieves all vehicles for a resident
	GetByResidentID(ctx context.Context, residentID int64) ([]*domain.ResidentVehicle, error)

	// GetByVehicleID retrieves all residents for a vehicle
	GetByVehicleID(ctx context.Context, vehicleID int64) ([]*domain.ResidentVehicle, error)

	// GetMainVehicle retrieves the main vehicle for a resident
	GetMainVehicle(ctx context.Context, residentID int64) (*domain.ResidentVehicle, error)

	// SetMainVehicle sets a vehicle as the main vehicle for a resident
	SetMainVehicle(ctx context.Context, residentID, vehicleID int64) error

	// Delete deletes a resident-vehicle relationship
	Delete(ctx context.Context, id int64) error

	// DeleteByResidentAndVehicle deletes a relationship by resident and vehicle IDs
	DeleteByResidentAndVehicle(ctx context.Context, residentID, vehicleID int64) error

	// Exists checks if a resident-vehicle relationship exists
	Exists(ctx context.Context, residentID, vehicleID int64) (bool, error)
}
