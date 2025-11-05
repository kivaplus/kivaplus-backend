package response

import (
	"github.com/aws/aws-lambda-go/events"
)

// Examples of how to use the response utilities

// ExampleSuccessResponses shows various success response patterns
func ExampleSuccessResponses() {
	// Simple success with data
	_ = OK(map[string]string{"name": "John Doe"}, "User retrieved successfully")

	// Created response
	_ = Created(map[string]interface{}{"id": 123, "name": "New User"}, "User created successfully")

	// Success without message (message is optional)
	_ = OK(map[string]string{"status": "active"})

	// No content response (for DELETE operations)
	_ = NoContent()

	// List response with pagination
	users := []map[string]interface{}{
		{"id": 1, "name": "User 1"},
		{"id": 2, "name": "User 2"},
	}
	_ = ListResponse(users, 50, 1, 10, "Users retrieved successfully")

	// Empty list
	_ = EmptyListResponse("No users found")
}

// ExampleErrorResponses shows various error response patterns
func ExampleErrorResponses() {
	// Basic error responses
	_ = BadRequest("Invalid input data")
	_ = Unauthorized("Invalid credentials")
	_ = Forbidden("Access denied")
	_ = NotFound("Resource not found")
	_ = InternalServerError("Something went wrong")

	// Error responses with details
	validationErrors := map[string][]string{
		"email":    {"Email is required", "Email format is invalid"},
		"password": {"Password must be at least 8 characters"},
	}
	_ = ValidationError(validationErrors)

	// Custom error with specific code
	_ = CustomError(418, "teapot_error", "I'm a teapot", map[string]string{"rfc": "2324"})
}

// ExampleBusinessResponses shows business-specific response patterns
func ExampleBusinessResponses() {
	// Authentication responses
	_ = InvalidCredentials()
	_ = InvalidRefreshToken()
	_ = ExpiredToken()

	// Profile responses
	_ = ProfileIncomplete("/complete-profile")

	// Resource-specific responses
	_ = CondominiumNotFound("123")
	_ = UserNotFound("456")
	_ = EmployeeNotFound("789")

	// Permission responses
	_ = InsufficientPermissions("admin access required")

	// Resource operation responses
	_ = ResourceCreated("Condominium", map[string]interface{}{"id": 1, "name": "Test Condo"})
	_ = ResourceUpdated("Employee", map[string]interface{}{"id": 1, "status": "active"})
	_ = ResourceDeleted("Vehicle")

	// Login success with tokens
	user := map[string]interface{}{"id": 1, "email": "user@example.com"}
	tokens := map[string]string{
		"access_token":  "eyJ...",
		"refresh_token": "eyJ...",
	}
	_ = LoginSuccess(user, tokens)
}

// ExampleHandlerUsage shows how to use responses in actual handlers
func ExampleHandlerUsage() events.APIGatewayProxyResponse {
	// Example: User registration handler

	// Validation failed
	if false { // some validation condition
		return BadRequest("Email is required")
	}

	// User already exists
	if false { // user exists check
		return ResourceAlreadyExists("User", "john@example.com")
	}

	// Internal error
	if false { // some error condition
		return InternalServerError("Failed to create user")
	}

	// Success case
	newUser := map[string]interface{}{
		"id":    123,
		"email": "john@example.com",
		"name":  "John Doe",
	}

	return ResourceCreated("User", newUser)
}

// ExampleWithCustomHeaders shows how to add custom headers
func ExampleWithCustomHeaders() events.APIGatewayProxyResponse {
	baseResponse := OK(map[string]string{"message": "Success"})

	customHeaders := map[string]string{
		"X-Custom-Header": "custom-value",
		"Cache-Control":   "no-cache",
	}

	return WithHeaders(baseResponse, customHeaders)
}

// ExampleErrorHandling shows different error handling patterns
func ExampleErrorHandling(err error) events.APIGatewayProxyResponse {
	if err == nil {
		return OK(map[string]string{"status": "success"})
	}

	// Handle different error types
	switch err.Error() {
	case "user not found":
		return UserNotFound()
	case "invalid credentials":
		return InvalidCredentials()
	case "profile incomplete":
		return ProfileIncomplete()
	case "validation failed":
		return UnprocessableEntity("Validation failed", err.Error())
	default:
		return InternalServerError("An unexpected error occurred")
	}
}
