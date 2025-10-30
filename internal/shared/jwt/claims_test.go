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

func TestClaims_HasRole(t *testing.T) {
	condominiumID := int64(100)
	claims := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin"},
			{ID: 3, Name: "sindico", CondominiumID: &condominiumID},
			{ID: 4, Name: "morador", CondominiumID: &condominiumID},
		},
	}

	// Test existing roles
	assert.True(t, claims.HasRole(1)) // super_admin
	assert.True(t, claims.HasRole(3)) // sindico
	assert.True(t, claims.HasRole(4)) // morador

	// Test non-existing role
	assert.False(t, claims.HasRole(2)) // admin (not in roles)
	assert.False(t, claims.HasRole(5)) // funcionario (not in roles)
}

func TestClaims_HasRole_EmptyRoles(t *testing.T) {
	claims := &Claims{
		Roles: []Role{},
	}

	assert.False(t, claims.HasRole(1))
	assert.False(t, claims.HasRole(3))
}

func TestClaims_HasRoleInCondominio(t *testing.T) {
	condominium1 := int64(100)
	condominium2 := int64(200)

	claims := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin", CondominiumID: nil}, // Global role
			{ID: 3, Name: "sindico", CondominiumID: &condominium1},
			{ID: 4, Name: "morador", CondominiumID: &condominium1},
			{ID: 4, Name: "morador", CondominiumID: &condominium2},
		},
	}

	// Test existing role in specific condominium
	assert.True(t, claims.HasRoleInCondominio(3, 100)) // sindico in condominium 100
	assert.True(t, claims.HasRoleInCondominio(4, 100)) // morador in condominium 100
	assert.True(t, claims.HasRoleInCondominio(4, 200)) // morador in condominium 200

	// Test non-existing combinations
	assert.False(t, claims.HasRoleInCondominio(3, 200)) // sindico not in condominium 200
	assert.False(t, claims.HasRoleInCondominio(1, 100)) // super_admin has no specific condominium
	assert.False(t, claims.HasRoleInCondominio(5, 100)) // funcionario role doesn't exist
	assert.False(t, claims.HasRoleInCondominio(4, 300)) // condominium 300 doesn't exist
}

func TestClaims_IsSuperAdmin(t *testing.T) {
	// Test with super admin role
	claimsWithSuperAdmin := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin"},
			{ID: 3, Name: "sindico"},
		},
	}
	assert.True(t, claimsWithSuperAdmin.IsSuperAdmin())

	// Test without super admin role
	claimsWithoutSuperAdmin := &Claims{
		Roles: []Role{
			{ID: 3, Name: "sindico"},
			{ID: 4, Name: "morador"},
		},
	}
	assert.False(t, claimsWithoutSuperAdmin.IsSuperAdmin())

	// Test with empty roles
	claimsEmpty := &Claims{
		Roles: []Role{},
	}
	assert.False(t, claimsEmpty.IsSuperAdmin())
}

func TestClaims_GetCondominiumsWhereSindico(t *testing.T) {
	condominium1 := int64(100)
	condominium2 := int64(200)
	condominium3 := int64(300)

	claims := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin", CondominiumID: nil},
			{ID: 2, Name: "sindico", CondominiumID: &condominium1},
			{ID: 2, Name: "sindico", CondominiumID: &condominium2},
			{ID: 3, Name: "morador", CondominiumID: &condominium3}, // Not sindico
		},
	}

	condominiums := claims.GetCondominiumsWhereSindico()

	assert.Len(t, condominiums, 2)
	assert.Contains(t, condominiums, int64(100))
	assert.Contains(t, condominiums, int64(200))
	assert.NotContains(t, condominiums, int64(300)) // Only morador, not sindico
}

func TestClaims_GetCondominiumsWhereSindico_Empty(t *testing.T) {
	claims := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin"},
			{ID: 4, Name: "morador"},
		},
	}

	condominiums := claims.GetCondominiumsWhereSindico()
	assert.Empty(t, condominiums)
}

func TestClaims_GetCondominiumsWhereMorador(t *testing.T) {
	condominium1 := int64(100)
	condominium2 := int64(200)
	condominium3 := int64(300)

	claims := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin", CondominiumID: nil},
			{ID: 3, Name: "morador", CondominiumID: &condominium1},
			{ID: 3, Name: "morador", CondominiumID: &condominium2},
			{ID: 2, Name: "sindico", CondominiumID: &condominium3}, // Not morador
		},
	}

	condominiums := claims.GetCondominiumsWhereMorador()

	assert.Len(t, condominiums, 2)
	assert.Contains(t, condominiums, int64(100))
	assert.Contains(t, condominiums, int64(200))
	assert.NotContains(t, condominiums, int64(300)) // Only sindico, not morador
}

func TestClaims_GetCondominiumsWhereMorador_Empty(t *testing.T) {
	claims := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin"},
			{ID: 3, Name: "sindico"},
		},
	}

	condominiums := claims.GetCondominiumsWhereMorador()
	assert.Empty(t, condominiums)
}

func TestClaims_CanAccessCondominium_SuperAdmin(t *testing.T) {
	claims := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin", CondominiumID: nil},
		},
	}

	// Super admin can access any condominium
	assert.True(t, claims.CanAccessCondominium(100))
	assert.True(t, claims.CanAccessCondominium(200))
	assert.True(t, claims.CanAccessCondominium(999))
}

func TestClaims_CanAccessCondominium_SpecificRoles(t *testing.T) {
	condominium1 := int64(100)
	condominium2 := int64(200)

	claims := &Claims{
		Roles: []Role{
			{ID: 3, Name: "sindico", CondominiumID: &condominium1},
			{ID: 4, Name: "morador", CondominiumID: &condominium2},
		},
	}

	// Can access condominiums where user has roles
	assert.True(t, claims.CanAccessCondominium(100))
	assert.True(t, claims.CanAccessCondominium(200))

	// Cannot access condominium where user has no roles
	assert.False(t, claims.CanAccessCondominium(300))
}

func TestClaims_CanAccessCondominium_NoRoles(t *testing.T) {
	claims := &Claims{
		Roles: []Role{},
	}

	// No roles means no access
	assert.False(t, claims.CanAccessCondominium(100))
	assert.False(t, claims.CanAccessCondominium(200))
}

func TestClaims_GetRolesForCondominium(t *testing.T) {
	condominium1 := int64(100)
	condominium2 := int64(200)

	claims := &Claims{
		Roles: []Role{
			{ID: 1, Name: "super_admin", CondominiumID: nil},
			{ID: 3, Name: "sindico", CondominiumID: &condominium1},
			{ID: 4, Name: "morador", CondominiumID: &condominium1},
			{ID: 4, Name: "morador", CondominiumID: &condominium2},
			{ID: 5, Name: "funcionario", CondominiumID: &condominium2},
		},
	}

	// Get roles for condominium 100
	roles100 := claims.GetRolesForCondominium(100)
	assert.Len(t, roles100, 2)

	roleIDs100 := make([]int, len(roles100))
	for i, role := range roles100 {
		roleIDs100[i] = role.ID
	}
	assert.Contains(t, roleIDs100, 3) // sindico
	assert.Contains(t, roleIDs100, 4) // morador

	// Get roles for condominium 200
	roles200 := claims.GetRolesForCondominium(200)
	assert.Len(t, roles200, 2)

	roleIDs200 := make([]int, len(roles200))
	for i, role := range roles200 {
		roleIDs200[i] = role.ID
	}
	assert.Contains(t, roleIDs200, 4) // morador
	assert.Contains(t, roleIDs200, 5) // funcionario

	// Get roles for non-existing condominium
	roles300 := claims.GetRolesForCondominium(300)
	assert.Empty(t, roles300)
}

func TestClaims_GetRolesForCondominium_NoRoles(t *testing.T) {
	claims := &Claims{
		Roles: []Role{},
	}

	roles := claims.GetRolesForCondominium(100)
	assert.Empty(t, roles)
}

// Integration test: Complex role scenario
func TestClaims_ComplexRoleScenario(t *testing.T) {
	condominium1 := int64(100)
	condominium2 := int64(200)
	condominium3 := int64(300)

	// User with multiple roles across different condominiums
	claims := &Claims{
		UserID: "complex-user",
		Email:  "complex@example.com",
		Roles: []Role{
			{ID: 2, Name: "sindico", CondominiumID: &condominium1, CondominiumName: "Building A"},
			{ID: 3, Name: "morador", CondominiumID: &condominium1, CondominiumName: "Building A"},
			{ID: 3, Name: "morador", CondominiumID: &condominium2, CondominiumName: "Building B"},
			{ID: 4, Name: "funcionario", CondominiumID: &condominium3, CondominiumName: "Building C"},
		},
	}

	// Test role checks
	assert.True(t, claims.HasRole(2))  // Has sindico role
	assert.True(t, claims.HasRole(3))  // Has morador role
	assert.True(t, claims.HasRole(4))  // Has funcionario role
	assert.False(t, claims.HasRole(1)) // No super admin role
	assert.False(t, claims.HasRole(5)) // No other role

	// Test condominium-specific role checks
	assert.True(t, claims.HasRoleInCondominio(2, 100))  // Sindico in building A
	assert.True(t, claims.HasRoleInCondominio(3, 100))  // Morador in building A
	assert.True(t, claims.HasRoleInCondominio(3, 200))  // Morador in building B
	assert.True(t, claims.HasRoleInCondominio(4, 300))  // Funcionario in building C
	assert.False(t, claims.HasRoleInCondominio(2, 200)) // Not sindico in building B

	// Test access permissions
	assert.True(t, claims.CanAccessCondominium(100))  // Has roles in building A
	assert.True(t, claims.CanAccessCondominium(200))  // Has roles in building B
	assert.True(t, claims.CanAccessCondominium(300))  // Has roles in building C
	assert.False(t, claims.CanAccessCondominium(400)) // No roles in building D

	// Test role collections
	sindicoCondominiums := claims.GetCondominiumsWhereSindico()
	assert.Len(t, sindicoCondominiums, 1)
	assert.Contains(t, sindicoCondominiums, int64(100))

	moradorCondominiums := claims.GetCondominiumsWhereMorador()
	assert.Len(t, moradorCondominiums, 2)
	assert.Contains(t, moradorCondominiums, int64(100))
	assert.Contains(t, moradorCondominiums, int64(200))

	// Test roles for specific condominium
	rolesBuilding1 := claims.GetRolesForCondominium(100)
	assert.Len(t, rolesBuilding1, 2) // Sindico and morador

	rolesBuilding2 := claims.GetRolesForCondominium(200)
	assert.Len(t, rolesBuilding2, 1) // Only morador

	rolesBuilding3 := claims.GetRolesForCondominium(300)
	assert.Len(t, rolesBuilding3, 1) // Only funcionario

	// Test super admin status
	assert.False(t, claims.IsSuperAdmin())
}

// Test edge cases
func TestClaims_EdgeCases(t *testing.T) {
	t.Run("Nil condominium ID handling", func(t *testing.T) {
		claims := &Claims{
			Roles: []Role{
				{ID: 1, Name: "super_admin", CondominiumID: nil},
			},
		}

		// Should not crash with nil condominium ID
		assert.False(t, claims.HasRoleInCondominio(1, 100))
		assert.Empty(t, claims.GetCondominiumsWhereSindico())
		assert.Empty(t, claims.GetCondominiumsWhereMorador())
		assert.Empty(t, claims.GetRolesForCondominium(100))
	})

	t.Run("Zero condominium ID", func(t *testing.T) {
		condominiumZero := int64(0)
		claims := &Claims{
			Roles: []Role{
				{ID: 3, Name: "morador", CondominiumID: &condominiumZero},
			},
		}

		assert.True(t, claims.HasRoleInCondominio(3, 0))
		assert.True(t, claims.CanAccessCondominium(0))
		assert.Contains(t, claims.GetCondominiumsWhereMorador(), int64(0))
	})

	t.Run("Negative condominium ID", func(t *testing.T) {
		condominiumNegative := int64(-1)
		claims := &Claims{
			Roles: []Role{
				{ID: 2, Name: "sindico", CondominiumID: &condominiumNegative},
			},
		}

		assert.True(t, claims.HasRoleInCondominio(2, -1))
		assert.True(t, claims.CanAccessCondominium(-1))
		assert.Contains(t, claims.GetCondominiumsWhereSindico(), int64(-1))
	})
}
