package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/citadel-agent/backend/internal/engine"

	// Import database drivers
	_ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// MigrateNodeConfig represents the configuration for a Migrate node
type MigrateNodeConfig struct {
	Type          string `json:"type"`           // "postgresql", "mysql", "sqlite"
	ConnectionURL string `json:"connection_url"` // database connection URL
	MigrationPath string `json:"migration_path"` // path to migration files
	Operation     string `json:"operation"`      // "up", "down", "goto", "drop", "force"
	TargetVersion uint   `json:"target_version"` // target version for goto operation
	ForceVersion  int    `json:"force_version"`  // version to force for force operation
	Transaction   bool   `json:"transaction"`    // whether to run migrations in a transaction
}

// MigrateNode represents a database migration node using golang-migrate
type MigrateNode struct {
	config  *MigrateNodeConfig
	migrate *migrate.Migrate
}

// NewMigrateNode creates a new Migrate node
func NewMigrateNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var migrateConfig MigrateNodeConfig
	err = json.Unmarshal(jsonData, &migrateConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if migrateConfig.ConnectionURL == "" {
		return nil, fmt.Errorf("connection_url is required for Migrate node")
	}

	if migrateConfig.MigrationPath == "" {
		return nil, fmt.Errorf("migration_path is required for Migrate node")
	}

	if migrateConfig.Type == "" {
		migrateConfig.Type = "sqlite" // default to sqlite
	}

	if migrateConfig.Operation == "" {
		migrateConfig.Operation = "up" // default operation
	}

	// Build the database URL for migrate
	dbURL := ""
	switch migrateConfig.Type {
	case "postgresql", "postgres":
		dbURL = "postgres://" + migrateConfig.ConnectionURL
	case "mysql":
		dbURL = "mysql://" + migrateConfig.ConnectionURL
	case "sqlite":
		dbURL = "sqlite3://" + migrateConfig.ConnectionURL
	default:
		return nil, fmt.Errorf("unsupported database type: %s", migrateConfig.Type)
	}

	// Create migrate instance
	m, err := migrate.New(
		"file://"+migrateConfig.MigrationPath,
		dbURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %v", err)
	}

	return &MigrateNode{
		config:  &migrateConfig,
		migrate: m,
	}, nil
}

// Execute implements the NodeInstance interface
func (m *MigrateNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	operation := m.config.Operation
	if inputOperation, ok := input["operation"].(string); ok && inputOperation != "" {
		operation = inputOperation
	}

	targetVersion := m.config.TargetVersion
	if inputVersion, ok := input["target_version"].(float64); ok {
		targetVersion = uint(inputVersion)
	}

	forceVersion := m.config.ForceVersion
	if inputForceVersion, ok := input["force_version"].(float64); ok {
		forceVersion = int(inputForceVersion)
	}

	var result interface{}
	var err error

	// Set context timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	switch operation {
	case "up":
		result, err = m.executeUp(ctxWithTimeout)
	case "down":
		result, err = m.executeDown(ctxWithTimeout)
	case "goto":
		result, err = m.executeGoto(ctxWithTimeout, targetVersion)
	case "drop":
		result, err = m.executeDrop(ctxWithTimeout)
	case "force":
		result, err = m.executeForce(ctxWithTimeout, forceVersion)
	default:
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     fmt.Sprintf("unsupported operation: %s", operation),
			Timestamp: time.Now(),
		}, nil
	}

	if err != nil {
		return &engine.ExecutionResult{
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data: map[string]interface{}{
			"result":         result,
			"operation":      operation,
			"target_version": targetVersion,
			"database_type":  m.config.Type,
			"timestamp":      time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// executeUp runs all available migrations
func (m *MigrateNode) executeUp(ctx context.Context) (interface{}, error) {
	err := m.migrate.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			return map[string]interface{}{
				"message": "no migrations to apply",
				"status":  "up_to_date",
			}, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"message": "migrations applied successfully",
		"status":  "success",
	}, nil
}

// executeDown runs one down migration
func (m *MigrateNode) executeDown(ctx context.Context) (interface{}, error) {
	err := m.migrate.Steps(-1) // Run one down migration
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "one migration reverted",
		"status":  "success",
	}, nil
}

// executeGoto migrates to a specific version
func (m *MigrateNode) executeGoto(ctx context.Context, version uint) (interface{}, error) {
	err := m.migrate.Migrate(version)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": fmt.Sprintf("migrated to version %d", version),
		"version": version,
		"status":  "success",
	}, nil
}

// executeDrop drops all database tables
func (m *MigrateNode) executeDrop(ctx context.Context) (interface{}, error) {
	err := m.migrate.Drop()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "all database tables dropped",
		"status":  "success",
	}, nil
}

// executeForce forces the version
func (m *MigrateNode) executeForce(ctx context.Context, version int) (interface{}, error) {
	err := m.migrate.Force(version)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": fmt.Sprintf("version forced to %d", version),
		"version": version,
		"status":  "success",
	}, nil
}

// GetType returns the type of the node
func (m *MigrateNode) GetType() string {
	return "migrate_database"
}

// GetID returns a unique ID for the node instance
func (m *MigrateNode) GetID() string {
	return "migrate_db_" + fmt.Sprintf("%d", time.Now().Unix())
}