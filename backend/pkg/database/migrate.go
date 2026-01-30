package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations executes all SQL migration files from the given directory.
// It tracks applied migrations in a schema_migrations table to avoid re-running them.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Create tracking table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) PRIMARY KEY,
			applied_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("⚠️  Migrations directory not found: %s (skipping)", migrationsDir)
			return nil
		}
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort .sql files
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		log.Println("📦 No migration files found")
		return nil
	}

	applied := 0
	for _, filename := range files {
		// Check if already applied
		var exists bool
		err := db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)",
			filename,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", filename, err)
		}
		if exists {
			continue
		}

		// Read SQL file
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute in transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		log.Printf("📦 Running migration: %s", filename)
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", filename, err)
		}

		// Record as applied
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (filename) VALUES ($1)",
			filename,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("✅ Migration applied: %s", filename)
		applied++
	}

	if applied == 0 {
		log.Println("📦 All migrations already applied")
	} else {
		log.Printf("📦 Applied %d migration(s)", applied)
	}

	return nil
}
