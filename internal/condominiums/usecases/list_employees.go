package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/concurrent"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	userPorts "github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// ListEmployees implements the list employees use case
type ListEmployees struct {
	employeeRepo ports.EmployeeRepository
	personRepo   userPorts.PersonRepository
	logger       logger.Logger
}

// NewListEmployees creates a new ListEmployees use case
func NewListEmployees(
	employeeRepo ports.EmployeeRepository,
	personRepo userPorts.PersonRepository,
	logger logger.Logger,
) *ListEmployees {
	return &ListEmployees{
		employeeRepo: employeeRepo,
		personRepo:   personRepo,
		logger:       logger,
	}
}

// ListEmployeesResponse represents the response for listing employees
type ListEmployeesResponse struct {
	Employees []*EmployeeData `json:"employees"`
	Message   string          `json:"message"`
}

// Execute lists all employees for a condominium with concurrent person data fetching
func (uc *ListEmployees) Execute(ctx context.Context, condominiumID int64, status *domain.EmployeeStatus) (*ListEmployeesResponse, error) {
	uc.logger.Info("Listing employees", "condominium_id", condominiumID, "status", status)

	if condominiumID <= 0 {
		return nil, fmt.Errorf("condominium ID must be greater than 0")
	}

	// Get employees based on status filter
	var employees []*domain.Employee
	var err error

	if status != nil {
		employees, err = uc.employeeRepo.GetByStatus(ctx, condominiumID, *status)
	} else {
		employees, err = uc.employeeRepo.GetByCondominiumID(ctx, condominiumID)
	}

	if err != nil {
		uc.logger.Error("Failed to get employees", "error", err, "condominium_id", condominiumID)
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	if len(employees) == 0 {
		return &ListEmployeesResponse{
			Employees: []*EmployeeData{},
			Message:   "No employees found.",
		}, nil
	}

	// Use concurrent processing to fetch person data for each employee
	employeeData, errors := concurrent.ProcessBatchGeneric(ctx, employees, func(ctx context.Context, employee *domain.Employee) (*EmployeeData, error) {
		// Get person data
		person, err := uc.personRepo.GetByID(ctx, employee.PersonID)
		if err != nil {
			uc.logger.Error("Failed to get person for employee", "error", err, "person_id", employee.PersonID)
			// Don't fail the entire operation, just skip person data
			person = nil
		}

		// Format employee data
		data := &EmployeeData{
			ID:            employee.ID,
			PersonID:      employee.PersonID,
			CondominiumID: employee.CondominiumID,
			Occupation:    employee.Occupation,
			Salary:        employee.Salary,
			StartDate:     employee.StartDate.Format("2006-01-02"),
			Status:        employee.Status,
			Observations:  employee.Observations,
		}

		if employee.EndDate != nil {
			endDate := employee.EndDate.Format("2006-01-02")
			data.EndDate = &endDate
		}

		if person != nil {
			data.Person = &PersonData{
				ID:          person.ID,
				Name:        person.Name,
				Email:       person.Email,
				PhoneNumber: person.PhoneNumber,
			}
		}

		return data, nil
	}, 5) // Use 5 concurrent workers

	// Log any errors but don't fail the operation
	if len(errors) > 0 {
		uc.logger.Warn("Some errors occurred while processing employees", "error_count", len(errors))
		for _, err := range errors {
			if err != nil {
				uc.logger.Error("Employee processing error", "error", err)
			}
		}
	}

	message := fmt.Sprintf("Found %d employees.", len(employeeData))
	if status != nil {
		message = fmt.Sprintf("Found %d employees with status '%s'.", len(employeeData), *status)
	}

	return &ListEmployeesResponse{
		Employees: employeeData,
		Message:   message,
	}, nil
}
