package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// BunDatabaseNodeConfig represents the configuration for a Bun database node
type BunDatabaseNodeConfig struct {
	Type          string                 `json:"type"`          // "postgresql", "mysql", "sqlite"
	ConnectionURL string                 `json:"connection_url"`
	MaxIdleConns  int                    `json:"max_idle_conns"` // max idle connections
	MaxOpenConns  int                    `json:"max_open_conns"` // max open connections
	ConnLifetime  int                    `json:"conn_lifetime"`  // connection lifetime in seconds
	Query         string                 `json:"query"`          // SQL query to execute
	QueryType     string                 `json:"query_type"`     // "select", "insert", "update", "delete", "raw"
	Model         string                 `json:"model"`          // model name for Bun operations
	Values        map[string]interface{} `json:"values"`         // values for insert/update operations
	Where         map[string]interface{} `json:"where"`          // conditions for where clause
}

// BunDatabaseNode represents a database operation node using Bun ORM
type BunDatabaseNode struct {
	config *BunDatabaseNodeConfig
	db     *bun.DB
}

// NewBunDatabaseNode creates a new Bun database node
func NewBunDatabaseNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var bunConfig BunDatabaseNodeConfig
	err = json.Unmarshal(jsonData, &bunConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if bunConfig.ConnectionURL == "" {
		return nil, fmt.Errorf("connection_url is required for Bun database node")
	}

	if bunConfig.Type == "" {
		bunConfig.Type = "sqlite" // default to sqlite
	}

	// Connect to database based on type
	var dialect bun.Dialect
	var sqldb *sql.DB

	switch bunConfig.Type {
	case "postgresql", "postgres":
		sqldb, err = sql.Open("pgx", bunConfig.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
		}
		dialect = pgdialect.New()
	case "mysql":
		sqldb, err = sql.Open("mysql", bunConfig.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
		}
		dialect = mysqldialect.New()
	case "sqlite":
		sqldb, err = sql.Open("sqlite_shim", bunConfig.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite: %v", err)
		}
		dialect = sqlitedialect.New()
	default:
		return nil, fmt.Errorf("unsupported database type: %s", bunConfig.Type)
	}

	// Create Bun DB instance
	db := bun.NewDB(sqldb, dialect)

	// Configure connection pool if specified
	if bunConfig.MaxIdleConns > 0 || bunConfig.MaxOpenConns > 0 || bunConfig.ConnLifetime > 0 {
		if bunConfig.MaxIdleConns > 0 {
			sqldb.SetMaxIdleConns(bunConfig.MaxIdleConns)
		}

		if bunConfig.MaxOpenConns > 0 {
			sqldb.SetMaxOpenConns(bunConfig.MaxOpenConns)
		}

		if bunConfig.ConnLifetime > 0 {
			sqldb.SetConnMaxLifetime(time.Duration(bunConfig.ConnLifetime) * time.Second)
		}
	}

	return &BunDatabaseNode{
		config: &bunConfig,
		db:     db,
	}, nil
}

// Execute implements the NodeInstance interface
func (b *BunDatabaseNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	query := b.config.Query
	if inputQuery, ok := input["query"].(string); ok && inputQuery != "" {
		query = inputQuery
	}

	queryType := b.config.QueryType
	if inputQueryType, ok := input["query_type"].(string); ok && inputQueryType != "" {
		queryType = inputQueryType
	}

	model := b.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	values := b.config.Values
	if inputValues, ok := input["values"].(map[string]interface{}); ok {
		values = inputValues
	}

	where := b.config.Where
	if inputWhere, ok := input["where"].(map[string]interface{}); ok {
		where = inputWhere
	}

	var result interface{}
	var err error

	switch queryType {
	case "select", "find":
		result, err = b.executeQuery(query, model, where)
	case "insert", "create":
		result, err = b.executeInsert(model, values)
	case "update":
		result, err = b.executeUpdate(model, values, where)
	case "delete":
		result, err = b.executeDelete(model, where)
	case "raw":
		result, err = b.executeRaw(query, values)
	default:
		// Default to select if query type is not specified
		result, err = b.executeQuery(query, model, where)
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
			"database":    b.config.Type,
			"timestamp":   time.Now().Unix(),
		},
		"timestamp": time.Now().Unix(),
	}, nil
}

// executeQuery executes a SELECT query
func (b *BunDatabaseNode) executeQuery(query string, model string, where map[string]interface{}) (interface{}, error) {
	// For now, execute a raw query since Bun would need specific model types
	rows, err := b.db.Query(query)
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
func (b *BunDatabaseNode) executeInsert(model string, values map[string]interface{}) (interface{}, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("values for insert are required")
	}

	// Since Bun requires specific struct types, we'll use a raw query approach
	query := fmt.Sprintf("INSERT INTO %s", model)
	
	// Build placeholders and values
	placeholders := make([]string, 0, len(values))
	insertValues := make([]interface{}, 0, len(values))
	for key, value := range values {
		placeholders = append(placeholders, key)
		insertValues = append(insertValues, value)
	}
	
	placeholderStr := "(" + strings.Join(placeholders, ", ") + ")"
	placeholdersFormat := "(" + strings.Repeat("?, ", len(placeholders)-1) + "?)"
	
	query = query + placeholderStr + " VALUES " + placeholdersFormat

	result, err := b.db.Exec(query, insertValues...)
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
func (b *BunDatabaseNode) executeUpdate(model string, values map[string]interface{}, where map[string]interface{}) (interface{}, error) {
	// For now using raw query approach
	setClause := ""
	setValues := []interface{}{}
	
	for key, value := range values {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = ?", key)
		setValues = append(setValues, value)
	}
	
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
	result, err := b.db.Exec(query, allValues...)
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
func (b *BunDatabaseNode) executeDelete(model string, where map[string]interface{}) (interface{}, error) {
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
	
	result, err := b.db.Exec(query, whereValues...)
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
func (b *BunDatabaseNode) executeRaw(query string, args map[string]interface{}) (interface{}, error) {
	argsList := make([]interface{}, 0, len(args))
	for _, v := range args {
		argsList = append(argsList, v)
	}
	
	result, err := b.db.Exec(query, argsList...)
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
func (b *BunDatabaseNode) GetType() string {
	return "bun_database"
}

// GetID returns a unique ID for the node instance
func (b *BunDatabaseNode) GetID() string {
	return "bun_db_" + fmt.Sprintf("%d", time.Now().Unix())
}