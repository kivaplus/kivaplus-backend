package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/kivaplus/kivaplus-backend/internal/condominiums/domain"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/validator"
	userPorts "github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// CreateEmployee implements the create employee use case
type CreateEmployee struct {
	employeeRepo ports.EmployeeRepository
	personRepo   userPorts.PersonRepository
	logger       logger.Logger
}

// NewCreateEmployee creates a new CreateEmployee use case
func NewCreateEmployee(
	employeeRepo ports.EmployeeRepository,
	personRepo userPorts.PersonRepository,
	logger logger.Logger,
) *CreateEmployee {
	return &CreateEmployee{
		employeeRepo: employeeRepo,
		personRepo:   personRepo,
		logger:       logger,
	}
}

// CreateEmployeeRequest represents the request to create an employee
type CreateEmployeeRequest struct {
	PersonID      int64   `json:"person_id"`
	CondominiumID int64   `json:"condominium_id"`
	Occupation    string  `json:"occupation"`
	Salary        float64 `json:"salary,omitempty"`
	StartDate     string  `json:"start_date"` // Format: YYYY-MM-DD
	Observations  string  `json:"observations,omitempty"`
}

// CreateEmployeeResponse represents the response after creating an employee
type CreateEmployeeResponse struct {
	Employee *EmployeeData `json:"employee"`
	Message  string        `json:"message"`
}

// EmployeeData represents employee data in responses
type EmployeeData struct {
	ID            int64                 `json:"id"`
	PersonID      int64                 `json:"person_id"`
	CondominiumID int64                 `json:"condominium_id"`
	Occupation    string                `json:"occupation"`
	Salary        *float64              `json:"salary,omitempty"`
	StartDate     string                `json:"start_date"`
	EndDate       *string               `json:"end_date,omitempty"`
	Status        domain.EmployeeStatus `json:"status"`
	Observations  string                `json:"observations,omitempty"`
	Person        *PersonData           `json:"person,omitempty"`
}

// PersonData represents person data in employee responses
type PersonData struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// Execute creates a new employee
func (uc *CreateEmployee) Execute(ctx context.Context, req CreateEmployeeRequest) (*CreateEmployeeResponse, error) {
	uc.logger.Info("Creating new employee", "person_id", req.PersonID, "condominium_id", req.CondominiumID)

	// Validate input
	v := validator.New()
	v.ValidateRequired("person_id", fmt.Sprintf("%d", req.PersonID))
	v.ValidateRequired("condominium_id", fmt.Sprintf("%d", req.CondominiumID))
	v.ValidateRequired("occupation", req.Occupation)
	v.ValidateMinLength("occupation", req.Occupation, 2)
	v.ValidateMaxLength("occupation", req.Occupation, 100)
	v.ValidateRequired("start_date", req.StartDate)

	if req.PersonID <= 0 {
		v.AddError("person_id", "must be greater than 0")
	}
	if req.CondominiumID <= 0 {
		v.AddError("condominium_id", "must be greater than 0")
	}
	if req.Salary < 0 {
		v.AddError("salary", "cannot be negative")
	}

	if v.HasErrors() {
		uc.logger.Error("Employee validation failed", "errors", v.Errors())
		return nil, fmt.Errorf("validation failed: %v", v.Errors())
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		uc.logger.Error("Invalid start date format", "start_date", req.StartDate, "error", err)
		return nil, fmt.Errorf("invalid start date format, expected YYYY-MM-DD")
	}

	// Check if person exists
	person, err := uc.personRepo.GetByID(ctx, req.PersonID)
	if err != nil {
		uc.logger.Error("Failed to get person", "error", err, "person_id", req.PersonID)
		return nil, fmt.Errorf("person not found")
	}

	// Check if employee already exists for this person in this condominium
	exists, err := uc.employeeRepo.Exists(ctx, req.PersonID, req.CondominiumID)
	if err != nil {
		uc.logger.Error("Failed to check employee existence", "error", err)
		return nil, fmt.Errorf("failed to check employee existence: %w", err)
	}

	if exists {
		uc.logger.Warn("Employee already exists", "person_id", req.PersonID, "condominium_id", req.CondominiumID)
		return nil, fmt.Errorf("employee already exists for this person in this condominium")
	}

	// Create the employee
	employee, err := domain.NewEmployee(req.PersonID, req.CondominiumID, req.Occupation, startDate)
	if err != nil {
		uc.logger.Error("Failed to create employee domain object", "error", err)
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	// Set optional fields
	if req.Salary > 0 {
		if err := employee.SetSalary(req.Salary); err != nil {
			uc.logger.Error("Failed to set employee salary", "error", err)
			return nil, fmt.Errorf("failed to set salary: %w", err)
		}
	}

	if req.Observations != "" {
		employee.SetObservations(req.Observations)
	}

	// Save to database
	if err := uc.employeeRepo.Create(ctx, employee); err != nil {
		uc.logger.Error("Failed to create employee", "error", err)
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	uc.logger.Info("Employee created successfully", "employee_id", employee.ID, "person_id", req.PersonID)

	// Format response
	employeeData := &EmployeeData{
		ID:            employee.ID,
		PersonID:      employee.PersonID,
		CondominiumID: employee.CondominiumID,
		Occupation:    employee.Occupation,
		Salary:        employee.Salary,
		StartDate:     employee.StartDate.Format("2006-01-02"),
		Status:        employee.Status,
		Observations:  employee.Observations,
		Person: &PersonData{
			ID:          person.ID,
			Name:        person.Name,
			Email:       person.Email,
			PhoneNumber: person.PhoneNumber,
		},
	}

	if employee.EndDate != nil {
		endDate := employee.EndDate.Format("2006-01-02")
		employeeData.EndDate = &endDate
	}

	return &CreateEmployeeResponse{
		Employee: employeeData,
		Message:  "Employee created successfully.",
	}, nil
}
