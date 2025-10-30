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
	_ "github.com/lib/pq"
)

// Connection represents a database connection
type Connection struct {
	DB *sql.DB
}

// DatabaseCredentials represents the structure of database credentials in AWS Secrets Manager
type DatabaseCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection() (*Connection, error) {
	// Try DATABASE_URL first (for local development)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		return connectWithURL(databaseURL)
	}

	// Use AWS Secrets Manager for production
	return connectWithAWSSecrets()
}

// connectWithURL connects using a direct DATABASE_URL
func connectWithURL(databaseURL string) (*Connection, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Connection{DB: db}, nil
}

// connectWithAWSSecrets connects using AWS Secrets Manager
func connectWithAWSSecrets() (*Connection, error) {
	ctx := context.Background()

	// Get environment variables
	dbSecretName := os.Getenv("DB_SECRET_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbSecretName == "" || dbHost == "" || dbPort == "" || dbName == "" {
		return nil, fmt.Errorf("missing required environment variables: DB_SECRET_NAME=%s, DB_HOST=%s, DB_PORT=%s, DB_NAME=%s",
			dbSecretName, dbHost, dbPort, dbName)
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Secrets Manager client
	secretsClient := secretsmanager.NewFromConfig(cfg)

	// Get database credentials from Secrets Manager
	result, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &dbSecretName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get database credentials from Secrets Manager (secret: %s): %w", dbSecretName, err)
	}

	// Parse credentials
	var creds DatabaseCredentials
	if err := json.Unmarshal([]byte(*result.SecretString), &creds); err != nil {
		return nil, fmt.Errorf("failed to parse database credentials: %w", err)
	}

	// Validate credentials
	if creds.Username == "" || creds.Password == "" {
		return nil, fmt.Errorf("invalid database credentials: username or password is empty")
	}

	// Build connection string with proper URL encoding
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		url.QueryEscape(creds.Username),
		url.QueryEscape(creds.Password),
		dbHost,
		dbPort,
		dbName,
	)

	// Log connection attempt (without password)
	fmt.Printf("Attempting to connect to database: postgres://%s:***@%s:%s/%s\n",
		creds.Username, dbHost, dbPort, dbName)

	// Connect to database
	return connectWithURL(databaseURL)
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// BeginTx starts a new transaction
func (c *Connection) BeginTx() (*sql.Tx, error) {
	return c.DB.Begin()
}
