package utility

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// Deduplicator removes duplicate items from data
type Deduplicator struct{}

// NewDeduplicator creates a new deduplicator
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{}
}

// DeduplicateByKey removes duplicates based on a specific key
func (d *Deduplicator) DeduplicateByKey(data []map[string]interface{}, key string, keepFirst bool) []map[string]interface{} {
	seen := make(map[interface{}]bool)
	var result []map[string]interface{}

	if keepFirst {
		// Keep first occurrence
		for _, row := range data {
			val, ok := row[key]
			if !ok {
				result = append(result, row)
				continue
			}

			if !seen[val] {
				seen[val] = true
				result = append(result, row)
			}
		}
	} else {
		// Keep last occurrence - reverse iterate
		for i := len(data) - 1; i >= 0; i-- {
			row := data[i]
			val, ok := row[key]
			if !ok {
				result = append([]map[string]interface{}{row}, result...)
				continue
			}

			if !seen[val] {
				seen[val] = true
				result = append([]map[string]interface{}{row}, result...)
			}
		}
	}

	return result
}

// DeduplicateByKeys removes duplicates based on multiple keys
func (d *Deduplicator) DeduplicateByKeys(data []map[string]interface{}, keys []string, keepFirst bool) []map[string]interface{} {
	seen := make(map[string]bool)
	var result []map[string]interface{}

	if keepFirst {
		for _, row := range data {
			hash := d.hashByKeys(row, keys)
			if !seen[hash] {
				seen[hash] = true
				result = append(result, row)
			}
		}
	} else {
		for i := len(data) - 1; i >= 0; i-- {
			row := data[i]
			hash := d.hashByKeys(row, keys)
			if !seen[hash] {
				seen[hash] = true
				result = append([]map[string]interface{}{row}, result...)
			}
		}
	}

	return result
}

// DeduplicateByHash removes duplicates based on entire row hash
func (d *Deduplicator) DeduplicateByHash(data []map[string]interface{}, keepFirst bool) []map[string]interface{} {
	seen := make(map[string]bool)
	var result []map[string]interface{}

	if keepFirst {
		for _, row := range data {
			hash := d.hashRow(row)
			if !seen[hash] {
				seen[hash] = true
				result = append(result, row)
			}
		}
	} else {
		for i := len(data) - 1; i >= 0; i-- {
			row := data[i]
			hash := d.hashRow(row)
			if !seen[hash] {
				seen[hash] = true
				result = append([]map[string]interface{}{row}, result...)
			}
		}
	}

	return result
}

// DeduplicateByFunction removes duplicates using custom comparison function
func (d *Deduplicator) DeduplicateByFunction(data []map[string]interface{}, keyFn func(map[string]interface{}) string, keepFirst bool) []map[string]interface{} {
	seen := make(map[string]bool)
	var result []map[string]interface{}

	if keepFirst {
		for _, row := range data {
			key := keyFn(row)
			if !seen[key] {
				seen[key] = true
				result = append(result, row)
			}
		}
	} else {
		for i := len(data) - 1; i >= 0; i-- {
			row := data[i]
			key := keyFn(row)
			if !seen[key] {
				seen[key] = true
				result = append([]map[string]interface{}{row}, result...)
			}
		}
	}

	return result
}

// FindDuplicates finds duplicate rows based on key
func (d *Deduplicator) FindDuplicates(data []map[string]interface{}, key string) []map[string]interface{} {
	seen := make(map[interface{}]int)

	// Count occurrences
	for _, row := range data {
		if val, ok := row[key]; ok {
			seen[val]++
		}
	}

	// Find duplicates
	var duplicates []map[string]interface{}
	for _, row := range data {
		if val, ok := row[key]; ok {
			if seen[val] > 1 {
				duplicates = append(duplicates, row)
			}
		}
	}

	return duplicates
}

// CountDuplicates counts duplicate occurrences for each key value
func (d *Deduplicator) CountDuplicates(data []map[string]interface{}, key string) map[interface{}]int {
	counts := make(map[interface{}]int)

	for _, row := range data {
		if val, ok := row[key]; ok {
			counts[val]++
		}
	}

	return counts
}

// hashByKeys creates a hash from specific keys
func (d *Deduplicator) hashByKeys(row map[string]interface{}, keys []string) string {
	values := make([]interface{}, len(keys))
	for i, key := range keys {
		values[i] = row[key]
	}

	jsonBytes, _ := json.Marshal(values)
	hash := sha256.Sum256(jsonBytes)
	return fmt.Sprintf("%x", hash)
}

// hashRow creates a hash from entire row
func (d *Deduplicator) hashRow(row map[string]interface{}) string {
	jsonBytes, _ := json.Marshal(row)
	hash := sha256.Sum256(jsonBytes)
	return fmt.Sprintf("%x", hash)
}
