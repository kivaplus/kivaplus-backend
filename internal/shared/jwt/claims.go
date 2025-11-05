package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims representa as informações do usuário no JWT
type Claims struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	TokenType     string `json:"token_type"`     // "access" ou "refresh"
	ProfileStatus string `json:"profile_status"` // "complete" ou "incomplete"
	TokenVersion  int64  `json:"token_version"`  // Version for lazy invalidation
	jwt.RegisteredClaims
}

// Role representa um papel do usuário em um condomínio específico
type Role struct {
	ID              int    `json:"id"`             // ID do papel (1=super_admin, 2=admin, 3=sindico, 4=morador, 5=funcionario)
	Name            string `json:"name"`           // Nome do papel
	CondominiumID   *int64 `json:"condominium_id"` // nil para super_admin, ID do condomínio para outros
	CondominiumName string `json:"condominium"`    // Nome do condomínio (para exibição)
}

// RoleType representa os tipos de papéis disponíveis
type RoleType int

const (
	RoleSuperAdmin  RoleType = 1
	RoleAdmin       RoleType = 2
	RoleSindico     RoleType = 3
	RoleMorador     RoleType = 4
	RoleFuncionario RoleType = 5
)

// String retorna o nome do papel
func (r RoleType) String() string {
	switch r {
	case RoleSuperAdmin:
		return "super_admin"
	case RoleAdmin:
		return "admin"
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

// Note: Role checking methods have been moved to EnhancedClaims in permissions package
// Use permissions.EnhancedChecker.ValidateRequestWithPermissions() to get EnhancedClaims with role methods
