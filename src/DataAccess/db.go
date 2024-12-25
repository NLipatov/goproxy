package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func ConnectDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	return db
}
