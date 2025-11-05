package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// EnhancedRoleRepository implements role operations with caching support
type EnhancedRoleRepository struct {
	db     *database.Connection
	logger logger.Logger
}

// NewEnhancedRoleRepository creates a new enhanced role repository
func NewEnhancedRoleRepository(db *database.Connection, logger logger.Logger) *EnhancedRoleRepository {
	return &EnhancedRoleRepository{
		db:     db,
		logger: logger,
	}
}

// GetUserRolesAndProfile retrieves user roles and profile status efficiently
func (r *EnhancedRoleRepository) GetUserRolesAndProfile(ctx context.Context, userID int64) ([]jwt.Role, string, error) {
	// Single query to get user roles and profile status
	query := `
		SELECT
			ur.role_id,
			ro.name as role_name,
			ur.condominium_id,
			c.name as condominium_name,
			CASE
				WHEN pd.id IS NOT NULL
				AND addr.id IS NOT NULL
				THEN 'complete'
				ELSE 'incomplete'
			END as profile_status
		FROM users_role ur
		JOIN role ro ON ur.role_id = ro.id
		LEFT JOIN condominium c ON ur.condominium_id = c.id
		JOIN users u ON ur.users_id = u.id
		LEFT JOIN person p ON u.id = p.user_id
		LEFT JOIN person_document pd ON p.id = pd.person_document_id
		LEFT JOIN address addr ON p.address_id = addr.id
		WHERE ur.users_id = $1
		ORDER BY ur.role_id, ur.condominium_id`

	rows, err := r.db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var roles []jwt.Role
	var profileStatus string

	for rows.Next() {
		var roleID int
		var roleName string
		var condominiumID sql.NullInt64
		var condominiumName sql.NullString
		var status string

		err := rows.Scan(&roleID, &roleName, &condominiumID, &condominiumName, &status)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan role: %w", err)
		}

		// Set profile status (same for all rows)
		if profileStatus == "" {
			profileStatus = status
		}

		// Create role
		role := jwt.Role{
			ID:   roleID,
			Name: roleName,
		}

		if condominiumID.Valid {
			role.CondominiumID = &condominiumID.Int64
			if condominiumName.Valid {
				role.CondominiumName = condominiumName.String
			}
		}

		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("error iterating roles: %w", err)
	}

	// If no roles found, still get profile status
	if len(roles) == 0 {
		profileStatus, err = r.getUserProfileStatus(ctx, userID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get profile status: %w", err)
		}
	}

	return roles, profileStatus, nil
}

// getUserProfileStatus gets profile status for users without roles
func (r *EnhancedRoleRepository) getUserProfileStatus(ctx context.Context, userID int64) (string, error) {
	query := `
		SELECT
			CASE
				WHEN pd.id IS NOT NULL
				AND addr.id IS NOT NULL
				THEN 'complete'
				ELSE 'incomplete'
			END as profile_status
		FROM users u
		LEFT JOIN person p ON u.id = p.user_id
		LEFT JOIN person_document pd ON p.id = pd.person_document_id
		LEFT JOIN address addr ON p.address_id = addr.id
		WHERE u.id = $1`

	var status string
	err := r.db.DB.QueryRowContext(ctx, query, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "incomplete", nil
		}
		return "", fmt.Errorf("failed to get profile status: %w", err)
	}

	return status, nil
}

// AddUserRole adds a role to a user (with cache invalidation trigger)
func (r *EnhancedRoleRepository) AddUserRole(ctx context.Context, userID int64, roleID int, condominiumID *int64) error {
	query := `
		INSERT INTO users_role (users_id, role_id, condominium_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (users_id, role_id, COALESCE(condominium_id, 0)) DO NOTHING`

	_, err := r.db.DB.ExecContext(ctx, query, userID, roleID, condominiumID)
	if err != nil {
		return fmt.Errorf("failed to add user role: %w", err)
	}

	r.logger.Info("Added role to user", "userID", userID, "roleID", roleID, "condominiumID", condominiumID)
	return nil
}

// RemoveUserRole removes a role from a user
func (r *EnhancedRoleRepository) RemoveUserRole(ctx context.Context, userID int64, roleID int, condominiumID *int64) error {
	var query string
	var args []interface{}

	if condominiumID != nil {
		query = `DELETE FROM users_role WHERE users_id = $1 AND role_id = $2 AND condominium_id = $3`
		args = []interface{}{userID, roleID, *condominiumID}
	} else {
		query = `DELETE FROM users_role WHERE users_id = $1 AND role_id = $2 AND condominium_id IS NULL`
		args = []interface{}{userID, roleID}
	}

	result, err := r.db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to remove user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found for user")
	}

	r.logger.Info("Removed role from user", "userID", userID, "roleID", roleID, "condominiumID", condominiumID)
	return nil
}

// GetUsersWithRole retrieves all users with a specific role (for bulk operations)
func (r *EnhancedRoleRepository) GetUsersWithRole(ctx context.Context, roleID int, condominiumID *int64) ([]int64, error) {
	var query string
	var args []any

	if condominiumID != nil {
		query = `SELECT users_id FROM users_role WHERE role_id = $1 AND condominium_id = $2`
		args = []any{roleID, *condominiumID}
	} else {
		query = `SELECT users_id FROM users_role WHERE role_id = $1 AND condominium_id IS NULL`
		args = []any{roleID}
	}

	rows, err := r.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users with role: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// GetUsersByRole retrieves all users with a specific role (alias for security checks)
func (r *EnhancedRoleRepository) GetUsersByRole(ctx context.Context, roleID int) ([]int64, error) {
	return r.GetUsersWithRole(ctx, roleID, nil)
}

// BulkAddUserRoles adds multiple roles efficiently
func (r *EnhancedRoleRepository) BulkAddUserRoles(ctx context.Context, userRoles []UserRoleAssignment) error {
	if len(userRoles) == 0 {
		return nil
	}

	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO users_role (users_id, role_id, condominium_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (users_id, role_id, COALESCE(condominium_id, 0)) DO NOTHING`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, ur := range userRoles {
		_, err = stmt.ExecContext(ctx, ur.UserID, ur.RoleID, ur.CondominiumID)
		if err != nil {
			return fmt.Errorf("failed to insert user role: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Bulk added user roles", "count", len(userRoles))
	return nil
}

// UserRoleAssignment represents a user role assignment
type UserRoleAssignment struct {
	UserID        int64  `json:"user_id"`
	RoleID        int    `json:"role_id"`
	CondominiumID *int64 `json:"condominium_id"`
}

// GetRolesByCondominium retrieves all roles for a specific condominium
func (r *EnhancedRoleRepository) GetRolesByCondominium(ctx context.Context, condominiumID int64) (map[int64][]jwt.Role, error) {
	query := `
		SELECT
			ur.users_id,
			ur.role_id,
			ro.name as role_name
		FROM users_role ur
		JOIN role ro ON ur.role_id = ro.id
		WHERE ur.condominium_id = $1
		ORDER BY ur.users_id, ur.role_id`

	rows, err := r.db.DB.QueryContext(ctx, query, condominiumID)
	if err != nil {
		return nil, fmt.Errorf("failed to query condominium roles: %w", err)
	}
	defer rows.Close()

	userRoles := make(map[int64][]jwt.Role)
	for rows.Next() {
		var userID int64
		var roleID int
		var roleName string

		err := rows.Scan(&userID, &roleID, &roleName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		role := jwt.Role{
			ID:              roleID,
			Name:            roleName,
			CondominiumID:   &condominiumID,
			CondominiumName: "", // Would need another join to get this
		}

		userRoles[userID] = append(userRoles[userID], role)
	}

	return userRoles, nil
}
