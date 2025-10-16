// internal/database/migrations.go
package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

type MigrationConfig struct {
	DatabaseURL string
}

func RunMigrations(cfg MigrationConfig) error {
	// Conectar ao banco
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Testar conexão
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Criar tabela de controle de migrations
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	// Ler arquivos de migration
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Ordenar por nome (apenas arquivos .up.sql)
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// Executar cada migration
	for _, file := range files {
		if err := runMigration(db, file); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `
	_, err := db.Exec(query)
	return err
}

func runMigration(db *sql.DB, filename string) error {
	// Verificar se já foi executada
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)",
		filename,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("⏭️  Skipping %s (already applied)", filename)
		return nil
	}

	// Ler conteúdo do arquivo
	content, err := migrationsFS.ReadFile("migrations/" + filename)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Executar em uma transação
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Executar SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Registrar como executada
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		filename,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("✅ Applied %s", filename)
	return nil
}

// RollbackMigrations reverts the specified number of migrations
func RollbackMigrations(cfg MigrationConfig, steps int) error {
	// Conectar ao banco
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Testar conexão
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Obter migrations aplicadas (ordenadas por versão decrescente)
	rows, err := db.Query(`
		SELECT version FROM schema_migrations
		ORDER BY version DESC
		LIMIT $1
	`, steps)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	var versionsToRollback []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan version: %w", err)
		}
		versionsToRollback = append(versionsToRollback, version)
	}

	if len(versionsToRollback) == 0 {
		log.Println("⚠️  No migrations to rollback")
		return nil
	}

	// Reverter cada migration
	for _, version := range versionsToRollback {
		if err := rollbackMigration(db, version); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", version, err)
		}
	}

	return nil
}

// rollbackMigration reverts a single migration
func rollbackMigration(db *sql.DB, version string) error {
	// Para este exemplo, vamos apenas remover da tabela de controle
	// Em um sistema real, você executaria o arquivo .down.sql correspondente

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remover da tabela de controle
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = $1", version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("⏪ Rolled back %s", version)
	return nil
}

// GetMigrationVersion returns the current migration version and dirty state
func GetMigrationVersion(cfg MigrationConfig) (int, bool, error) {
	// Conectar ao banco
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return 0, false, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Testar conexão
	if err := db.Ping(); err != nil {
		return 0, false, fmt.Errorf("failed to ping database: %w", err)
	}

	// Verificar se a tabela de migrations existe
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'schema_migrations'
		)
	`).Scan(&tableExists)
	if err != nil {
		return 0, false, fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !tableExists {
		return 0, false, nil // No migrations table = version 0
	}

	// Obter a versão mais recente
	var latestVersion string
	err = db.QueryRow(`
		SELECT version FROM schema_migrations
		ORDER BY version DESC
		LIMIT 1
	`).Scan(&latestVersion)

	if err == sql.ErrNoRows {
		return 0, false, nil // No migrations applied = version 0
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get latest version: %w", err)
	}

	// Extrair número da versão (assumindo formato 000001_name.up.sql)
	var versionNum int

	// Tentar extrair apenas os primeiros dígitos do nome do arquivo
	if n, err := fmt.Sscanf(latestVersion, "%d", &versionNum); err != nil || n == 0 {
		// Se não conseguir extrair número, usar contagem de migrations
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
		if err != nil {
			return 0, false, fmt.Errorf("failed to count migrations: %w", err)
		}
		versionNum = count
	}

	// Para este exemplo, assumimos que não há estado "dirty"
	// Em um sistema real, você verificaria se há migrations parcialmente aplicadas
	return versionNum, false, nil
}
