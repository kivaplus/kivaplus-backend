package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/jwt"
)

// RolePostgresRepository implementa o RoleRepository usando PostgreSQL
type RolePostgresRepository struct {
	db *database.Connection
}

// NewRolePostgresRepository cria um novo repositório de papéis PostgreSQL
func NewRolePostgresRepository(db *database.Connection) *RolePostgresRepository {
	return &RolePostgresRepository{
		db: db,
	}
}

// GetUserRoles retorna todos os papéis de um usuário
func (r *RolePostgresRepository) GetUserRoles(ctx context.Context, userID int64) ([]jwt.Role, error) {
	// Use a safer query that handles NULL condominium_id properly
	query := `
		SELECT
			ur.role_id,
			r.name,
			ur.condominium_id,
			CASE
				WHEN ur.condominium_id IS NOT NULL THEN COALESCE(c.name, '')
				ELSE ''
			END as condominium_name
		FROM users_role ur
		JOIN role r ON ur.role_id = r.id
		LEFT JOIN condominium c ON ur.condominium_id = c.id
		WHERE ur.users_id = $1
		ORDER BY ur.role_id, ur.condominium_id`

	fmt.Printf("GetUserRoles: Querying roles for userID=%d\n", userID)
	rows, err := r.db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		fmt.Printf("GetUserRoles: Query failed for userID=%d, error=%v\n", userID, err)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []jwt.Role
	for rows.Next() {
		var role jwt.Role
		var condominiumID sql.NullInt64
		var condominiumName string

		err := rows.Scan(
			&role.ID,
			&role.Name,
			&condominiumID,
			&condominiumName,
		)
		if err != nil {
			fmt.Printf("GetUserRoles: Scan failed for userID=%d, error=%v\n", userID, err)
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		// Handle condominium ID and name if present
		if condominiumID.Valid {
			role.CondominiumID = &condominiumID.Int64
			role.CondominiumName = condominiumName
		}

		fmt.Printf("GetUserRoles: Found role ID=%d, Name=%s, CondominiumID=%v, CondominiumName=%s\n",
			role.ID, role.Name, role.CondominiumID, role.CondominiumName)
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// AddUserRole adiciona um papel a um usuário
func (r *RolePostgresRepository) AddUserRole(ctx context.Context, userID int64, roleID int, condominiumID *int64) error {
	query := `
		INSERT INTO users_role (users_id, role_id, condominium_id, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (users_id, role_id, condominium_id) DO NOTHING`

	_, err := r.db.DB.ExecContext(ctx, query, userID, roleID, condominiumID)
	if err != nil {
		return fmt.Errorf("failed to add user role: %w", err)
	}

	return nil
}

// RemoveUserRole remove um papel de um usuário
func (r *RolePostgresRepository) RemoveUserRole(ctx context.Context, userID int64, roleID int, condominiumID *int64) error {
	query := `
		DELETE FROM users_role
		WHERE users_id = $1 AND role_id = $2 AND condominium_id = $3`

	_, err := r.db.DB.ExecContext(ctx, query, userID, roleID, condominiumID)
	if err != nil {
		return fmt.Errorf("failed to remove user role: %w", err)
	}

	return nil
}

// RemoveUserRoleWithoutCondominium remove um papel de um usuário que não tem condominium_id (NULL)
func (r *RolePostgresRepository) RemoveUserRoleWithoutCondominium(ctx context.Context, userID int64, roleID int) error {
	query := `
		DELETE FROM users_role
		WHERE users_id = $1 AND role_id = $2 AND condominium_id IS NULL`

	_, err := r.db.DB.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove user role without condominium: %w", err)
	}

	return nil
}

// GetUserRolesInCondominio retorna os papéis de um usuário em um condomínio específico
func (r *RolePostgresRepository) GetUserRolesInCondominium(ctx context.Context, userID int64, condominiumID int64) ([]jwt.Role, error) {
	query := `
		SELECT
			up.role_id,
			p.name,
			up.condominium_id,
			c.name as condominium_name
		FROM users_role up
		JOIN role p ON up.role_id = p.id
		JOIN condominium c ON up.condominium_id = c.id
		WHERE up.users_id = $1 AND up.condominium_id = $2
		ORDER BY up.role_id`

	rows, err := r.db.DB.QueryContext(ctx, query, userID, condominiumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles in condominium: %w", err)
	}
	defer rows.Close()

	var roles []jwt.Role
	for rows.Next() {
		var role jwt.Role
		var condID int64
		var condominiumName string

		err := rows.Scan(
			&role.ID,
			&role.Name,
			&condID,
			&condominiumName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		role.CondominiumID = &condID
		role.CondominiumName = condominiumName
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// GetUsersByRole retrieves all users with a specific role (for security validation)
func (r *RolePostgresRepository) GetUsersByRole(ctx context.Context, roleID int) ([]int64, error) {
	query := `SELECT users_id FROM users_role WHERE role_id = $1 AND condominium_id IS NULL`

	rows, err := r.db.DB.QueryContext(ctx, query, roleID)
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
