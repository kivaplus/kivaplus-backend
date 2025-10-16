package types

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUser struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type User struct {
	Email     string `json:"email"`
	SenhaHash string `json:"senha"`
}

func NewUser(registerUser RegisterUser) (User, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(registerUser.Senha), 10)
	if err != nil {
		return User{}, err
	}

	return User{
		Email:     registerUser.Email,
		SenhaHash: string(hashPassword),
	}, nil
}

func ValidatePassword(hashPassword, plainTextPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(plainTextPassword))
	return err == nil
}

func CreateToken(user User) string {
	now := time.Now()
	validUntil := now.Add(time.Hour * 1).Unix()

	claims := jwt.MapClaims{
		"user":    user.Email,
		"expires": validUntil,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims, nil)

	// Get JWT secret from AWS Secrets Manager or fallback to env var
	secret := getJWTSecret()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}

	return tokenString
}

// getJWTSecret retrieves the JWT secret, with fallback to environment variable
func getJWTSecret() string {
	// First try to get from environment variable (for local development)
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}

	// For production, you should use AWS Secrets Manager
	// For now, use a default secret (NOT RECOMMENDED FOR PRODUCTION)
	secretName := os.Getenv("JWT_SECRET_NAME")
	if secretName != "" {
		// This is a placeholder - in production you should retrieve from Secrets Manager
		return "your-jwt-secret-key-here"
	}

	// Fallback to a default (NOT SECURE - only for development)
	return "default-jwt-secret-key"
}
