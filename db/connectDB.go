package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func Connect() error {
	var err error

	dbPath := "data/server_check.db"
	dsn := "file:" + dbPath + "?cache=shared&mode=rwc"

	DB, err = sqlx.Connect("sqlite3", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return err
}
