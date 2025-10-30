package routes

// getDefaultRoutesConfiguration returns the default route configuration
// This serves as both the fallback and the template for customization
func getDefaultRoutesConfiguration() *RoutesConfiguration {
	return &RoutesConfiguration{
		Version: "1.0.0",
		Groups: []RouteGroup{
			{
				Name:        "authentication",
				Description: "Public authentication routes",
				Routes: []RouteConfig{
					{
						Name:        "register",
						Method:      "POST",
						Path:        "/register",
						PathPattern: "^/register$",
						Public:      true,
						Lambda:      "auth",
						Description: "User registration",
					},
					{
						Name:        "login",
						Method:      "POST",
						Path:        "/login",
						PathPattern: "^/login$",
						Public:      true,
						Lambda:      "auth",
						Description: "User login",
					},
				},
			},
			{
				Name:        "health",
				Description: "Health check routes",
				Routes: []RouteConfig{
					{
						Name:        "health",
						Method:      "GET",
						Path:        "/health",
						PathPattern: "^/health$",
						Public:      true,
						Lambda:      "auth",
						Description: "Health check endpoint",
					},
				},
			},
			{
				Name:        "profile",
				Description: "User profile management",
				Routes: []RouteConfig{
					{
						Name:        "get_profile",
						Method:      "GET",
						Path:        "/profile",
						PathPattern: "^/profile$",
						Public:      false,
						Roles:       []int{1, 2, 3, 4}, // All roles
						Resource:    "profile",
						Action:      "read",
						Scope:       "own",
						Lambda:      "users",
						Description: "Get user profile",
					},
					{
						Name:        "update_profile",
						Method:      "PUT",
						Path:        "/profile",
						PathPattern: "^/profile$",
						Public:      false,
						Roles:       []int{1, 2, 3, 4}, // All roles
						Resource:    "profile",
						Action:      "update",
						Scope:       "own",
						Lambda:      "users",
						Description: "Update user profile",
					},
				},
			},
			{
				Name:        "users",
				Description: "User management routes",
				Routes: []RouteConfig{
					{
						Name:        "list_users",
						Method:      "GET",
						Path:        "/usuarios",
						PathPattern: "^/usuarios$",
						Public:      false,
						Roles:       []int{1, 2}, // Admin and Sindico
						Resource:    "usuarios",
						Action:      "list",
						Scope:       "global",
						Lambda:      "users",
						Description: "List all users",
					},
					{
						Name:        "get_user",
						Method:      "GET",
						Path:        "/usuarios/{id}",
						PathPattern: "^/usuarios/\\w+$",
						Public:      false,
						Roles:       []int{1, 2}, // Admin and Sindico
						Resource:    "usuarios",
						Action:      "read",
						Scope:       "own",
						Lambda:      "users",
						Description: "Get specific user",
					},
					{
						Name:        "update_user",
						Method:      "PUT",
						Path:        "/usuarios/{id}",
						PathPattern: "^/usuarios/\\w+$",
						Public:      false,
						Roles:       []int{1, 2, 3, 4}, // All roles
						Resource:    "usuarios",
						Action:      "update",
						Scope:       "own",
						Lambda:      "users",
						Description: "Update user",
					},
				},
			},
			{
				Name:        "condominiums",
				Description: "Condominium management routes",
				Routes: []RouteConfig{
					{
						Name:        "list_condominiums",
						Method:      "GET",
						Path:        "/condominios",
						PathPattern: "^/condominios$",
						Public:      false,
						Roles:       []int{1}, // Admin only
						Resource:    "condominios",
						Action:      "list",
						Scope:       "global",
						Lambda:      "users",
						Description: "List all condominiums",
					},
					{
						Name:        "create_condominium",
						Method:      "POST",
						Path:        "/condominios",
						PathPattern: "^/condominios$",
						Public:      false,
						Roles:       []int{1}, // Admin only
						Resource:    "condominios",
						Action:      "create",
						Scope:       "global",
						Lambda:      "users",
						Description: "Create new condominium",
					},
					{
						Name:        "get_condominium",
						Method:      "GET",
						Path:        "/condominios/{id}",
						PathPattern: "^/condominios/\\d+$",
						Public:      false,
						Roles:       []int{1, 2, 3, 4}, // All roles
						Resource:    "condominios",
						Action:      "read",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "Get specific condominium",
					},
					{
						Name:        "update_condominium",
						Method:      "PUT",
						Path:        "/condominios/{id}",
						PathPattern: "^/condominios/\\d+$",
						Public:      false,
						Roles:       []int{1, 2}, // Admin and Sindico
						Resource:    "condominios",
						Action:      "update",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "Update condominium",
					},
				},
			},
			{
				Name:        "units",
				Description: "Unit management routes",
				Routes: []RouteConfig{
					{
						Name:        "list_units",
						Method:      "GET",
						Path:        "/condominios/{id}/unidades",
						PathPattern: "^/condominios/\\d+/unidades$",
						Public:      false,
						Roles:       []int{1, 2, 3, 4}, // All roles
						Resource:    "unidades",
						Action:      "list",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "List units in condominium",
					},
					{
						Name:        "create_unit",
						Method:      "POST",
						Path:        "/condominios/{id}/unidades",
						PathPattern: "^/condominios/\\d+/unidades$",
						Public:      false,
						Roles:       []int{1, 2}, // Admin and Sindico
						Resource:    "unidades",
						Action:      "create",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "Create unit in condominium",
					},
				},
			},
			{
				Name:        "residents",
				Description: "Resident management routes",
				Routes: []RouteConfig{
					{
						Name:        "list_residents",
						Method:      "GET",
						Path:        "/condominios/{id}/moradores",
						PathPattern: "^/condominios/\\d+/moradores$",
						Public:      false,
						Roles:       []int{1, 2, 3}, // Admin, Sindico, Funcionario
						Resource:    "moradores",
						Action:      "list",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "List residents in condominium",
					},
					{
						Name:        "create_resident",
						Method:      "POST",
						Path:        "/condominios/{id}/moradores",
						PathPattern: "^/condominios/\\d+/moradores$",
						Public:      false,
						Roles:       []int{1, 2}, // Admin and Sindico
						Resource:    "moradores",
						Action:      "create",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "Create resident in condominium",
					},
				},
			},
			{
				Name:        "employees",
				Description: "Employee management routes",
				Routes: []RouteConfig{
					{
						Name:        "list_employees",
						Method:      "GET",
						Path:        "/condominios/{id}/funcionarios",
						PathPattern: "^/condominios/\\d+/funcionarios$",
						Public:      false,
						Roles:       []int{1, 2}, // Admin and Sindico
						Resource:    "funcionarios",
						Action:      "list",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "List employees in condominium",
					},
					{
						Name:        "create_employee",
						Method:      "POST",
						Path:        "/condominios/{id}/funcionarios",
						PathPattern: "^/condominios/\\d+/funcionarios$",
						Public:      false,
						Roles:       []int{1, 2}, // Admin and Sindico
						Resource:    "funcionarios",
						Action:      "create",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "Create employee in condominium",
					},
				},
			},
			{
				Name:        "vehicles",
				Description: "Vehicle management routes",
				Routes: []RouteConfig{
					{
						Name:        "list_vehicles",
						Method:      "GET",
						Path:        "/condominios/{id}/veiculos",
						PathPattern: "^/condominios/\\d+/veiculos$",
						Public:      false,
						Roles:       []int{1, 2, 3}, // Admin, Sindico, Funcionario
						Resource:    "veiculos",
						Action:      "list",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "List vehicles in condominium",
					},
					{
						Name:        "create_vehicle",
						Method:      "POST",
						Path:        "/condominios/{id}/veiculos",
						PathPattern: "^/condominios/\\d+/veiculos$",
						Public:      false,
						Roles:       []int{1, 2, 4}, // Admin, Sindico, Morador
						Resource:    "veiculos",
						Action:      "create",
						Scope:       "condominium",
						Lambda:      "users",
						Description: "Create vehicle in condominium",
					},
				},
			},
		},
	}
}
