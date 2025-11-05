package ports

import (
	"context"

	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
)

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// CreateUserResponse represents the response from creating a user
type CreateUserResponse struct {
	User   *domain.User   `json:"user"`
	Person *domain.Person `json:"person"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *domain.User `json:"user"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UpdateProfileRequest represents the request to update user profile
type UpdateProfileRequest struct {
	Name          string               `json:"name"`
	Document      string               `json:"document,omitempty"`
	DocumentType  domain.DocumentType  `json:"document_type,omitempty"`
	PersonType    domain.PersonType    `json:"person_type"`
	PhoneNumber   string               `json:"phone_number,omitempty"`
	Email         string               `json:"email,omitempty"`
	Occupation    string               `json:"occupation,omitempty"`
	MaritalStatus domain.MaritalStatus `json:"marital_status,omitempty"`
	BirthdayDate  string               `json:"birthday_date,omitempty"`
	Observations  string               `json:"observations,omitempty"`
	Address       CreateAddressRequest `json:"address"`
}

type CreateAddressRequest struct {
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
}

// ProfileData represents the formatted profile data for responses
type ProfileData struct {
	Name          string               `json:"name"`
	Email         string               `json:"email"`
	Document      string               `json:"document,omitempty"`
	DocumentType  domain.DocumentType  `json:"document_type,omitempty"`
	PhoneNumber   string               `json:"phone_number,omitempty"`
	Occupation    string               `json:"occupation,omitempty"`
	MaritalStatus domain.MaritalStatus `json:"marital_status,omitempty"`
	BirthdayDate  string               `json:"birthday_date,omitempty"`
	Address       *AddressData         `json:"address,omitempty"`
}

// AddressData represents the formatted address data for responses
type AddressData struct {
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
}

type CompleteProfileResponse struct {
	Profile      *ProfileData `json:"profile"`
	Message      string       `json:"message"`
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
}

type GetProfileResponse struct {
	Profile *ProfileData `json:"profile"`
	Message string       `json:"message"`
}

// CreateUserService defines the interface for creating users
type CreateUserService interface {
	Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)
}

// LoginService defines the interface for user authentication
type LoginService interface {
	Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

// RefreshTokenService defines the interface for token refresh
type RefreshTokenService interface {
	Execute(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error)
}

// CompleteProfileService defines the interface for completing user profile
type CompleteProfileService interface {
	Execute(ctx context.Context, userID int64, req UpdateProfileRequest) (*CompleteProfileResponse, error)
}

type GetProfileService interface {
	Execute(ctx context.Context, userID int64) (*GetProfileResponse, error)
}
