package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/kivaplus/kivaplus-backend/lambda/types"
	_ "github.com/lib/pq"
)

type PostgresClient struct {
	db *sql.DB
}

// NewPostgresClient creates a new PostgreSQL client that implements UserStore interface
func NewPostgresClient() (*PostgresClient, error) {
	databaseURL := os.Getenv("DATABASE_URL")

	// If DATABASE_URL is not set, try to build it from individual components
	if databaseURL == "" {
		databaseURL = buildDatabaseURLFromComponents()
		if databaseURL == "" {
			return nil, fmt.Errorf("DATABASE_URL environment variable not set and cannot build from components")
		}
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresClient{db: db}, nil
}

// Close closes the database connection
func (p *PostgresClient) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// DoesUserExist checks if a user exists by email (username)
func (p *PostgresClient) DoesUserExist(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM usuario WHERE email = $1 AND ativo = true AND deleted_at IS NULL)`

	err := p.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

// InsertUser creates a new user in the usuario table
func (p *PostgresClient) InsertUser(user types.User) error {
	query := `
		INSERT INTO usuario (email, senha_hash, ativo, created_at, updated_at)
		VALUES ($1, $2, true, NOW(), NOW())
	`

	_, err := p.db.Exec(query, user.Email, user.SenhaHash)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by email (username)
func (p *PostgresClient) GetUser(email string) (types.User, error) {
	var user types.User
	query := `
		SELECT email, senha_hash
		FROM usuario
		WHERE email = $1 AND ativo = true AND deleted_at IS NULL
	`

	err := p.db.QueryRow(query, email).Scan(&user.Email, &user.SenhaHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("user not found")
		}
		return user, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateLastAccess updates the ultimo_acesso field for a user
func (p *PostgresClient) UpdateLastAccess(email string) error {
	query := `
		UPDATE usuario
		SET ultimo_acesso = NOW(), updated_at = NOW()
		WHERE email = $1 AND ativo = true AND deleted_at IS NULL
	`

	result, err := p.db.Exec(query, email)
	if err != nil {
		return fmt.Errorf("failed to update last access: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or inactive")
	}

	return nil
}

// DeactivateUser soft deletes a user by setting ativo = false
func (p *PostgresClient) DeactivateUser(email string) error {
	query := `
		UPDATE usuario
		SET ativo = false, updated_at = NOW()
		WHERE email = $1 AND deleted_at IS NULL
	`

	result, err := p.db.Exec(query, email)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser performs a soft delete by setting deleted_at timestamp
func (p *PostgresClient) DeleteUser(email string) error {
	query := `
		UPDATE usuario
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE email = $1 AND deleted_at IS NULL
	`

	result, err := p.db.Exec(query, email)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// buildDatabaseURLFromComponents builds a DATABASE_URL from individual environment variables
func buildDatabaseURLFromComponents() string {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbPort == "" || dbName == "" {
		return ""
	}

	// Try to get credentials from Secrets Manager
	dbSecretName := os.Getenv("DB_SECRET_NAME")
	if dbSecretName != "" {
		username, password, err := getCredentialsFromSecretsManager(dbSecretName)
		if err == nil && username != "" && password != "" {
			// URL encode the password to handle special characters
			encodedPassword := url.QueryEscape(password)
			return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require",
				username, encodedPassword, dbHost, dbPort, dbName)
		}
	}

	// Fallback: try to get from direct environment variables (less secure)
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbUser != "" && dbPassword != "" {
		encodedPassword := url.QueryEscape(dbPassword)
		return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require",
			dbUser, encodedPassword, dbHost, dbPort, dbName)
	}

	return ""
}

// getCredentialsFromSecretsManager retrieves database credentials from AWS Secrets Manager
func getCredentialsFromSecretsManager(secretName string) (string, string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	result, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret: %w", err)
	}

	if result.SecretString == nil {
		return "", "", fmt.Errorf("secret string is nil")
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err = json.Unmarshal([]byte(*result.SecretString), &credentials)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	return credentials.Username, credentials.Password, nil
}
