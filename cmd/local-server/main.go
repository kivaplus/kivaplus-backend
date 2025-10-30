package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joho/godotenv"
	"github.com/kivaplus/kivaplus-backend/internal/shared/authorization"
	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/users/handlers"
	"github.com/kivaplus/kivaplus-backend/internal/users/repository"
	"github.com/kivaplus/kivaplus-backend/internal/users/usecases"
)

var (
	userHandler  *handlers.HTTPHandler
	jwtService   *jwt.Service
	authzService *authorization.AuthorizationServiceV2
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// Initialize logger
	logger := logger.New()

	// Initialize database connection
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserPostgresRepository(db)
	personRepo := repository.NewPersonPostgresRepository(db)
	roleRepo := repository.NewRolePostgresRepository(db)
	addressRepo := repository.NewAddressPostgresRepository(db)

	// Initialize use cases
	createUserUC := usecases.NewCreateUser(userRepo, personRepo, roleRepo, logger)
	loginUC := usecases.NewLogin(userRepo, personRepo, roleRepo, logger)
	refreshTokenUC := usecases.NewRefreshToken(userRepo, personRepo, roleRepo, logger)
	completeProfileUC := usecases.NewCompleteProfile(userRepo, personRepo, addressRepo, logger)

	// Initialize JWT service
	jwtService = jwt.NewJWTService()

	// Initialize authorization service (uses centralized routes configuration)
	authzService, err = authorization.NewAuthorizationServiceV2("configs/routes.json")
	if err != nil {
		log.Fatalf("❌ Failed to initialize authorization service: %v", err)
	}

	// Initialize handler
	userHandler = handlers.NewHTTPHandler(createUserUC, loginUC, refreshTokenUC, completeProfileUC, logger)

	log.Println("✅ Local server initialized successfully")
	log.Println("🔐 Authorization system loaded with centralized routes configuration")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup routes with authentication middleware (matching centralized configuration)
	http.HandleFunc("/register", authHandler(handleRegister))     // Public route
	http.HandleFunc("/login", authHandler(handleLogin))           // Public route
	http.HandleFunc("/token/refresh", authHandler(handleRefresh)) // Public route
	http.HandleFunc("/profile", authHandler(handleProfile))       // Protected route
	http.HandleFunc("/usuarios", authHandler(handleUsers))        // Protected route (matches config)
	http.HandleFunc("/health", authHandler(handleHealth))         // Public route

	log.Printf("🚀 Local server starting on port %s", port)
	log.Printf("📍 Available endpoints:")
	log.Printf("   POST http://localhost:%s/register (public)", port)
	log.Printf("   POST http://localhost:%s/login (public)", port)
	log.Printf("   POST http://localhost:%s/token/refresh (public)", port)
	log.Printf("   GET  http://localhost:%s/profile (🔐 protected)", port)
	log.Printf("   PUT  http://localhost:%s/profile (🔐 protected)", port)
	log.Printf("   GET  http://localhost:%s/usuarios (🔐 protected)", port)
	log.Printf("   GET  http://localhost:%s/health (public)", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

// corsHandler adds CORS headers to all responses
func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// authHandler adds authentication middleware using centralized routes configuration
func authHandler(next http.HandlerFunc) http.HandlerFunc {
	return corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Check if route is public using centralized configuration
		isPublic := isPublicRoute(r.Method, r.URL.Path)

		if isPublic {
			log.Printf("🌐 Public route accessed: %s %s", r.Method, r.URL.Path)
			next(w, r)
			return
		}

		// Protected route - require authentication
		log.Printf("🔐 Protected route accessed: %s %s", r.Method, r.URL.Path)

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("❌ Missing Authorization header for %s %s", r.Method, r.URL.Path)
			http.Error(w, `{"error":"Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			log.Printf("❌ Invalid Authorization header format for %s %s", r.Method, r.URL.Path)
			http.Error(w, `{"error":"Invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}

		// Validate JWT token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			log.Printf("❌ Invalid token for %s %s: %v", r.Method, r.URL.Path, err)
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Check if profile is incomplete and redirect to complete profile
		if claims.ProfileStatus == "incomplete" && !isProfileCompletionRoute(r.Method, r.URL.Path) {
			log.Printf("⚠️ User has incomplete profile, should complete profile first (user: %s)", claims.Email)
			http.Error(w, `{"error":"Profile incomplete. Please complete your profile first.","redirect":"/profile"}`, http.StatusForbidden)
			return
		}

		// Check permissions using centralized authorization
		hasPermission, err := authzService.CheckPermission(r.Method, r.URL.Path, claims)
		if err != nil || !hasPermission {
			log.Printf("❌ Access denied for %s %s: %v", r.Method, r.URL.Path, err)
			http.Error(w, `{"error":"Access denied"}`, http.StatusForbidden)
			return
		}

		log.Printf("✅ Access granted for %s %s (user: %s)", r.Method, r.URL.Path, claims.Email)

		// Add claims to request context for use in handlers
		ctx := context.WithValue(r.Context(), "claims", claims)
		r = r.WithContext(ctx)

		next(w, r)
	})
}

// isPublicRoute checks if a route is public using the centralized configuration
func isPublicRoute(method, path string) bool {
	// Use the authorization service to check if route is public
	// This uses the centralized routes configuration

	// For now, hardcode the known public routes, but this should use the centralized config
	publicRoutes := map[string][]string{
		"POST": {"/register", "/login", "/token/refresh"},
		"GET":  {"/health"},
	}

	if routes, exists := publicRoutes[method]; exists {
		for _, route := range routes {
			if path == route {
				return true
			}
		}
	}

	return false
}

// Convert HTTP request to Lambda event format
func httpToLambdaEvent(r *http.Request) events.APIGatewayProxyRequest {
	// Read body
	body := ""
	if r.Body != nil {
		bodyBytes := make([]byte, r.ContentLength)
		r.Body.Read(bodyBytes)
		body = string(bodyBytes)
	}

	// Convert headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return events.APIGatewayProxyRequest{
		HTTPMethod: r.Method,
		Path:       r.URL.Path,
		Headers:    headers,
		Body:       body,
	}
}

// Convert Lambda response to HTTP response
func lambdaToHTTPResponse(w http.ResponseWriter, response events.APIGatewayProxyResponse) {
	// Set headers
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	// Set status code
	w.WriteHeader(response.StatusCode)

	// Write body
	w.Write([]byte(response.Body))
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	event := httpToLambdaEvent(r)
	response, err := userHandler.Register(context.Background(), event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lambdaToHTTPResponse(w, response)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	event := httpToLambdaEvent(r)
	response, err := userHandler.Login(context.Background(), event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lambdaToHTTPResponse(w, response)
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	event := httpToLambdaEvent(r)
	var response events.APIGatewayProxyResponse
	var err error

	switch r.Method {
	case "GET":
		response, err = userHandler.GetProfile(context.Background(), event)
	case "PUT":
		response, err = userHandler.UpdateProfile(context.Background(), event)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lambdaToHTTPResponse(w, response)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	event := httpToLambdaEvent(r)
	response, err := userHandler.ListUsers(context.Background(), event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lambdaToHTTPResponse(w, response)
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	event := httpToLambdaEvent(r)
	response, err := userHandler.RefreshToken(context.Background(), event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lambdaToHTTPResponse(w, response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// isProfileCompletionRoute checks if a route is for profile completion
func isProfileCompletionRoute(method, path string) bool {
	profileRoutes := map[string][]string{
		"GET": {"/profile"},
		"PUT": {"/profile"},
	}

	if routes, exists := profileRoutes[method]; exists {
		for _, route := range routes {
			if path == route {
				return true
			}
		}
	}

	return false
}
