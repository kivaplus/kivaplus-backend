//go:build integration
// +build integration

package main

import (
	"fmt"
	"os"

	"github.com/kivaplus/kivaplus-backend/internal/shared/database"
)

func main() {
	fmt.Println("🔍 Database Connection Debug")
	fmt.Println("===========================")

	// Check environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	fmt.Printf("DATABASE_URL: %s\n", databaseURL)

	if databaseURL == "" {
		fmt.Println("❌ DATABASE_URL is not set!")
		fmt.Println("💡 Set it with: export DATABASE_URL='postgres://kivaplus:kivaplus123@localhost:5432/kivaplus_local?sslmode=disable'")
		os.Exit(1)
	}

	fmt.Println("✅ DATABASE_URL is set")
	fmt.Println("")

	// Try to connect to the database
	fmt.Println("🔗 Attempting database connection...")
	db, err := database.NewPostgresConnection()
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		fmt.Println("")
		fmt.Println("🔍 Debugging info:")
		fmt.Printf("   - DATABASE_URL: %s\n", databaseURL)
		fmt.Println("   - Make sure PostgreSQL is running: docker compose ps postgres")
		fmt.Println("   - Test direct connection: docker compose exec postgres psql -U kivaplus -d kivaplus_local")
		os.Exit(1)
	}

	fmt.Println("✅ Database connection successful")

	// Test a simple query
	fmt.Println("🧪 Testing database query...")
	var result int
	err = db.DB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		fmt.Printf("❌ Database query failed: %v\n", err)
		os.Exit(1)
	}

	if result != 1 {
		fmt.Printf("❌ Database query returned unexpected result: %d\n", result)
		os.Exit(1)
	}

	fmt.Println("✅ Database query successful")
	fmt.Println("")
	fmt.Println("🎉 All database tests passed!")
	fmt.Println("")
	fmt.Println("📋 Connection Details:")
	fmt.Printf("   - URL: %s\n", databaseURL)
	fmt.Println("   - Status: Connected and working")

	os.Exit(0)
}
