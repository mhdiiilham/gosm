package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectPGSQL establishes a connection to a PostgreSQL database using the provided connection URL.
// It allows configuring connection pool settings and ensures proper health checks before returning the database connection.
func ConnectPGSQL(dbURL string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set database connection pool configurations
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Context with timeout for Ping to avoid indefinite blocking
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
