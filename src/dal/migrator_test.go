package dal

import (
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	setEnvErr := os.Setenv("DB_DATABASE", "proxy")
	if setEnvErr != nil {
		t.Fatal(setEnvErr)
	}

	defer func() {
		_ = os.Unsetenv("DB_DATABASE")
	}()

	_, db, cleanup := SetupPostgresContainer(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	Migrate(db)

	var schemaMigrationsTblExist bool
	err := db.
		QueryRow("SELECT EXISTS (" +
			"SELECT 1 FROM information_schema.tables " +
			"WHERE table_schema = 'public' AND table_name = 'schema_migrations');").
		Scan(&schemaMigrationsTblExist)

	if err != nil {
		t.Fatalf("Failed to check if schema migrations table exists: %v", err)
	}

	if !schemaMigrationsTblExist {
		t.Fatalf("Schema migrations table does not exist")
	}

	var appliedVersion int
	err = db.
		QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations;").
		Scan(&appliedVersion)

	if err != nil {
		t.Fatalf("Failed to get schema migration version: %v", err)
	}

	migrationsDir := getMigrationsPath()
	files, err := filepath.Glob(migrationsDir)

	if appliedVersion != len(files) {
		t.Fatalf("Unexpected applied migration version. Expected %d, got %d", appliedVersion, len(files))
	}
}
