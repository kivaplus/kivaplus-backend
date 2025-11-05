package domain

import (
	"fmt"
	"time"
)

// EmployeeStatus represents the status of an employee
type EmployeeStatus string

const (
	EmployeeStatusActive   EmployeeStatus = "ativo"
	EmployeeStatusInactive EmployeeStatus = "inativo"
	EmployeeStatusLeave    EmployeeStatus = "afastado"
	EmployeeStatusVacation EmployeeStatus = "ferias"
)

// Employee represents an employee working in a condominium
type Employee struct {
	ID            int64          `json:"id"`
	PersonID      int64          `json:"person_id"`
	CondominiumID int64          `json:"condominium_id"`
	Occupation    string         `json:"occupation"`
	Salary        *float64       `json:"salary,omitempty"`
	StartDate     time.Time      `json:"start_date"`
	EndDate       *time.Time     `json:"end_date,omitempty"`
	Status        EmployeeStatus `json:"status"`
	Observations  string         `json:"observations,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     *time.Time     `json:"deleted_at,omitempty"`
}

// NewEmployee creates a new employee
func NewEmployee(personID, condominiumID int64, occupation string, startDate time.Time) (*Employee, error) {
	if personID <= 0 {
		return nil, fmt.Errorf("person ID must be greater than 0")
	}
	if condominiumID <= 0 {
		return nil, fmt.Errorf("condominium ID must be greater than 0")
	}
	if occupation == "" {
		return nil, fmt.Errorf("occupation is required")
	}

	now := time.Now()
	return &Employee{
		PersonID:      personID,
		CondominiumID: condominiumID,
		Occupation:    occupation,
		StartDate:     startDate,
		Status:        EmployeeStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// SetSalary sets the employee salary
func (e *Employee) SetSalary(salary float64) error {
	if salary < 0 {
		return fmt.Errorf("salary cannot be negative")
	}
	e.Salary = &salary
	e.UpdatedAt = time.Now()
	return nil
}

// SetStatus sets the employee status
func (e *Employee) SetStatus(status EmployeeStatus) error {
	validStatuses := map[EmployeeStatus]bool{
		EmployeeStatusActive:   true,
		EmployeeStatusInactive: true,
		EmployeeStatusLeave:    true,
		EmployeeStatusVacation: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid employee status: %s", status)
	}

	e.Status = status
	e.UpdatedAt = time.Now()
	return nil
}

// SetEndDate sets the employee end date
func (e *Employee) SetEndDate(endDate time.Time) error {
	if endDate.Before(e.StartDate) {
		return fmt.Errorf("end date cannot be before start date")
	}
	e.EndDate = &endDate
	e.UpdatedAt = time.Now()
	return nil
}

// SetObservations sets observations
func (e *Employee) SetObservations(observations string) {
	e.Observations = observations
	e.UpdatedAt = time.Now()
}

// IsActive returns true if the employee is active
func (e *Employee) IsActive() bool {
	return e.Status == EmployeeStatusActive && e.DeletedAt == nil
}

// IsDeleted returns true if the employee is soft deleted
func (e *Employee) IsDeleted() bool {
	return e.DeletedAt != nil
}

// Terminate terminates the employee
func (e *Employee) Terminate(endDate time.Time) error {
	if err := e.SetEndDate(endDate); err != nil {
		return err
	}
	return e.SetStatus(EmployeeStatusInactive)
}
