# Route Permission Factory

This package provides a factory pattern for building route permissions in a clean, maintainable way.

## Overview

The `RoutePermissionFactory` provides a fluent interface for creating `[]RoutePermission` slices, making it easier to define and maintain API route permissions.

## Basic Usage

### 1. Simple Factory Usage

```go
factory := NewRoutePermissionFactory()

// Add a public route
factory.AddPublicRoute("GET", "^/health$", "system", "health")

// Add a protected route
factory.
    AddRoute("GET", "^/users$").
    WithResource("users").
    WithAction("list").
    WithRoles(jwt.RoleSuperAdmin, jwt.RoleSindico).
    GlobalScope().
    Build()

permissions := factory.Build()
```

### 2. Using Helper Functions

```go
factory := NewRoutePermissionFactory()

factory.
    AddRoute("GET", "^/condominios/\\d+/reports$").
    WithResource("reports").
    WithAction("read").
    WithRoles(AdminAndSindico()...).  // Helper function
    CondominiumScope().
    Build()
```

## Available Helper Functions

- `AdminOnly()` - Only super admin
- `AdminAndSindico()` - Super admin and sindico
- `AdminSindicoMorador()` - Super admin, sindico, and morador
- `AdminSindicoFuncionario()` - Super admin, sindico, and funcionario
- `AllRoles()` - All available roles

## Scope Types

- `GlobalScope()` - Global access (no specific condominium)
- `CondominiumScope()` - Access limited to specific condominium
- `OwnScope()` - Access limited to own data

## Fluent Interface Methods

### RoutePermissionFactory Methods

- `AddRoute(method, pathPattern string)` - Start building a new route
- `AddPublicRoute(method, pathPattern, resource, action string)` - Add a public route
- `Build()` - Return the final `[]RoutePermission`

### RouteBuilder Methods

- `WithResource(resource string)` - Set the resource name
- `WithAction(action string)` - Set the action name
- `WithRoles(roles ...jwt.RoleType)` - Set allowed roles
- `WithScope(scope string)` - Set custom scope
- `GlobalScope()` - Set global scope
- `CondominiumScope()` - Set condominium scope
- `OwnScope()` - Set own data scope
- `Build()` - Complete the route and return to factory

## Examples

### Example 1: Public Routes

```go
factory := NewRoutePermissionFactory()

factory.
    AddPublicRoute("POST", "^/register$", "auth", "register").
    AddPublicRoute("POST", "^/login$", "auth", "login").
    AddPublicRoute("GET", "^/health$", "system", "health")

return factory.Build()
```

### Example 2: User Management Routes

```go
factory := NewRoutePermissionFactory()

// List all users (admin only)
factory.
    AddRoute("GET", "^/users$").
    WithResource("users").WithAction("list").
    WithRoles(AdminOnly()...).GlobalScope().Build()

// Get specific user (admin and sindico)
factory.
    AddRoute("GET", "^/users/\\w+$").
    WithResource("users").WithAction("read").
    WithRoles(AdminAndSindico()...).OwnScope().Build()

// Update user (own data)
factory.
    AddRoute("PUT", "^/users/\\w+$").
    WithResource("users").WithAction("update").
    WithRoles(AllRoles()...).OwnScope().Build()

return factory.Build()
```

### Example 3: Condominium Routes

```go
factory := NewRoutePermissionFactory()

// List condominiums (admin only)
factory.
    AddRoute("GET", "^/condominios$").
    WithResource("condominios").WithAction("list").
    WithRoles(AdminOnly()...).GlobalScope().Build()

// Get specific condominium (all roles with condominium access)
factory.
    AddRoute("GET", "^/condominios/\\d+$").
    WithResource("condominios").WithAction("read").
    WithRoles(AllRoles()...).CondominiumScope().Build()

// Update condominium (admin and sindico only)
factory.
    AddRoute("PUT", "^/condominios/\\d+$").
    WithResource("condominios").WithAction("update").
    WithRoles(AdminAndSindico()...).CondominiumScope().Build()

return factory.Build()
```

### Example 4: Complex Nested Routes

```go
factory := NewRoutePermissionFactory()

// Condominium units
factory.
    AddRoute("GET", "^/condominios/\\d+/unidades$").
    WithResource("unidades").WithAction("list").
    WithRoles(AllRoles()...).CondominiumScope().Build().

    AddRoute("POST", "^/condominios/\\d+/unidades$").
    WithResource("unidades").WithAction("create").
    WithRoles(AdminAndSindico()...).CondominiumScope().Build()

// Condominium residents
factory.
    AddRoute("GET", "^/condominios/\\d+/moradores$").
    WithResource("moradores").WithAction("list").
    WithRoles(AdminSindicoFuncionario()...).CondominiumScope().Build().

    AddRoute("POST", "^/condominios/\\d+/moradores$").
    WithResource("moradores").WithAction("create").
    WithRoles(AdminAndSindico()...).CondominiumScope().Build()

return factory.Build()
```

## Benefits

1. **Readability** - Clear, fluent interface makes permissions easy to understand
2. **Maintainability** - Easy to add, modify, or remove permissions
3. **Type Safety** - Compile-time checking of role types
4. **Reusability** - Helper functions for common role combinations
5. **Consistency** - Standardized way to define permissions across the application

## Integration

The factory is used in the `getDefaultRoutePermissions()` function to build the default route permissions for the authorization service:

```go
func getDefaultRoutePermissions() []RoutePermission {
    factory := NewRoutePermissionFactory()

    // Build all your routes using the factory
    // ... route definitions ...

    return factory.Build()
}
```

This approach makes the permission system much more maintainable and easier to extend as your application grows.
