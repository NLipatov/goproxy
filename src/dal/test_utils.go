package dal

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"os"
	"testing"
	"time"
)

func SetupPostgresContainer(t *testing.T) (testcontainers.Container, *sql.DB, func()) {
	setEnv(t)
	ctx := context.Background()

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

	for i := 0; i < 10; i++ {
		if err = db.Ping(); err == nil {
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
