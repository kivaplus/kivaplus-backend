package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

// checkServerRunning checks if local server is running
func checkServerRunning(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skip("⚠️  Local server is not running. Start with: make run-local")
		return
	}
	resp.Body.Close()
}

// TestHTTPUserRegistrationAndLogin tests the complete flow via HTTP
func TestHTTPUserRegistrationAndLogin(t *testing.T) {
	checkServerRunning(t)

	// Generate unique email for this test
	testEmail := fmt.Sprintf("test-%d@example.com", time.Now().Unix())
	testPassword := "password123"
	testName := "HTTP Integration Test User"

	t.Run("UserRegistration", func(t *testing.T) {
		payload := map[string]string{
			"name":             testName,
			"email":            testEmail,
			"password":         testPassword,
			"confirm_password": testPassword,
		}

		resp, body := makeHTTPRequest(t, "POST", "/register", payload)
		assert.Equal(t, 201, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		require.NoError(t, err)

		assert.Contains(t, response, "data")
		assert.Contains(t, response, "message")

		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "user")
		assert.Contains(t, data, "person")
	})

	t.Run("UserLogin", func(t *testing.T) {
		payload := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		resp, body := makeHTTPRequest(t, "POST", "/login", payload)
		assert.Equal(t, 200, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		require.NoError(t, err)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "access_token")
		assert.Contains(t, data, "refresh_token")
		assert.Contains(t, data, "user")

		// Verify token is not empty
		accessToken := data["access_token"].(string)
		assert.NotEmpty(t, accessToken)
		assert.True(t, len(accessToken) > 50) // JWT tokens are typically longer
	})

	t.Run("ProfileAccess", func(t *testing.T) {
		// First login to get token
		loginPayload := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		resp, body := makeHTTPRequest(t, "POST", "/login", loginPayload)
		require.Equal(t, 200, resp.StatusCode)

		var loginResponse map[string]interface{}
		err := json.Unmarshal(body, &loginResponse)
		require.NoError(t, err)

		data := loginResponse["data"].(map[string]interface{})
		accessToken := data["access_token"].(string)

		// Now access profile
		resp, body = makeAuthenticatedHTTPRequest(t, "GET", "/profile", nil, accessToken)
		assert.Equal(t, 200, resp.StatusCode)

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

// makeHTTPRequest is a helper function to make HTTP requests
func makeHTTPRequest(t *testing.T, method, endpoint string, payload interface{}) (*http.Response, []byte) {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, bytes.NewBuffer(body))
	require.NoError(t, err)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	responseBody := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			responseBody = append(responseBody, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return resp, responseBody
}

// makeAuthenticatedHTTPRequest is a helper function to make authenticated HTTP requests
func makeAuthenticatedHTTPRequest(t *testing.T, method, endpoint string, payload interface{}, token string) (*http.Response, []byte) {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, bytes.NewBuffer(body))
	require.NoError(t, err)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	responseBody := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			responseBody = append(responseBody, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return resp, responseBody
}
