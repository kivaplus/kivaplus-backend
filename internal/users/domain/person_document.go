package domain

import "time"

// DocumentType represents the type of document
type DocumentType string

const (
	CPF  DocumentType = "CPF"
	RG   DocumentType = "RG"
	CNH  DocumentType = "CNH"
	CNPJ DocumentType = "CNPJ"
)

// PersonDocument represents a document associated with a person
type PersonDocument struct {
	ID               int64        `json:"id"`
	PersonID         int64        `json:"person_id"`
	Type             DocumentType `json:"type"`
	Number           string       `json:"number"`
	IssuingAuthority string       `json:"issuing_authority,omitempty"`
	IssueDate        *time.Time   `json:"issue_date,omitempty"`
	ExpirationDate   *time.Time   `json:"expiration_date,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// NewPersonDocument creates a new person document
func NewPersonDocument(personID int64, docType DocumentType, number string) (*PersonDocument, error) {
	now := time.Now()
	return &PersonDocument{
		PersonID:  personID,
		Type:      docType,
		Number:    number,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SetIssuingDetails sets the issuing authority and dates
func (pd *PersonDocument) SetIssuingDetails(authority string, issueDate, expirationDate *time.Time) {
	pd.IssuingAuthority = authority
	pd.IssueDate = issueDate
	pd.ExpirationDate = expirationDate
	pd.UpdatedAt = time.Now()
}
