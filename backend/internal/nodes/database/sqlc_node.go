package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
)

// SQLCNodeConfig represents the configuration for a SQLC database node
type SQLCNodeConfig struct {
	Type          string                 `json:"type"`          // "postgresql", "mysql", "sqlite"
	ConnectionURL string                 `json:"connection_url"`
	MaxIdleConns  int                    `json:"max_idle_conns"` // max idle connections
	MaxOpenConns  int                    `json:"max_open_conns"` // max open connections
	ConnLifetime  int                    `json:"conn_lifetime"`  // connection lifetime in seconds
	Query         string                 `json:"query"`          // SQL query to execute
	QueryType     string                 `json:"query_type"`     // "select", "insert", "update", "delete", "raw"
	Params        map[string]interface{} `json:"params"`         // parameters for the query
	SchemaPath    string                 `json:"schema_path"`    // path to SQLC schema
	SQLFilePath   string                 `json:"sql_file_path"`  // path to SQL file
	QueryName     string                 `json:"query_name"`     // name of the specific query to run
}

// SQLCNode represents a database operation node using SQLC approach
type SQLCNode struct {
	config *SQLCNodeConfig
	db     *sql.DB
}

// NewSQLCNode creates a new SQLC database node
func NewSQLCNode(config map[string]interface{}) (engine.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var sqlcConfig SQLCNodeConfig
	err = json.Unmarshal(jsonData, &sqlcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if sqlcConfig.ConnectionURL == "" {
		return nil, fmt.Errorf("connection_url is required for SQLC database node")
	}

	if sqlcConfig.Type == "" {
		sqlcConfig.Type = "sqlite" // default to sqlite
	}

	// Connect to database based on type
	var driverName string

	switch sqlcConfig.Type {
	case "postgresql", "postgres":
		driverName = "pgx"
	case "mysql":
		driverName = "mysql"
	case "sqlite":
		driverName = "sqlite3"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", sqlcConfig.Type)
	}

	db, err := sql.Open(driverName, sqlcConfig.ConnectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool if specified
	if sqlcConfig.MaxIdleConns > 0 || sqlcConfig.MaxOpenConns > 0 || sqlcConfig.ConnLifetime > 0 {
		if sqlcConfig.MaxIdleConns > 0 {
			db.SetMaxIdleConns(sqlcConfig.MaxIdleConns)
		}

		if sqlcConfig.MaxOpenConns > 0 {
			db.SetMaxOpenConns(sqlcConfig.MaxOpenConns)
		}

		if sqlcConfig.ConnLifetime > 0 {
			db.SetConnMaxLifetime(time.Duration(sqlcConfig.ConnLifetime) * time.Second)
		}
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &SQLCNode{
		config: &sqlcConfig,
		db:     db,
	}, nil
}

// Execute implements the NodeInstance interface
func (s *SQLCNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	// Override configuration with input values if provided
	query := s.config.Query
	if inputQuery, ok := input["query"].(string); ok && inputQuery != "" {
		query = inputQuery
	}

	queryType := s.config.QueryType
	if inputQueryType, ok := input["query_type"].(string); ok && inputQueryType != "" {
		queryType = inputQueryType
	}

	params := s.config.Params
	if inputParams, ok := input["params"].(map[string]interface{}); ok {
		params = inputParams
	}

	var result interface{}
	var err error

	// If SQL file path is provided and query is empty, try to read query from file
	if query == "" && s.config.SQLFilePath != "" {
		// In a real implementation, this would read the query from the specified file
		// For now, we'll skip this feature and rely on query in config
	}

	switch queryType {
	case "select", "find":
		result, err = s.executeSelect(query, params)
	case "insert", "create":
		result, err = s.executeInsert(query, params)
	case "update":
		result, err = s.executeUpdate(query, params)
	case "delete":
		result, err = s.executeDelete(query, params)
	case "raw":
		result, err = s.executeRaw(query, params)
	default:
		// Default to select if query type is not specified
		result, err = s.executeSelect(query, params)
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
			"result":      result,
			"query_type":  queryType,
			"database":    s.config.Type,
			"query":       query,
			"timestamp":   time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}, nil
}

// executeSelect executes a SELECT query
func (s *SQLCNode) executeSelect(query string, params map[string]interface{}) (interface{}, error) {
	// Convert named parameters to positional parameters
	convertedQuery, args := s.convertNamedParams(query, params)

	rows, err := s.db.Query(convertedQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Iterate through the rows
	var result []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the result into the column pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create our map and retrieve the value for each column from the pointers slice
		entry := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				entry[col] = string(b)
			} else {
				entry[col] = val
			}
		}
		result = append(result, entry)
	}

	return result, nil
}

// executeInsert executes an INSERT query
func (s *SQLCNode) executeInsert(query string, params map[string]interface{}) (interface{}, error) {
	// Convert named parameters to positional parameters
	convertedQuery, args := s.convertNamedParams(query, params)

	result, err := s.db.Exec(convertedQuery, args...)
	if err != nil {
		return nil, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		// Not all databases support LastInsertId, so we'll just return 0 if not supported
		id = 0
	}
	
	return map[string]interface{}{
		"rows_affected": rowsAffected,
		"last_insert_id": id,
	}, nil
}

// executeUpdate executes an UPDATE query
func (s *SQLCNode) executeUpdate(query string, params map[string]interface{}) (interface{}, error) {
	// Convert named parameters to positional parameters
	convertedQuery, args := s.convertNamedParams(query, params)

	result, err := s.db.Exec(convertedQuery, args...)
	if err != nil {
		return nil, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"rows_affected": rowsAffected,
	}, nil
}

// executeDelete executes a DELETE query
func (s *SQLCNode) executeDelete(query string, params map[string]interface{}) (interface{}, error) {
	// Convert named parameters to positional parameters
	convertedQuery, args := s.convertNamedParams(query, params)

	result, err := s.db.Exec(convertedQuery, args...)
	if err != nil {
		return nil, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"rows_affected": rowsAffected,
	}, nil
}

// executeRaw executes a raw SQL query
func (s *SQLCNode) executeRaw(query string, params map[string]interface{}) (interface{}, error) {
	// Convert named parameters to positional parameters
	convertedQuery, args := s.convertNamedParams(query, params)

	result, err := s.db.Exec(convertedQuery, args...)
	if err != nil {
		return nil, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"rows_affected": rowsAffected,
	}, nil
}

// convertNamedParams converts named parameters (:param) to positional parameters (?, ?, ?)
// and returns the converted query and the argument slice
func (s *SQLCNode) convertNamedParams(query string, params map[string]interface{}) (string, []interface{}) {
	args := make([]interface{}, 0, len(params))
	
	// Simple replacement of named parameters with positional ones
	// In a real SQLC implementation, you would use the generated code
	convertedQuery := query
	
	for key, value := range params {
		placeholder := ":" + key
		if strings.Contains(convertedQuery, placeholder) {
			// Replace named parameter with positional parameter (?)
			convertedQuery = strings.ReplaceAll(convertedQuery, placeholder, "?")
			args = append(args, value)
		}
	}
	
	return convertedQuery, args
}

// GetType returns the type of the node
func (s *SQLCNode) GetType() string {
	return "sqlc_database"
}

// GetID returns a unique ID for the node instance
func (s *SQLCNode) GetID() string {
	return "sqlc_db_" + fmt.Sprintf("%d", time.Now().Unix())
}