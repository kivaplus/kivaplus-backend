package permissions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// PermissionChecker handles both JSONB and granular permission systems
type PermissionChecker struct {
	db     *database.Connection
	logger logger.Logger
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(db *database.Connection, logger logger.Logger) *PermissionChecker {
	return &PermissionChecker{
		db:     db,
		logger: logger,
	}
}

// Role represents a role with its permission type
type Role struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	PermissionType string          `json:"permission_type"` // 'jsonb' or 'granular'
	Permissions    json.RawMessage `json:"permissions"`     // JSONB permissions
	CondominiumID  *int64          `json:"condominium_id"`  // For scoped roles
}

// Permission represents a granular permission
type Permission struct {
	ID            int64  `json:"id"`
	RoleID        int64  `json:"role_id"`
	Resource      string `json:"resource"`
	Action        string `json:"action"`
	Allowed       bool   `json:"allowed"`
	Scope         string `json:"scope"`          // 'global', 'condominium', 'own'
	CondominiumID *int64 `json:"condominium_id"` // For condominium-scoped permissions
}

// JSONBPermissions represents the structure of JSONB permissions
type JSONBPermissions map[string]map[string]bool

// HasPermission checks if a role has permission for a resource and action
func (pc *PermissionChecker) HasPermission(ctx context.Context, roleID int64, resource, action string, condominiumID *int64) (bool, error) {
	// First, get the role to determine permission type
	role, err := pc.getRole(ctx, roleID)
	if err != nil {
		return false, fmt.Errorf("failed to get role: %w", err)
	}

	switch role.PermissionType {
	case "jsonb":
		return pc.checkJSONBPermission(role, resource, action)
	case "granular":
		return pc.checkGranularPermission(ctx, roleID, resource, action, condominiumID)
	default:
		return false, fmt.Errorf("unknown permission type: %s", role.PermissionType)
	}
}

// checkJSONBPermission checks permissions stored in JSONB format
func (pc *PermissionChecker) checkJSONBPermission(role *Role, resource, action string) (bool, error) {
	// First, check for super admin "all" permission using raw JSON
	// This handles the case where permissions is {"all": true}
	var rawPerms map[string]interface{}
	if err := json.Unmarshal(role.Permissions, &rawPerms); err != nil {
		return false, fmt.Errorf("failed to parse JSONB permissions: %w", err)
	}

	// Check for super admin "all" permission
	if allValue, exists := rawPerms["all"]; exists {
		if allBool, ok := allValue.(bool); ok && allBool {
			return true, nil
		}
	}

	// Parse as structured permissions for resource-specific checks
	var permissions JSONBPermissions
	if err := json.Unmarshal(role.Permissions, &permissions); err != nil {
		// If it fails to parse as structured permissions, it might be the {"all": true} format
		// which we already checked above, so return false
		return false, nil
	}

	// Check specific resource permission
	if resourcePerms, exists := permissions[resource]; exists {
		if allowed, exists := resourcePerms[action]; exists {
			return allowed, nil
		}
	}

	return false, nil
}

// checkGranularPermission checks permissions stored in the permissions table
func (pc *PermissionChecker) checkGranularPermission(ctx context.Context, roleID int64, resource, action string, condominiumID *int64) (bool, error) {
	query := `
		SELECT allowed, scope, condominium_id
		FROM permissions
		WHERE role_id = $1 AND resource = $2 AND action = $3 AND allowed = true
		ORDER BY
			CASE scope
				WHEN 'condominium' THEN 1
				WHEN 'own' THEN 2
				WHEN 'global' THEN 3
			END`

	rows, err := pc.db.DB.QueryContext(ctx, query, roleID, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var allowed bool
		var scope string
		var permCondominiumID sql.NullInt64

		if err := rows.Scan(&allowed, &scope, &permCondominiumID); err != nil {
			return false, fmt.Errorf("failed to scan permission: %w", err)
		}

		if !allowed {
			continue
		}

		switch scope {
		case "global":
			return true, nil
		case "condominium":
			// Check if the permission is for the specific condominium
			if condominiumID != nil && permCondominiumID.Valid && permCondominiumID.Int64 == *condominiumID {
				return true, nil
			}
			// If no specific condominium in permission, allow for any condominium the user has access to
			if !permCondominiumID.Valid {
				return true, nil
			}
		case "own":
			// This would require additional context about ownership
			// For now, treat as condominium-scoped
			return condominiumID != nil, nil
		}
	}

	return false, nil
}

// getRole retrieves a role by ID
func (pc *PermissionChecker) getRole(ctx context.Context, roleID int64) (*Role, error) {
	query := `
		SELECT id, name, description, permission_type, permissions
		FROM role
		WHERE id = $1`

	var role Role
	err := pc.db.DB.QueryRowContext(ctx, query, roleID).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.PermissionType,
		&role.Permissions,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// CreateCustomRole creates a new custom role with granular permissions
func (pc *PermissionChecker) CreateCustomRole(ctx context.Context, name, description string, permissions []Permission) (*Role, error) {
	tx, err := pc.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create the role
	var roleID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO role (name, description, permission_type, permissions)
		VALUES ($1, $2, 'granular', '{}')
		RETURNING id`,
		name, description).Scan(&roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Insert granular permissions
	for _, perm := range permissions {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO permissions (role_id, resource, action, allowed, scope, condominium_id)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			roleID, perm.Resource, perm.Action, perm.Allowed, perm.Scope, perm.CondominiumID)
		if err != nil {
			return nil, fmt.Errorf("failed to create permission: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the created role
	return pc.getRole(ctx, roleID)
}

// GetRolePermissions retrieves all permissions for a role (works for both types)
func (pc *PermissionChecker) GetRolePermissions(ctx context.Context, roleID int64) (map[string]map[string]bool, error) {
	role, err := pc.getRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	switch role.PermissionType {
	case "jsonb":
		var permissions JSONBPermissions
		if err := json.Unmarshal(role.Permissions, &permissions); err != nil {
			return nil, fmt.Errorf("failed to parse JSONB permissions: %w", err)
		}
		return permissions, nil

	case "granular":
		return pc.getGranularPermissions(ctx, roleID)

	default:
		return nil, fmt.Errorf("unknown permission type: %s", role.PermissionType)
	}
}

// getGranularPermissions converts granular permissions to the same format as JSONB
func (pc *PermissionChecker) getGranularPermissions(ctx context.Context, roleID int64) (map[string]map[string]bool, error) {
	query := `
		SELECT resource, action, allowed
		FROM permissions
		WHERE role_id = $1`

	rows, err := pc.db.DB.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query granular permissions: %w", err)
	}
	defer rows.Close()

	permissions := make(map[string]map[string]bool)
	for rows.Next() {
		var resource, action string
		var allowed bool

		if err := rows.Scan(&resource, &action, &allowed); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		if permissions[resource] == nil {
			permissions[resource] = make(map[string]bool)
		}
		permissions[resource][action] = allowed
	}

	return permissions, nil
}
