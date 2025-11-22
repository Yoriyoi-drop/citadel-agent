package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// EntDatabaseNodeConfig represents the configuration for an Ent database node
type EntDatabaseNodeConfig struct {
	Type          string                 `json:"type"`          // "postgresql", "mysql", "sqlite"
	ConnectionURL string                 `json:"connection_url"`
	MaxIdleConns  int                    `json:"max_idle_conns"` // max idle connections
	MaxOpenConns  int                    `json:"max_open_conns"` // max open connections
	ConnLifetime  int                    `json:"conn_lifetime"`  // connection lifetime in seconds
	Query         string                 `json:"query"`          // SQL query to execute
	QueryType     string                 `json:"query_type"`     // "select", "insert", "update", "delete", "raw"
	Model         string                 `json:"model"`          // model name for Ent operations
	Values        map[string]interface{} `json:"values"`         // values for insert/update operations
	Where         map[string]interface{} `json:"where"`          // conditions for where clause
	SchemaPath    string                 `json:"schema_path"`    // path to Ent schema
}

// EntDatabaseNode represents a database operation node using Ent ORM
type EntDatabaseNode struct {
	config *EntDatabaseNodeConfig
	db     *entsql.Driver
}

// NewEntDatabaseNode creates a new Ent database node
func NewEntDatabaseNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var entConfig EntDatabaseNodeConfig
	err = json.Unmarshal(jsonData, &entConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if entConfig.ConnectionURL == "" {
		return nil, fmt.Errorf("connection_url is required for Ent database node")
	}

	if entConfig.Type == "" {
		entConfig.Type = "sqlite" // default to sqlite
	}

	// Connect to database based on type
	var dbDriver dialect.Driver
	var sqlDB *sql.DB

	switch entConfig.Type {
	case "postgresql", "postgres":
		sqlDB, err = sql.Open("pgx", entConfig.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
		}
	case "mysql":
		sqlDB, err = sql.Open("mysql", entConfig.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
		}
	case "sqlite":
		sqlDB, err = sql.Open("sqlite3", entConfig.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", entConfig.Type)
	}

	// Create Ent SQL driver
	dbDriver = entsql.OpenDB(sqlDB)

	// Configure connection pool if specified
	if entConfig.MaxIdleConns > 0 || entConfig.MaxOpenConns > 0 || entConfig.ConnLifetime > 0 {
		if entConfig.MaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(entConfig.MaxIdleConns)
		}

		if entConfig.MaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(entConfig.MaxOpenConns)
		}

		if entConfig.ConnLifetime > 0 {
			sqlDB.SetConnMaxLifetime(time.Duration(entConfig.ConnLifetime) * time.Second)
		}
	}

	return &EntDatabaseNode{
		config: &entConfig,
		db:     dbDriver,
	}, nil
}

// Execute implements the NodeInstance interface
func (e *EntDatabaseNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	query := e.config.Query
	if inputQuery, ok := input["query"].(string); ok && inputQuery != "" {
		query = inputQuery
	}

	queryType := e.config.QueryType
	if inputQueryType, ok := input["query_type"].(string); ok && inputQueryType != "" {
		queryType = inputQueryType
	}

	model := e.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	values := e.config.Values
	if inputValues, ok := input["values"].(map[string]interface{}); ok {
		values = inputValues
	}

	where := e.config.Where
	if inputWhere, ok := input["where"].(map[string]interface{}); ok {
		where = inputWhere
	}

	var result interface{}
	var err error

	switch queryType {
	case "select", "find":
		result, err = e.executeQuery(query, model, where)
	case "insert", "create":
		result, err = e.executeInsert(model, values)
	case "update":
		result, err = e.executeUpdate(model, values, where)
	case "delete":
		result, err = e.executeDelete(model, where)
	case "raw":
		result, err = e.executeRaw(query, values)
	default:
		// Default to select if query type is not specified
		result, err = e.executeQuery(query, model, where)
	}

	if err != nil {
		return map[string]interface{}{
			"status":    "error",
			"error":     err.Error(),
			"timestamp": time.Now().Unix(),
		}, nil
	}

	return map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"result":      result,
			"query_type":  queryType,
			"model":       model,
			"database":    e.config.Type,
			"timestamp":   time.Now().Unix(),
		},
		"timestamp": time.Now().Unix(),
	}, nil
}

// executeQuery executes a SELECT query
func (e *EntDatabaseNode) executeQuery(query string, model string, where map[string]interface{}) (interface{}, error) {
	// Since Ent requires generated code for specific models, we'll execute raw SQL
	// In a real implementation, you would have generated Ent models and use their methods
	db := e.db.DB()
	
	rows, err := db.Query(query)
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
func (e *EntDatabaseNode) executeInsert(model string, values map[string]interface{}) (interface{}, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("values for insert are required")
	}

	// Since Ent requires generated code for specific models, we'll execute raw SQL
	db := e.db.DB()
	
	// Build placeholders and values
	placeholders := make([]string, 0, len(values))
	insertValues := make([]interface{}, 0, len(values))
	for key, value := range values {
		placeholders = append(placeholders, key)
		insertValues = append(insertValues, value)
	}
	
	placeholderStr := "(" + strings.Join(placeholders, ", ") + ")"
	placeholdersFormat := "(" + strings.Repeat("?, ", len(placeholders)-1) + "?)"
	
	query := fmt.Sprintf("INSERT INTO %s %s VALUES %s", model, placeholderStr, placeholdersFormat)
	
	result, err := db.Exec(query, insertValues...)
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

// executeUpdate executes an UPDATE query
func (e *EntDatabaseNode) executeUpdate(model string, values map[string]interface{}, where map[string]interface{}) (interface{}, error) {
	db := e.db.DB()
	
	// Build SET clause
	setClause := ""
	setValues := []interface{}{}
	
	for key, value := range values {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = ?", key)
		setValues = append(setValues, value)
	}
	
	// Build WHERE clause
	whereClause := ""
	whereValues := []interface{}{}
	
	for key, value := range where {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("%s = ?", key)
		whereValues = append(whereValues, value)
	}
	
	query := fmt.Sprintf("UPDATE %s SET %s", model, setClause)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	
	allValues := append(setValues, whereValues...)
	result, err := db.Exec(query, allValues...)
	if err != nil {
		return nil, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"rows_affected": rowsAffected,
		"updated_fields": len(values),
	}, nil
}

// executeDelete executes a DELETE query
func (e *EntDatabaseNode) executeDelete(model string, where map[string]interface{}) (interface{}, error) {
	db := e.db.DB()
	
	query := fmt.Sprintf("DELETE FROM %s", model)
	
	whereValues := []interface{}{}
	whereClause := ""
	
	for key, value := range where {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("%s = ?", key)
		whereValues = append(whereValues, value)
	}
	
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	
	result, err := db.Exec(query, whereValues...)
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
func (e *EntDatabaseNode) executeRaw(query string, args map[string]interface{}) (interface{}, error) {
	db := e.db.DB()
	
	argsList := make([]interface{}, 0, len(args))
	for _, v := range args {
		argsList = append(argsList, v)
	}
	
	result, err := db.Exec(query, argsList...)
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

// GetType returns the type of the node
func (e *EntDatabaseNode) GetType() string {
	return "ent_database"
}

// GetID returns a unique ID for the node instance
func (e *EntDatabaseNode) GetID() string {
	return "ent_db_" + fmt.Sprintf("%d", time.Now().Unix())
}