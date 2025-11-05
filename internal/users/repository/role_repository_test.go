package repository

import (
	"context"
	"testing"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleRepository_GetUserRoles(t *testing.T) {
	// Skip if no DATABASE_URL is set (for CI/CD)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database connection
	db, err := database.NewPostgresConnection()
	require.NoError(t, err, "Failed to connect to test database")
	defer db.Close()

	repo := NewRolePostgresRepository(db)
	ctx := context.Background()

	t.Run("GetUserRoles_NonExistentUser", func(t *testing.T) {
		// Test with a user ID that doesn't exist
		roles, err := repo.GetUserRoles(ctx, 99999)

		// Should not error, just return empty slice
		assert.NoError(t, err)
		assert.Empty(t, roles)
	})

	t.Run("AddUserRole_And_GetUserRoles", func(t *testing.T) {
		// This test requires a user to exist, so we'll use a known user ID
		// In a real test, you'd create a test user first
		userID := int64(1)
		roleID := 3 // Sindico role (default)

		// Add role to user
		err := repo.AddUserRole(ctx, userID, roleID, nil)
		assert.NoError(t, err)

		// Get user roles
		roles, err := repo.GetUserRoles(ctx, userID)
		assert.NoError(t, err)

		// Should have at least one role
		assert.NotEmpty(t, roles)

		// Check if our role is in the list
		found := false
		for _, role := range roles {
			if role.ID == roleID {
				found = true
				assert.Equal(t, "sindico", role.Name)
				assert.Nil(t, role.CondominiumID) // No condominium assigned
				break
			}
		}
		assert.True(t, found, "Role should be found in user roles")
	})
}

func TestRoleRepository_AddUserRole(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := database.NewPostgresConnection()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRolePostgresRepository(db)
	ctx := context.Background()

	t.Run("AddUserRole_Success", func(t *testing.T) {
		userID := int64(1)
		roleID := 4 // Morador role

		err := repo.AddUserRole(ctx, userID, roleID, nil)
		assert.NoError(t, err)

		// Verify the role was added
		roles, err := repo.GetUserRoles(ctx, userID)
		assert.NoError(t, err)

		found := false
		for _, role := range roles {
			if role.ID == roleID {
				found = true
				break
			}
		}
		assert.True(t, found, "Role should be added to user")
	})

	t.Run("AddUserRole_Duplicate", func(t *testing.T) {
		userID := int64(1)
		roleID := 4 // Same role as above

		// Adding the same role again should not error (ON CONFLICT DO NOTHING)
		err := repo.AddUserRole(ctx, userID, roleID, nil)
		assert.NoError(t, err)
	})
}
