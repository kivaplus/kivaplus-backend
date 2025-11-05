package response

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// defaultHeaders returns the standard CORS headers
func defaultHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Origin, Content-Type, Authorization",
	}
}

// buildResponse creates an APIGatewayProxyResponse with the given parameters
func buildResponse(statusCode int, body interface{}, headers map[string]string) events.APIGatewayProxyResponse {
	if headers == nil {
		headers = defaultHeaders()
	} else {
		// Merge with default headers
		defaultHdrs := defaultHeaders()
		for k, v := range defaultHdrs {
			if _, exists := headers[k]; !exists {
				headers[k] = v
			}
		}
	}

	bodyBytes, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(bodyBytes),
	}
}

// Success responses

// OK returns a 200 OK response
func OK(data interface{}, message ...string) events.APIGatewayProxyResponse {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	response := SuccessResponse{
		Data:    data,
		Message: msg,
	}
	return buildResponse(http.StatusOK, response, nil)
}

// Created returns a 201 Created response
func Created(data interface{}, message ...string) events.APIGatewayProxyResponse {
	msg := "Resource created successfully"
	if len(message) > 0 {
		msg = message[0]
	}

	response := SuccessResponse{
		Data:    data,
		Message: msg,
	}
	return buildResponse(http.StatusCreated, response, nil)
}

// NoContent returns a 204 No Content response
func NoContent() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers:    defaultHeaders(),
		Body:       "",
	}
}

// Error responses

// BadRequest returns a 400 Bad Request response
func BadRequest(message string, details ...interface{}) events.APIGatewayProxyResponse {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	response := ErrorResponse{
		Error:   "bad_request",
		Message: message,
		Details: detail,
		Code:    "BAD_REQUEST",
	}
	return buildResponse(http.StatusBadRequest, response, nil)
}

// Unauthorized returns a 401 Unauthorized response
func Unauthorized(message string, details ...interface{}) events.APIGatewayProxyResponse {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	response := ErrorResponse{
		Error:   "unauthorized",
		Message: message,
		Details: detail,
		Code:    "UNAUTHORIZED",
	}
	return buildResponse(http.StatusUnauthorized, response, nil)
}

// Forbidden returns a 403 Forbidden response
func Forbidden(message string, details ...interface{}) events.APIGatewayProxyResponse {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	response := ErrorResponse{
		Error:   "forbidden",
		Message: message,
		Details: detail,
		Code:    "FORBIDDEN",
	}
	return buildResponse(http.StatusForbidden, response, nil)
}

// NotFound returns a 404 Not Found response
func NotFound(message string, details ...interface{}) events.APIGatewayProxyResponse {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	response := ErrorResponse{
		Error:   "not_found",
		Message: message,
		Details: detail,
		Code:    "NOT_FOUND",
	}
	return buildResponse(http.StatusNotFound, response, nil)
}

// Conflict returns a 409 Conflict response
func Conflict(message string, details ...interface{}) events.APIGatewayProxyResponse {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	response := ErrorResponse{
		Error:   "conflict",
		Message: message,
		Details: detail,
		Code:    "CONFLICT",
	}
	return buildResponse(http.StatusConflict, response, nil)
}

// UnprocessableEntity returns a 422 Unprocessable Entity response
func UnprocessableEntity(message string, details ...interface{}) events.APIGatewayProxyResponse {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	response := ErrorResponse{
		Error:   "validation_failed",
		Message: message,
		Details: detail,
		Code:    "VALIDATION_FAILED",
	}
	return buildResponse(http.StatusUnprocessableEntity, response, nil)
}

// InternalServerError returns a 500 Internal Server Error response
func InternalServerError(message string, details ...interface{}) events.APIGatewayProxyResponse {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	response := ErrorResponse{
		Error:   "internal_error",
		Message: message,
		Details: detail,
		Code:    "INTERNAL_ERROR",
	}
	return buildResponse(http.StatusInternalServerError, response, nil)
}

// NotImplemented returns a 501 Not Implemented response
func NotImplemented(message string) events.APIGatewayProxyResponse {
	response := ErrorResponse{
		Error:   "not_implemented",
		Message: message,
		Code:    "NOT_IMPLEMENTED",
	}
	return buildResponse(http.StatusNotImplemented, response, nil)
}

// Custom error response with specific error code and status
func CustomError(statusCode int, errorCode, message string, details ...interface{}) events.APIGatewayProxyResponse {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	response := ErrorResponse{
		Error:   errorCode,
		Message: message,
		Details: detail,
		Code:    errorCode,
	}
	return buildResponse(statusCode, response, nil)
}

// WithHeaders allows adding custom headers to any response
func WithHeaders(response events.APIGatewayProxyResponse, headers map[string]string) events.APIGatewayProxyResponse {
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}

	for k, v := range headers {
		response.Headers[k] = v
	}

	return response
}

// ValidationError creates a structured validation error response
func ValidationError(validationErrors map[string][]string) events.APIGatewayProxyResponse {
	response := ErrorResponse{
		Error:   "validation_failed",
		Message: "The request contains validation errors",
		Details: validationErrors,
		Code:    "VALIDATION_FAILED",
	}
	return buildResponse(http.StatusBadRequest, response, nil)
}
