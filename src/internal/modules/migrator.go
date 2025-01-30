package modules

import (
	"database/sql"
	"goproxy/dal"
	"log"
)

type Migrator struct {
}

func NewMigrator() *Migrator {
	return &Migrator{}
}

func (m *Migrator) MigrateDb() {
	db, err := dal.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	dal.Migrate(db)
	return
}
