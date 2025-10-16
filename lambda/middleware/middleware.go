package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
)

func ValidatePasswordJWTMiddleware(next func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		tokenString := extractTokenFromHeaders(request.Headers)
		if tokenString == "" {
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("Missing Auth Token. Headers: %v", request.Headers),
				StatusCode: http.StatusUnauthorized,
			}, nil
		}

		claims, err := parseToken(tokenString)
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("Unauthorized: %v. Token: %s", err, tokenString[:20]+"..."),
				StatusCode: http.StatusUnauthorized,
			}, nil
		}

		expires := int64(claims["expires"].(float64))

		if time.Now().Unix() > expires {
			return events.APIGatewayProxyResponse{
				Body:       "Token Expired",
				StatusCode: http.StatusUnauthorized,
			}, nil
		}

		return next(request)

	}
}

func extractTokenFromHeaders(headers map[string]string) string {
	authHeader, ok := headers["Authorization"]
	if !ok {
		return ""
	}

	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}

	return splitToken[1]
}

func parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		secret := getJWTSecret()
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("unauthorized")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("claims in unauthorized type")
	}

	return claims, nil
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
