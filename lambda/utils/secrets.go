package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// GetJWTSecret retrieves the JWT secret from AWS Secrets Manager
func GetJWTSecret() (string, error) {
	secretName := os.Getenv("JWT_SECRET_NAME")
	if secretName == "" {
		return "", fmt.Errorf("JWT_SECRET_NAME environment variable not set")
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Get the secret
	result, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	// Parse the secret value
	var secretData map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		// If it's not JSON, assume it's a plain string
		return *result.SecretString, nil
	}

	// Look for common JWT secret keys
	if secret, ok := secretData["jwt_secret"]; ok {
		return secret, nil
	}
	if secret, ok := secretData["JWT_SECRET"]; ok {
		return secret, nil
	}
	if secret, ok := secretData["secret"]; ok {
		return secret, nil
	}

	// If no specific key found, return the raw secret string
	return *result.SecretString, nil
}

// Cache for JWT secret to avoid multiple AWS calls
var jwtSecretCache string

// GetCachedJWTSecret returns the cached JWT secret or retrieves it if not cached
func GetCachedJWTSecret() (string, error) {
	if jwtSecretCache == "" {
		secret, err := GetJWTSecret()
		if err != nil {
			return "", err
		}
		jwtSecretCache = secret
	}
	return jwtSecretCache, nil
}
