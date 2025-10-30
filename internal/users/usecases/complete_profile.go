package usecases

import (
	"context"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/validator"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// CompleteProfile implements the complete user profile use case
type CompleteProfile struct {
	userRepo    ports.UserRepository
	personRepo  ports.PersonRepository
	addressRepo ports.AddressRepository
	logger      logger.Logger
}

// NewCompleteProfile creates a new CompleteProfile use case
func NewCompleteProfile(userRepo ports.UserRepository, personRepo ports.PersonRepository, addressRepo ports.AddressRepository, logger logger.Logger) *CompleteProfile {
	return &CompleteProfile{
		userRepo:    userRepo,
		personRepo:  personRepo,
		addressRepo: addressRepo,
		logger:      logger,
	}
}

// Execute completes the user profile by creating or updating person information
func (uc *CompleteProfile) Execute(ctx context.Context, userID int64, req ports.UpdateProfileRequest) (*ports.CompleteProfileResponse, error) {
	uc.logger.Info("Completing user profile", "user_id", userID)

	// Validate input
	v := validator.New()
	v.ValidateRequired("name", req.Name)
	v.ValidateMinLength("name", req.Name, 2)
	v.ValidateMaxLength("name", req.Name, 255)

	if req.Email != "" {
		v.ValidateEmail("email", req.Email)
	}

	if v.HasErrors() {
		uc.logger.Error("Profile validation failed", "errors", v.Errors())
		return nil, fmt.Errorf("validation failed: %v", v.Errors())
	}

	// Check if user exists
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user by ID", "error", err, "usuario_id", userID)
		return nil, fmt.Errorf("user not found")
	}

	// Check if person already exists for this user
	hasPerson, err := uc.personRepo.GetByUserID(ctx, userID)
	if err != nil && err.Error() != "pessoa not found" {
		uc.logger.Error("Failed to check existing person", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}

	// Create address from request
	address, err := domain.NewAddress(
		req.Address.ZipCode,
		req.Address.Street,
		req.Address.Number,
		req.Address.Complement,
		req.Address.Neighborhood,
		req.Address.City,
		req.Address.State,
		req.Address.Country,
	)
	if err != nil {
		uc.logger.Error("Failed to create address domain object", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	createdAddress, err := uc.addressRepo.Create(ctx, address)
	if err != nil {
		uc.logger.Error("Failed to create address", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	var person *domain.Person

	if hasPerson != nil {
		// Update existing person
		hasPerson.UpdateProfile(req.Name, req.PhoneNumber, req.Email, req.Occupation, &createdAddress.ID)

		if err := uc.personRepo.Update(ctx, hasPerson); err != nil {
			uc.logger.Error("Failed to update person", "error", err, "usuario_id", userID)
			return nil, fmt.Errorf("failed to update profile: %w", err)
		}

		person = hasPerson
		uc.logger.Info("Profile updated successfully", "user_id", userID, "person_id", person.ID)
	} else {
		// Create new person with the created address
		person, err = domain.NewPerson(req.Name, req.Email, userID)
		if err != nil {
			uc.logger.Error("Failed to create person domain object", "error", err, "user_id", userID)
			return nil, fmt.Errorf("failed to create person: %w", err)
		}

		person.PhoneNumber = req.PhoneNumber
		person.Email = req.Email
		person.Occupation = req.Occupation
		person.AddressID = &createdAddress.ID
		person.LinkToUser(userID)

		if err := uc.personRepo.Create(ctx, person); err != nil {
			uc.logger.Error("Failed to create person", "error", err, "usuario_id", userID)
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}

		uc.logger.Info("Profile created successfully", "usuario_id", userID, "pessoa_id", person.ID)
	}

	return &ports.CompleteProfileResponse{
		User:    user,
		Person:  person,
		Message: "Profile complete successfully.",
	}, nil
}
