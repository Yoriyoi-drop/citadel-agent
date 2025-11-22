package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// GORMDatabaseNodeConfig represents the configuration for a GORM database node
type GORMDatabaseNodeConfig struct {
	Type          string                 `json:"type"`          // "postgresql", "mysql", "sqlite"
	ConnectionURL string                 `json:"connection_url"`
	AutoMigrate   bool                   `json:"auto_migrate"`   // whether to run auto-migration
	MaxIdleConns  int                    `json:"max_idle_conns"` // max idle connections
	MaxOpenConns  int                    `json:"max_open_conns"` // max open connections
	ConnLifetime  int                    `json:"conn_lifetime"`  // connection lifetime in seconds
	Query         string                 `json:"query"`          // SQL query to execute
	QueryType     string                 `json:"query_type"`     // "select", "insert", "update", "delete", "raw"
	Model         string                 `json:"model"`          // model name for GORM operations
	Values        map[string]interface{} `json:"values"`         // values for insert/update operations
	Where         map[string]interface{} `json:"where"`          // conditions for where clause
}

// GORMDatabaseNode represents a database operation node using GORM
type GORMDatabaseNode struct {
	config *GORMDatabaseNodeConfig
	db     *gorm.DB
}

// NewGORMDatabaseNode creates a new GORM database node
func NewGORMDatabaseNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert interface{} map to JSON and back to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	var gormConfig GORMDatabaseNodeConfig
	err = json.Unmarshal(jsonData, &gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// Validate required fields
	if gormConfig.ConnectionURL == "" {
		return nil, fmt.Errorf("connection_url is required for GORM database node")
	}

	if gormConfig.Type == "" {
		gormConfig.Type = "sqlite" // default to sqlite
	}

	// Connect to database based on type
	var dialector gorm.Dialector
	switch gormConfig.Type {
	case "postgresql", "postgres":
		dialector = postgres.Open(gormConfig.ConnectionURL)
	case "mysql":
		dialector = mysql.Open(gormConfig.ConnectionURL)
	case "sqlite":
		dialector = sqlite.Open(gormConfig.ConnectionURL)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", gormConfig.Type)
	}

	// Initialize GORM database instance
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool if specified
	if gormConfig.MaxIdleConns > 0 || gormConfig.MaxOpenConns > 0 || gormConfig.ConnLifetime > 0 {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get sql.DB: %v", err)
		}

		if gormConfig.MaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(gormConfig.MaxIdleConns)
		}

		if gormConfig.MaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(gormConfig.MaxOpenConns)
		}

		if gormConfig.ConnLifetime > 0 {
			sqlDB.SetConnMaxLifetime(time.Duration(gormConfig.ConnLifetime) * time.Second)
		}
	}

	// Run auto-migration if enabled
	if gormConfig.AutoMigrate {
		// In a real implementation, you would migrate specific models
		// This is a simplified version
	}

	return &GORMDatabaseNode{
		config: &gormConfig,
		db:     db,
	}, nil
}

// Execute implements the NodeInstance interface
func (g *GORMDatabaseNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Override configuration with input values if provided
	query := g.config.Query
	if inputQuery, ok := input["query"].(string); ok && inputQuery != "" {
		query = inputQuery
	}

	queryType := g.config.QueryType
	if inputQueryType, ok := input["query_type"].(string); ok && inputQueryType != "" {
		queryType = inputQueryType
	}

	model := g.config.Model
	if inputModel, ok := input["model"].(string); ok && inputModel != "" {
		model = inputModel
	}

	values := g.config.Values
	if inputValues, ok := input["values"].(map[string]interface{}); ok {
		values = inputValues
	}

	where := g.config.Where
	if inputWhere, ok := input["where"].(map[string]interface{}); ok {
		where = inputWhere
	}

	var result interface{}
	var err error

	switch queryType {
	case "select", "find":
		result, err = g.executeQuery(query, model, where)
	case "insert", "create":
		result, err = g.executeInsert(model, values)
	case "update":
		result, err = g.executeUpdate(model, values, where)
	case "delete":
		result, err = g.executeDelete(model, where)
	case "raw":
		result, err = g.executeRaw(query, values)
	default:
		// Default to select if query type is not specified
		result, err = g.executeQuery(query, model, where)
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
			"database":    g.config.Type,
			"timestamp":   time.Now().Unix(),
		},
		"timestamp": time.Now().Unix(),
	}, nil
}

// executeQuery executes a SELECT query
func (g *GORMDatabaseNode) executeQuery(query string, model string, where map[string]interface{}) (interface{}, error) {
	// This is a simplified implementation
	// In a real system, you would need to define model structures dynamically or use a generic approach
	var result []map[string]interface{}

	// Build the query with where conditions
	db := g.db
	for key, value := range where {
		db = db.Where(key, value)
	}

	// For now, execute the raw query as GORM might need specific model types
	rows, err := db.Raw(query).Rows()
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
func (g *GORMDatabaseNode) executeInsert(model string, values map[string]interface{}) (interface{}, error) {
	// For now, we'll execute a raw insert
	// In a real implementation, you would use GORM's Create method with dynamic models
	if len(values) == 0 {
		return nil, fmt.Errorf("values for insert are required")
	}

	// Since we don't have specific model types, we'll build a raw query
	query := fmt.Sprintf("INSERT INTO %s", model)
	
	// Build placeholders and values
	placeholders := make([]string, 0, len(values))
	insertValues := make([]interface{}, 0, len(values))
	for key, value := range values {
		placeholders = append(placeholders, key)
		insertValues = append(insertValues, value)
	}
	
	placeholderStr := "(" + strings.Join(placeholders, ", ") + ")"
	placeholdersFormat := "(" + strings.Repeat("?, ", len(placeholders)-1) + "?"
	
	// Execute raw query
	result := g.db.Exec("INSERT INTO ? (?) VALUES ?", model, strings.Join(placeholders, ", "), insertValues...)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return map[string]interface{}{
		"rows_affected": result.RowsAffected,
		"id":            result.Statement.Vars, // This is a simplification
	}, nil
}

// executeUpdate executes an UPDATE query
func (g *GORMDatabaseNode) executeUpdate(model string, values map[string]interface{}, where map[string]interface{}) (interface{}, error) {
	db := g.db.Model(&model) // This is a simplification

	// Apply where conditions
	for key, value := range where {
		db = db.Where(key, value)
	}

	// Update with the provided values
	result := db.Updates(values)
	if result.Error != nil {
		return nil, result.Error
	}

	return map[string]interface{}{
		"rows_affected": result.RowsAffected,
		"updated_data":  values,
	}, nil
}

// executeDelete executes a DELETE query
func (g *GORMDatabaseNode) executeDelete(model string, where map[string]interface{}) (interface{}, error) {
	db := g.db // Using raw DB for deletion
	
	// For deletion, we'll use a raw query approach
	whereClause := ""
	whereValues := []interface{}{}
	
	for key, value := range where {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("%s = ?", key)
		whereValues = append(whereValues, value)
	}
	
	query := fmt.Sprintf("DELETE FROM %s", model)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	
	result := db.Exec(query, whereValues...)
	if result.Error != nil {
		return nil, result.Error
	}
	
	return map[string]interface{}{
		"rows_affected": result.RowsAffected,
	}, nil
}

// executeRaw executes a raw SQL query
func (g *GORMDatabaseNode) executeRaw(query string, args []interface{}) (interface{}, error) {
	result := g.db.Exec(query, args...)
	if result.Error != nil {
		return nil, result.Error
	}
	
	return map[string]interface{}{
		"rows_affected": result.RowsAffected,
	}, nil
}

// GetType returns the type of the node
func (g *GORMDatabaseNode) GetType() string {
	return "gorm_database"
}

// GetID returns a unique ID for the node instance
func (g *GORMDatabaseNode) GetID() string {
	return "gorm_db_" + fmt.Sprintf("%d", time.Now().Unix())
}