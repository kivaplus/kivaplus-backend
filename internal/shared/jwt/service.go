package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/golang-jwt/jwt/v5"
)

// Service handles JWT operations with AWS integration and caching
type Service struct {
	secret         string
	secretsManager *secretsmanager.Client
	paramStore     *ssm.Client
	secretCache    *sync.Map
	cacheMutex     *sync.RWMutex
	cacheExpiry    time.Time
}

// Singleton pattern for JWT service
var (
	jwtServiceInstance *Service
	jwtServiceOnce     sync.Once
)

// GetJWTService returns a singleton JWT service instance
func GetJWTService() *Service {
	jwtServiceOnce.Do(func() {
		jwtServiceInstance = newJWTService()
	})
	return jwtServiceInstance
}

// NewJWTService creates a new JWT service (for backward compatibility)
func NewJWTService() *Service {
	return GetJWTService()
}

// newJWTService creates a new JWT service with AWS integration
func newJWTService() *Service {
	service := &Service{
		secretCache: &sync.Map{},
		cacheMutex:  &sync.RWMutex{},
	}

	// Initialize AWS clients if in AWS environment
	if isAWSEnvironment() {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err == nil {
			service.secretsManager = secretsmanager.NewFromConfig(cfg)
			service.paramStore = ssm.NewFromConfig(cfg)
		}
	}

	// Get initial secret
	service.secret = service.getJWTSecret()

	return service
}

// GenerateToken generates a new JWT token with minimal claims (no roles)
func (s *Service) GenerateToken(userID, userEmail string, profileStatus string, tokenVersion int64) (string, error) {
	now := time.Now()
	validUntil := now.Add(time.Hour * 1) // Shorter-lived tokens (1 hour)

	claims := &Claims{
		UserID:        userID,
		Email:         userEmail,
		TokenType:     "access",
		ProfileStatus: profileStatus,
		TokenVersion:  tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(validUntil),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kivaplus-backend",
			Subject:   userID,
		},
	}

	secret := s.getJWTSecretCached()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Note: GenerateTokenWithRoles method removed - use GenerateToken + EnhancedChecker instead

// GenerateRefreshToken generates a refresh token
func (s *Service) GenerateRefreshToken(userID, userEmail string) (string, error) {
	now := time.Now()
	validUntil := now.Add(time.Hour * 24 * 30) // 30 days validity

	claims := &Claims{
		UserID:        userID,
		Email:         userEmail,
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
	secret := s.getJWTSecretCached()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
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

// ExtractUserIDFromAuthHeader extracts user ID from Authorization header
func (s *Service) ExtractUserIDFromAuthHeader(authHeader string) (int64, error) {
	claims, err := s.ExtractClaimsFromAuthHeader(authHeader)
	if err != nil {
		return 0, err
	}

	userID, err := strconv.ParseInt(claims.UserID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return userID, nil
}

// ExtractClaimsFromAuthHeader extracts full claims from Authorization header
func (s *Service) ExtractClaimsFromAuthHeader(authHeader string) (*Claims, error) {
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Extract token from "Bearer <token>" format
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		return nil, fmt.Errorf("invalid authorization header format, expected 'Bearer <token>'")
	}

	tokenString := tokenParts[1]

	// Validate and parse the JWT token
	claims, err := s.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	return claims, nil
}

// ExtractUserIDFromRequest extracts user ID from APIGatewayProxyRequest
func (s *Service) ExtractUserIDFromRequest(request events.APIGatewayProxyRequest) (int64, error) {
	authHeader := s.getAuthHeaderFromRequest(request)
	return s.ExtractUserIDFromAuthHeader(authHeader)
}

// ExtractClaimsFromRequest extracts full claims from APIGatewayProxyRequest
func (s *Service) ExtractClaimsFromRequest(request events.APIGatewayProxyRequest) (*Claims, error) {
	authHeader := s.getAuthHeaderFromRequest(request)
	return s.ExtractClaimsFromAuthHeader(authHeader)
}

// getAuthHeaderFromRequest gets the Authorization header from APIGatewayProxyRequest, handling case variations
func (s *Service) getAuthHeaderFromRequest(request events.APIGatewayProxyRequest) string {
	// Try standard case first
	if authHeader := request.Headers["Authorization"]; authHeader != "" {
		return authHeader
	}
	// Try lowercase (some clients send it this way)
	return request.Headers["authorization"]
}

// getJWTSecretCached retrieves JWT secret with caching
func (s *Service) getJWTSecretCached() string {
	s.cacheMutex.RLock()
	if time.Now().Before(s.cacheExpiry) && s.secret != "" {
		secret := s.secret
		s.cacheMutex.RUnlock()
		return secret
	}
	s.cacheMutex.RUnlock()

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if time.Now().Before(s.cacheExpiry) && s.secret != "" {
		return s.secret
	}

	// Refresh secret
	s.secret = s.getJWTSecret()
	s.cacheExpiry = time.Now().Add(5 * time.Minute) // Cache for 5 minutes

	return s.secret
}

// getJWTSecret retrieves the JWT secret from various sources
func (s *Service) getJWTSecret() string {
	// Try environment variable first
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}

	// Try AWS Secrets Manager
	if secretName := os.Getenv("JWT_SECRET_NAME"); secretName != "" && s.secretsManager != nil {
		if secret := s.getSecretFromAWS(secretName); secret != "" {
			return secret
		}
	}

	// Try AWS Parameter Store
	if paramName := os.Getenv("JWT_SECRET_PARAM"); paramName != "" && s.paramStore != nil {
		if secret := s.getParameterFromAWS(paramName); secret != "" {
			return secret
		}
	}

	// Fallback (NOT SECURE - only for development)
	return "default-jwt-secret-key-for-development-only"
}

// getSecretFromAWS retrieves secret from AWS Secrets Manager
func (s *Service) getSecretFromAWS(secretName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.secretsManager.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return ""
	}

	if result.SecretString != nil {
		// Try to parse as JSON first
		var secretData map[string]string
		if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err == nil {
			if secret, ok := secretData["jwt_secret"]; ok {
				return secret
			}
		}
		// Return as plain string
		return *result.SecretString
	}

	return ""
}

// getParameterFromAWS retrieves parameter from AWS Parameter Store
func (s *Service) getParameterFromAWS(paramName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.paramStore.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return ""
	}

	if result.Parameter != nil && result.Parameter.Value != nil {
		return *result.Parameter.Value
	}

	return ""
}

// isAWSEnvironment checks if running in AWS environment
func isAWSEnvironment() bool {
	return os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" ||
		os.Getenv("AWS_EXECUTION_ENV") != "" ||
		os.Getenv("AWS_REGION") != ""
}
