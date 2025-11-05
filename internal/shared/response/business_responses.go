package response

import (
	"github.com/aws/aws-lambda-go/events"
)

// Business-specific response helpers

// ProfileIncomplete returns a response indicating the user needs to complete their profile
func ProfileIncomplete(redirectPath ...string) events.APIGatewayProxyResponse {
	redirect := "/profile"
	if len(redirectPath) > 0 {
		redirect = redirectPath[0]
	}

	details := map[string]interface{}{
		"redirect": redirect,
		"action":   "complete_profile",
	}

	return CustomError(403, "profile_incomplete",
		"Please complete your profile to access this resource", details)
}

// InvalidCredentials returns a response for invalid login credentials
func InvalidCredentials() events.APIGatewayProxyResponse {
	return Unauthorized("Invalid email or password")
}

// InvalidRefreshToken returns a response for invalid refresh tokens
func InvalidRefreshToken() events.APIGatewayProxyResponse {
	return Unauthorized("Invalid refresh token")
}

// ExpiredToken returns a response for expired tokens
func ExpiredToken() events.APIGatewayProxyResponse {
	details := map[string]interface{}{
		"action": "refresh_token",
	}
	return CustomError(401, "token_expired", "Token has expired", details)
}

// ResourceAlreadyExists returns a response when trying to create a duplicate resource
func ResourceAlreadyExists(resourceType, identifier string) events.APIGatewayProxyResponse {
	details := map[string]interface{}{
		"resource_type": resourceType,
		"identifier":    identifier,
	}
	return Conflict("Resource already exists", details)
}

// InsufficientPermissions returns a response for permission denied scenarios
func InsufficientPermissions(requiredPermission ...string) events.APIGatewayProxyResponse {
	message := "You don't have permission to perform this action"
	var details interface{}

	if len(requiredPermission) > 0 {
		details = map[string]interface{}{
			"required_permission": requiredPermission[0],
		}
	}

	return Forbidden(message, details)
}

// ResourceNotFound returns a response when a requested resource is not found
func ResourceNotFound(resourceType string, identifier ...string) events.APIGatewayProxyResponse {
	message := resourceType + " not found"
	var details interface{}

	if len(identifier) > 0 {
		details = map[string]interface{}{
			"resource_type": resourceType,
			"identifier":    identifier[0],
		}
	}

	return NotFound(message, details)
}

// CondominiumNotFound returns a specific response for condominium not found
func CondominiumNotFound(condominiumID ...interface{}) events.APIGatewayProxyResponse {
	if len(condominiumID) > 0 {
		return ResourceNotFound("Condominium", condominiumID[0].(string))
	}
	return ResourceNotFound("Condominium")
}

// UserNotFound returns a specific response for user not found
func UserNotFound(userID ...interface{}) events.APIGatewayProxyResponse {
	if len(userID) > 0 {
		return ResourceNotFound("User", userID[0].(string))
	}
	return ResourceNotFound("User")
}

// EmployeeNotFound returns a specific response for employee not found
func EmployeeNotFound(employeeID ...interface{}) events.APIGatewayProxyResponse {
	if len(employeeID) > 0 {
		return ResourceNotFound("Employee", employeeID[0].(string))
	}
	return ResourceNotFound("Employee")
}

// LoginSuccess returns a successful login response
func LoginSuccess(user interface{}, tokens map[string]string) events.APIGatewayProxyResponse {
	data := map[string]interface{}{
		"user":          user,
		"access_token":  tokens["access_token"],
		"refresh_token": tokens["refresh_token"],
	}

	return OK(data, "Login successful")
}

// ProfileUpdated returns a successful profile update response
func ProfileUpdated(profile interface{}, newTokens ...map[string]string) events.APIGatewayProxyResponse {
	data := map[string]interface{}{
		"profile": profile,
	}

	// Include new tokens if profile completion triggered token refresh
	if len(newTokens) > 0 && newTokens[0] != nil {
		data["access_token"] = newTokens[0]["access_token"]
		data["refresh_token"] = newTokens[0]["refresh_token"]
	}

	return OK(data, "Profile updated successfully")
}

// ResourceCreated returns a successful resource creation response
func ResourceCreated(resourceType string, resource interface{}) events.APIGatewayProxyResponse {
	return Created(resource, resourceType+" created successfully")
}

// ResourceUpdated returns a successful resource update response
func ResourceUpdated(resourceType string, resource interface{}) events.APIGatewayProxyResponse {
	return OK(resource, resourceType+" updated successfully")
}

// ResourceDeleted returns a successful resource deletion response
func ResourceDeleted(resourceType string) events.APIGatewayProxyResponse {
	return OK(nil, resourceType+" deleted successfully")
}

// ListResponse returns a paginated list response
func ListResponse(items interface{}, total int, page, limit int, message ...string) events.APIGatewayProxyResponse {
	msg := "Items retrieved successfully"
	if len(message) > 0 {
		msg = message[0]
	}

	data := map[string]interface{}{
		"items": items,
		"meta": map[string]interface{}{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit, // Ceiling division
		},
	}

	return OK(data, msg)
}

// EmptyListResponse returns an empty list response
func EmptyListResponse(message string) events.APIGatewayProxyResponse {
	data := map[string]interface{}{
		"items": []interface{}{},
		"meta": map[string]interface{}{
			"total":       0,
			"page":        1,
			"limit":       10,
			"total_pages": 0,
		},
	}

	return OK(data, message)
}
