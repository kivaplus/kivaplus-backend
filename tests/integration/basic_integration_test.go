package integration

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
	handler *handlers.AuthHandler
	db      *database.Connection
}

// SetupIntegrationTest sets up the test environment
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Use the same database as the running system
	testDatabaseURL := os.Getenv("DATABASE_URL")
	if testDatabaseURL == "" {
		// Fallback to the standard KivaPlus database URL
		testDatabaseURL = "postgres://kivaplus:kivaplus123@localhost:5432/kivaplus_local?sslmode=disable"
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
	roleRepo := repository.NewRolePostgresRepository(db)

	// Setup use cases
	log := logger.New()
	createUserUC := usecases.NewCreateUser(userRepo, personRepo, roleRepo, log)
	loginUC := usecases.NewLogin(userRepo, personRepo, roleRepo, log)
	refreshTokenUC := usecases.NewRefreshToken(userRepo, personRepo, roleRepo, log)

	// Setup handler
	authHandler := handlers.NewAuthHandler(createUserUC, loginUC, refreshTokenUC, log)

	return &IntegrationTestSuite{
		handler: authHandler,
		db:      db,
	}
}

// TestBasicUserRegistration tests basic user registration functionality
func TestBasicUserRegistration(t *testing.T) {
	suite := SetupIntegrationTest(t)

	// Test data
	requestBody := map[string]interface{}{
		"name":             "Test User",
		"email":            "test@example.com",
		"password":         "password123",
		"confirm_password": "password123",
	}

	bodyBytes, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Create Lambda event
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/register",
		Body:       string(bodyBytes),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// Execute handler
	response, err := suite.handler.Register(context.Background(), event)
	require.NoError(t, err)

	// Assertions
	assert.Equal(t, 201, response.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "data")
	assert.Contains(t, responseBody, "message")
}

// TestBasicUserLogin tests basic user login functionality
func TestBasicUserLogin(t *testing.T) {
	suite := SetupIntegrationTest(t)

	// First register a user
	registerBody := map[string]interface{}{
		"name":             "Login Test User",
		"email":            "login@example.com",
		"password":         "password123",
		"confirm_password": "password123",
	}

	registerBytes, err := json.Marshal(registerBody)
	require.NoError(t, err)

	registerEvent := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/register",
		Body:       string(registerBytes),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	_, err = suite.handler.Register(context.Background(), registerEvent)
	require.NoError(t, err)

	// Now test login
	loginBody := map[string]interface{}{
		"email":    "login@example.com",
		"password": "password123",
	}

	loginBytes, err := json.Marshal(loginBody)
	require.NoError(t, err)

	loginEvent := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/login",
		Body:       string(loginBytes),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// Execute login
	response, err := suite.handler.Login(context.Background(), loginEvent)
	require.NoError(t, err)

	// Assertions
	assert.Equal(t, 200, response.StatusCode)

	var responseBody map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &responseBody)
	require.NoError(t, err)

	assert.Contains(t, responseBody, "data")
	data := responseBody["data"].(map[string]interface{})
	assert.Contains(t, data, "access_token")
	assert.Contains(t, data, "refresh_token")
}
