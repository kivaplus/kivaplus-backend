package authorization

import "github.com/kivaplus/kivaplus-backend/internal/shared/jwt"

// ExampleUsage demonstrates how to use the RoutePermissionFactory
func ExampleUsage() []RoutePermission {
	factory := NewRoutePermissionFactory()

	// Example 1: Simple public route
	factory.AddPublicRoute("GET", "^/health$", "system", "health")

	// Example 2: Admin-only route with fluent interface
	factory.
		AddRoute("DELETE", "^/admin/users/\\d+$").
		WithResource("users").
		WithAction("delete").
		WithRoles(jwt.RoleSuperAdmin).
		GlobalScope().
		Build()

	// Example 3: Complex route with multiple roles and condominium scope
	factory.
		AddRoute("POST", "^/condominios/\\d+/contracts$").
		WithResource("contracts").
		WithAction("create").
		WithRoles(jwt.RoleSuperAdmin, jwt.RoleSindico).
		CondominiumScope().
		Build()

	// Example 4: Using helper functions for common role combinations
	factory.
		AddRoute("GET", "^/reports/financial$").
		WithResource("reports").
		WithAction("read").
		WithRoles(AdminAndSindico()...).
		CondominiumScope().
		Build()

	// Example 5: Own data access
	factory.
		AddRoute("PUT", "^/users/\\w+/password$").
		WithResource("users").
		WithAction("update_password").
		WithRoles(AllRoles()...).
		OwnScope().
		Build()

	return factory.Build()
}

// CustomRoutePermissions shows how to create custom route permissions for specific modules
func CustomRoutePermissions() []RoutePermission {
	factory := NewRoutePermissionFactory()

	// Financial module routes
	factory.
		AddRoute("GET", "^/condominios/\\d+/financial/statements$").
		WithResource("financial").WithAction("read_statements").
		WithRoles(AdminAndSindico()...).CondominiumScope().Build().
		AddRoute("POST", "^/condominios/\\d+/financial/payments$").
		WithResource("financial").WithAction("create_payment").
		WithRoles(AdminSindicoMorador()...).CondominiumScope().Build().
		AddRoute("GET", "^/condominios/\\d+/financial/reports$").
		WithResource("financial").WithAction("read_reports").
		WithRoles(AdminAndSindico()...).CondominiumScope().Build()

	// Maintenance module routes
	factory.
		AddRoute("POST", "^/condominios/\\d+/maintenance/requests$").
		WithResource("maintenance").WithAction("create_request").
		WithRoles(AllRoles()...).CondominiumScope().Build().
		AddRoute("PUT", "^/condominios/\\d+/maintenance/requests/\\d+/status$").
		WithResource("maintenance").WithAction("update_status").
		WithRoles(AdminSindicoFuncionario()...).CondominiumScope().Build()

	return factory.Build()
}
