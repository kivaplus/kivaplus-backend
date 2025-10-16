package database

import (
	"os"
	"testing"

	"github.com/kivaplus/kivaplus-backend/lambda/types"
)

// TestPostgresClient tests the PostgreSQL client functionality
func TestPostgresClient(t *testing.T) {
	// Skip test if DATABASE_URL is not set
	if os.Getenv("DB_HOST") == "" {
		t.Skip("DATABASE_URL not set, skipping PostgreSQL tests")
	}

	// Create client
	client, err := NewPostgresClient()
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL client: %v", err)
	}
	defer client.Close()

	t.Log("✅ PostgreSQL client created successfully")

	// Test data
	testEmail := "test-user@kivaplus.com"
	testPassword := "testpassword123"

	// Clean up any existing test user
	t.Cleanup(func() {
		client.DeleteUser(testEmail)
	})

	t.Run("UserDoesNotExistInitially", func(t *testing.T) {
		exists, err := client.DoesUserExist(testEmail)
		if err != nil {
			t.Fatalf("Error checking if user exists: %v", err)
		}
		if exists {
			t.Errorf("User should not exist initially, but DoesUserExist returned true")
		}
		t.Logf("✅ User %s does not exist initially", testEmail)
	})

	t.Run("InsertUser", func(t *testing.T) {
		// Create user
		registerUser := types.RegisterUser{
			Email: testEmail,
			Senha: testPassword,
		}

		user, err := types.NewUser(registerUser)
		if err != nil {
			t.Fatalf("Failed to create user struct: %v", err)
		}

		// Insert user
		err = client.InsertUser(user)
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}
		t.Logf("✅ User %s inserted successfully", testEmail)
	})

	t.Run("UserExistsAfterInsert", func(t *testing.T) {
		exists, err := client.DoesUserExist(testEmail)
		if err != nil {
			t.Fatalf("Error checking if user exists: %v", err)
		}
		if !exists {
			t.Errorf("User should exist after insert, but DoesUserExist returned false")
		}
		t.Logf("✅ User %s exists after insert", testEmail)
	})

	t.Run("GetUser", func(t *testing.T) {
		user, err := client.GetUser(testEmail)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}
		if user.Email != testEmail {
			t.Errorf("Expected username %s, got %s", testEmail, user.Email)
		}
		if user.SenhaHash == "" {
			t.Errorf("Password hash should not be empty")
		}
		t.Logf("✅ User %s retrieved successfully", user.Email)
	})

	t.Run("ValidatePassword", func(t *testing.T) {
		user, err := client.GetUser(testEmail)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		// Test correct password
		isValid := types.ValidatePassword(user.SenhaHash, testPassword)
		if !isValid {
			t.Errorf("Password validation should succeed for correct password")
		}
		t.Log("✅ Correct password validation passed")

		// Test incorrect password
		isInvalid := types.ValidatePassword(user.SenhaHash, "wrongpassword")
		if isInvalid {
			t.Errorf("Password validation should fail for incorrect password")
		}
		t.Log("✅ Incorrect password validation failed as expected")
	})

	t.Run("UpdateLastAccess", func(t *testing.T) {
		err := client.UpdateLastAccess(testEmail)
		if err != nil {
			t.Fatalf("Failed to update last access: %v", err)
		}
		t.Logf("✅ Last access updated for user %s", testEmail)
	})

	t.Run("DeactivateUser", func(t *testing.T) {
		err := client.DeactivateUser(testEmail)
		if err != nil {
			t.Fatalf("Failed to deactivate user: %v", err)
		}
		t.Logf("✅ User %s deactivated", testEmail)

		// User should not exist when checking (because we only check active users)
		exists, err := client.DoesUserExist(testEmail)
		if err != nil {
			t.Fatalf("Error checking if deactivated user exists: %v", err)
		}
		if exists {
			t.Errorf("Deactivated user should not be found by DoesUserExist")
		}
		t.Log("✅ Deactivated user is not found by DoesUserExist")
	})

	t.Run("GetDeactivatedUser", func(t *testing.T) {
		// Should not be able to get deactivated user
		_, err := client.GetUser(testEmail)
		if err == nil {
			t.Errorf("Should not be able to get deactivated user")
		}
		t.Log("✅ Cannot get deactivated user as expected")
	})
}

// TestPostgresClientErrors tests error conditions
func TestPostgresClientErrors(t *testing.T) {
	// Skip test if DATABASE_URL is not set
	if os.Getenv("DB_HOST") == "" {
		t.Skip("DATABASE_URL not set, skipping PostgreSQL tests")
	}

	client, err := NewPostgresClient()
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL client: %v", err)
	}
	defer client.Close()

	t.Run("GetNonExistentUser", func(t *testing.T) {
		_, err := client.GetUser("nonexistent@example.com")
		if err == nil {
			t.Errorf("Should return error when getting non-existent user")
		}
		t.Log("✅ Getting non-existent user returns error as expected")
	})

	t.Run("UpdateLastAccessNonExistentUser", func(t *testing.T) {
		err := client.UpdateLastAccess("nonexistent@example.com")
		if err == nil {
			t.Errorf("Should return error when updating last access for non-existent user")
		}
		t.Log("✅ Updating last access for non-existent user returns error as expected")
	})

	t.Run("DeactivateNonExistentUser", func(t *testing.T) {
		err := client.DeactivateUser("nonexistent@example.com")
		if err == nil {
			t.Errorf("Should return error when deactivating non-existent user")
		}
		t.Log("✅ Deactivating non-existent user returns error as expected")
	})
}

// TestPostgresClientConnection tests connection handling
func TestPostgresClientConnection(t *testing.T) {
	t.Run("InvalidDatabaseURL", func(t *testing.T) {
		// Save original DATABASE_URL
		originalURL := os.Getenv("DB_HOST")
		defer func() {
			if originalURL != "" {
				os.Setenv("DB_HOST", originalURL)
			} else {
				os.Unsetenv("DB_HOST")
			}
		}()

		// Test with invalid URL
		os.Setenv("DATABASE_URL", "invalid://url")
		_, err := NewPostgresClient()
		if err == nil {
			t.Errorf("Should return error with invalid database URL")
		}
		t.Log("✅ Invalid database URL returns error as expected")
	})

	t.Run("MissingDatabaseURL", func(t *testing.T) {
		// Save original DATABASE_URL
		originalURL := os.Getenv("DB_HOST")
		defer func() {
			if originalURL != "" {
				os.Setenv("DB_HOST", originalURL)
			}
		}()

		// Test with missing URL
		os.Unsetenv("DB_HOST")
		_, err := NewPostgresClient()
		if err == nil {
			t.Errorf("Should return error when DATABASE_URL is not set")
		}
		t.Log("✅ Missing DATABASE_URL returns error as expected")
	})
}

// BenchmarkPostgresOperations benchmarks common database operations
func BenchmarkPostgresOperations(b *testing.B) {
	if os.Getenv("DB_HOST") == "" {
		b.Skip("DATABASE_URL not set, skipping PostgreSQL benchmarks")
	}

	client, err := NewPostgresClient()
	if err != nil {
		b.Fatalf("Failed to create PostgreSQL client: %v", err)
	}
	defer client.Close()

	// Setup test user
	testEmail := "benchmark-user@kivaplus.com"
	registerUser := types.RegisterUser{
		Email: testEmail,
		Senha: "benchmarkpassword",
	}
	user, _ := types.NewUser(registerUser)
	client.InsertUser(user)

	// Clean up
	defer client.DeleteUser(testEmail)

	b.Run("DoesUserExist", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.DoesUserExist(testEmail)
		}
	})

	b.Run("GetUser", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.GetUser(testEmail)
		}
	})

	b.Run("UpdateLastAccess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.UpdateLastAccess(testEmail)
		}
	})
}
