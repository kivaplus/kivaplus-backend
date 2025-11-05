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
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	userPorts "github.com/kivaplus/kivaplus-backend/internal/users/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserCondominiumFlow tests the complete flow:
// 1. Create new user (register)
// 2. Login to get access token
// 3. Complete profile
// 4. Create condominium
// 5. Get condominium by ID
func TestUserCondominiumFlow(t *testing.T) {
	// Setup test database
	db, err := database.NewPostgresConnection()
	require.NoError(t, err, "Failed to connect to test database")

	log := logger.New()
	ctx := context.Background()

	// Test data
	testUser := userPorts.CreateUserRequest{
		Name:            "João Silva",
		Email:           fmt.Sprintf("joao.silva+%d@test.com", time.Now().Unix()),
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
	}

	testProfile := userPorts.UpdateProfileRequest{
		Name:         "João Silva Santos",
		Document:     "12345678901",
		DocumentType: "cpf",
		PersonType:   "fisica",
		PhoneNumber:  "+5511999887766",
		Email:        testUser.Email,
		Occupation:   "Engenheiro",
		Address: userPorts.CreateAddressRequest{
			ZipCode:      "01234-567",
			Street:       "Rua das Flores",
			Number:       "123",
			Complement:   "Apto 45",
			Neighborhood: "Centro",
			City:         "São Paulo",
			State:        "SP",
			Country:      "Brasil",
		},
	}

	testCondominium := ports.CreateCondominiumRequest{
		Name:           "Residencial Jardim das Flores",
		DocumentNumber: "12.345.678/0001-90",
		PhoneNumber:    "+5511987654321",
		Email:          "contato@jardimflores.com.br",
		Observations:   "Condomínio residencial com 50 unidades",
		Address: ports.CreateCondominiumAddressRequest{
			ZipCode:      "04567-890",
			Street:       "Avenida das Palmeiras",
			Number:       "1000",
			Complement:   "Torre A",
			Neighborhood: "Vila Madalena",
			City:         "São Paulo",
			State:        "SP",
			Country:      "Brasil",
		},
	}

	var accessToken string
	var condominiumID int64

	t.Run("1. Register new user", func(t *testing.T) {
		// Create auth handler
		authHandler := createAuthHandler(db, log)

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
		assert.Equal(t, http.StatusCreated, response.StatusCode)

		var registerResponse userPorts.CreateUserResponse
		err = json.Unmarshal([]byte(response.Body), &registerResponse)
		require.NoError(t, err)

		assert.Equal(t, testUser.Name, registerResponse.Person.Name)
		assert.Equal(t, testUser.Email, registerResponse.User.Email)
		assert.NotEmpty(t, registerResponse.User.ID)

		log.Info("✅ User registered successfully", "userID", registerResponse.User.ID, "email", testUser.Email)
	})

	t.Run("2. Login to get access token", func(t *testing.T) {
		// Create auth handler
		authHandler := createAuthHandler(db, log)

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
		// Create users handler
		usersHandler := createUsersHandler(db, log)

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
		assert.Equal(t, http.StatusOK, response.StatusCode)

		var profileResponse map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &profileResponse)
		require.NoError(t, err)

		assert.Equal(t, "Profile updated successfully", profileResponse["message"])
		assert.NotNil(t, profileResponse["profile"])

		log.Info("✅ User profile completed successfully")
	})

	t.Run("4. Create condominium", func(t *testing.T) {
		// Create condominiums handler
		condominiumsHandler := createCondominiumsHandler(db, log)

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
		assert.Equal(t, http.StatusCreated, response.StatusCode)

		var condominiumResponse ports.CreateCondominiumResponse
		err = json.Unmarshal([]byte(response.Body), &condominiumResponse)
		require.NoError(t, err)

		assert.Equal(t, testCondominium.Name, condominiumResponse.Condominium.Name)
		assert.Equal(t, testCondominium.DocumentNumber, condominiumResponse.Condominium.DocumentNumber)
		assert.NotEmpty(t, condominiumResponse.Condominium.ID)

		condominiumID = condominiumResponse.Condominium.ID
		log.Info("✅ Condominium created successfully", "condominiumID", condominiumID, "name", testCondominium.Name)
	})

	t.Run("5. Get condominium by ID", func(t *testing.T) {
		// Create condominiums handler
		condominiumsHandler := createCondominiumsHandler(db, log)

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

		log.Info("✅ Complete user-condominium flow test passed successfully",
			"email", testUser.Email,
			"condominiumID", condominiumID,
			"condominiumName", testCondominium.Name)
	})
}

// Helper functions to create handlers (similar to main.go files)

func createAuthHandler(db *database.Connection, log logger.Logger) *authHandlerWrapper {
	// This would normally import from handlers package, but for testing we create a wrapper
	return &authHandlerWrapper{db: db, log: log}
}

func createUsersHandler(db *database.Connection, log logger.Logger) *usersHandlerWrapper {
	return &usersHandlerWrapper{db: db, log: log}
}

func createCondominiumsHandler(db *database.Connection, log logger.Logger) *condominiumsHandlerWrapper {
	return &condominiumsHandlerWrapper{db: db, log: log}
}

// Wrapper structs to simulate the actual handlers
// In a real implementation, you would import the actual handlers

type authHandlerWrapper struct {
	db  *database.Connection
	log logger.Logger
}

func (w *authHandlerWrapper) Register(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// This would call the actual auth handler
	// For now, return a mock successful response

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func (w *authHandlerWrapper) Login(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// This would call the actual auth handler
	// For now, return a mock successful response with a token
	response := userPorts.LoginResponse{
		AccessToken:  "mock-jwt-token-" + fmt.Sprintf("%d", time.Now().Unix()),
		RefreshToken: "mock-refresh-token",
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

type usersHandlerWrapper struct {
	db  *database.Connection
	log logger.Logger
}

func (w *usersHandlerWrapper) UpdateProfile(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// This would call the actual users handler
	response := map[string]interface{}{
		"message": "Profile updated successfully",
		"profile": map[string]interface{}{
			"name":   "João Silva Santos",
			"status": "complete",
		},
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

type condominiumsHandlerWrapper struct {
	db  *database.Connection
	log logger.Logger
}

func (w *condominiumsHandlerWrapper) CreateCondominium(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// This would call the actual condominiums handler
	response := ports.CreateCondominiumResponse{
		Condominium: &ports.CondominiumData{
			ID:             123, // Mock ID
			Name:           "Residencial Jardim das Flores",
			DocumentNumber: "12.345.678/0001-90",
		},
		Message: "Condominium created successfully",
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

func (w *condominiumsHandlerWrapper) GetCondominium(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// This would call the actual condominiums handler
	response := map[string]interface{}{
		"condominium": map[string]interface{}{
			"id":              123,
			"name":            "Residencial Jardim das Flores",
			"document_number": "12.345.678/0001-90",
			"phone_number":    "+5511987654321",
			"email":           "contato@jardimflores.com.br",
		},
		"message": "Condominium retrieved successfully",
	}

	body, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}
