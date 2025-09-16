package migration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

type Migration struct {
	Version string
	Up      string
	Down    string
}

type MigrationRunner struct {
	db *sql.DB
}

func NewMigrationRunner(db *sql.DB) *MigrationRunner {
	return &MigrationRunner{db: db}
}

func (m *MigrationRunner) CreateMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`
	_, err := m.db.Exec(query)
	return err
}

func (m *MigrationRunner) LoadMigrations(migrationsDir string) ([]Migration, error) {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return nil, err
		}

		migration := parseMigration(string(content))
		migration.Version = strings.TrimSuffix(file.Name(), ".sql")
		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigration(content string) Migration {
	lines := strings.Split(content, "\n")
	var upLines, downLines []string
	var inUp, inDown bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "-- +migrate Up") {
			inUp = true
			inDown = false
			continue
		}
		if strings.Contains(line, "-- +migrate Down") {
			inUp = false
			inDown = true
			continue
		}

		if inUp && line != "" && !strings.HasPrefix(line, "--") {
			upLines = append(upLines, line)
		}
		if inDown && line != "" && !strings.HasPrefix(line, "--") {
			downLines = append(downLines, line)
		}
	}

	return Migration{
		Up:   strings.Join(upLines, "\n"),
		Down: strings.Join(downLines, "\n"),
	}
}

func (m *MigrationRunner) GetAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := m.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, nil
}

func (m *MigrationRunner) ApplyMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(migration.Up); err != nil {
		return fmt.Errorf("failed to apply migration %s: %v", migration.Version, err)
	}

	// Record migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
		return fmt.Errorf("failed to record migration %s: %v", migration.Version, err)
	}

	return tx.Commit()
}

func (m *MigrationRunner) RunMigrations(migrationsDir string) error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	migrations, err := m.LoadMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %v", err)
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	for _, migration := range migrations {
		if applied[migration.Version] {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Applying migration %s", migration.Version)
		if err := m.ApplyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %v", migration.Version, err)
		}
		log.Printf("Migration %s applied successfully", migration.Version)
	}

	return nil
}
