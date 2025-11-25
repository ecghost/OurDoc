// db/db.go
package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var (
	DBOg1 *sql.DB
	DBOg2 *sql.DB
)

func InitDB() {
	var err error
	DBOg1, err = sql.Open("postgres", "host=localhost port=5432 user=gaussdb password=Sakura030523! dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to og1: %v", err)
	}
	DBOg2, err = sql.Open("postgres", "host=localhost port=5433 user=gaussdb password=Sakura030523! dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to og2: %v", err)
	}

	if err = DBOg1.Ping(); err != nil {
		log.Fatalf("Ping og1 failed: %v", err)
	}
	if err = DBOg2.Ping(); err != nil {
		log.Fatalf("Ping og2 failed: %v", err)
	}
}
