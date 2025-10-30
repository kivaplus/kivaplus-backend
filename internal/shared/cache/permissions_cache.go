package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/kivaplus/kivaplus-backend/internal/shared/logger"
)

// Permission representa uma permissão específica
type Permission map[string]map[string]bool // resource -> action -> allowed

// Role representa um papel com suas permissões
type Role struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Permissions Permission `json:"permissions"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PermissionsCache gerencia o cache de permissões dos papéis
type PermissionsCache struct {
	roles       map[int]*Role
	lastUpdated time.Time
	mutex       sync.RWMutex
	db          *database.Connection
	logger      logger.Logger
	cacheTTL    time.Duration
}

// NewPermissionsCache cria um novo cache de permissões
func NewPermissionsCache(db *database.Connection, logger logger.Logger) *PermissionsCache {
	cache := &PermissionsCache{
		roles:    make(map[int]*Role),
		db:       db,
		logger:   logger,
		cacheTTL: 30 * time.Minute, // Cache por 30 minutos
	}

	// Carrega as permissões inicialmente
	if err := cache.loadRoles(context.Background()); err != nil {
		logger.Error("Failed to load initial roles", "error", err)
	}

	return cache
}

// GetRole retorna um papel por ID, carregando do cache ou DB conforme necessário
func (c *PermissionsCache) GetRole(ctx context.Context, roleID int) (*Role, error) {
	c.mutex.RLock()

	// Verifica se o cache está válido
	if time.Since(c.lastUpdated) > c.cacheTTL {
		c.mutex.RUnlock()
		// Cache expirado, recarrega
		if err := c.loadRoles(ctx); err != nil {
			return nil, fmt.Errorf("failed to reload roles: %w", err)
		}
		c.mutex.RLock()
	}

	role, exists := c.roles[roleID]
	c.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("role with ID %d not found", roleID)
	}

	return role, nil
}

// GetAllRoles retorna todos os papéis do cache
func (c *PermissionsCache) GetAllRoles(ctx context.Context) (map[int]*Role, error) {
	c.mutex.RLock()

	// Verifica se o cache está válido
	if time.Since(c.lastUpdated) > c.cacheTTL {
		c.mutex.RUnlock()
		// Cache expirado, recarrega
		if err := c.loadRoles(ctx); err != nil {
			return nil, fmt.Errorf("failed to reload roles: %w", err)
		}
		c.mutex.RLock()
	}

	// Cria uma cópia para evitar modificações concorrentes
	rolesCopy := make(map[int]*Role)
	for id, role := range c.roles {
		rolesCopy[id] = role
	}
	c.mutex.RUnlock()

	return rolesCopy, nil
}

// HasPermission verifica se um papel tem uma permissão específica
func (c *PermissionsCache) HasPermission(ctx context.Context, roleID int, resource, action string) (bool, error) {
	role, err := c.GetRole(ctx, roleID)
	if err != nil {
		return false, err
	}

	// Super admin tem acesso a tudo
	if role.Name == "super_admin" {
		if allPerm, exists := role.Permissions["all"]; exists {
			if allowed, ok := allPerm["true"]; ok && allowed {
				return true, nil
			}
		}
		return true, nil // Assume que super_admin sempre tem acesso
	}

	// Verifica permissão específica
	if resourcePerms, exists := role.Permissions[resource]; exists {
		if allowed, ok := resourcePerms[action]; ok {
			return allowed, nil
		}
	}

	return false, nil
}

// InvalidateCache força a recarga do cache na próxima consulta
func (c *PermissionsCache) InvalidateCache() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.lastUpdated = time.Time{} // Força recarga
	c.logger.Info("Permissions cache invalidated")
}

// loadRoles carrega os papéis do banco de dados
func (c *PermissionsCache) loadRoles(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.logger.Info("Loading roles from database")

	query := `
		SELECT id, nome, descricao, permissoes, created_at
		FROM papel
		ORDER BY id`

	rows, err := c.db.DB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	newRoles := make(map[int]*Role)

	for rows.Next() {
		var role Role
		var permissionsJSON string

		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&permissionsJSON,
			&role.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan role: %w", err)
		}

		// Parse das permissões JSON
		if err := json.Unmarshal([]byte(permissionsJSON), &role.Permissions); err != nil {
			c.logger.Error("Failed to parse permissions JSON", "roleID", role.ID, "error", err)
			// Continua com permissões vazias em caso de erro
			role.Permissions = make(Permission)
		}

		newRoles[role.ID] = &role
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating roles: %w", err)
	}

	// Atualiza o cache
	c.roles = newRoles
	c.lastUpdated = time.Now()

	c.logger.Info("Roles loaded successfully", "count", len(newRoles))
	return nil
}

// GetCacheStats retorna estatísticas do cache
func (c *PermissionsCache) GetCacheStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"roles_count":  len(c.roles),
		"last_updated": c.lastUpdated,
		"cache_ttl":    c.cacheTTL,
		"is_expired":   time.Since(c.lastUpdated) > c.cacheTTL,
	}
}
