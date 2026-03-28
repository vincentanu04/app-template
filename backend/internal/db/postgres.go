package db

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// New opens a postgres connection using DATABASE_URL from the environment.
// go-jet uses database/sql with the lib/pq driver.
func New() (*sql.DB, error) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
