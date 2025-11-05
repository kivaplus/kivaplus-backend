package domain

import (
	"fmt"
	"time"
)

// ServiceFrequency represents the frequency of a service
type ServiceFrequency string

const (
	ServiceFrequencyOnce    ServiceFrequency = "unico"
	ServiceFrequencyDaily   ServiceFrequency = "diario"
	ServiceFrequencyWeekly  ServiceFrequency = "semanal"
	ServiceFrequencyMonthly ServiceFrequency = "mensal"
	ServiceFrequencyYearly  ServiceFrequency = "anual"
)

// Service represents a service provided under an outsourcing contract
type Service struct {
	ID                    int64            `json:"id"`
	OutsourcingContractID int64            `json:"outsourcing_contract_id"`
	Name                  string           `json:"name"`
	Description           string           `json:"description,omitempty"`
	Value                 float64          `json:"value"`
	Frequency             ServiceFrequency `json:"frequency"`
	Observations          string           `json:"observations,omitempty"`
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
}

// NewService creates a new service
func NewService(
	outsourcingContractID int64,
	name string,
	value float64,
	frequency ServiceFrequency,
) (*Service, error) {
	if outsourcingContractID <= 0 {
		return nil, fmt.Errorf("outsourcing contract ID must be greater than 0")
	}
	if name == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if value < 0 {
		return nil, fmt.Errorf("service value cannot be negative")
	}
	if !isValidFrequency(frequency) {
		return nil, fmt.Errorf("invalid service frequency: %s", frequency)
	}

	now := time.Now()
	return &Service{
		OutsourcingContractID: outsourcingContractID,
		Name:                  name,
		Value:                 value,
		Frequency:             frequency,
		CreatedAt:             now,
		UpdatedAt:             now,
	}, nil
}

// isValidFrequency checks if the frequency is valid
func isValidFrequency(frequency ServiceFrequency) bool {
	validFrequencies := map[ServiceFrequency]bool{
		ServiceFrequencyOnce:    true,
		ServiceFrequencyDaily:   true,
		ServiceFrequencyWeekly:  true,
		ServiceFrequencyMonthly: true,
		ServiceFrequencyYearly:  true,
	}
	return validFrequencies[frequency]
}

// SetDescription sets the service description
func (s *Service) SetDescription(description string) {
	s.Description = description
	s.UpdatedAt = time.Now()
}

// UpdateValue updates the service value
func (s *Service) UpdateValue(value float64) error {
	if value < 0 {
		return fmt.Errorf("service value cannot be negative")
	}
	s.Value = value
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateFrequency updates the service frequency
func (s *Service) UpdateFrequency(frequency ServiceFrequency) error {
	if !isValidFrequency(frequency) {
		return fmt.Errorf("invalid service frequency: %s", frequency)
	}
	s.Frequency = frequency
	s.UpdatedAt = time.Now()
	return nil
}

// SetObservations sets observations
func (s *Service) SetObservations(observations string) {
	s.Observations = observations
	s.UpdatedAt = time.Now()
}

// UpdateName updates the service name
func (s *Service) UpdateName(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	s.Name = name
	s.UpdatedAt = time.Now()
	return nil
}
