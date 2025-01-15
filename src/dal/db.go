package dal

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func ConnectDB() (*sql.DB, error) {
	connString, err := buildConnectionString()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func buildConnectionString() (string, error) {
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Fatalf("DB_USER environment variable not set")
	}

	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		log.Fatalf("DB_PASS environment variable not set")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Fatalf("DB_HOST environment variable not set")
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		log.Fatalf("DB_PORT environment variable not set")
	}

	dbDatabase := os.Getenv("DB_DATABASE")
	if dbDatabase == "" {
		log.Fatalf("DB_DATABASE environment variable not set")
	}

	dataSourceName := fmt.
		Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbDatabase)

	return dataSourceName, nil
}
