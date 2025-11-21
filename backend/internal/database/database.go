package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents the database connection
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB creates a new database connection
func NewDB(databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		Pool: pool,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}