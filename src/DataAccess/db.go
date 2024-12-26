package data_access

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
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
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbDatabase := os.Getenv("DB_DATABASE")

	if dbUser == "" || dbPass == "" || dbHost == "" || dbPort == "" || dbDatabase == "" {
		return "", fmt.Errorf("invalid db credentials")
	}

	dataSourceName := fmt.
		Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbDatabase)

	return dataSourceName, nil
}
