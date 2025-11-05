package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/handlers"
	condominiumPorts "github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	condominiumRepo "github.com/kivaplus/kivaplus-backend/internal/condominiums/repository"
	condominiumUsecases "github.com/kivaplus/kivaplus-backend/internal/condominiums/usecases"
	"github.com/kivaplus/kivaplus-backend/internal/shared/cache"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/permissions"
	"github.com/kivaplus/kivaplus-backend/internal/users/domain"
	userHandlers "github.com/kivaplus/kivaplus-backend/internal/users/handlers"
	userPorts "github.com/kivaplus/kivaplus-backend/internal/users/ports"
	userRepo "github.com/kivaplus/kivaplus-backend/internal/users/repository"
	userUsecases "github.com/kivaplus/kivaplus-backend/internal/users/usecases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealUserCondominiumFlow tests the complete flow using real handlers and use cases
func TestRealUserCondominiumFlow(t *testing.T) {
	// Setup test database
	db, err := database.NewPostgresConnection()
	require.NoError(t, err, "Failed to connect to test database")

	log := logger.New()
	ctx := context.Background()

	// Initialize repositories
	userRepository := userRepo.NewUserPostgresRepository(db)
	personRepository := userRepo.NewPersonPostgresRepository(db)
	addressRepository := userRepo.NewAddressPostgresRepository(db)
	personDocRepository := userRepo.NewPersonDocumentPostgresRepository(db)
	roleRepository := userRepo.NewRolePostgresRepository(db)
	condominiumRepository := condominiumRepo.NewCondominiumPostgresRepository(db)

	// Initialize JWT service (singleton)
	jwtService := jwt.GetJWTService()

	// Initialize Redis cache (optional for tests)
	redisService := cache.NewRedisService()
	if err := redisService.Ping(); err != nil {
		log.Warn("Redis connection failed in test, continuing without cache", "error", err)
		redisService = nil
	}

	// Initialize enhanced role repository
	enhancedRoleRepository := userRepo.NewEnhancedRoleRepository(db, log)

	// Initialize permission service
	var permissionService *permissions.Service
	if redisService != nil {
		permissionService = permissions.NewPermissionService(redisService, enhancedRoleRepository, log)
	} else {
		permissionService = permissions.NewPermissionService(nil, enhancedRoleRepository, log)
	}

	// Initialize enhanced permission checker
	permissionChecker := permissions.NewEnhancedChecker(jwtService, permissionService, redisService, log)

	// Initialize use cases
	createUserUC := userUsecases.NewCreateUser(userRepository, personRepository, roleRepository, log)
	loginUC := userUsecases.NewLogin(userRepository, personRepository, roleRepository, log)
	refreshTokenUC := userUsecases.NewRefreshToken(userRepository, personRepository, roleRepository, log)
	getProfileUC := userUsecases.NewGetProfile(userRepository, personRepository, addressRepository, personDocRepository, log)
	completeProfileUC := userUsecases.NewCompleteProfile(userRepository, personRepository, addressRepository, personDocRepository, roleRepository, jwtService, log)
	createCondominiumUC := condominiumUsecases.NewCreateCondominium(condominiumRepository, personRepository, addressRepository, roleRepository, userRepository, jwtService, log)
	listCondominiumsUC := condominiumUsecases.NewListCondominiums(condominiumRepository, personRepository, addressRepository, roleRepository, log)
	getCondominiumUC := condominiumUsecases.NewGetCondominium(condominiumRepository, personRepository, addressRepository, roleRepository, log)

	// Initialize handlers with enhanced permissions
	authHandler := userHandlers.NewAuthHandler(createUserUC, loginUC, refreshTokenUC, log)
	usersHandler := userHandlers.NewHTTPHandler(getProfileUC, completeProfileUC, permissionChecker, permissionService, redisService, log)
	condominiumsHandler := handlers.NewHTTPHandler(createCondominiumUC, listCondominiumsUC, getCondominiumUC, permissionChecker, permissionService, redisService, log)

	// Test data
	testUser := userPorts.CreateUserRequest{
		Name:            "Maria Santos",
		Email:           fmt.Sprintf("maria.santos+%d@test.com", time.Now().Unix()),
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
	}

	// Generate unique document number based on timestamp
	uniqueDoc := fmt.Sprintf("987654321%02d", time.Now().Unix()%100)

	testProfile := userPorts.UpdateProfileRequest{
		Name:          "Maria Santos Silva",
		Document:      uniqueDoc,
		DocumentType:  domain.CPF,              // Use domain constant
		PersonType:    domain.IndividualEntity, // Use domain constant "PF"
		PhoneNumber:   "+5511888777666",
		Email:         testUser.Email,
		Occupation:    "Arquiteta",
		MaritalStatus: domain.Single, // Use domain constant (now "solteiro")
		BirthdayDate:  "1990-05-15",  // Add birthday date
		Address: userPorts.CreateAddressRequest{
			ZipCode:      "05678-901",
			Street:       "Rua das Acácias",
			Number:       "456",
			Complement:   "Casa 2",
			Neighborhood: "Jardim Europa",
			City:         "São Paulo",
			State:        "SP",
			Country:      "BR",
		},
	}

	// Generate unique document number based on timestamp
	uniqueCondominiumDoc := fmt.Sprintf("98.765.432/0001-%02d", time.Now().Unix()%100)

	testCondominium := condominiumPorts.CreateCondominiumRequest{
		Name:           "Condomínio Residencial Primavera",
		DocumentNumber: uniqueCondominiumDoc,
		PhoneNumber:    "+5511876543210",
		Email:          "contato@primavera.com.br",
		Observations:   "Condomínio com área de lazer completa",
		Address: condominiumPorts.CreateCondominiumAddressRequest{
			ZipCode:      "06789-012",
			Street:       "Avenida das Magnólias",
			Number:       "2000",
			Complement:   "Bloco Principal",
			Neighborhood: "Alto da Boa Vista",
			City:         "São Paulo",
			State:        "SP",
			Country:      "BR",
		},
	}

	var accessToken string
	var condominiumID int64

	t.Run("1. Register new user", func(t *testing.T) {
		// Prepare request
		requestBody, err := json.Marshal(testUser)
		require.NoError(t, err)

		request := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       "/register",
			Body:       string(requestBody),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		// Execute request
		response, err := authHandler.Register(ctx, request)
		require.NoError(t, err)

		// Verify response
		t.Logf("Register response status: %d", response.StatusCode)
		t.Logf("Register response body: %s", response.Body)

		assert.Equal(t, http.StatusCreated, response.StatusCode)

		var registerResponse userPorts.CreateUserResponse
		err = json.Unmarshal([]byte(response.Body), &registerResponse)
		require.NoError(t, err)

		// Check that both user and person were created
		require.NotNil(t, registerResponse.User, "User should not be nil in response")
		require.NotNil(t, registerResponse.Person, "Person should not be nil in response")
		assert.Equal(t, testUser.Name, registerResponse.Person.Name)
		assert.Equal(t, testUser.Email, registerResponse.User.Email)
		assert.NotEmpty(t, registerResponse.User.ID)

		log.Info("✅ User registered successfully", "userID", registerResponse.User.ID, "email", testUser.Email)
	})

	t.Run("2. Login to get access token", func(t *testing.T) {
		// Prepare login request
		loginRequest := userPorts.LoginRequest{
			Email:    testUser.Email,
			Password: testUser.Password,
		}

		requestBody, err := json.Marshal(loginRequest)
		require.NoError(t, err)

		request := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       "/login",
			Body:       string(requestBody),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		// Execute request
		response, err := authHandler.Login(ctx, request)
		require.NoError(t, err)

		// Verify response
		t.Logf("Login response status: %d", response.StatusCode)
		t.Logf("Login response body: %s", response.Body)

		assert.Equal(t, http.StatusOK, response.StatusCode)

		var loginResponse userPorts.LoginResponse
		err = json.Unmarshal([]byte(response.Body), &loginResponse)
		require.NoError(t, err)

		assert.NotEmpty(t, loginResponse.AccessToken)
		assert.NotEmpty(t, loginResponse.RefreshToken)
		assert.Equal(t, testUser.Email, loginResponse.User.Email)

		accessToken = loginResponse.AccessToken
		log.Info("✅ User logged in successfully", "email", testUser.Email)
	})

	t.Run("3. Complete user profile", func(t *testing.T) {
		// Prepare profile update request
		requestBody, err := json.Marshal(testProfile)
		require.NoError(t, err)

		request := events.APIGatewayProxyRequest{
			HTTPMethod: "PUT",
			Path:       "/profile",
			Body:       string(requestBody),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer " + accessToken,
			},
		}

		// Execute request
		response, err := usersHandler.UpdateProfile(ctx, request)
		require.NoError(t, err)

		// Verify response
		t.Logf("Profile update response status: %d", response.StatusCode)
		t.Logf("Profile update response body: %s", response.Body)

		assert.Equal(t, http.StatusOK, response.StatusCode)

		var profileResponse map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &profileResponse)
		require.NoError(t, err)

		// The message can be either "Profile updated successfully" or "Profile completed successfully."
		message := profileResponse["message"].(string)
		assert.True(t, message == "Profile updated successfully" || message == "Profile completed successfully.", "Expected profile completion message")
		assert.NotNil(t, profileResponse["profile"])

		// Update access token if new one was provided (it's nested inside the profile object)
		if profileData, ok := profileResponse["profile"].(map[string]interface{}); ok {
			if newAccessToken, ok := profileData["access_token"].(string); ok && newAccessToken != "" {
				accessToken = newAccessToken
				log.Info("✅ Updated access token after profile completion", "tokenLength", len(accessToken))
			} else {
				log.Warn("⚠️ No new access token found in profile data")
			}
		} else {
			log.Warn("⚠️ No profile data found in response")
		}

		log.Info("✅ User profile completed successfully")
	})

	t.Run("4. Create condominium", func(t *testing.T) {
		// Prepare condominium creation request
		requestBody, err := json.Marshal(testCondominium)
		require.NoError(t, err)

		request := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       "/condominios",
			Body:       string(requestBody),
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer " + accessToken,
			},
		}

		// Execute request
		response, err := condominiumsHandler.CreateCondominium(ctx, request)
		require.NoError(t, err)

		// Verify response
		t.Logf("Create condominium response status: %d", response.StatusCode)
		t.Logf("Create condominium response body: %s", response.Body)

		assert.Equal(t, http.StatusCreated, response.StatusCode)

		var condominiumResponse condominiumPorts.CreateCondominiumResponse
		err = json.Unmarshal([]byte(response.Body), &condominiumResponse)
		require.NoError(t, err)

		require.NotNil(t, condominiumResponse.Condominium, "Condominium should not be nil in response")
		assert.Equal(t, testCondominium.Name, condominiumResponse.Condominium.Name)
		assert.Equal(t, testCondominium.DocumentNumber, condominiumResponse.Condominium.DocumentNumber)
		assert.NotEmpty(t, condominiumResponse.Condominium.ID)

		condominiumID = condominiumResponse.Condominium.ID

		// Update access token if new one was provided
		if condominiumResponse.AccessToken != "" {
			accessToken = condominiumResponse.AccessToken
			log.Info("✅ Updated access token after condominium creation", "tokenLength", len(accessToken))
		} else {
			log.Warn("⚠️ No new access token provided in condominium creation response")
		}

		log.Info("✅ Condominium created successfully", "condominiumID", condominiumID, "name", testCondominium.Name)
	})

	t.Run("5. Get condominium by ID", func(t *testing.T) {
		// Prepare get condominium request
		request := events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			Path:       fmt.Sprintf("/condominios/%d", condominiumID),
			PathParameters: map[string]string{
				"id": strconv.FormatInt(condominiumID, 10),
			},
			Headers: map[string]string{
				"Authorization": "Bearer " + accessToken,
			},
		}

		// Execute request
		response, err := condominiumsHandler.GetCondominium(ctx, request)
		require.NoError(t, err)

		// Verify response
		t.Logf("Get condominium response status: %d", response.StatusCode)
		t.Logf("Get condominium response body: %s", response.Body)

		assert.Equal(t, http.StatusOK, response.StatusCode)

		var getCondominiumResponse map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &getCondominiumResponse)
		require.NoError(t, err)

		// Verify condominium data
		condominiumData, ok := getCondominiumResponse["condominium"].(map[string]interface{})
		require.True(t, ok, "Expected condominium data in response")

		assert.Equal(t, float64(condominiumID), condominiumData["id"])
		assert.Equal(t, testCondominium.Name, condominiumData["name"])
		assert.Equal(t, testCondominium.DocumentNumber, condominiumData["document_number"])

		log.Info("✅ Condominium retrieved successfully", "condominiumID", condominiumID)
	})

	t.Run("6. Verify complete flow integrity", func(t *testing.T) {
		// Additional verification that the complete flow worked
		assert.NotEmpty(t, accessToken, "Access token should be available")
		assert.Greater(t, condominiumID, int64(0), "Condominium ID should be valid")

		// Verify we can get the user profile
		request := events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			Path:       "/profile",
			Headers: map[string]string{
				"Authorization": "Bearer " + accessToken,
			},
		}

		response, err := usersHandler.GetProfile(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		log.Info("✅ Complete user-condominium flow test passed successfully",
			"email", testUser.Email,
			"condominiumID", condominiumID,
			"condominiumName", testCondominium.Name)
	})
}

// TestFlowWithValidation tests the flow with various validation scenarios
func TestFlowWithValidation(t *testing.T) {
	// Setup test database
	db, err := database.NewPostgresConnection()
	require.NoError(t, err, "Failed to connect to test database")

	log := logger.New()
	ctx := context.Background()

	// Initialize minimal handlers for validation tests
	userRepository := userRepo.NewUserPostgresRepository(db)
	personRepository := userRepo.NewPersonPostgresRepository(db)
	roleRepository := userRepo.NewRolePostgresRepository(db)

	createUserUC := userUsecases.NewCreateUser(userRepository, personRepository, roleRepository, log)
	loginUC := userUsecases.NewLogin(userRepository, personRepository, roleRepository, log)
	refreshTokenUC := userUsecases.NewRefreshToken(userRepository, personRepository, roleRepository, log)

	authHandler := userHandlers.NewAuthHandler(createUserUC, loginUC, refreshTokenUC, log)

	t.Run("Invalid registration - missing fields", func(t *testing.T) {
		invalidUser := userPorts.CreateUserRequest{
			Name:     "Test User",
			Email:    "", // Missing email
			Password: "password123",
		}

		requestBody, err := json.Marshal(invalidUser)
		require.NoError(t, err)

		request := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       "/register",
			Body:       string(requestBody),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		response, err := authHandler.Register(ctx, request)
		require.NoError(t, err)

		// Should return validation error
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		log.Info("✅ Validation test passed - invalid registration rejected")
	})

	t.Run("Invalid login - wrong credentials", func(t *testing.T) {
		loginRequest := userPorts.LoginRequest{
			Email:    "nonexistent@test.com",
			Password: "wrongpassword",
		}

		requestBody, err := json.Marshal(loginRequest)
		require.NoError(t, err)

		request := events.APIGatewayProxyRequest{
			HTTPMethod: "POST",
			Path:       "/login",
			Body:       string(requestBody),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		response, err := authHandler.Login(ctx, request)
		require.NoError(t, err)

		// Should return authentication error
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
		log.Info("✅ Validation test passed - invalid login rejected")
	})
}
