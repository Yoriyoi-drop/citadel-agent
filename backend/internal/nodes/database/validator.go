package database

import (
	"fmt"
	"regexp"
	"strings"
)

// SQLValidator provides SQL injection protection
type SQLValidator struct {
	// Whitelist of allowed table names
	allowedTables map[string]bool
	// Whitelist of allowed column names
	allowedColumns map[string]bool
	// Pattern for valid identifiers (alphanumeric + underscore)
	identifierPattern *regexp.Regexp
}

// NewSQLValidator creates a new SQL validator
func NewSQLValidator() *SQLValidator {
	return &SQLValidator{
		allowedTables:     make(map[string]bool),
		allowedColumns:    make(map[string]bool),
		identifierPattern: regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`),
	}
}

// RegisterTable adds a table to the whitelist
func (v *SQLValidator) RegisterTable(tableName string) {
	v.allowedTables[tableName] = true
}

// RegisterColumn adds a column to the whitelist
func (v *SQLValidator) RegisterColumn(columnName string) {
	v.allowedColumns[columnName] = true
}

// ValidateTableName validates a table name against whitelist and pattern
func (v *SQLValidator) ValidateTableName(tableName string) error {
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Check if it matches valid identifier pattern
	if !v.identifierPattern.MatchString(tableName) {
		return fmt.Errorf("invalid table name format: %s", tableName)
	}

	// If whitelist is configured, check against it
	if len(v.allowedTables) > 0 && !v.allowedTables[tableName] {
		return fmt.Errorf("table name not in whitelist: %s", tableName)
	}

	return nil
}

// ValidateColumnName validates a column name
func (v *SQLValidator) ValidateColumnName(columnName string) error {
	if columnName == "" {
		return fmt.Errorf("column name cannot be empty")
	}

	// Check if it matches valid identifier pattern
	if !v.identifierPattern.MatchString(columnName) {
		return fmt.Errorf("invalid column name format: %s", columnName)
	}

	// If whitelist is configured, check against it
	if len(v.allowedColumns) > 0 && !v.allowedColumns[columnName] {
		return fmt.Errorf("column name not in whitelist: %s", columnName)
	}

	return nil
}

// ValidateOrderBy validates ORDER BY clause
func (v *SQLValidator) ValidateOrderBy(orderBy string) error {
	if orderBy == "" {
		return nil
	}

	// Split by comma for multiple columns
	parts := strings.Split(orderBy, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for ASC/DESC
		tokens := strings.Fields(part)
		if len(tokens) == 0 {
			continue
		}

		columnName := tokens[0]
		if err := v.ValidateColumnName(columnName); err != nil {
			return fmt.Errorf("invalid ORDER BY column: %w", err)
		}

		// Validate direction if present
		if len(tokens) > 1 {
			direction := strings.ToUpper(tokens[1])
			if direction != "ASC" && direction != "DESC" {
				return fmt.Errorf("invalid ORDER BY direction: %s", direction)
			}
		}
	}

	return nil
}

// SanitizeWhereClause performs basic sanitization on WHERE clause
// Note: This is NOT a complete solution - parameterized queries should be used
func (v *SQLValidator) SanitizeWhereClause(where string) (string, error) {
	if where == "" {
		return "", nil
	}

	// Check for dangerous SQL keywords
	dangerousPatterns := []string{
		"DROP", "DELETE", "TRUNCATE", "ALTER", "CREATE",
		"EXEC", "EXECUTE", "SCRIPT", "UNION",
		"--", "/*", "*/", ";",
	}

	upperWhere := strings.ToUpper(where)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(upperWhere, pattern) {
			return "", fmt.Errorf("potentially dangerous SQL keyword detected: %s", pattern)
		}
	}

	return where, nil
}

// ValidateLimit validates LIMIT value
func (v *SQLValidator) ValidateLimit(limit int) error {
	if limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if limit > 10000 {
		return fmt.Errorf("limit too large (max 10000)")
	}
	return nil
}
