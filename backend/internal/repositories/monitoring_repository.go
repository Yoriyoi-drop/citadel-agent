// backend/internal/repositories/monitoring_repository.go
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MonitoringRepository handles monitoring-related database operations
type MonitoringRepository struct {
	db *pgxpool.Pool
}

// NewMonitoringRepository creates a new monitoring repository
func NewMonitoringRepository(db *pgxpool.Pool) *MonitoringRepository {
	return &MonitoringRepository{
		db: db,
	}
}

// Metric represents a metric record in the database
type Metric struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels"`
	Description string                 `json:"description"`
	Unit        string                 `json:"unit"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Alert represents an alert record in the database
type Alert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	FiredAt     *time.Time             `json:"fired_at,omitempty"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	MetricID    *string                `json:"metric_id,omitempty"`
	Condition   string                 `json:"condition"`
	Value       *float64               `json:"value,omitempty"`
	Threshold   *float64               `json:"threshold,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// LogEntry represents a log entry record in the database
type LogEntry struct {
	ID        string                 `json:"id"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
	Tags      []string               `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// CreateMetric creates a new metric record
func (mr *MonitoringRepository) CreateMetric(ctx context.Context, metric *Metric) (*Metric, error) {
	// Serialize JSON fields
	labelsJSON, err := json.Marshal(metric.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal labels: %w", err)
	}

	tagsJSON, err := json.Marshal(metric.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(metric.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO metrics (
			id, name, type, value, labels, description, unit, 
			created_at, updated_at, tags, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, name, type, value, description, unit, created_at, updated_at
	`

	var createdMetric Metric
	err = mr.db.QueryRow(ctx, query,
		metric.ID,
		metric.Name,
		metric.Type,
		metric.Value,
		labelsJSON,
		metric.Description,
		metric.Unit,
		metric.CreatedAt,
		metric.UpdatedAt,
		tagsJSON,
		metadataJSON,
	).Scan(
		&createdMetric.ID,
		&createdMetric.Name,
		&createdMetric.Type,
		&createdMetric.Value,
		&createdMetric.Description,
		&createdMetric.Unit,
		&createdMetric.CreatedAt,
		&createdMetric.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create metric: %w", err)
	}

	// Deserialize the JSON fields for return
	createdMetric.Labels = metric.Labels
	createdMetric.Tags = metric.Tags
	createdMetric.Metadata = metric.Metadata

	return &createdMetric, nil
}

// GetMetricByID retrieves a metric by ID
func (mr *MonitoringRepository) GetMetricByID(ctx context.Context, id string) (*Metric, error) {
	query := `
		SELECT id, name, type, value, labels, description, unit, 
		       created_at, updated_at, tags, metadata
		FROM metrics
		WHERE id = $1
	`

	var metric Metric
	var labelsJSON, tagsJSON, metadataJSON []byte

	err := mr.db.QueryRow(ctx, query, id).Scan(
		&metric.ID,
		&metric.Name,
		&metric.Type,
		&metric.Value,
		&labelsJSON,
		&metric.Description,
		&metric.Unit,
		&metric.CreatedAt,
		&metric.UpdatedAt,
		&tagsJSON,
		&metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metric not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	// Deserialize JSON fields
	if labelsJSON != nil {
		if err := json.Unmarshal(labelsJSON, &metric.Labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}
	}

	if tagsJSON != nil {
		var tags []string
		if err := json.Unmarshal(tagsJSON, &tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
		metric.Tags = tags
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &metric.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &metric, nil
}

// GetMetricsByTimeRange retrieves metrics within a time range
func (mr *MonitoringRepository) GetMetricsByTimeRange(ctx context.Context, start, end time.Time, name *string, limit int) ([]*Metric, error) {
	query := `
		SELECT id, name, type, value, labels, description, unit, 
		       created_at, updated_at, tags, metadata
		FROM metrics
		WHERE created_at >= $1 AND created_at <= $2
	`
	
	args := []interface{}{start, end}
	argCount := 3
	
	if name != nil {
		query += fmt.Sprintf(" AND name = $%d", argCount)
		args = append(args, *name)
		argCount++
	}
	
	query += " ORDER BY created_at DESC LIMIT $3"
	args = append(args, limit)
	
	rows, err := mr.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()
	
	var metrics []*Metric
	for rows.Next() {
		var metric Metric
		var labelsJSON, tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&metric.ID,
			&metric.Name,
			&metric.Type,
			&metric.Value,
			&labelsJSON,
			&metric.Description,
			&metric.Unit,
			&metric.CreatedAt,
			&metric.UpdatedAt,
			&tagsJSON,
			&metadataJSON,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}

		// Deserialize JSON fields
		if labelsJSON != nil {
			if err := json.Unmarshal(labelsJSON, &metric.Labels); err != nil {
				return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
			}
		}

		if tagsJSON != nil {
			var tags []string
			if err := json.Unmarshal(tagsJSON, &tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
			metric.Tags = tags
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &metric.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		metrics = append(metrics, &metric)
	}
	
	return metrics, nil
}

// CreateAlert creates a new alert record
func (mr *MonitoringRepository) CreateAlert(ctx context.Context, alert *Alert) (*Alert, error) {
	// Serialize JSON fields
	labelsJSON, err := json.Marshal(alert.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal labels: %w", err)
	}

	annotationsJSON, err := json.Marshal(alert.Annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal annotations: %w", err)
	}

	metadataJSON, err := json.Marshal(alert.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO alerts (
			id, name, severity, status, message, labels, annotations,
			created_at, updated_at, fired_at, resolved_at, metric_id,
			condition, value, threshold, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, name, severity, status, message, 
		          created_at, updated_at, fired_at, resolved_at, metric_id, condition
	`

	var createdAlert Alert
	err = mr.db.QueryRow(ctx, query,
		alert.ID,
		alert.Name,
		alert.Severity,
		alert.Status,
		alert.Message,
		labelsJSON,
		annotationsJSON,
		alert.CreatedAt,
		alert.UpdatedAt,
		alert.FiredAt,
		alert.ResolvedAt,
		alert.MetricID,
		alert.Condition,
		alert.Value,
		alert.Threshold,
		metadataJSON,
	).Scan(
		&createdAlert.ID,
		&createdAlert.Name,
		&createdAlert.Severity,
		&createdAlert.Status,
		&createdAlert.Message,
		&createdAlert.CreatedAt,
		&createdAlert.UpdatedAt,
		&createdAlert.FiredAt,
		&createdAlert.ResolvedAt,
		&createdAlert.MetricID,
		&createdAlert.Condition,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	// Set the deserialized fields
	createdAlert.Labels = alert.Labels
	createdAlert.Annotations = alert.Annotations
	createdAlert.Metadata = alert.Metadata

	return &createdAlert, nil
}

// GetAlertByID retrieves an alert by ID
func (mr *MonitoringRepository) GetAlertByID(ctx context.Context, id string) (*Alert, error) {
	query := `
		SELECT id, name, severity, status, message, labels, annotations,
		       created_at, updated_at, fired_at, resolved_at, metric_id,
		       condition, value, threshold, metadata
		FROM alerts
		WHERE id = $1
	`

	var alert Alert
	var labelsJSON, annotationsJSON, metadataJSON []byte

	err := mr.db.QueryRow(ctx, query, id).Scan(
		&alert.ID,
		&alert.Name,
		&alert.Severity,
		&alert.Status,
		&alert.Message,
		&labelsJSON,
		&annotationsJSON,
		&alert.CreatedAt,
		&alert.UpdatedAt,
		&alert.FiredAt,
		&alert.ResolvedAt,
		&alert.MetricID,
		&alert.Condition,
		&alert.Value,
		&alert.Threshold,
		&metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	// Deserialize JSON fields
	if labelsJSON != nil {
		if err := json.Unmarshal(labelsJSON, &alert.Labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}
	}

	if annotationsJSON != nil {
		if err := json.Unmarshal(annotationsJSON, &alert.Annotations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
		}
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &alert.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &alert, nil
}

// UpdateAlert updates an existing alert
func (mr *MonitoringRepository) UpdateAlert(ctx context.Context, alert *Alert) (*Alert, error) {
	// Serialize JSON fields
	labelsJSON, err := json.Marshal(alert.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal labels: %w", err)
	}

	annotationsJSON, err := json.Marshal(alert.Annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal annotations: %w", err)
	}

	metadataJSON, err := json.Marshal(alert.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE alerts
		SET severity = $2, status = $3, message = $4, labels = $5, annotations = $6,
		    updated_at = $7, fired_at = $8, resolved_at = $9, metric_id = $10,
		    condition = $11, value = $12, threshold = $13, metadata = $14
		WHERE id = $1
		RETURNING id, name, severity, status, message, 
		          created_at, updated_at, fired_at, resolved_at, metric_id, condition
	`

	var updatedAlert Alert
	err = mr.db.QueryRow(ctx, query,
		alert.ID,
		alert.Severity,
		alert.Status,
		alert.Message,
		labelsJSON,
		annotationsJSON,
		alert.UpdatedAt,
		alert.FiredAt,
		alert.ResolvedAt,
		alert.MetricID,
		alert.Condition,
		alert.Value,
		alert.Threshold,
		metadataJSON,
	).Scan(
		&updatedAlert.ID,
		&updatedAlert.Name,
		&updatedAlert.Severity,
		&updatedAlert.Status,
		&updatedAlert.Message,
		&updatedAlert.CreatedAt,
		&updatedAlert.UpdatedAt,
		&updatedAlert.FiredAt,
		&updatedAlert.ResolvedAt,
		&updatedAlert.MetricID,
		&updatedAlert.Condition,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	// Set the deserialized fields
	updatedAlert.Labels = alert.Labels
	updatedAlert.Annotations = alert.Annotations
	updatedAlert.Metadata = alert.Metadata

	return &updatedAlert, nil
}

// GetAlertsByStatus retrieves alerts by status
func (mr *MonitoringRepository) GetAlertsByStatus(ctx context.Context, status string, limit, offset int) ([]*Alert, error) {
	query := `
		SELECT id, name, severity, status, message, labels, annotations,
		       created_at, updated_at, fired_at, resolved_at, metric_id,
		       condition, value, threshold, metadata
		FROM alerts
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := mr.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()
	
	var alerts []*Alert
	for rows.Next() {
		var alert Alert
		var labelsJSON, annotationsJSON, metadataJSON []byte

		err := rows.Scan(
			&alert.ID,
			&alert.Name,
			&alert.Severity,
			&alert.Status,
			&alert.Message,
			&labelsJSON,
			&annotationsJSON,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&alert.FiredAt,
			&alert.ResolvedAt,
			&alert.MetricID,
			&alert.Condition,
			&alert.Value,
			&alert.Threshold,
			&metadataJSON,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		// Deserialize JSON fields
		if labelsJSON != nil {
			if err := json.Unmarshal(labelsJSON, &alert.Labels); err != nil {
				return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
			}
		}

		if annotationsJSON != nil {
			if err := json.Unmarshal(annotationsJSON, &alert.Annotations); err != nil {
				return nil, fmt.Errorf("failed to unmarshal annotations: %w", err)
			}
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &alert.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		alerts = append(alerts, &alert)
	}
	
	return alerts, nil
}

// CreateLogEntry creates a new log entry record
func (mr *MonitoringRepository) CreateLogEntry(ctx context.Context, entry *LogEntry) (*LogEntry, error) {
	// Serialize JSON fields
	fieldsJSON, err := json.Marshal(entry.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields: %w", err)
	}

	tagsJSON, err := json.Marshal(entry.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO log_entries (
			id, level, message, service, timestamp, fields, tags, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, level, message, service, timestamp
	`

	var createdEntry LogEntry
	err = mr.db.QueryRow(ctx, query,
		entry.ID,
		entry.Level,
		entry.Message,
		entry.Service,
		entry.Timestamp,
		fieldsJSON,
		tagsJSON,
		metadataJSON,
	).Scan(
		&createdEntry.ID,
		&createdEntry.Level,
		&createdEntry.Message,
		&createdEntry.Service,
		&createdEntry.Timestamp,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create log entry: %w", err)
	}

	// Set deserialized fields
	createdEntry.Fields = entry.Fields
	createdEntry.Tags = entry.Tags
	createdEntry.Metadata = entry.Metadata

	return &createdEntry, nil
}

// GetLogEntriesByTimeRange retrieves log entries within a time range
func (mr *MonitoringRepository) GetLogEntriesByTimeRange(ctx context.Context, start, end time.Time, level, service *string, limit, offset int) ([]*LogEntry, error) {
	query := `
		SELECT id, level, message, service, timestamp, fields, tags, metadata
		FROM log_entries
		WHERE timestamp >= $1 AND timestamp <= $2
	`
	
	args := []interface{}{start, end}
	argCount := 3

	if level != nil {
		query += fmt.Sprintf(" AND level = $%d", argCount)
		args = append(args, *level)
		argCount++
	}

	if service != nil {
		query += fmt.Sprintf(" AND service = $%d", argCount)
		args = append(args, *service)
		argCount++
	}
	
	query += " ORDER BY timestamp DESC LIMIT $3 OFFSET $4"
	args = append(args, limit, offset)
	
	rows, err := mr.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query log entries: %w", err)
	}
	defer rows.Close()
	
	var logEntries []*LogEntry
	for rows.Next() {
		var entry LogEntry
		var fieldsJSON, tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.Level,
			&entry.Message,
			&entry.Service,
			&entry.Timestamp,
			&fieldsJSON,
			&tagsJSON,
			&metadataJSON,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		// Deserialize JSON fields
		if fieldsJSON != nil {
			if err := json.Unmarshal(fieldsJSON, &entry.Fields); err != nil {
				return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
			}
		}

		if tagsJSON != nil {
			var tags []string
			if err := json.Unmarshal(tagsJSON, &tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
			entry.Tags = tags
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &entry.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		logEntries = append(logEntries, &entry)
	}
	
	return logEntries, nil
}