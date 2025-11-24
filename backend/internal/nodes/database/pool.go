package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// ConnectionPool manages database connections
type ConnectionPool struct {
	pools map[string]*sql.DB
	mu    sync.RWMutex
}

// NewConnectionPool creates a new connection pool manager
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		pools: make(map[string]*sql.DB),
	}
}

// GetConnection gets or creates a database connection
func (cp *ConnectionPool) GetConnection(dbType, connectionString string) (*sql.DB, error) {
	// Create a unique key for this connection
	key := fmt.Sprintf("%s:%s", dbType, connectionString)

	// Check if connection already exists
	cp.mu.RLock()
	if db, exists := cp.pools[key]; exists {
		cp.mu.RUnlock()
		// Verify connection is still alive
		if err := db.Ping(); err == nil {
			return db, nil
		}
		// Connection is dead, remove it
		cp.mu.Lock()
		delete(cp.pools, key)
		cp.mu.Unlock()
	} else {
		cp.mu.RUnlock()
	}

	// Create new connection
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Double-check after acquiring write lock
	if db, exists := cp.pools[key]; exists {
		return db, nil
	}

	// Open new connection
	db, err := sql.Open(dbType, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
	db.SetConnMaxIdleTime(1 * time.Minute) // Maximum idle time of a connection

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Store in pool
	cp.pools[key] = db

	return db, nil
}

// Close closes all connections in the pool
func (cp *ConnectionPool) Close() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	var errs []error
	for key, db := range cp.pools {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection %s: %w", key, err))
		}
	}

	cp.pools = make(map[string]*sql.DB)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// HealthCheck checks the health of all connections
func (cp *ConnectionPool) HealthCheck(ctx context.Context) map[string]error {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	results := make(map[string]error)
	for key, db := range cp.pools {
		results[key] = db.PingContext(ctx)
	}

	return results
}

// Stats returns statistics for all connections
func (cp *ConnectionPool) Stats() map[string]sql.DBStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	stats := make(map[string]sql.DBStats)
	for key, db := range cp.pools {
		stats[key] = db.Stats()
	}

	return stats
}

// Global connection pool instance
var globalPool *ConnectionPool
var poolOnce sync.Once

// GetGlobalPool returns the global connection pool instance
func GetGlobalPool() *ConnectionPool {
	poolOnce.Do(func() {
		globalPool = NewConnectionPool()
	})
	return globalPool
}
