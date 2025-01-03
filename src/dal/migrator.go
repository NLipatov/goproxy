package dal

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func Migrate(db *sql.DB) {
	appliedVersion := getAppliedVersion(db)

	migrationsDir := getMigrationsPath()
	files, err := filepath.Glob(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migration files: %s", err)
	}

	if len(files) == 0 {
		log.Fatalf("No migrations found in %s", migrationsDir)
	}

	sort.Strings(files)

	migrationCounter := 0
	for _, file := range files {
		versionStr := strings.TrimPrefix(filepath.Base(file), "v")
		versionStr = strings.Split(versionStr, "_")[0]
		version, convErr := strconv.Atoi(versionStr)
		if convErr != nil {
			log.Fatalf("Invalid migration file name: %s", file)
		}

		if version > appliedVersion {
			content, readFileErr := os.ReadFile(file)
			if readFileErr != nil {
				log.Fatalf("Failed to read migration file %s: %v", file, readFileErr)
			}

			applyMigration(db, version, string(content))
			migrationCounter++
		}
	}

	log.Printf("%d migrations applied", migrationCounter)
}

func getMigrationsPath() string {
	databaseName := os.Getenv("DB_DATABASE")
	if databaseName == "" {
		log.Fatalf("DB_DATABASE env variable must be set.")
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Failed to get current file path")
	}
	return filepath.Join(filepath.Dir(filename), "migrations", databaseName, "v*.sql")
}

func getAppliedVersion(db *sql.DB) int {
	var appliedVersion int
	err := db.
		QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations;").
		Scan(&appliedVersion)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			log.Println("schema_migrations table does not exist, returning version 0")
			return 0
		}
		log.Fatalf("Error getting applied version: %v", err)
	}

	return appliedVersion
}

func applyMigration(db *sql.DB, version int, script string) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}

	// apply migration
	_, err = tx.Exec(script)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("Error applying migration: %v", err)
	}

	// increase applied migration counter
	_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("Error applying migration: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("Error committing migration: %v", err)
	}

	log.Printf("Applied version: %d", version)
}
