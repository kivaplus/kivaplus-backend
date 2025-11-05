package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const apiURL = "http://localhost:4566/restapis/4gk5dposvy/local/_user_request_"

// TestSimpleIntegrationFlow tests the basic flow via API Gateway
func TestSimpleIntegrationFlow(t *testing.T) {
	// Check if LocalStack is running
	resp, err := http.Get("http://localhost:4566/_localstack/health")
	if err != nil {
		t.Skip("⚠️  LocalStack is not running. Start with: make setup")
		return
	}
	resp.Body.Close()

	// Generate unique test data
	timestamp := time.Now().Unix()
	testEmail := fmt.Sprintf("test-integration-%d@example.com", timestamp)
	testPassword := "SecurePass123!"
	testName := "Integration Test User"

	t.Run("UserRegistration", func(t *testing.T) {
		payload := fmt.Sprintf(`{
			"name": "%s",
			"email": "%s",
			"password": "%s",
			"confirm_password": "%s"
		}`, testName, testEmail, testPassword, testPassword)

		resp, err := http.Post(apiURL+"/register", "application/json", strings.NewReader(payload))
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, 201, resp.StatusCode, "Registration should succeed. Response: %s", string(body))

		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)

		assert.Contains(t, response, "data")
		assert.Contains(t, response, "message")

		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "user")
		assert.Contains(t, data, "person")
	})

	t.Run("UserLogin", func(t *testing.T) {
		payload := fmt.Sprintf(`{
			"email": "%s",
			"password": "%s"
		}`, testEmail, testPassword)

		resp, err := http.Post(apiURL+"/login", "application/json", strings.NewReader(payload))
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode, "Login should succeed. Response: %s", string(body))

		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "access_token")
		assert.Contains(t, data, "refresh_token")
		assert.Contains(t, data, "user")

		// Verify token is not empty
		accessToken := data["access_token"].(string)
		assert.NotEmpty(t, accessToken)
		assert.True(t, len(accessToken) > 50, "JWT token should be substantial length")

		// Store token for next test
		t.Setenv("TEST_ACCESS_TOKEN", accessToken)
	})

	t.Run("ProfileAccess", func(t *testing.T) {
		// First get token from login
		payload := fmt.Sprintf(`{
			"email": "%s",
			"password": "%s"
		}`, testEmail, testPassword)

		resp, err := http.Post(apiURL+"/login", "application/json", strings.NewReader(payload))
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var loginResponse map[string]interface{}
		err = json.Unmarshal(body, &loginResponse)
		require.NoError(t, err)

		data := loginResponse["data"].(map[string]interface{})
		accessToken := data["access_token"].(string)

		// Now access profile
		req, err := http.NewRequest("GET", apiURL+"/profile", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode, "Profile access should succeed. Response: %s", string(body))

		var profileResponse map[string]interface{}
		err = json.Unmarshal(body, &profileResponse)
		require.NoError(t, err)

		assert.Contains(t, profileResponse, "data")
		profileData := profileResponse["data"].(map[string]interface{})
		assert.Contains(t, profileData, "name")
		assert.Contains(t, profileData, "email")
		assert.Equal(t, testName, profileData["name"])
		assert.Equal(t, testEmail, profileData["email"])
	})
}
