package test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/handlers"
	"github.com/kivaplus/kivaplus-backend/internal/users/repository"
	"github.com/kivaplus/kivaplus-backend/internal/users/usecases"
)

// IntegrationTestSuite contains the test setup
type IntegrationTestSuite struct {
	handler *handlers.HTTPHandler
	db      *database.Connection
}

// SetupIntegrationTest sets up the test environment
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Set test database URL - you should use a separate test database
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		// Fallback to a default test database URL
		testDatabaseURL = "postgres://postgres:password@localhost:5432/kivaplus_test?sslmode=disable"
	}

	// Set the DATABASE_URL for the test
	os.Setenv("DATABASE_URL", testDatabaseURL)

	// Set JWT secret for tests
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-for-integration-tests")
	}

	// Setup test database connection
	db, err := database.NewPostgresConnection()
	require.NoError(t, err)

	// Setup repositories
	userRepo := repository.NewUserPostgresRepository(db)
	personRepo := repository.NewPersonPostgresRepository(db)
	addressRepo := repository.NewAddressPostgresRepository(db)
	roleRepo := repository.NewRolePostgresRepository(db)

	// Setup logger
	logger := logger.New()

	// Setup use cases
	createUserUC := usecases.NewCreateUser(userRepo, personRepo, roleRepo, logger)
	loginUC := usecases.NewLogin(userRepo, personRepo, roleRepo, logger)
	completeProfileUC := usecases.NewCompleteProfile(userRepo, personRepo, addressRepo, logger)
	refreshTokenUC := usecases.NewRefreshToken(userRepo, personRepo, roleRepo, logger)

	// Setup handler
	handler := handlers.NewHTTPHandler(createUserUC, loginUC, refreshTokenUC, completeProfileUC, logger)

	return &IntegrationTestSuite{
		handler: handler,
		db:      db,
	}
}

func TestIntegration_UserRegistrationFlow(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.db.Close()

	// Test 1: Register a new user
	t.Run("Register User", func(t *testing.T) {
		requestBody := map[string]string{
			"name":             "Integration Test User",
			"email":            "integration@test.com",
			"password":         "password123",
			"confirm_password": "password123",
		}

		body, _ := json.Marshal(requestBody)
		request := events.APIGatewayProxyRequest{
			Body: string(body),
		}

		response, err := suite.handler.Register(context.Background(), request)

		assert.NoError(t, err)
		assert.Equal(t, 201, response.StatusCode)
		assert.Contains(t, response.Body, "User created successfully")

		// Parse response to get user data
		var responseData map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &responseData)
		assert.NoError(t, err)
		assert.Contains(t, responseData, "user")
	})

	// Test 2: Login with the created user
	t.Run("Login User", func(t *testing.T) {
		requestBody := map[string]string{
			"email":    "integration@test.com",
			"password": "password123",
		}

		body, _ := json.Marshal(requestBody)
		request := events.APIGatewayProxyRequest{
			Body: string(body),
		}

		response, err := suite.handler.Login(context.Background(), request)

		assert.NoError(t, err)
		assert.Equal(t, 200, response.StatusCode)

		// Parse response to get access token
		var responseData map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &responseData)
		assert.NoError(t, err)
		assert.Contains(t, responseData, "access_token")
		assert.NotEmpty(t, responseData["access_token"])
	})

	// Test 3: Try to register duplicate user
	t.Run("Duplicate Registration", func(t *testing.T) {
		requestBody := map[string]string{
			"name":             "Another User",
			"email":            "integration@test.com", // Same email
			"password":         "password123",
			"confirm_password": "password123",
		}

		body, _ := json.Marshal(requestBody)
		request := events.APIGatewayProxyRequest{
			Body: string(body),
		}

		response, err := suite.handler.Register(context.Background(), request)

		assert.NoError(t, err)
		assert.Equal(t, 400, response.StatusCode)
		assert.Contains(t, response.Body, "already exists")
	})

	// Test 4: Invalid login
	t.Run("Invalid Login", func(t *testing.T) {
		requestBody := map[string]string{
			"email":    "integration@test.com",
			"password": "wrongpassword",
		}

		body, _ := json.Marshal(requestBody)
		request := events.APIGatewayProxyRequest{
			Body: string(body),
		}

		response, err := suite.handler.Login(context.Background(), request)

		assert.NoError(t, err)
		assert.Equal(t, 401, response.StatusCode)
		assert.Contains(t, response.Body, "Invalid credentials")
	})
}

func TestIntegration_ValidationErrors(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.db.Close()

	testCases := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Missing Name",
			requestBody: map[string]string{
				"email":            "test@example.com",
				"password":         "password123",
				"confirm_password": "password123",
			},
			expectedStatus: 400,
			expectedError:  "validation failed",
		},
		{
			name: "Invalid Email",
			requestBody: map[string]string{
				"name":             "Test User",
				"email":            "invalid-email",
				"password":         "password123",
				"confirm_password": "password123",
			},
			expectedStatus: 400,
			expectedError:  "validation failed",
		},
		{
			name: "Password Mismatch",
			requestBody: map[string]string{
				"name":             "Test User",
				"email":            "test@example.com",
				"password":         "password123",
				"confirm_password": "different123",
			},
			expectedStatus: 400,
			expectedError:  "passwords do not match",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.requestBody)
			request := events.APIGatewayProxyRequest{
				Body: string(body),
			}

			response, err := suite.handler.Register(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, response.StatusCode)
			assert.Contains(t, response.Body, tc.expectedError)
		})
	}
}
