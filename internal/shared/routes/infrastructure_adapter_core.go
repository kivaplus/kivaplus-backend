package routes

// InfrastructureAdapter adapts route configuration for infrastructure deployment
// This version doesn't depend on AWS CDK to avoid import issues in CLI tools
type InfrastructureAdapter struct {
	config *RoutesConfiguration
}

// NewInfrastructureAdapter creates a new infrastructure adapter
func NewInfrastructureAdapter(config *RoutesConfiguration) *InfrastructureAdapter {
	return &InfrastructureAdapter{config: config}
}

// APIRouteInfo represents route information for infrastructure setup
type APIRouteInfo struct {
	Path      string
	Method    string
	Lambda    string // Lambda function name
	Protected bool
	Name      string
}

// LambdaMapping represents which Lambda function handles which routes
type LambdaMapping struct {
	Name   string
	Routes []RouteConfig
}

// GetAPIRouteInfos returns routes formatted for API Gateway creation
func (a *InfrastructureAdapter) GetAPIRouteInfos() []APIRouteInfo {
	var apiRoutes []APIRouteInfo

	for _, route := range a.config.GetAllRoutes() {
		// Skip health routes or other routes that don't need API Gateway setup
		if route.Lambda == "" {
			continue
		}

		apiRoutes = append(apiRoutes, APIRouteInfo{
			Path:      route.Path,
			Method:    route.Method,
			Lambda:    route.Lambda,
			Protected: !route.Public,
			Name:      route.Name,
		})
	}

	return apiRoutes
}

// GetLambdaMappings returns which routes each Lambda should handle
func (a *InfrastructureAdapter) GetLambdaMappings() map[string]LambdaMapping {
	mappings := make(map[string]LambdaMapping)

	for _, route := range a.config.GetAllRoutes() {
		if route.Lambda == "" {
			continue
		}

		if mapping, exists := mappings[route.Lambda]; exists {
			mapping.Routes = append(mapping.Routes, route)
			mappings[route.Lambda] = mapping
		} else {
			mappings[route.Lambda] = LambdaMapping{
				Name:   route.Lambda,
				Routes: []RouteConfig{route},
			}
		}
	}

	return mappings
}

// GetPublicAPIRouteInfos returns only public routes for API Gateway
func (a *InfrastructureAdapter) GetPublicAPIRouteInfos() []APIRouteInfo {
	var apiRoutes []APIRouteInfo

	for _, route := range a.config.GetPublicRoutes() {
		if route.Lambda == "" {
			continue
		}

		apiRoutes = append(apiRoutes, APIRouteInfo{
			Path:      route.Path,
			Method:    route.Method,
			Lambda:    route.Lambda,
			Protected: false,
			Name:      route.Name,
		})
	}

	return apiRoutes
}

// GetProtectedAPIRouteInfos returns only protected routes for API Gateway
func (a *InfrastructureAdapter) GetProtectedAPIRouteInfos() []APIRouteInfo {
	var apiRoutes []APIRouteInfo

	for _, route := range a.config.GetProtectedRoutes() {
		if route.Lambda == "" {
			continue
		}

		apiRoutes = append(apiRoutes, APIRouteInfo{
			Path:      route.Path,
			Method:    route.Method,
			Lambda:    route.Lambda,
			Protected: true,
			Name:      route.Name,
		})
	}

	return apiRoutes
}

// GetRequiredLambdaFunctions returns a list of Lambda functions that need to be created
func (a *InfrastructureAdapter) GetRequiredLambdaFunctions() []string {
	lambdaSet := make(map[string]bool)

	for _, route := range a.config.GetAllRoutes() {
		if route.Lambda != "" {
			lambdaSet[route.Lambda] = true
		}
	}

	var lambdas []string
	for lambda := range lambdaSet {
		lambdas = append(lambdas, lambda)
	}

	return lambdas
}

// ValidateConfiguration validates that all routes have valid Lambda mappings
func (a *InfrastructureAdapter) ValidateConfiguration() []string {
	var errors []string

	requiredLambdas := a.GetRequiredLambdaFunctions()
	if len(requiredLambdas) == 0 {
		errors = append(errors, "no Lambda functions defined in routes")
	}

	// Add more validation as needed
	for _, route := range a.config.GetAllRoutes() {
		if !route.Public && len(route.Roles) == 0 {
			errors = append(errors, "protected route "+route.Name+" has no roles defined")
		}
	}

	return errors
}

// GenerateInfrastructureCode generates code snippets for infrastructure setup
func (a *InfrastructureAdapter) GenerateInfrastructureCode() map[string]string {
	code := make(map[string]string)

	// Generate public routes setup
	publicRoutes := a.GetPublicAPIRouteInfos()
	publicCode := "// Public routes setup\n"
	for _, route := range publicRoutes {
		publicCode += "CreateLambdaIntegration(api, \"" + route.Path + "\", " + route.Lambda + "Lambda, \"" + route.Method + "\")\n"
	}
	code["public_routes"] = publicCode

	// Generate protected routes setup
	protectedRoutes := a.GetProtectedAPIRouteInfos()
	protectedCode := "// Protected routes setup\n"
	for _, route := range protectedRoutes {
		protectedCode += "CreateProtectedLambdaIntegration(api, \"" + route.Path + "\", " + route.Lambda + "Lambda, \"" + route.Method + "\", authorizer)\n"
	}
	code["protected_routes"] = protectedCode

	return code
}
