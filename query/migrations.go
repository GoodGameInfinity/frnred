// Yep. Absolute raw dog. I've been debugging golang-migrate for an hour straight and I'm fed up with "SQLite database busy" errors.
// If somebody knows a good migration engine that | a) supports libSQL b) isn't a pain to implement | please tell me.
// I've been searching for it for an absurd amount of time and now have lost all motivation so I've asked ChatGPT to do everything for me.
// Enjoy it, AI haters!
//
// Maybe some day I'll make this into an actual library, maybe :)

package query

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migration represents a single migration file.
type Migration struct {
	Version  string
	Up       bool   // true for "up" migrations, false for "down"
	Filename string // Full filename of the migration
}

// RunMigrations applies raw SQL migrations in the correct order.
func RunMigrations(db *sql.DB) error {
	// Read all migration files from the embedded filesystem
	files, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Parse and sort migration files
	migrations := parseMigrationFiles(files)
	sortMigrations(migrations)

	// Create a table to track applied migrations if it doesn't exist
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	// Get the list of already applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to fetch applied migrations: %w", err)
	}

	// Apply migrations in order
	for _, migration := range migrations {
		if migration.Up && !appliedMigrations[migration.Version] {
			log.Printf("Applying migration: %s", migration.Filename)
			if err := applyMigration(db, migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Filename, err)
			}
			if err := recordMigration(db, migration.Version); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// parseMigrationFiles parses migration filenames into Migration structs.
func parseMigrationFiles(files []fs.DirEntry) []Migration {
	var migrations []Migration
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue // Skip invalid filenames
		}
		version := parts[0]
		if strings.HasSuffix(name, ".up.sql") {
			migrations = append(migrations, Migration{Version: version, Up: true, Filename: name})
		}
	}
	return migrations
}

// sortMigrations sorts migrations by version in ascending order.
func sortMigrations(migrations []Migration) {
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
}

// ensureMigrationsTable creates the migrations table if it doesn't exist.
func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			version TEXT PRIMARY KEY
		)
	`)
	return err
}

// getAppliedMigrations fetches the list of applied migrations from the database.
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT version FROM migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, nil
}

// applyMigration executes the SQL statements in a migration file.
func applyMigration(db *sql.DB, migration Migration) error {
	content, err := migrationFiles.ReadFile("migrations/" + migration.Filename)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", migration.Filename, err)
	}

	// Split the file into individual SQL statements
	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement in %s: %w", migration.Filename, err)
		}
	}
	return nil
}

// recordMigration records a migration as applied in the database.
func recordMigration(db *sql.DB, version string) error {
	_, err := db.Exec(`INSERT INTO migrations (version) VALUES (?)`, version)
	return err
}
