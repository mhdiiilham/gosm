package database

import (
	"database/sql"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectPGSQL establishes a connection to a PostgreSQL database using the provided connection URL.
// It returns a pointer to the SQL database connection and an error if the connection fails.
func ConnectPGSQL(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
