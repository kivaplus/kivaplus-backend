package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDocument(t *testing.T) {
	tests := []struct {
		name         string
		document     string
		docType      string
		expectErrors bool
		errorMessage string
	}{
		{
			name:         "Valid CPF with 11 digits",
			document:     "12345678901",
			docType:      "CPF",
			expectErrors: false,
		},
		{
			name:         "Valid CPF with formatting",
			document:     "123.456.789-01",
			docType:      "CPF",
			expectErrors: false,
		},
		{
			name:         "Invalid CPF with 10 digits",
			document:     "1234567890",
			docType:      "CPF",
			expectErrors: true,
			errorMessage: "CPF must have exactly 11 digits",
		},
		{
			name:         "Invalid CPF with 12 digits",
			document:     "123456789012",
			docType:      "CPF",
			expectErrors: true,
			errorMessage: "CPF must have exactly 11 digits",
		},
		{
			name:         "Valid CNPJ with 14 digits",
			document:     "12345678901234",
			docType:      "CNPJ",
			expectErrors: false,
		},
		{
			name:         "Valid CNPJ with formatting",
			document:     "12.345.678/9012-34",
			docType:      "CNPJ",
			expectErrors: false,
		},
		{
			name:         "Invalid CNPJ with 13 digits",
			document:     "1234567890123",
			docType:      "CNPJ",
			expectErrors: true,
			errorMessage: "CNPJ must have exactly 14 digits",
		},
		{
			name:         "Empty document",
			document:     "",
			docType:      "CPF",
			expectErrors: true,
			errorMessage: "document is required",
		},
		{
			name:         "Invalid document type",
			document:     "12345678901",
			docType:      "RG",
			expectErrors: true,
			errorMessage: "document type must be CPF or CNPJ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.ValidateDocument("document", tt.document, tt.docType)

			if tt.expectErrors {
				assert.True(t, v.HasErrors(), "Expected validation errors")
				if tt.errorMessage != "" {
					found := false
					for _, err := range v.Errors() {
						if err.Message == tt.errorMessage {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message: %s", tt.errorMessage)
				}
			} else {
				assert.False(t, v.HasErrors(), "Expected no validation errors")
			}
		})
	}
}
