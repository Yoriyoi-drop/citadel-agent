// backend/internal/database/database.go
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

// DB represents a database connection
type DB struct {
	pool   *pgxpool.Pool
	config *Config
}

// New creates a new database connection
func New(ctx context.Context, config *Config) (*DB, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Apply connection pool settings from our config
	poolConfig.MaxConns = int32(config.MaxConns)
	poolConfig.MinConns = int32(config.MinConns)
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheck
	poolConfig.LazyConnect = false // Connect immediately to verify connection

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		pool:   pool,
		config: config,
	}

	return db, nil
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Config returns the database configuration
func (db *DB) Config() *Config {
	return db.config
}

// Ping tests the database connection
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Exec executes a query without returning any rows
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (pgx.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}

// Query executes a query that returns rows
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

// Begin starts a new database transaction
func (db *DB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// BeginTx starts a new database transaction with specific options
func (db *DB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return db.pool.BeginTx(ctx, txOptions)
}

// Conn returns a single connection from the pool
func (db *DB) Conn(ctx context.Context) (*pgx.Conn, error) {
	return db.pool.Acquire(ctx)
}

// Stats returns database connection pool statistics
func (db *DB) Stats() pgxpool.Stat {
	return db.pool.Stat()
}

// WaitForConnection waits for a connection to be available
func (db *DB) WaitForConnection(ctx context.Context) error {
	// This is just a simple implementation - in reality, any query would wait for a connection
	_, err := db.pool.Acquire(ctx)
	return err
}

// HealthCheck performs a health check on the database
func (db *DB) HealthCheck(ctx context.Context) error {
	// Check if we can acquire a connection
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Check if we can ping the database
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to execute simple query: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: %d", result)
	}

	return nil
}

// Migrate executes database migrations
func (db *DB) Migrate(ctx context.Context) error {
	// This is a placeholder for migration logic
	// In a real implementation, you would use a migration library like golang-migrate
	fmt.Println("Executing database migrations...")
	
	// Example: run a simple migration check
	_, err := db.Exec(ctx, "SELECT 1 FROM workflows LIMIT 1")
	if err != nil {
		// If the workflows table doesn't exist, we would run migrations
		// For now, let's just check if it's a "does not exist" error
		if pgErr, ok := err.(interface{ Code() string }); ok {
			if pgErr.(interface{ Code() string }).Code() == "42P01" { // undefined_table
				fmt.Println("Workflows table does not exist, need to run migrations")
				// Here you would execute your migrations
				return db.RunInitialMigrations(ctx)
			}
		}
		return err
	}

	fmt.Println("Database migration check completed")
	return nil
}

// RunInitialMigrations executes the initial set of migrations
func (db *DB) RunInitialMigrations(ctx context.Context) error {
	fmt.Println("Running initial database migrations...")

	// In a real implementation, you would read and execute the SQL migration files
	// This is a simplified example of what the migrations might do:

	// 1. Create the workflows table
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS workflows (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			nodes JSONB NOT NULL DEFAULT '[]',
			connections JSONB NOT NULL DEFAULT '[]',
			config JSONB,
			status VARCHAR(50) NOT NULL DEFAULT 'draft',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			version INTEGER NOT NULL DEFAULT 1,
			owner_id UUID NOT NULL,
			team_id UUID,
			tags TEXT[] DEFAULT '{}'
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create workflows table: %w", err)
	}

	// 2. Create indexes for the workflows table
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_workflows_owner_id ON workflows(owner_id);
		CREATE INDEX IF NOT EXISTS idx_workflows_team_id ON workflows(team_id);
		CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status);
		CREATE INDEX IF NOT EXISTS idx_workflows_created_at ON workflows(created_at);
	`)
	if err != nil {
		return fmt.Errorf("failed to create workflows indexes: %w", err)
	}

	// 3. Create the executions table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS executions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			name VARCHAR(255),
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP WITH TIME ZONE,
			variables JSONB,
			node_results JSONB,
			error TEXT,
			triggered_by VARCHAR(100) NOT NULL DEFAULT 'manual',
			trigger_params JSONB,
			parent_id UUID REFERENCES executions(id) ON DELETE CASCADE,
			retry_count INTEGER NOT NULL DEFAULT 0,
			user_id UUID NOT NULL,
			team_id UUID,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create executions table: %w", err)
	}

	// 4. Create indexes for the executions table
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_executions_workflow_id ON executions(workflow_id);
		CREATE INDEX IF NOT EXISTS idx_executions_status ON executions(status);
		CREATE INDEX IF NOT EXISTS idx_executions_started_at ON executions(started_at);
		CREATE INDEX IF NOT EXISTS idx_executions_user_id ON executions(user_id);
		CREATE INDEX IF NOT EXISTS idx_executions_team_id ON executions(team_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create executions indexes: %w", err)
	}

	// 5. Create the execution_logs table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS execution_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
			node_id UUID,
			status VARCHAR(50),
			action VARCHAR(100) NOT NULL,
			message TEXT,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			parameters JSONB,
			details JSONB,
			user_id UUID,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create execution_logs table: %w", err)
	}

	// 6. Create indexes for the execution_logs table
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_execution_logs_execution_id ON execution_logs(execution_id);
		CREATE INDEX IF NOT EXISTS idx_execution_logs_node_id ON execution_logs(node_id);
		CREATE INDEX IF NOT EXISTS idx_execution_logs_timestamp ON execution_logs(timestamp);
		CREATE INDEX IF NOT EXISTS idx_execution_logs_action ON execution_logs(action);
	`)
	if err != nil {
		return fmt.Errorf("failed to create execution_logs indexes: %w", err)
	}

	// 7. Create the users table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'viewer',
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			last_login_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			profile JSONB,
			preferences JSONB
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// 8. Create indexes for the users table
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
		CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
	`)
	if err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// 9. Create the teams table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS teams (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			owner_id UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			settings JSONB
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create teams table: %w", err)
	}

	// 10. Create indexes for the teams table
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams(owner_id);
		CREATE INDEX IF NOT EXISTS idx_teams_created_at ON teams(created_at);
	`)
	if err != nil {
		return fmt.Errorf("failed to create teams indexes: %w", err)
	}

	// 11. Create the team_members table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS team_members (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE(team_id, user_id)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create team_members table: %w", err)
	}

	// 12. Create indexes for the team_members table
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
		CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
		CREATE INDEX IF NOT EXISTS idx_team_members_role ON team_members(role);
	`)
	if err != nil {
		return fmt.Errorf("failed to create team_members indexes: %w", err)
	}

	fmt.Println("Database migrations completed successfully")
	return nil
}

// CheckTableExists checks if a table exists in the database
func (db *DB) CheckTableExists(ctx context.Context, tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		);
	`

	var exists bool
	err := db.QueryRow(ctx, query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if table exists: %w", err)
	}

	return exists, nil
}

// GetTableCount returns the number of records in a table
func (db *DB) GetTableCount(ctx context.Context, tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	
	var count int64
	err := db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count records in table %s: %w", tableName, err)
	}

	return count, nil
}

// EnableExtensions enables required PostgreSQL extensions
func (db *DB) EnableExtensions(ctx context.Context) error {
	extensions := []string{
		"uuid-ossp",  // For generating UUIDs
		"pgcrypto",  // For cryptographic functions
		"pg_stat_statements",  // For query statistics
	}

	for _, ext := range extensions {
		query := fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\";", ext)
		_, err := db.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to enable extension %s: %w", ext, err)
		}
	}

	return nil
}