// backend/internal/nodes/integrations/database_node.go
package integrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"

	_ "github.com/lib/pq"        // PostgreSQL driver
	_ "github.com/go-sql-driver/mysql"  // MySQL driver
	_ "github.com/mattn/go-sqlite3"     // SQLite driver
	"github.com/jmoiron/sqlx"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypeSQLServer  DatabaseType = "sqlserver"
	DatabaseTypeOracle     DatabaseType = "oracle"
)

// DatabaseNodeConfig represents the configuration for a database node
type DatabaseNodeConfig struct {
	Type          DatabaseType     `json:"type"`
	ConnectionURL string           `json:"connection_url"`
	Query         string           `json:"query"`
	QueryType     string           `json:"query_type"` // "select", "insert", "update", "delete", "raw"
	Parameters    map[string]interface{} `json:"parameters"`
	Timeout       time.Duration    `json:"timeout"`
	MaxRetries    int              `json:"max_retries"`
	ConnectionPoolSize int         `json:"connection_pool_size"`
	SSLMode       string           `json:"ssl_mode"` // For PostgreSQL
	EnableSSL     bool             `json:"enable_ssl"`
}

// DatabaseNode represents a database operation node
type DatabaseNode struct {
	config *DatabaseNodeConfig
	db     *sqlx.DB
}

// NewDatabaseNode creates a new database node
func NewDatabaseNode(config *DatabaseNodeConfig) (*DatabaseNode, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.ConnectionPoolSize == 0 {
		config.ConnectionPoolSize = 10
	}

	// Initialize database connection based on type
	var driverName string
	connectionURL := config.ConnectionURL

	switch config.Type {
	case DatabaseTypePostgreSQL:
		driverName = "postgres"
		if config.EnableSSL && config.SSLMode != "" {
			connectionURL += fmt.Sprintf("?sslmode=%s", config.SSLMode)
		}
	case DatabaseTypeMySQL:
		driverName = "mysql"
	case DatabaseTypeSQLite:
		driverName = "sqlite3"
	case DatabaseTypeSQLServer:
		driverName = "sqlserver"
	case DatabaseTypeOracle:
		driverName = "godror" // Oracle driver
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	// Create database connection
	db, err := sqlx.Connect(driverName, connectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.ConnectionPoolSize)
	db.SetMaxIdleConns(config.ConnectionPoolSize / 2)
	db.SetConnMaxLifetime(30 * time.Minute)

	node := &DatabaseNode{
		config: config,
		db:     db,
	}

	return node, nil
}

// Execute executes the database operation
func (dn *DatabaseNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	query := dn.config.Query
	if q, exists := inputs["query"]; exists {
		if qStr, ok := q.(string); ok {
			query = qStr
		}
	}

	queryType := dn.config.QueryType
	if qt, exists := inputs["query_type"]; exists {
		if qtStr, ok := qt.(string); ok {
			queryType = qtStr
		}
	}

	// Prepare parameters
	params := make(map[string]interface{})
	
	// Start with configured parameters
	for k, v := range dn.config.Parameters {
		params[k] = v
	}
	
	// Override with inputs
	for k, v := range inputs {
		if k != "query" && k != "query_type" && k != "connection_url" {
			params[k] = v
		}
	}

	// Validate query
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Execute query with retry logic
	var result map[string]interface{}
	var err error
	
	for attempt := 0; attempt <= dn.config.MaxRetries; attempt++ {
		result, err = dn.executeQuery(ctx, query, queryType, params)
		if err == nil {
			break // Success
		}
		
		if attempt < dn.config.MaxRetries {
			// Wait before retry (exponential backoff)
			waitTime := time.Duration(attempt+1) * time.Second
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("database operation failed after %d attempts: %w", dn.config.MaxRetries+1, err)
	}

	return result, nil
}

// executeQuery executes the actual query
func (dn *DatabaseNode) executeQuery(ctx context.Context, query, queryType string, params map[string]interface{}) (map[string]interface{}, error) {
	// Validate query type
	if queryType == "" {
		// Determine query type from query itself
		queryType = dn.inferQueryType(query)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, dn.config.Timeout)
	defer cancel()

	// Execute based on query type
	switch strings.ToLower(queryType) {
	case "select", "get", "read":
		return dn.executeSelect(ctx, query, params)
	case "insert", "create":
		return dn.executeInsert(ctx, query, params)
	case "update":
		return dn.executeUpdate(ctx, query, params)
	case "delete", "remove":
		return dn.executeDelete(ctx, query, params)
	case "raw", "execute":
		return dn.executeRaw(ctx, query, params)
	default:
		return nil, fmt.Errorf("unsupported query type: %s", queryType)
	}
}

// inferQueryType infers the query type from the query string
func (dn *DatabaseNode) inferQueryType(query string) string {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	
	switch {
	case strings.HasPrefix(trimmed, "SELECT"):
		return "select"
	case strings.HasPrefix(trimmed, "INSERT"):
		return "insert"
	case strings.HasPrefix(trimmed, "UPDATE"):
		return "update"
	case strings.HasPrefix(trimmed, "DELETE"):
		return "delete"
	default:
		return "raw"
	}
}

// executeSelect executes a SELECT query
func (dn *DatabaseNode) executeSelect(ctx context.Context, query string, params map[string]interface{}) (map[string]interface{}, error) {
	// Prepare query with named parameters
	stmt, err := dn.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	// Execute query
	rows, err := stmt.QueryxContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Collect results
	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"results": results,
		"row_count": len(results),
		"query_type": "select",
		"timestamp": time.Now().Unix(),
	}, nil
}

// executeInsert executes an INSERT query
func (dn *DatabaseNode) executeInsert(ctx context.Context, query string, params map[string]interface{}) (map[string]interface{}, error) {
	// Prepare query with named parameters
	stmt, err := dn.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	// Execute query
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	insertId, err := result.LastInsertId()
	if err != nil {
		// Not all databases support LastInsertId, so we'll ignore this error for now
		insertId = -1
	}

	return map[string]interface{}{
		"success": true,
		"rows_affected": rowsAffected,
		"insert_id": insertId,
		"query_type": "insert",
		"timestamp": time.Now().Unix(),
	}, nil
}

// executeUpdate executes an UPDATE query
func (dn *DatabaseNode) executeUpdate(ctx context.Context, query string, params map[string]interface{}) (map[string]interface{}, error) {
	// Prepare query with named parameters
	stmt, err := dn.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	// Execute query
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"rows_affected": rowsAffected,
		"query_type": "update",
		"timestamp": time.Now().Unix(),
	}, nil
}

// executeDelete executes a DELETE query
func (dn *DatabaseNode) executeDelete(ctx context.Context, query string, params map[string]interface{}) (map[string]interface{}, error) {
	// Prepare query with named parameters
	stmt, err := dn.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	// Execute query
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"rows_affected": rowsAffected,
		"query_type": "delete",
		"timestamp": time.Now().Unix(),
	}, nil
}

// executeRaw executes a raw query (any type)
func (dn *DatabaseNode) executeRaw(ctx context.Context, query string, params map[string]interface{}) (map[string]interface{}, error) {
	// Determine the query type for return information
	queryType := strings.ToLower(dn.inferQueryType(query))

	// Prepare query with named parameters
	stmt, err := dn.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	// Execute query
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Some queries don't affect rows (like CREATE, DROP)
		rowsAffected = 0
	}

	response := map[string]interface{}{
		"success": true,
		"rows_affected": rowsAffected,
		"query_type": queryType,
		"timestamp": time.Now().Unix(),
	}

	// For SELECT-like queries, also return results
	if queryType == "select" {
		// Execute as select to get results
		selectResult, selectErr := dn.executeSelect(ctx, query, params)
		if selectErr == nil {
			response["results"] = selectResult["results"]
			response["row_count"] = selectResult["row_count"]
		}
	}

	return response, nil
}

// Close closes the database connection
func (dn *DatabaseNode) Close() error {
	if dn.db != nil {
		return dn.db.Close()
	}
	return nil
}

// ValidateConnection validates the database connection
func (dn *DatabaseNode) ValidateConnection(ctx context.Context) error {
	return dn.db.PingContext(ctx)
}

// GetConnectionInfo returns information about the database connection
func (dn *DatabaseNode) GetConnectionInfo() map[string]interface{} {
	info := map[string]interface{}{
		"type": dn.config.Type,
		"max_open_connections": dn.db.Stats().MaxOpenConnections,
		"open_connections": dn.db.Stats().OpenConnections,
		"in_use": dn.db.Stats().InUse,
		"idle": dn.db.Stats().Idle,
		"wait_count": dn.db.Stats().WaitCount,
		"wait_duration": dn.db.Stats().WaitDuration.String(),
		"max_idle_closed": dn.db.Stats().MaxIdleClosed,
		"max_lifetime_closed": dn.db.Stats().MaxLifetimeClosed,
	}

	return info
}

// RegisterDatabaseNode registers the database node type with the engine
func RegisterDatabaseNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("database_operation", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var dbType DatabaseType
		if typ, exists := config["type"]; exists {
			if typStr, ok := typ.(string); ok {
				dbType = DatabaseType(typStr)
			}
		} else {
			return nil, fmt.Errorf("database type is required")
		}

		var connURL string
		if url, exists := config["connection_url"]; exists {
			if urlStr, ok := url.(string); ok {
				connURL = urlStr
			}
		} else {
			return nil, fmt.Errorf("connection URL is required")
		}

		var query string
		if q, exists := config["query"]; exists {
			if qStr, ok := q.(string); ok {
				query = qStr
			}
		}

		var queryType string
		if qt, exists := config["query_type"]; exists {
			if qtStr, ok := qt.(string); ok {
				queryType = qtStr
			}
		}

		var timeout float64
		if t, exists := config["timeout_seconds"]; exists {
			if tFloat, ok := t.(float64); ok {
				timeout = tFloat
			}
		}

		var maxRetries float64
		if retries, exists := config["max_retries"]; exists {
			if retriesFloat, ok := retries.(float64); ok {
				maxRetries = retriesFloat
			}
		}

		var poolSize float64
		if size, exists := config["connection_pool_size"]; exists {
			if sizeFloat, ok := size.(float64); ok {
				poolSize = sizeFloat
			}
		}

		var sslMode string
		if mode, exists := config["ssl_mode"]; exists {
			if modeStr, ok := mode.(string); ok {
				sslMode = modeStr
			}
		}

		var enableSSL bool
		if ssl, exists := config["enable_ssl"]; exists {
			if sslBool, ok := ssl.(bool); ok {
				enableSSL = sslBool
			}
		}

		var parameters map[string]interface{}
		if params, exists := config["parameters"]; exists {
			if paramsMap, ok := params.(map[string]interface{}); ok {
				parameters = paramsMap
			}
		}

		nodeConfig := &DatabaseNodeConfig{
			Type:               dbType,
			ConnectionURL:      connURL,
			Query:              query,
			QueryType:          queryType,
			Parameters:         parameters,
			Timeout:            time.Duration(timeout) * time.Second,
			MaxRetries:         int(maxRetries),
			ConnectionPoolSize: int(poolSize),
			SSLMode:            sslMode,
			EnableSSL:          enableSSL,
		}

		return NewDatabaseNode(nodeConfig)
	})
}