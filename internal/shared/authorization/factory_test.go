package authorization

import (
	"testing"

	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
)

func TestRoutePermissionFactory(t *testing.T) {
	factory := NewRoutePermissionFactory()

	// Test public route
	factory.AddPublicRoute("GET", "^/health$", "system", "health")

	// Test protected route
	factory.
		AddRoute("GET", "^/users$").
		WithResource("users").
		WithAction("list").
		WithRoles(jwt.RoleSuperAdmin).
		GlobalScope().
		Build()

	// Test complex route with multiple roles
	factory.
		AddRoute("POST", "^/condominios/\\d+/unidades$").
		WithResource("unidades").
		WithAction("create").
		WithRoles(AdminAndSindico()...).
		CondominiumScope().
		Build()

	permissions := factory.Build()

	// Verify we have the expected number of permissions
	if len(permissions) != 3 {
		t.Errorf("Expected 3 permissions, got %d", len(permissions))
	}

	// Test public route
	publicRoute := permissions[0]
	if publicRoute.Method != "GET" {
		t.Errorf("Expected method GET, got %s", publicRoute.Method)
	}
	if publicRoute.PathPattern != "^/health$" {
		t.Errorf("Expected path ^/health$, got %s", publicRoute.PathPattern)
	}
	if len(publicRoute.Permission.Roles) != 0 {
		t.Errorf("Expected 0 roles for public route, got %d", len(publicRoute.Permission.Roles))
	}

	// Test protected route
	protectedRoute := permissions[1]
	if protectedRoute.Permission.Resource != "users" {
		t.Errorf("Expected resource users, got %s", protectedRoute.Permission.Resource)
	}
	if protectedRoute.Permission.Action != "list" {
		t.Errorf("Expected action list, got %s", protectedRoute.Permission.Action)
	}
	if len(protectedRoute.Permission.Roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(protectedRoute.Permission.Roles))
	}
	if protectedRoute.Permission.Roles[0] != int(jwt.RoleSuperAdmin) {
		t.Errorf("Expected role %d, got %d", int(jwt.RoleSuperAdmin), protectedRoute.Permission.Roles[0])
	}

	// Test complex route
	complexRoute := permissions[2]
	if complexRoute.Permission.Scope != "condominio" {
		t.Errorf("Expected scope condominio, got %s", complexRoute.Permission.Scope)
	}
	if len(complexRoute.Permission.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(complexRoute.Permission.Roles))
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test AdminOnly
	adminRoles := AdminOnly()
	if len(adminRoles) != 1 {
		t.Errorf("Expected 1 role in AdminOnly, got %d", len(adminRoles))
	}
	if adminRoles[0] != jwt.RoleSuperAdmin {
		t.Errorf("Expected RoleSuperAdmin, got %v", adminRoles[0])
	}

	// Test AdminAndSindico
	adminSindicoRoles := AdminAndSindico()
	if len(adminSindicoRoles) != 2 {
		t.Errorf("Expected 2 roles in AdminAndSindico, got %d", len(adminSindicoRoles))
	}

	// Test AllRoles
	allRoles := AllRoles()
	if len(allRoles) != 4 {
		t.Errorf("Expected 4 roles in AllRoles, got %d", len(allRoles))
	}
}

func TestFluentInterface(t *testing.T) {
	factory := NewRoutePermissionFactory()

	// Test method chaining
	result := factory.
		AddRoute("GET", "^/test$").
		WithResource("test").
		WithAction("read").
		WithRoles(jwt.RoleSuperAdmin).
		GlobalScope().
		Build().
		AddRoute("POST", "^/test$").
		WithResource("test").
		WithAction("create").
		WithRoles(jwt.RoleSuperAdmin, jwt.RoleSindico).
		CondominiumScope().
		Build()

	// Verify the factory is returned for chaining
	if result != factory {
		t.Error("Expected fluent interface to return the same factory instance")
	}

	permissions := factory.Build()
	if len(permissions) != 2 {
		t.Errorf("Expected 2 permissions after chaining, got %d", len(permissions))
	}
}

func BenchmarkFactoryCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		factory := NewRoutePermissionFactory()

		factory.
			AddPublicRoute("GET", "^/health$", "system", "health").
			AddRoute("GET", "^/users$").
			WithResource("users").WithAction("list").
			WithRoles(AdminOnly()...).GlobalScope().Build().
			AddRoute("POST", "^/condominios/\\d+/unidades$").
			WithResource("unidades").WithAction("create").
			WithRoles(AdminAndSindico()...).CondominiumScope().Build()

		_ = factory.Build()
	}
}
