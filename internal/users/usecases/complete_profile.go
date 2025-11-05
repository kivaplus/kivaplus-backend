package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/validator"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	"github.com/kivaplus/kivaplus-backend/internal/users/ports"
)

// CompleteProfile implements the complete user profile use case
type CompleteProfile struct {
	userRepo           ports.UserRepository
	personRepo         ports.PersonRepository
	addressRepo        ports.AddressRepository
	personDocumentRepo ports.PersonDocumentRepository
	roleRepo           ports.RoleRepository
	jwtService         *jwt.Service
	logger             logger.Logger
}

// NewCompleteProfile creates a new CompleteProfile use case
func NewCompleteProfile(userRepo ports.UserRepository, personRepo ports.PersonRepository, addressRepo ports.AddressRepository, personDocumentRepo ports.PersonDocumentRepository, roleRepo ports.RoleRepository, jwtService *jwt.Service, logger logger.Logger) *CompleteProfile {
	return &CompleteProfile{
		userRepo:           userRepo,
		personRepo:         personRepo,
		addressRepo:        addressRepo,
		personDocumentRepo: personDocumentRepo,
		roleRepo:           roleRepo,
		jwtService:         jwtService,
		logger:             logger,
	}
}

// Execute completes the user profile by creating or updating person information
func (uc *CompleteProfile) Execute(ctx context.Context, userID int64, req ports.UpdateProfileRequest) (*ports.CompleteProfileResponse, error) {
	uc.logger.Info("Completing user profile", "user_id", userID)

	// Check if user exists first
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user by ID", "error", err, "usuario_id", userID)
		return nil, fmt.Errorf("user not found")
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("name", req.Name)
	v.ValidateMinLength("name", req.Name, 2)
	v.ValidateMaxLength("name", req.Name, 255)

	if req.Email != "" {
		v.ValidateEmail("email", req.Email)
	}

	// Validate document if provided
	if req.Document != "" {
		v.ValidateDocument("document", req.Document, string(req.DocumentType))

		// Check if document already exists for another user
		if !v.HasErrors() {
			// Get current person to exclude from uniqueness check
			currentPerson, err := uc.personRepo.GetByUserID(ctx, userID)
			if err != nil {
				uc.logger.Error("Debug: GetByUserID error", "error", err, "error_string", err.Error(), "userID", userID)
				if !isPersonNotFoundError(err) {
					uc.logger.Error("Failed to check existing person for document validation", "error", err)
					return nil, fmt.Errorf("failed to validate document ownership: %w", err)
				}
			}

			var excludePersonID int64 = -1 // Use -1 if no person exists yet
			if currentPerson != nil {
				excludePersonID = currentPerson.ID
			}

			// Check if document exists for a different person
			exists, err := uc.personDocumentRepo.ExistsByNumberAndTypeForDifferentPerson(ctx, req.Document, req.DocumentType, excludePersonID)
			if err != nil {
				uc.logger.Error("Failed to check document uniqueness", "error", err)
				return nil, fmt.Errorf("failed to validate document uniqueness: %w", err)
			}

			if exists {
				v.AddError("document", fmt.Sprintf("Document %s already exists for another user", req.Document))
			}
		}
	}

	if v.HasErrors() {
		uc.logger.Error("Profile validation failed", "errors", v.Errors())
		return nil, fmt.Errorf("validation failed: %v", v.Errors())
	}

	// Get existing person (we may have already checked this during validation)
	hasPerson, err := uc.personRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("Debug: GetByUserID error in main flow", "error", err, "error_string", err.Error(), "userID", userID)
		if !isPersonNotFoundError(err) {
			uc.logger.Error("Failed to check existing person", "error", err, "user_id", userID)
			return nil, fmt.Errorf("failed to check existing profile: %w", err)
		}
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

	// Parse birthday date if provided
	var birthdayDate *time.Time
	if req.BirthdayDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.BirthdayDate)
		if err != nil {
			uc.logger.Error("Invalid birthday date format", "error", err, "date", req.BirthdayDate)
			return nil, fmt.Errorf("invalid birthday date format: %w", err)
		}
		birthdayDate = &parsedDate
	}

	if hasPerson != nil {
		// Update existing person
		hasPerson.UpdateProfile(
			req.Name,
			req.PhoneNumber,
			req.Email,
			req.Occupation,
			req.MaritalStatus,
			birthdayDate,
			&createdAddress.ID,
		)

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
		person.MaritalStatus = req.MaritalStatus
		person.BirthdayDate = birthdayDate
		person.AddressID = &createdAddress.ID
		person.LinkToUser(userID)

		if err := uc.personRepo.Create(ctx, person); err != nil {
			uc.logger.Error("Failed to create person", "error", err, "usuario_id", userID)
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}

		uc.logger.Info("Profile created successfully", "user_id", userID, "person_id", person.ID)
	}

	// Handle document creation/update if provided
	if req.Document != "" && req.DocumentType != "" {
		// Check if document already exists for this person
		existingDoc, err := uc.personDocumentRepo.GetByPersonIDAndType(ctx, person.ID, req.DocumentType)
		if err != nil && err.Error() != "document not found" {
			uc.logger.Error("Failed to check existing document", "error", err, "person_id", person.ID)
			return nil, fmt.Errorf("failed to check existing document: %w", err)
		}

		if existingDoc != nil {
			// Update existing document only if the number is different
			if existingDoc.Number != req.Document {
				existingDoc.Number = req.Document
				existingDoc.UpdatedAt = time.Now()
				if err := uc.personDocumentRepo.Update(ctx, existingDoc); err != nil {
					uc.logger.Error("Failed to update document", "error", err, "person_id", person.ID)
					return nil, fmt.Errorf("failed to update document: %w", err)
				}
				uc.logger.Info("Document updated successfully", "person_id", person.ID, "document_type", req.DocumentType)
			}
		} else {
			// Create new document
			document, err := domain.NewPersonDocument(person.ID, req.DocumentType, req.Document)
			if err != nil {
				uc.logger.Error("Failed to create document domain object", "error", err, "person_id", person.ID)
				return nil, fmt.Errorf("failed to create document: %w", err)
			}

			if err := uc.personDocumentRepo.Create(ctx, document); err != nil {
				uc.logger.Error("Failed to create document", "error", err, "person_id", person.ID)
				return nil, fmt.Errorf("failed to create document: %w", err)
			}
			uc.logger.Info("Document created successfully", "person_id", person.ID, "document_type", req.DocumentType)
		}
	}

	// Get document information
	var document string
	var documentType domain.DocumentType
	if req.Document != "" && req.DocumentType != "" {
		// Get the document we just created/updated
		doc, err := uc.personDocumentRepo.GetByPersonIDAndType(ctx, person.ID, req.DocumentType)
		if err == nil {
			document = doc.Number
			documentType = doc.Type
		}
	}

	// Format the response
	profileData := &ports.ProfileData{
		Name:          person.Name,
		Email:         person.Email,
		Document:      document,
		DocumentType:  documentType,
		PhoneNumber:   person.PhoneNumber,
		Occupation:    person.Occupation,
		MaritalStatus: person.MaritalStatus,
		Address: &ports.AddressData{
			ZipCode:      createdAddress.ZipCode,
			Street:       createdAddress.Street,
			Number:       createdAddress.Number,
			Complement:   createdAddress.Complement,
			Neighborhood: createdAddress.Neighborhood,
			City:         createdAddress.City,
			State:        createdAddress.State,
			Country:      createdAddress.Country,
		},
	}

	if person.BirthdayDate != nil {
		profileData.BirthdayDate = person.BirthdayDate.Format("2006-01-02")
	}

	// Generate new JWT tokens with updated profile status
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user for JWT generation", "error", err, "userID", userID)
		// Don't fail the whole operation, just return without tokens
		return &ports.CompleteProfileResponse{
			Profile: profileData,
			Message: "Profile completed successfully.",
		}, nil
	}

	// Note: Role validation is now handled by the JWT service internally

	// Profile is now complete
	profileStatus := "complete"

	// Generate new access token - Use optimized method
	userIDStr := fmt.Sprintf("%d", user.ID)
	tokenVersion := time.Now().Unix() // Generate new version for token invalidation
	accessToken, err := uc.jwtService.GenerateToken(userIDStr, user.Email, profileStatus, tokenVersion)
	if err != nil {
		uc.logger.Error("Failed to generate access token", "error", err, "userID", userID)
		// Don't fail the whole operation, just return without tokens
		return &ports.CompleteProfileResponse{
			Profile: profileData,
			Message: "Profile completed successfully.",
		}, nil
	}

	// Generate new refresh token
	refreshToken, err := uc.jwtService.GenerateRefreshToken(userIDStr, user.Email)
	if err != nil {
		uc.logger.Error("Failed to generate refresh token", "error", err, "userID", userID)
		// Don't fail the whole operation, just return without tokens
		return &ports.CompleteProfileResponse{
			Profile:     profileData,
			Message:     "Profile completed successfully.",
			AccessToken: accessToken, // At least return the access token
		}, nil
	}

	return &ports.CompleteProfileResponse{
		Profile:      profileData,
		Message:      "Profile completed successfully.",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
