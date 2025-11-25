package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3"    // SQLite driver

	"citadel-agent/backend/internal/interfaces"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MySQL      DatabaseType = "mysql"
	SQLite     DatabaseType = "sqlite"
	MongoDB    DatabaseType = "mongodb" // Would need a different implementation
)

// QueryType represents the type of database operation
type QueryType string

const (
	QuerySelect QueryType = "select"
	QueryInsert QueryType = "insert"
	QueryUpdate QueryType = "update"
	QueryDelete QueryType = "delete"
	QueryRaw    QueryType = "raw"
)

// DatabaseConfig represents the configuration for a database node
type DatabaseConfig struct {
	DBType           DatabaseType           `json:"db_type"`
	ConnectionString string                 `json:"connection_string"`
	Query            string                 `json:"query"`
	QueryType        QueryType              `json:"query_type"`
	Parameters       map[string]interface{} `json:"parameters"`
	TableName        string                 `json:"table_name"`
	Fields           []string               `json:"fields"`
	WhereClause      string                 `json:"where_clause"`
	OrderBy          string                 `json:"order_by"`
	Limit            int                    `json:"limit"`
	EnableCaching    bool                   `json:"enable_caching"`
	CacheTTL         int                    `json:"cache_ttl"` // in seconds
	EnableProfiling  bool                   `json:"enable_profiling"`
	ReturnRawResults bool                   `json:"return_raw_results"`
	CustomParams     map[string]interface{} `json:"custom_params"`
	Timeout          int                    `json:"timeout"` // in seconds
}

// DatabaseNode represents a database operation node
type DatabaseNode struct {
	config    *DatabaseConfig
	validator *SQLValidator
}

// NewDatabaseNode creates a new database node
func NewDatabaseNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	// Convert config map to struct
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var dbConfig DatabaseConfig
	if err := json.Unmarshal(jsonData, &dbConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate and set defaults
	if dbConfig.DBType == "" {
		dbConfig.DBType = PostgreSQL
	}

	if dbConfig.QueryType == "" {
		dbConfig.QueryType = QuerySelect
	}

	if dbConfig.Timeout == 0 {
		dbConfig.Timeout = 30 // 30 seconds default
	}

	if dbConfig.CacheTTL == 0 {
		dbConfig.CacheTTL = 3600 // 1 hour default cache TTL
	}

	if dbConfig.Parameters == nil {
		dbConfig.Parameters = make(map[string]interface{})
	}

	return &DatabaseNode{
		config:    &dbConfig,
		validator: NewSQLValidator(),
	}, nil
}

// Execute executes the database operation
func (dn *DatabaseNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()

	// Override config values with inputs if provided
	dbType := dn.config.DBType
	if inputDBType, exists := inputs["db_type"]; exists {
		if dbTypeStr, ok := inputDBType.(string); ok && dbTypeStr != "" {
			switch dbTypeStr {
			case "postgresql", "postgres":
				dbType = PostgreSQL
			case "mysql":
				dbType = MySQL
			case "sqlite":
				dbType = SQLite
			case "mongodb":
				dbType = MongoDB
			}
		}
	}

	connectionString := dn.config.ConnectionString
	if inputConnStr, exists := inputs["connection_string"]; exists {
		if connStr, ok := inputConnStr.(string); ok {
			connectionString = connStr
		}
	}

	query := dn.config.Query
	if inputQuery, exists := inputs["query"]; exists {
		if q, ok := inputQuery.(string); ok {
			query = q
		}
	}

	queryType := dn.config.QueryType
	if inputQueryType, exists := inputs["query_type"]; exists {
		if qType, ok := inputQueryType.(string); ok && qType != "" {
			switch qType {
			case "select":
				queryType = QuerySelect
			case "insert":
				queryType = QueryInsert
			case "update":
				queryType = QueryUpdate
			case "delete":
				queryType = QueryDelete
			case "raw":
				queryType = QueryRaw
			}
		}
	}

	tableName := dn.config.TableName
	if inputTable, exists := inputs["table_name"]; exists {
		if table, ok := inputTable.(string); ok {
			tableName = table
		}
	}

	parameters := dn.config.Parameters
	if inputParams, exists := inputs["parameters"]; exists {
		if paramMap, ok := inputParams.(map[string]interface{}); ok {
			parameters = make(map[string]interface{})
			for k, v := range paramMap {
				parameters[k] = v
			}
		}
	}

	whereClause := dn.config.WhereClause
	if inputWhere, exists := inputs["where_clause"]; exists {
		if where, ok := inputWhere.(string); ok {
			whereClause = where
		}
	}

	orderBy := dn.config.OrderBy
	if inputOrder, exists := inputs["order_by"]; exists {
		if order, ok := inputOrder.(string); ok {
			orderBy = order
		}
	}

	limit := dn.config.Limit
	if inputLimit, exists := inputs["limit"]; exists {
		if limitFloat, ok := inputLimit.(float64); ok {
			limit = int(limitFloat)
		}
	}

	timeout := dn.config.Timeout
	if inputTimeout, exists := inputs["timeout"]; exists {
		if timeoutFloat, ok := inputTimeout.(float64); ok {
			timeout = int(timeoutFloat)
		}
	}

	enableProfiling := dn.config.EnableProfiling
	if inputEnableProfiling, exists := inputs["enable_profiling"]; exists {
		if prof, ok := inputEnableProfiling.(bool); ok {
			enableProfiling = prof
		}
	}

	returnRawResults := dn.config.ReturnRawResults
	if inputReturnRaw, exists := inputs["return_raw_results"]; exists {
		if raw, ok := inputReturnRaw.(bool); ok {
			returnRawResults = raw
		}
	}

	// Connect to database using connection pool
	pool := GetGlobalPool()
	db, err := pool.GetConnection(string(dbType), connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	// Don't defer db.Close() - connection is managed by pool

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Validate inputs before building query
	if tableName != "" {
		if err := dn.validator.ValidateTableName(tableName); err != nil {
			return nil, fmt.Errorf("invalid table name: %w", err)
		}
	}

	if orderBy != "" {
		if err := dn.validator.ValidateOrderBy(orderBy); err != nil {
			return nil, fmt.Errorf("invalid ORDER BY clause: %w", err)
		}
	}

	if whereClause != "" {
		sanitized, err := dn.validator.SanitizeWhereClause(whereClause)
		if err != nil {
			return nil, fmt.Errorf("invalid WHERE clause: %w", err)
		}
		whereClause = sanitized
	}

	if limit > 0 {
		if err := dn.validator.ValidateLimit(limit); err != nil {
			return nil, fmt.Errorf("invalid LIMIT: %w", err)
		}
	}

	// Build query if not provided
	if query == "" {
		query, err = dn.buildQuery(queryType, tableName, whereClause, orderBy, limit, parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to build query: %w", err)
		}
	}

	// Execute the query
	var result interface{}
	switch queryType {
	case QuerySelect:
		result, err = dn.executeQuery(ctx, db, query, parameters)
	case QueryInsert, QueryUpdate, QueryDelete:
		result, err = dn.executeUpdate(ctx, db, query, parameters)
	case QueryRaw:
		// Determine if it's a SELECT or other type by checking the query string
		lowerQuery := strings.ToLower(strings.TrimSpace(query))
		if strings.HasPrefix(lowerQuery, "select") {
			result, err = dn.executeQuery(ctx, db, query, parameters)
		} else {
			result, err = dn.executeUpdate(ctx, db, query, parameters)
		}
	default:
		return nil, fmt.Errorf("unsupported query type: %s", queryType)
	}

	if err != nil {
		return nil, fmt.Errorf("database operation failed: %w", err)
	}

	// Prepare output
	output := make(map[string]interface{})
	output["success"] = true
	output["db_type"] = string(dbType)
	output["query_type"] = string(queryType)
	output["query_executed"] = query
	output["result"] = result
	output["input_parameters"] = parameters
	output["connection_used"] = connectionString // Note: In production, never expose connection string
	output["timestamp"] = time.Now().Unix()
	output["execution_time"] = time.Since(startTime).Seconds()

	if returnRawResults {
		output["raw_query"] = query
		output["raw_result"] = result
		output["raw_inputs"] = inputs
	}

	// Add profiling data if enabled
	if enableProfiling {
		output["profiling"] = map[string]interface{}{
			"start_time":    startTime.Unix(),
			"end_time":      time.Now().Unix(),
			"duration":      time.Since(startTime).Seconds(),
			"db_type":       string(dbType),
			"query_type":    string(queryType),
			"query":         query,
			"rows_affected": 0, // This would be populated based on result
		}
	}

	return output, nil
}

// buildQuery builds a SQL query based on the configuration
func (dn *DatabaseNode) buildQuery(qType QueryType, table, where, orderBy string, limit int, params map[string]interface{}) (string, error) {
	switch qType {
	case QuerySelect:
		query := "SELECT "

		if len(params) > 0 {
			// Check if fields are specified in the params
			fieldsParam, exists := params["fields"]
			if exists {
				if fields, ok := fieldsParam.([]interface{}); ok {
					fieldNames := make([]string, len(fields))
					for i, f := range fields {
						fieldNames[i] = fmt.Sprintf("%v", f)
					}
					query += fmt.Sprintf("%s", joinStrings(fieldNames, ", "))
				} else {
					// If fields not provided in params, use all fields
					query += "*"
				}
			} else {
				// Use all fields
				query += "*"
			}
		} else {
			query += "*"
		}

		query += fmt.Sprintf(" FROM %s", table)

		if where != "" {
			query += fmt.Sprintf(" WHERE %s", where)
		}

		if orderBy != "" {
			query += fmt.Sprintf(" ORDER BY %s", orderBy)
		}

		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", limit)
		}

		return query, nil

	case QueryInsert:
		query := fmt.Sprintf("INSERT INTO %s ", table)

		if len(params) > 0 {
			// Build column list and value placeholders
			columns := make([]string, 0, len(params))
			values := make([]string, 0, len(params))

			for k := range params {
				if k != "fields" && k != "where" { // Skip special parameters
					columns = append(columns, k)
					// Add placeholder - in a real implementation, proper escaping would be used
					values = append(values, fmt.Sprintf("$%d", len(values)+1))
				}
			}

			query += fmt.Sprintf("(%s) VALUES (%s)",
				joinStrings(columns, ", "),
				joinStrings(values, ", "))
		}

		return query, nil

	case QueryUpdate:
		query := fmt.Sprintf("UPDATE %s SET ", table)

		if len(params) > 0 {
			setParts := make([]string, 0, len(params))
			paramIndex := 1

			for k := range params {
				if k != "fields" && k != "where" && k != "id" { // Skip special parameters
					setParts = append(setParts, fmt.Sprintf("%s = $%d", k, paramIndex))
					paramIndex++
				}
			}

			query += joinStrings(setParts, ", ")
		}

		if where != "" {
			query += fmt.Sprintf(" WHERE %s", where)
		}

		return query, nil

	case QueryDelete:
		query := fmt.Sprintf("DELETE FROM %s", table)

		if where != "" {
			query += fmt.Sprintf(" WHERE %s", where)
		}

		return query, nil

	default:
		return "", fmt.Errorf("unsupported query type for building: %s", qType)
	}
}

// executeQuery executes a SELECT query
func (dn *DatabaseNode) executeQuery(ctx context.Context, db *sql.DB, query string, params map[string]interface{}) (interface{}, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare result slice
	var results []map[string]interface{}

	// Iterate through rows
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column's value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the result into the column pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create our map and retrieve the value for each column from the pointers slice
		entry := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				entry[col] = nil
			} else {
				// Handle different types appropriately
				switch valTyped := val.(type) {
				case []byte:
					entry[col] = string(valTyped)
				default:
					entry[col] = valTyped
				}
			}
		}

		results = append(results, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

// executeUpdate executes INSERT, UPDATE, DELETE queries
func (dn *DatabaseNode) executeUpdate(ctx context.Context, db *sql.DB, query string, params map[string]interface{}) (interface{}, error) {
	// Prepare parameters for execution (this is simplified)
	// This is where we'd extract parameter values in the right order
	// For now, we'll just execute without parameters
	result, err := db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return map[string]interface{}{
		"rows_affected": rowsAffected,
		"success":       true,
	}, nil
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

// GetType returns the type of node
func (dn *DatabaseNode) GetType() string {
	return "database_query"
}

// GetID returns the unique ID of the node instance
func (dn *DatabaseNode) GetID() string {
	return fmt.Sprintf("db_%s_%s_%d", dn.config.DBType, dn.config.QueryType, time.Now().Unix())
}
