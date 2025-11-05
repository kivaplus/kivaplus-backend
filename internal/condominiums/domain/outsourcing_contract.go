package domain

import (
	"fmt"
	"time"
)

// ContractStatus represents the status of an outsourcing contract
type ContractStatus string

const (
	ContractStatusActive    ContractStatus = "ativo"
	ContractStatusSuspended ContractStatus = "suspenso"
	ContractStatusCanceled  ContractStatus = "cancelado"
	ContractStatusFinalized ContractStatus = "finalizado"
)

// OutsourcingContract represents a contract between a condominium and an outsourcing company
type OutsourcingContract struct {
	ID                   int64          `json:"id"`
	CondominiumID        int64          `json:"condominium_id"`
	OutsourcingCompanyID int64          `json:"outsourcing_company_id"`
	ContractNumber       string         `json:"contract_number"`
	MonthlyAmount        float64        `json:"monthly_amount"`
	DueDay               int            `json:"due_day"`
	StartDate            time.Time      `json:"start_date"`
	EndDate              *time.Time     `json:"end_date,omitempty"`
	Status               ContractStatus `json:"status"`
	Observations         string         `json:"observations,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            *time.Time     `json:"deleted_at,omitempty"`
}

// NewOutsourcingContract creates a new outsourcing contract
func NewOutsourcingContract(
	condominiumID, outsourcingCompanyID int64,
	contractNumber string,
	monthlyAmount float64,
	dueDay int,
	startDate time.Time,
) (*OutsourcingContract, error) {
	if condominiumID <= 0 {
		return nil, fmt.Errorf("condominium ID must be greater than 0")
	}
	if outsourcingCompanyID <= 0 {
		return nil, fmt.Errorf("outsourcing company ID must be greater than 0")
	}
	if contractNumber == "" {
		return nil, fmt.Errorf("contract number is required")
	}
	if monthlyAmount < 0 {
		return nil, fmt.Errorf("monthly amount cannot be negative")
	}
	if dueDay < 1 || dueDay > 31 {
		return nil, fmt.Errorf("due day must be between 1 and 31")
	}

	now := time.Now()
	return &OutsourcingContract{
		CondominiumID:        condominiumID,
		OutsourcingCompanyID: outsourcingCompanyID,
		ContractNumber:       contractNumber,
		MonthlyAmount:        monthlyAmount,
		DueDay:               dueDay,
		StartDate:            startDate,
		Status:               ContractStatusActive,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

// SetStatus sets the contract status
func (oc *OutsourcingContract) SetStatus(status ContractStatus) error {
	validStatuses := map[ContractStatus]bool{
		ContractStatusActive:    true,
		ContractStatusSuspended: true,
		ContractStatusCanceled:  true,
		ContractStatusFinalized: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid contract status: %s", status)
	}

	oc.Status = status
	oc.UpdatedAt = time.Now()
	return nil
}

// SetEndDate sets the contract end date
func (oc *OutsourcingContract) SetEndDate(endDate time.Time) error {
	if endDate.Before(oc.StartDate) {
		return fmt.Errorf("end date cannot be before start date")
	}
	oc.EndDate = &endDate
	oc.UpdatedAt = time.Now()
	return nil
}

// UpdateMonthlyAmount updates the monthly amount
func (oc *OutsourcingContract) UpdateMonthlyAmount(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("monthly amount cannot be negative")
	}
	oc.MonthlyAmount = amount
	oc.UpdatedAt = time.Now()
	return nil
}

// UpdateDueDay updates the due day
func (oc *OutsourcingContract) UpdateDueDay(dueDay int) error {
	if dueDay < 1 || dueDay > 31 {
		return fmt.Errorf("due day must be between 1 and 31")
	}
	oc.DueDay = dueDay
	oc.UpdatedAt = time.Now()
	return nil
}

// SetObservations sets observations
func (oc *OutsourcingContract) SetObservations(observations string) {
	oc.Observations = observations
	oc.UpdatedAt = time.Now()
}

// IsActive returns true if the contract is active
func (oc *OutsourcingContract) IsActive() bool {
	return oc.Status == ContractStatusActive && oc.DeletedAt == nil
}

// IsDeleted returns true if the contract is soft deleted
func (oc *OutsourcingContract) IsDeleted() bool {
	return oc.DeletedAt != nil
}

// Cancel cancels the contract
func (oc *OutsourcingContract) Cancel() error {
	return oc.SetStatus(ContractStatusCanceled)
}

// Suspend suspends the contract
func (oc *OutsourcingContract) Suspend() error {
	return oc.SetStatus(ContractStatusSuspended)
}

// Finalize finalizes the contract
func (oc *OutsourcingContract) Finalize(endDate time.Time) error {
	if err := oc.SetEndDate(endDate); err != nil {
		return err
	}
	return oc.SetStatus(ContractStatusFinalized)
}
