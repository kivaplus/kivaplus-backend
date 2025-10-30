package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, ", ")
}

// Validator provides validation methods
type Validator struct {
	errors ValidationErrors
}

// New creates a new validator
func New() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(field, email string) {
	if email == "" {
		v.AddError(field, "email is required")
		return
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		v.AddError(field, "invalid email format")
	}
}

// ValidatePassword validates a password
func (v *Validator) ValidatePassword(field, password string) {
	if password == "" {
		v.AddError(field, "password is required")
		return
	}

	if len(password) < 6 {
		v.AddError(field, "password must be at least 6 characters long")
	}
}

// ValidateRequired validates that a field is not empty
func (v *Validator) ValidateRequired(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, fmt.Sprintf("%s is required", field))
	}
}

// ValidateMinLength validates minimum length
func (v *Validator) ValidateMinLength(field, value string, minLength int) {
	if len(value) < minLength {
		v.AddError(field, fmt.Sprintf("%s must be at least %d characters long", field, minLength))
	}
}

// ValidateMaxLength validates maximum length
func (v *Validator) ValidateMaxLength(field, value string, maxLength int) {
	if len(value) > maxLength {
		v.AddError(field, fmt.Sprintf("%s must be at most %d characters long", field, maxLength))
	}
}
