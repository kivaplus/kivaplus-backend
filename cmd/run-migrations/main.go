package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/kivaplus/kivaplus-backend/internal/database"
)

type SecretData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handler(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
	log.Printf("Event Type: %s", event.RequestType)

	if event.RequestType == cfn.RequestCreate || event.RequestType == cfn.RequestUpdate {
		if err := runMigrations(ctx, event); err != nil {
			return "", nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return "MigrationComplete", map[string]interface{}{
		"Status": "Success",
	}, nil
}

func runMigrations(ctx context.Context, event cfn.Event) error {
	// Obter credenciais
	secretArn := os.Getenv("DB_SECRET_ARN")
	dbName := os.Getenv("DB_NAME")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	secretsClient := secretsmanager.NewFromConfig(cfg)
	result, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretArn,
	})
	if err != nil {
		return err
	}

	var secret SecretData
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return err
	}

	// Construir DATABASE_URL com URL encoding
	endpoint := event.ResourceProperties["DatabaseEndpoint"].(string)
	encodedUsername := url.QueryEscape(secret.Username)
	encodedPassword := url.QueryEscape(secret.Password)

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:5432/%s?sslmode=require",
		encodedUsername,
		encodedPassword,
		endpoint,
		dbName,
	)

	// Executar migrations
	log.Println("🚀 Starting database migrations...")

	if err := database.RunMigrations(database.MigrationConfig{
		DatabaseURL: databaseURL,
	}); err != nil {
		return err
	}

	log.Println("✅ Migrations completed successfully!")
	return nil
}

func main() {
	lambda.Start(cfn.LambdaWrap(handler))
}
