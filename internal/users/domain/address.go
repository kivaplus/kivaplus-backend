package domain

import "time"

type Address struct {
	ID           int64     `json:"id"`
	ZipCode      string    `json:"zip_code"`
	Street       string    `json:"street"`
	Number       string    `json:"number"`
	Complement   string    `json:"complement"`
	Neighborhood string    `json:"neighborhood"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewAddress(zipCode, street, number, complement, neighborhood, city, state, country string) (*Address, error) {
	now := time.Now()
	return &Address{
		ZipCode:      zipCode,
		Street:       street,
		Number:       number,
		Complement:   complement,
		Neighborhood: neighborhood,
		City:         city,
		State:        state,
		Country:      country,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
