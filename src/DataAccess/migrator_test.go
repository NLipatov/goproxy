package data_access

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
)

func setEnv(t *testing.T) {
	envVars := map[string]string{
		"DB_USER":     "test_user",
		"DB_PASS":     "test_pass",
		"DB_HOST":     "127.0.0.1",
		"DB_PORT":     "6543",
		"DB_DATABASE": "test_db",
	}

	for key, value := range envVars {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
	}
}

func TestMigrate(t *testing.T) {
	_, db, cleanup := setupPostgresContainer(t)
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

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	migrationsPath := filepath.Join(wd, "migrations", "v*.sql")

	files, err := filepath.Glob(migrationsPath)

	if appliedVersion != len(files) {
		t.Fatalf("Unexpected applied migration version. Expected %d, got %d", appliedVersion, len(files))
	}
}

func setupPostgresContainer(t *testing.T) (testcontainers.Container, *sql.DB, func()) {
	setEnv(t)
	ctx := context.Background()

	// Container request with explicit port mapping
	req := testcontainers.ContainerRequest{
		Image: "postgres:latest",
		ExposedPorts: []string{
			"5432/tcp",
		},
		Env: map[string]string{
			"POSTGRES_USER":     os.Getenv("DB_USER"),
			"POSTGRES_PASSWORD": os.Getenv("DB_PASS"),
			"POSTGRES_DB":       os.Getenv("DB_DATABASE"),
		},
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get host and port
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	mappedPort, err := postgresContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		host,
		mappedPort.Port(),
		os.Getenv("DB_DATABASE"))

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Ping the database to confirm it's ready
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		if i == 9 {
			t.Fatalf("Database is not ready after retries: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	cleanup := func() {
		_ = db.Close()
		_ = postgresContainer.Terminate(ctx)
	}

	return postgresContainer, db, cleanup
}
