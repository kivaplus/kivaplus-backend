package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kivaplus/kivaplus-backend/internal/condominiums/ports"
	"github.com/kivaplus/kivaplus-backend/internal/shared/cache"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
	"github.com/kivaplus/kivaplus-backend/internal/shared/permissions"
	"github.com/kivaplus/kivaplus-backend/internal/shared/response"
)

// HTTPHandler handles HTTP requests for condominium operations
type HTTPHandler struct {
	createCondominiumService ports.CreateCondominiumService
	listCondominiumsService  ports.ListCondominiumsService
	getCondominiumService    ports.GetCondominiumService
	permissionChecker        *permissions.EnhancedChecker
	permissionService        *permissions.Service
	cache                    *cache.RedisService
	logger                   logger.Logger
}

// NewHTTPHandler creates a new condominium handler with enhanced permissions
func NewHTTPHandler(
	createCondominiumService ports.CreateCondominiumService,
	listCondominiumsService ports.ListCondominiumsService,
	getCondominiumService ports.GetCondominiumService,
	permissionChecker *permissions.EnhancedChecker,
	permissionService *permissions.Service,
	cache *cache.RedisService,
	logger logger.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		createCondominiumService: createCondominiumService,
		listCondominiumsService:  listCondominiumsService,
		getCondominiumService:    getCondominiumService,
		permissionChecker:        permissionChecker,
		permissionService:        permissionService,
		cache:                    cache,
		logger:                   logger,
	}
}

// CreateCondominium handles POST /condominios with optimized validation
func (h *HTTPHandler) CreateCondominium(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing create condominium request")

	// Validate request with complete profile requirement
	validator := h.permissionChecker.NewValidator(ctx, request)
	enhancedClaims, err := validator.RequireCompleteProfile()
	if err != nil {
		h.logger.Error("Profile validation failed", "error", err)
		if err.Error() == "profile must be completed" {
			return response.ProfileIncomplete(), nil
		}
		return response.Unauthorized("Access denied"), nil
	}

	userID, err := enhancedClaims.GetUserID()
	if err != nil {
		h.logger.Error("Invalid user ID in token", "error", err)
		return response.Unauthorized("Invalid token"), nil
	}

	// Parse request body
	var req ports.CreateCondominiumRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		h.logger.Error("Invalid request body", "error", err)
		return response.BadRequest("Invalid request body"), nil
	}

	// Execute condominium creation
	result, err := h.createCondominiumService.Execute(ctx, userID, req)
	if err != nil {
		h.logger.Error("Failed to create condominium", "error", err, "user_id", userID)

		// Handle specific error types
		switch err.Error() {
		case "user profile not found":
			return response.ProfileIncomplete("/profile"), nil
		case "condominium already exists":
			return response.ResourceAlreadyExists("Condominium", req.DocumentNumber), nil
		default:
			return response.InternalServerError("Failed to create condominium"), nil
		}
	}

	// Invalidate user permissions cache after condominium creation
	if err := h.permissionService.InvalidateUserPermissions(ctx, enhancedClaims.Claims.UserID); err != nil {
		h.logger.Warn("Failed to invalidate user permissions cache", "userID", userID, "error", err)
	}

	h.logger.Info("Condominium created successfully", "condominium_id", result.Condominium.ID, "user_id", userID)

	return response.ResourceCreated("Condominium", result), nil
}

// ListCondominiums handles GET /condominios with cached permissions
func (h *HTTPHandler) ListCondominiums(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing list condominiums request")

	// Validate request with permission checking
	validator := h.permissionChecker.NewValidator(ctx, request)
	enhancedClaims, err := validator.RequirePermission("condominios", "list", nil)
	if err != nil {
		h.logger.Error("Permission validation failed", "error", err)
		return response.Unauthorized("Access denied"), nil
	}

	userID, err := enhancedClaims.GetUserID()
	if err != nil {
		h.logger.Error("Invalid user ID in token", "error", err)
		return response.Unauthorized("Invalid token"), nil
	}

	// Execute list condominiums
	result, err := h.listCondominiumsService.Execute(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to list condominiums", "error", err, "user_id", userID)
		return response.InternalServerError("Failed to retrieve condominiums"), nil
	}

	h.logger.Info("Listed condominiums successfully", "user_id", userID, "count", len(result.Condominiums))
	return response.OK(result, "Condominiums retrieved successfully"), nil
}

// GetCondominium handles GET /condominios/:id with enhanced permission checking
func (h *HTTPHandler) GetCondominium(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Processing get condominium request")

	// Get condominium ID from path parameters
	condominiumID, err := h.getCondominiumIDFromPath(request.PathParameters)
	if err != nil {
		h.logger.Error("Invalid condominium ID in path", "error", err)
		return response.BadRequest("Invalid condominium ID"), nil
	}

	// Validate request with condominium-specific permission
	validator := h.permissionChecker.NewValidator(ctx, request)
	enhancedClaims, err := validator.RequireCondominiumAccess(condominiumID)
	if err != nil {
		h.logger.Error("Condominium access validation failed", "error", err, "condominiumID", condominiumID)
		return response.InsufficientPermissions("condominium access required"), nil
	}

	userID, err := enhancedClaims.GetUserID()
	if err != nil {
		h.logger.Error("Invalid user ID in token", "error", err)
		return response.Unauthorized("Invalid token"), nil
	}

	// Execute get condominium
	result, err := h.getCondominiumService.Execute(ctx, userID, condominiumID)
	if err != nil {
		h.logger.Error("Failed to get condominium", "error", err, "user_id", userID, "condominium_id", condominiumID)

		if err.Error() == "condominium not found" {
			return response.CondominiumNotFound(), nil
		}

		return response.InternalServerError("Failed to retrieve condominium"), nil
	}

	h.logger.Info("Retrieved condominium successfully", "user_id", userID, "condominium_id", condominiumID)
	return response.OK(result, "Condominium retrieved successfully"), nil
}

// ListUnits handles GET /condominios/{id}/units
func (h *HTTPHandler) ListUnits(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement list units handler
	return response.NotImplemented("List units not implemented yet"), nil
}

// CreateUnit handles POST /condominios/{id}/units
func (h *HTTPHandler) CreateUnit(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement create unit handler
	return response.NotImplemented("Create unit not implemented yet"), nil
}

// GetUnit handles GET /condominios/{id}/units/{unit_id}
func (h *HTTPHandler) GetUnit(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement get unit handler
	return response.NotImplemented("Get unit not implemented yet"), nil
}

// UpdateUnit handles PUT /condominios/{id}/units/{unit_id}
func (h *HTTPHandler) UpdateUnit(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement update unit handler
	return response.NotImplemented("Update unit not implemented yet"), nil
}

// DeleteUnit handles DELETE /condominios/{id}/units/{unit_id}
func (h *HTTPHandler) DeleteUnit(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement delete unit handler
	return response.NotImplemented("Delete unit not implemented yet"), nil
}

// ListVehicles handles GET /condominios/{id}/vehicles
func (h *HTTPHandler) ListVehicles(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement list vehicles handler
	return response.NotImplemented("List vehicles not implemented yet"), nil
}

// CreateVehicle handles POST /condominios/{id}/vehicles
func (h *HTTPHandler) CreateVehicle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement create vehicle handler
	return response.NotImplemented("Create vehicle not implemented yet"), nil
}

// ListParkingSpaces handles GET /condominios/{id}/parking-spaces
func (h *HTTPHandler) ListParkingSpaces(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement list parking spaces handler
	return response.NotImplemented("List parking spaces not implemented yet"), nil
}

// ListResidents handles GET /condominios/{id}/residents
func (h *HTTPHandler) ListResidents(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Implement list residents handler
	return response.NotImplemented("List residents not implemented yet"), nil
}

// Helper methods

// getCondominiumIDFromPath extracts condominium ID from path parameters
func (h *HTTPHandler) getCondominiumIDFromPath(pathParameters map[string]string) (int64, error) {
	condominiumIDStr, exists := pathParameters["condominium_id"]
	if !exists {
		condominiumIDStr, exists = pathParameters["id"]
		if !exists {
			return 0, fmt.Errorf("condominium ID not found in path")
		}
	}

	condominiumID, err := strconv.ParseInt(condominiumIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid condominium ID format: %w", err)
	}

	return condominiumID, nil
}

// Old helper methods removed - now using shared response utilities
