package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims representa as informações do usuário no JWT
type Claims struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Roles         []Role `json:"roles"`          // Array de papéis em diferentes condomínios
	TokenType     string `json:"token_type"`     // "access" ou "refresh"
	ProfileStatus string `json:"profile_status"` // "complete" ou "incomplete"
	jwt.RegisteredClaims
}

// Role representa um papel do usuário em um condomínio específico
type Role struct {
	ID              int    `json:"id"`             // ID do papel (1=super_admin, 2=administrator, 3=sindico, 4=morador, 5=funcionario)
	Name            string `json:"name"`           // Nome do papel
	CondominiumID   *int64 `json:"condominium_id"` // nil para super_admin, ID do condomínio para outros
	CondominiumName string `json:"condominium"`    // Nome do condomínio (para exibição)
}

// RoleType representa os tipos de papéis disponíveis
type RoleType int

const (
	RoleSuperAdmin  RoleType = 1
	RoleSindico     RoleType = 2
	RoleMorador     RoleType = 3
	RoleFuncionario RoleType = 4
)

// String retorna o nome do papel
func (r RoleType) String() string {
	switch r {
	case RoleSuperAdmin:
		return "super_admin"
	case RoleSindico:
		return "sindico"
	case RoleMorador:
		return "morador"
	case RoleFuncionario:
		return "funcionario"
	default:
		return "unknown"
	}
}

// HasRole verifica se o usuário tem um papel específico
func (c *Claims) HasRole(roleID int) bool {
	for _, role := range c.Roles {
		if role.ID == roleID {
			return true
		}
	}
	return false
}

// HasRoleInCondominio verifica se o usuário tem um papel específico em um condomínio
func (c *Claims) HasRoleInCondominio(roleID int, condominiumID int64) bool {
	for _, role := range c.Roles {
		if role.ID == roleID && role.CondominiumID != nil && *role.CondominiumID == condominiumID {
			return true
		}
	}
	return false
}

// IsSuperAdmin verifica se o usuário é super admin
func (c *Claims) IsSuperAdmin() bool {
	return c.HasRole(int(RoleSuperAdmin))
}

// GetCondominiosWhereSindico retorna os IDs dos condomínios onde o usuário é síndico
func (c *Claims) GetCondominiumsWhereSindico() []int64 {
	var condominiums []int64
	for _, role := range c.Roles {
		if role.ID == int(RoleSindico) && role.CondominiumID != nil {
			condominiums = append(condominiums, *role.CondominiumID)
		}
	}
	return condominiums
}

// GetCondominiosWhereMorador retorna os IDs dos condomínios onde o usuário é morador
func (c *Claims) GetCondominiumsWhereMorador() []int64 {
	var condominiums []int64
	for _, role := range c.Roles {
		if role.ID == int(RoleMorador) && role.CondominiumID != nil {
			condominiums = append(condominiums, *role.CondominiumID)
		}
	}
	return condominiums
}

// CanAccessCondominio verifica se o usuário pode acessar um condomínio específico
func (c *Claims) CanAccessCondominium(condominiumID int64) bool {
	// Super admin pode acessar qualquer condomínio
	if c.IsSuperAdmin() {
		return true
	}

	// Verifica se tem algum papel no condomínio específico
	for _, role := range c.Roles {
		if role.CondominiumID != nil && *role.CondominiumID == condominiumID {
			return true
		}
	}

	return false
}

// GetRolesForCondominio retorna todos os papéis do usuário em um condomínio específico
func (c *Claims) GetRolesForCondominium(condominiumID int64) []Role {
	var roles []Role
	for _, role := range c.Roles {
		if role.CondominiumID != nil && *role.CondominiumID == condominiumID {
			roles = append(roles, role)
		}
	}
	return roles
}
