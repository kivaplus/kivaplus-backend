package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

func TestMain(m *testing.M) {
	// Check if local server is running
	resp, err := http.Get(baseURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("❌ Local server is not running. Start it with: make run-local")
		os.Exit(1)
	}
	resp.Body.Close()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestUserRegistrationAndLogin(t *testing.T) {
	// Generate unique email for this test
	testEmail := fmt.Sprintf("test-%d@example.com", time.Now().Unix())
	testPassword := "password123"
	testName := "Integration Test User"

	t.Run("UserRegistration", func(t *testing.T) {
		payload := map[string]string{
			"name":             testName,
			"email":            testEmail,
			"password":         testPassword,
			"confirm_password": testPassword,
		}

		resp, body := makeRequest(t, "POST", "/register", payload)
		assert.Equal(t, 201, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")
		assert.Contains(t, response, "user")
	})

	t.Run("UserLogin", func(t *testing.T) {
		payload := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		resp, body := makeRequest(t, "POST", "/login", payload)
		assert.Equal(t, 200, resp.StatusCode)

		var response map[string]interface{}
		err := json.Unmarshal(body, &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "user")
		assert.NotEmpty(t, response["access_token"])
	})

	t.Run("DuplicateRegistration", func(t *testing.T) {
		payload := map[string]string{
			"name":             "Another User",
			"email":            testEmail, // Same email
			"password":         testPassword,
			"confirm_password": testPassword,
		}

		resp, _ := makeRequest(t, "POST", "/register", payload)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("InvalidLogin", func(t *testing.T) {
		payload := map[string]string{
			"email":    testEmail,
			"password": "wrongpassword",
		}

		resp, _ := makeRequest(t, "POST", "/login", payload)
		assert.Equal(t, 401, resp.StatusCode)
	})
}

func TestValidationErrors(t *testing.T) {
	testCases := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name: "MissingName",
			payload: map[string]string{
				"email":            "test@example.com",
				"password":         "password123",
				"confirm_password": "password123",
			},
			expectedStatus: 400,
		},
		{
			name: "InvalidEmail",
			payload: map[string]string{
				"name":             "Test User",
				"email":            "invalid-email",
				"password":         "password123",
				"confirm_password": "password123",
			},
			expectedStatus: 400,
		},
		{
			name: "PasswordMismatch",
			payload: map[string]string{
				"name":             "Test User",
				"email":            "test@example.com",
				"password":         "password123",
				"confirm_password": "different123",
			},
			expectedStatus: 400,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := makeRequest(t, "POST", "/register", tc.payload)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func makeRequest(t *testing.T, method, endpoint string, payload interface{}) (*http.Response, []byte) {
	var body bytes.Buffer
	if payload != nil {
		err := json.NewEncoder(&body).Encode(payload)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, &body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	responseBody := make([]byte, resp.ContentLength)
	if resp.ContentLength > 0 {
		resp.Body.Read(responseBody)
	}
	resp.Body.Close()

	return resp, responseBody
}
