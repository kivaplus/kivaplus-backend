package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleType_String(t *testing.T) {
	tests := []struct {
		role     RoleType
		expected string
	}{
		{RoleSuperAdmin, "super_admin"},
		{RoleAdmin, "admin"},
		{RoleSindico, "sindico"},
		{RoleMorador, "morador"},
		{RoleFuncionario, "funcionario"},
		{RoleType(999), "unknown"}, // Invalid role
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.role.String()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestClaims_BasicFields(t *testing.T) {
	claims := &Claims{
		UserID:        "123",
		Email:         "test@example.com",
		TokenType:     "access",
		ProfileStatus: "complete",
		TokenVersion:  1234567890,
	}

	assert.Equal(t, "123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, "complete", claims.ProfileStatus)
	assert.Equal(t, int64(1234567890), claims.TokenVersion)
}

func TestClaims_EmptyStruct(t *testing.T) {
	claims := &Claims{}

	assert.Equal(t, "", claims.UserID)
	assert.Equal(t, "", claims.Email)
	assert.Equal(t, "", claims.TokenType)
	assert.Equal(t, "", claims.ProfileStatus)
	assert.Equal(t, int64(0), claims.TokenVersion)
}

func TestClaims_RefreshTokenType(t *testing.T) {
	claims := &Claims{
		UserID:    "456",
		Email:     "refresh@example.com",
		TokenType: "refresh",
	}

	assert.Equal(t, "456", claims.UserID)
	assert.Equal(t, "refresh@example.com", claims.Email)
	assert.Equal(t, "refresh", claims.TokenType)
}

func TestRole_Structure(t *testing.T) {
	condominiumID := int64(100)
	role := Role{
		ID:              3,
		Name:            "sindico",
		CondominiumID:   &condominiumID,
		CondominiumName: "Edifício Teste",
	}

	assert.Equal(t, 3, role.ID)
	assert.Equal(t, "sindico", role.Name)
	assert.NotNil(t, role.CondominiumID)
	assert.Equal(t, int64(100), *role.CondominiumID)
	assert.Equal(t, "Edifício Teste", role.CondominiumName)
}

func TestRole_GlobalRole(t *testing.T) {
	role := Role{
		ID:              1,
		Name:            "super_admin",
		CondominiumID:   nil, // Global role
		CondominiumName: "",
	}

	assert.Equal(t, 1, role.ID)
	assert.Equal(t, "super_admin", role.Name)
	assert.Nil(t, role.CondominiumID)
	assert.Equal(t, "", role.CondominiumName)
}

// Note: Role checking methods have been moved to EnhancedClaims in permissions package
// Use permissions.EnhancedChecker for role validation instead of Claims methods
