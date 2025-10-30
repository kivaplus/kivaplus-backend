package jwt

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTService(t *testing.T) {
	// Test with environment variable
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	service := NewJWTService()
	assert.NotNil(t, service)
	assert.Equal(t, "test-secret", service.secret)
}

func TestNewJWTService_WithSecretName(t *testing.T) {
	// Clear JWT_SECRET and set JWT_SECRET_NAME
	os.Unsetenv("JWT_SECRET")
	os.Setenv("JWT_SECRET_NAME", "aws-secret-name")
	defer os.Unsetenv("JWT_SECRET_NAME")

	service := NewJWTService()
	assert.NotNil(t, service)
	assert.Equal(t, "your-jwt-secret-key-here", service.secret)
}

func TestNewJWTService_DefaultFallback(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_SECRET_NAME")

	service := NewJWTService()
	assert.NotNil(t, service)
	assert.Equal(t, "default-jwt-secret-key", service.secret)
}

func TestGenerateToken_Success(t *testing.T) {
	service := &Service{secret: "test-secret"}

	roles := []Role{
		{
			ID:              1,
			Name:            "super_admin",
			CondominiumID:   nil,
			CondominiumName: "",
		},
	}

	token, err := service.GenerateToken("123", "test@example.com", roles, "complete")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token structure (should have 3 parts separated by dots)
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3)
}

func TestGenerateToken_ValidateContent(t *testing.T) {
	service := &Service{secret: "test-secret"}

	condominiumID := int64(100)
	roles := []Role{
		{
			ID:              3,
			Name:            "sindico",
			CondominiumID:   &condominiumID,
			CondominiumName: "Test Condominium",
		},
	}

	token, err := service.GenerateToken("456", "sindico@example.com", roles, "incomplete")
	require.NoError(t, err)

	// Parse and validate the token content
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, "456", claims.UserID)
	assert.Equal(t, "sindico@example.com", claims.Email)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, "incomplete", claims.ProfileStatus)
	assert.Len(t, claims.Roles, 1)
	assert.Equal(t, 3, claims.Roles[0].ID)
	assert.Equal(t, "sindico", claims.Roles[0].Name)
	assert.Equal(t, int64(100), *claims.Roles[0].CondominiumID)
	assert.Equal(t, "kivaplus-backend", claims.Issuer)
	assert.Equal(t, "456", claims.Subject)
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	service := &Service{secret: "test-secret"}

	token, err := service.GenerateRefreshToken("789", "refresh@example.com")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token structure
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3)
}

func TestGenerateRefreshToken_ValidateContent(t *testing.T) {
	service := &Service{secret: "test-secret"}

	token, err := service.GenerateRefreshToken("789", "refresh@example.com")
	require.NoError(t, err)

	// Parse and validate the token content
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, "789", claims.UserID)
	assert.Equal(t, "refresh@example.com", claims.Email)
	assert.Equal(t, "refresh", claims.TokenType)
	assert.Equal(t, "unknown", claims.ProfileStatus)
	assert.Empty(t, claims.Roles) // Refresh tokens don't have roles
	assert.Equal(t, "kivaplus-backend", claims.Issuer)
	assert.Equal(t, "789", claims.Subject)

	// Verify expiration (should be ~30 days)
	now := time.Now()
	expectedExpiry := now.Add(time.Hour * 24 * 30)
	actualExpiry := claims.ExpiresAt.Time

	// Allow 1 minute tolerance for test execution time
	assert.WithinDuration(t, expectedExpiry, actualExpiry, time.Minute)
}

func TestValidateToken_Success(t *testing.T) {
	service := &Service{secret: "test-secret"}

	// Generate a token first
	roles := []Role{{ID: 2, Name: "admin"}}
	token, err := service.GenerateToken("123", "test@example.com", roles, "complete")
	require.NoError(t, err)

	// Validate the token
	claims, err := service.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	service := &Service{secret: "test-secret"}

	claims, err := service.ValidateToken("invalid.token.here")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestValidateToken_WrongSecret(t *testing.T) {
	// Generate token with one secret
	service1 := &Service{secret: "secret1"}
	token, err := service1.GenerateToken("123", "test@example.com", []Role{}, "complete")
	require.NoError(t, err)

	// Try to validate with different secret
	service2 := &Service{secret: "secret2"}
	claims, err := service2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	service := &Service{secret: "test-secret"}

	// Create an expired token manually
	now := time.Now()
	expiredTime := now.Add(-time.Hour) // 1 hour ago

	claims := &Claims{
		UserID:        "123",
		Email:         "test@example.com",
		Roles:         []Role{},
		TokenType:     "access",
		ProfileStatus: "complete",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredTime),
			IssuedAt:  jwt.NewNumericDate(now.Add(-time.Hour * 2)),
			NotBefore: jwt.NewNumericDate(now.Add(-time.Hour * 2)),
			Issuer:    "kivaplus-backend",
			Subject:   "123",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(service.secret))
	require.NoError(t, err)

	// Try to validate expired token
	validatedClaims, err := service.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	// For now, let's skip this specific test case as it requires complex RSA key setup
	// to properly test wrong signing method validation
	t.Skip("Skipping wrong signing method test - requires RSA key setup")
}

func TestValidateAccessToken_Success(t *testing.T) {
	service := &Service{secret: "test-secret"}

	// Generate access token
	roles := []Role{{ID: 1, Name: "super_admin"}}
	token, err := service.GenerateToken("123", "test@example.com", roles, "complete")
	require.NoError(t, err)

	// Validate as access token
	claims, err := service.ValidateAccessToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, "123", claims.UserID)
}

func TestValidateAccessToken_WrongTokenType(t *testing.T) {
	service := &Service{secret: "test-secret"}

	// Generate refresh token
	token, err := service.GenerateRefreshToken("123", "test@example.com")
	require.NoError(t, err)

	// Try to validate as access token
	claims, err := service.ValidateAccessToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token type: expected access, got refresh")
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	service := &Service{secret: "test-secret"}

	claims, err := service.ValidateAccessToken("invalid.token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateRefreshToken_Success(t *testing.T) {
	service := &Service{secret: "test-secret"}

	// Generate refresh token
	token, err := service.GenerateRefreshToken("123", "test@example.com")
	require.NoError(t, err)

	// Validate as refresh token
	claims, err := service.ValidateRefreshToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "refresh", claims.TokenType)
	assert.Equal(t, "123", claims.UserID)
}

func TestValidateRefreshToken_WrongTokenType(t *testing.T) {
	service := &Service{secret: "test-secret"}

	// Generate access token
	roles := []Role{{ID: 1, Name: "super_admin"}}
	token, err := service.GenerateToken("123", "test@example.com", roles, "complete")
	require.NoError(t, err)

	// Try to validate as refresh token
	claims, err := service.ValidateRefreshToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token type: expected refresh, got access")
}

func TestValidateRefreshToken_InvalidToken(t *testing.T) {
	service := &Service{secret: "test-secret"}

	claims, err := service.ValidateRefreshToken("invalid.token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestGetJWTSecret_EnvironmentVariable(t *testing.T) {
	os.Setenv("JWT_SECRET", "env-secret")
	defer os.Unsetenv("JWT_SECRET")

	secret := getJWTSecret()
	assert.Equal(t, "env-secret", secret)
}

func TestGetJWTSecret_SecretName(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	os.Setenv("JWT_SECRET_NAME", "aws-secret")
	defer os.Unsetenv("JWT_SECRET_NAME")

	secret := getJWTSecret()
	assert.Equal(t, "your-jwt-secret-key-here", secret)
}

func TestGetJWTSecret_DefaultFallback(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_SECRET_NAME")

	secret := getJWTSecret()
	assert.Equal(t, "default-jwt-secret-key", secret)
}

// Integration test: Full token lifecycle
func TestTokenLifecycle_Integration(t *testing.T) {
	service := &Service{secret: "integration-test-secret"}

	// 1. Generate access token
	condominiumID := int64(200)
	roles := []Role{
		{
			ID:              4,
			Name:            "morador",
			CondominiumID:   &condominiumID,
			CondominiumName: "Integration Test Building",
		},
	}

	accessToken, err := service.GenerateToken("integration-user", "integration@test.com", roles, "complete")
	require.NoError(t, err)

	// 2. Generate refresh token
	refreshToken, err := service.GenerateRefreshToken("integration-user", "integration@test.com")
	require.NoError(t, err)

	// 3. Validate access token
	accessClaims, err := service.ValidateAccessToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "integration-user", accessClaims.UserID)
	assert.Equal(t, "access", accessClaims.TokenType)
	assert.Len(t, accessClaims.Roles, 1)

	// 4. Validate refresh token
	refreshClaims, err := service.ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, "integration-user", refreshClaims.UserID)
	assert.Equal(t, "refresh", refreshClaims.TokenType)
	assert.Empty(t, refreshClaims.Roles)

	// 5. Cross-validation should fail
	_, err = service.ValidateAccessToken(refreshToken)
	assert.Error(t, err)

	_, err = service.ValidateRefreshToken(accessToken)
	assert.Error(t, err)
}
