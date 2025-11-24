package utility

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrEmptyData     = errors.New("empty data")
	ErrInvalidField  = errors.New("invalid field")
	ErrInvalidMethod = errors.New("invalid aggregation method")
)

// Aggregator performs aggregation operations on data
type Aggregator struct{}

// NewAggregator creates a new aggregator
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// Sum calculates the sum of numeric values
func (a *Aggregator) Sum(data []map[string]interface{}, field string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	var sum float64
	for _, row := range data {
		val, ok := row[field]
		if !ok {
			continue
		}

		numVal, err := toFloat64(val)
		if err != nil {
			continue
		}
		sum += numVal
	}

	return sum, nil
}

// Average calculates the average of numeric values
func (a *Aggregator) Average(data []map[string]interface{}, field string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	sum, err := a.Sum(data, field)
	if err != nil {
		return 0, err
	}

	count := a.countNonNull(data, field)
	if count == 0 {
		return 0, errors.New("no valid numeric values")
	}

	return sum / float64(count), nil
}

// Min finds the minimum value
func (a *Aggregator) Min(data []map[string]interface{}, field string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	min := math.MaxFloat64
	found := false

	for _, row := range data {
		val, ok := row[field]
		if !ok {
			continue
		}

		numVal, err := toFloat64(val)
		if err != nil {
			continue
		}

		if numVal < min {
			min = numVal
			found = true
		}
	}

	if !found {
		return 0, errors.New("no valid numeric values")
	}

	return min, nil
}

// Max finds the maximum value
func (a *Aggregator) Max(data []map[string]interface{}, field string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	max := -math.MaxFloat64
	found := false

	for _, row := range data {
		val, ok := row[field]
		if !ok {
			continue
		}

		numVal, err := toFloat64(val)
		if err != nil {
			continue
		}

		if numVal > max {
			max = numVal
			found = true
		}
	}

	if !found {
		return 0, errors.New("no valid numeric values")
	}

	return max, nil
}

// Count counts the number of rows
func (a *Aggregator) Count(data []map[string]interface{}) int {
	return len(data)
}

// CountDistinct counts distinct values in a field
func (a *Aggregator) CountDistinct(data []map[string]interface{}, field string) int {
	seen := make(map[interface{}]bool)

	for _, row := range data {
		if val, ok := row[field]; ok {
			seen[val] = true
		}
	}

	return len(seen)
}

// GroupBy groups data by field and applies aggregation
func (a *Aggregator) GroupBy(data []map[string]interface{}, groupField string, aggField string, aggMethod string) ([]map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}

	// Group data
	groups := make(map[interface{}][]map[string]interface{})
	for _, row := range data {
		key, ok := row[groupField]
		if !ok {
			continue
		}
		groups[key] = append(groups[key], row)
	}

	// Apply aggregation to each group
	var result []map[string]interface{}
	for key, group := range groups {
		aggValue, err := a.applyAggregation(group, aggField, aggMethod)
		if err != nil {
			continue
		}

		result = append(result, map[string]interface{}{
			groupField: key,
			aggField:   aggValue,
		})
	}

	return result, nil
}

// Median calculates the median value
func (a *Aggregator) Median(data []map[string]interface{}, field string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	var values []float64
	for _, row := range data {
		val, ok := row[field]
		if !ok {
			continue
		}

		numVal, err := toFloat64(val)
		if err != nil {
			continue
		}
		values = append(values, numVal)
	}

	if len(values) == 0 {
		return 0, errors.New("no valid numeric values")
	}

	// Sort values
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	mid := len(values) / 2
	if len(values)%2 == 0 {
		return (values[mid-1] + values[mid]) / 2, nil
	}
	return values[mid], nil
}

// Percentile calculates the percentile value
func (a *Aggregator) Percentile(data []map[string]interface{}, field string, percentile float64) (float64, error) {
	if percentile < 0 || percentile > 100 {
		return 0, errors.New("percentile must be between 0 and 100")
	}

	if len(data) == 0 {
		return 0, ErrEmptyData
	}

	var values []float64
	for _, row := range data {
		val, ok := row[field]
		if !ok {
			continue
		}

		numVal, err := toFloat64(val)
		if err != nil {
			continue
		}
		values = append(values, numVal)
	}

	if len(values) == 0 {
		return 0, errors.New("no valid numeric values")
	}

	// Sort values
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	index := (percentile / 100) * float64(len(values)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return values[lower], nil
	}

	// Linear interpolation
	weight := index - float64(lower)
	return values[lower]*(1-weight) + values[upper]*weight, nil
}

// applyAggregation applies aggregation method to data
func (a *Aggregator) applyAggregation(data []map[string]interface{}, field string, method string) (interface{}, error) {
	switch method {
	case "sum":
		return a.Sum(data, field)
	case "avg", "average":
		return a.Average(data, field)
	case "min":
		return a.Min(data, field)
	case "max":
		return a.Max(data, field)
	case "count":
		return a.Count(data), nil
	case "count_distinct":
		return a.CountDistinct(data, field), nil
	case "median":
		return a.Median(data, field)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidMethod, method)
	}
}

// countNonNull counts non-null numeric values
func (a *Aggregator) countNonNull(data []map[string]interface{}, field string) int {
	count := 0
	for _, row := range data {
		val, ok := row[field]
		if !ok {
			continue
		}
		if _, err := toFloat64(val); err == nil {
			count++
		}
	}
	return count
}

// toFloat64 converts interface{} to float64
func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	default:
		return 0, errors.New("not a numeric value")
	}
}
