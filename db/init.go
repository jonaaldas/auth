package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var db *sql.DB

func InitDb() *sql.DB {
	url := os.Getenv("URL")

	var err error
	db, err = sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to db: %s", err)
		os.Exit(1)
	}

	log.Println("Database connected successfully")

	return db
}
