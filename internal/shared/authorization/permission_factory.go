package authorization

import "github.com/kivaplus/kivaplus-backend/internal/shared/jwt"

// RoutePermissionFactory provides a fluent interface for building route permissions
type RoutePermissionFactory struct {
	permissions []RoutePermission
}

// NewRoutePermissionFactory creates a new factory instance
func NewRoutePermissionFactory() *RoutePermissionFactory {
	return &RoutePermissionFactory{
		permissions: make([]RoutePermission, 0),
	}
}

// AddRoute adds a new route permission using a fluent interface
func (f *RoutePermissionFactory) AddRoute(method, pathPattern string) *RouteBuilder {
	return &RouteBuilder{
		factory:     f,
		method:      method,
		pathPattern: pathPattern,
	}
}

// AddPublicRoute adds a public route (no authentication required)
func (f *RoutePermissionFactory) AddPublicRoute(method, pathPattern, resource, action string) *RoutePermissionFactory {
	f.permissions = append(f.permissions, RoutePermission{
		Method:      method,
		PathPattern: pathPattern,
		Permission: Permission{
			Resource: resource,
			Action:   action,
			Roles:    []int{}, // Public access
			Scope:    "global",
		},
	})
	return f
}

// Build returns the final slice of RoutePermission
func (f *RoutePermissionFactory) Build() []RoutePermission {
	return f.permissions
}

// RouteBuilder provides a fluent interface for building individual route permissions
type RouteBuilder struct {
	factory     *RoutePermissionFactory
	method      string
	pathPattern string
	resource    string
	action      string
	roles       []int
	scope       string
}

// WithResource sets the resource for the permission
func (b *RouteBuilder) WithResource(resource string) *RouteBuilder {
	b.resource = resource
	return b
}

// WithAction sets the action for the permission
func (b *RouteBuilder) WithAction(action string) *RouteBuilder {
	b.action = action
	return b
}

// WithRoles sets the allowed roles for the permission
func (b *RouteBuilder) WithRoles(roles ...jwt.RoleType) *RouteBuilder {
	b.roles = make([]int, len(roles))
	for i, role := range roles {
		b.roles[i] = int(role)
	}
	return b
}

// WithScope sets the scope for the permission
func (b *RouteBuilder) WithScope(scope string) *RouteBuilder {
	b.scope = scope
	return b
}

// GlobalScope sets the scope to "global"
func (b *RouteBuilder) GlobalScope() *RouteBuilder {
	return b.WithScope("global")
}

// CondominiumScope sets the scope to "condominium"
func (b *RouteBuilder) CondominiumScope() *RouteBuilder {
	return b.WithScope("condominium")
}

// OwnScope sets the scope to "own"
func (b *RouteBuilder) OwnScope() *RouteBuilder {
	return b.WithScope("own")
}

// Build completes the route permission and adds it to the factory
func (b *RouteBuilder) Build() *RoutePermissionFactory {
	permission := RoutePermission{
		Method:      b.method,
		PathPattern: b.pathPattern,
		Permission: Permission{
			Resource: b.resource,
			Action:   b.action,
			Roles:    b.roles,
			Scope:    b.scope,
		},
	}

	b.factory.permissions = append(b.factory.permissions, permission)
	return b.factory
}

// Helper functions for common role combinations
func AdminOnly() []jwt.RoleType {
	return []jwt.RoleType{jwt.RoleSuperAdmin}
}

func AdminAndSindico() []jwt.RoleType {
	return []jwt.RoleType{jwt.RoleSuperAdmin, jwt.RoleSindico}
}

func AdminSindicoMorador() []jwt.RoleType {
	return []jwt.RoleType{jwt.RoleSuperAdmin, jwt.RoleSindico, jwt.RoleMorador}
}

func AdminSindicoFuncionario() []jwt.RoleType {
	return []jwt.RoleType{jwt.RoleSuperAdmin, jwt.RoleSindico, jwt.RoleFuncionario}
}

func AllRoles() []jwt.RoleType {
	return []jwt.RoleType{jwt.RoleSuperAdmin, jwt.RoleSindico, jwt.RoleMorador, jwt.RoleFuncionario}
}
