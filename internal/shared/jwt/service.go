package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service handles JWT operations
type Service struct {
	secret string
}

// NewJWTService creates a new JWT service
func NewJWTService() *Service {
	secret := getJWTSecret()
	return &Service{
		secret: secret,
	}
}

// GenerateToken generates a new JWT token for a user with roles
func (s *Service) GenerateToken(userID, userEmail string, roles []Role, profileStatus string) (string, error) {
	now := time.Now()
	validUntil := now.Add(time.Hour * 24) // 24 hours validity

	claims := &Claims{
		UserID:        userID,
		Email:         userEmail,
		Roles:         roles,
		TokenType:     "access",
		ProfileStatus: profileStatus,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(validUntil),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kivaplus-backend",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token
func (s *Service) GenerateRefreshToken(userID, userEmail string) (string, error) {
	now := time.Now()
	validUntil := now.Add(time.Hour * 24 * 30) // 30 days validity

	claims := &Claims{
		UserID:        userID,
		Email:         userEmail,
		Roles:         []Role{}, // Refresh tokens don't need roles
		TokenType:     "refresh",
		ProfileStatus: "unknown", // Refresh tokens don't need profile status
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(validUntil),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kivaplus-backend",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ValidateAccessToken validates specifically an access token
func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.TokenType)
	}

	return claims, nil
}

// ValidateRefreshToken validates specifically a refresh token
func (s *Service) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}

	return claims, nil
}

// getJWTSecret retrieves the JWT secret from environment
func getJWTSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}

	// For production, retrieve from AWS Secrets Manager
	if secretName := os.Getenv("JWT_SECRET_NAME"); secretName != "" {
		// TODO: Implement AWS Secrets Manager retrieval
		return "your-jwt-secret-key-here"
	}

	// Fallback (NOT SECURE - only for development)
	return "default-jwt-secret-key"
}
