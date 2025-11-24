package transform

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	ErrInvalidCSV    = errors.New("invalid CSV")
	ErrEmptyCSV      = errors.New("empty CSV")
	ErrInvalidHeader = errors.New("invalid header")
)

// CSVTransformer handles CSV transformation operations
type CSVTransformer struct {
	delimiter rune
	hasHeader bool
}

// CSVConfig holds CSV configuration
type CSVConfig struct {
	Delimiter rune
	HasHeader bool
}

// NewCSVTransformer creates a new CSV transformer
func NewCSVTransformer(config CSVConfig) *CSVTransformer {
	if config.Delimiter == 0 {
		config.Delimiter = ','
	}
	return &CSVTransformer{
		delimiter: config.Delimiter,
		hasHeader: config.HasHeader,
	}
}

// Parse parses CSV string to array of maps
func (t *CSVTransformer) Parse(csvStr string) ([]map[string]interface{}, error) {
	reader := csv.NewReader(strings.NewReader(csvStr))
	reader.Comma = t.delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCSV, err)
	}

	if len(records) == 0 {
		return nil, ErrEmptyCSV
	}

	var headers []string
	startIdx := 0

	if t.hasHeader {
		if len(records) < 2 {
			return nil, errors.New("CSV has header but no data rows")
		}
		headers = records[0]
		startIdx = 1
	} else {
		// Generate default headers
		headers = make([]string, len(records[0]))
		for i := range headers {
			headers[i] = fmt.Sprintf("column_%d", i)
		}
	}

	result := make([]map[string]interface{}, 0, len(records)-startIdx)
	for _, record := range records[startIdx:] {
		row := make(map[string]interface{})
		for i, value := range record {
			if i < len(headers) {
				row[headers[i]] = t.inferType(value)
			}
		}
		result = append(result, row)
	}

	return result, nil
}

// Stringify converts array of maps to CSV string
func (t *CSVTransformer) Stringify(data []map[string]interface{}) (string, error) {
	if len(data) == 0 {
		return "", ErrEmptyCSV
	}

	// Extract headers from first row
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	writer.Comma = t.delimiter

	// Write header if configured
	if t.hasHeader {
		if err := writer.Write(headers); err != nil {
			return "", err
		}
	}

	// Write data rows
	for _, row := range data {
		record := make([]string, len(headers))
		for i, header := range headers {
			if val, ok := row[header]; ok {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return builder.String(), nil
}

// ParseToArray parses CSV to array of arrays
func (t *CSVTransformer) ParseToArray(csvStr string) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(csvStr))
	reader.Comma = t.delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCSV, err)
	}

	return records, nil
}

// StringifyFromArray converts array of arrays to CSV string
func (t *CSVTransformer) StringifyFromArray(data [][]string) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	writer.Comma = t.delimiter

	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return builder.String(), nil
}

// Filter filters CSV rows based on condition
func (t *CSVTransformer) Filter(data []map[string]interface{}, filterFn func(map[string]interface{}) bool) []map[string]interface{} {
	var filtered []map[string]interface{}
	for _, row := range data {
		if filterFn(row) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

// Map transforms CSV rows
func (t *CSVTransformer) Map(data []map[string]interface{}, mapFn func(map[string]interface{}) map[string]interface{}) []map[string]interface{} {
	mapped := make([]map[string]interface{}, len(data))
	for i, row := range data {
		mapped[i] = mapFn(row)
	}
	return mapped
}

// inferType attempts to infer the type of a string value
func (t *CSVTransformer) inferType(value string) interface{} {
	// Try integer
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Try boolean
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Default to string
	return value
}

// ParseStream parses CSV from a reader (for large files)
func (t *CSVTransformer) ParseStream(reader io.Reader, callback func(map[string]interface{}) error) error {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = t.delimiter

	var headers []string
	if t.hasHeader {
		var err error
		headers, err = csvReader.Read()
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidHeader, err)
		}
	}

	rowNum := 0
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row %d: %w", rowNum, err)
		}

		// Generate headers if not provided
		if headers == nil {
			headers = make([]string, len(record))
			for i := range headers {
				headers[i] = fmt.Sprintf("column_%d", i)
			}
		}

		// Convert to map
		row := make(map[string]interface{})
		for i, value := range record {
			if i < len(headers) {
				row[headers[i]] = t.inferType(value)
			}
		}

		// Call callback
		if err := callback(row); err != nil {
			return err
		}

		rowNum++
	}

	return nil
}

// GetColumn extracts a specific column from CSV data
func (t *CSVTransformer) GetColumn(data []map[string]interface{}, columnName string) []interface{} {
	column := make([]interface{}, 0, len(data))
	for _, row := range data {
		if val, ok := row[columnName]; ok {
			column = append(column, val)
		}
	}
	return column
}

// AddColumn adds a new column to CSV data
func (t *CSVTransformer) AddColumn(data []map[string]interface{}, columnName string, valueFn func(map[string]interface{}) interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(data))
	for i, row := range data {
		newRow := make(map[string]interface{})
		for k, v := range row {
			newRow[k] = v
		}
		newRow[columnName] = valueFn(row)
		result[i] = newRow
	}
	return result
}
