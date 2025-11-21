// backend/internal/repositories/workflow_repository.go
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/models"
	"github.com/google/uuid"
)

// WorkflowRepository handles database operations for workflows
type WorkflowRepository struct {
	db *sql.DB
}

// NewWorkflowRepository creates a new workflow repository
func NewWorkflowRepository(db *sql.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// Create creates a new workflow
func (wr *WorkflowRepository) Create(ctx context.Context, workflow *models.Workflow) (*models.Workflow, error) {
	query := `
		INSERT INTO workflows (id, name, description, nodes, connections, config, status, created_at, updated_at, version, owner_id, team_id, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, name, description, nodes, connections, config, status, created_at, updated_at, version, owner_id, team_id, tags
	`

	// Convert nodes, connections, and tags to JSON
	nodesJSON, err := json.Marshal(workflow.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nodes: %w", err)
	}

	connectionsJSON, err := json.Marshal(workflow.Connections)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal connections: %w", err)
	}

	configJSON, err := json.Marshal(workflow.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	tagsJSON, err := json.Marshal(workflow.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	var createdWorkflow models.Workflow
	var nodesBytes, connectionsBytes, configBytes, tagsBytes []byte

	err = wr.db.QueryRowContext(ctx, query,
		uuid.New().String(), // Generate new ID
		workflow.Name,
		workflow.Description,
		nodesJSON,
		connectionsJSON,
		configJSON,
		workflow.Status,
		workflow.CreatedAt,
		workflow.UpdatedAt,
		workflow.Version,
		workflow.OwnerID,
		workflow.TeamID,
		tagsJSON,
	).Scan(
		&createdWorkflow.ID,
		&createdWorkflow.Name,
		&createdWorkflow.Description,
		&nodesBytes,
		&connectionsBytes,
		&configBytes,
		&createdWorkflow.Status,
		&createdWorkflow.CreatedAt,
		&createdWorkflow.UpdatedAt,
		&createdWorkflow.Version,
		&createdWorkflow.OwnerID,
		&createdWorkflow.TeamID,
		&tagsBytes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(nodesBytes, &createdWorkflow.Nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	if err := json.Unmarshal(connectionsBytes, &createdWorkflow.Connections); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connections: %w", err)
	}

	if err := json.Unmarshal(configBytes, &createdWorkflow.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(tagsBytes, &createdWorkflow.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &createdWorkflow, nil
}

// GetByID retrieves a workflow by ID
func (wr *WorkflowRepository) GetByID(ctx context.Context, id string) (*models.Workflow, error) {
	query := `
		SELECT id, name, description, nodes, connections, config, status, created_at, updated_at, version, owner_id, team_id, tags
		FROM workflows
		WHERE id = $1
	`

	var workflow models.Workflow
	var nodesBytes, connectionsBytes, configBytes, tagsBytes []byte

	err := wr.db.QueryRowContext(ctx, query, id).Scan(
		&workflow.ID,
		&workflow.Name,
		&workflow.Description,
		&nodesBytes,
		&connectionsBytes,
		&configBytes,
		&workflow.Status,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
		&workflow.Version,
		&workflow.OwnerID,
		&workflow.TeamID,
		&tagsBytes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(nodesBytes, &workflow.Nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	if err := json.Unmarshal(connectionsBytes, &workflow.Connections); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connections: %w", err)
	}

	if err := json.Unmarshal(configBytes, &workflow.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(tagsBytes, &workflow.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &workflow, nil
}

// Update updates an existing workflow
func (wr *WorkflowRepository) Update(ctx context.Context, id string, workflow *models.Workflow) (*models.Workflow, error) {
	query := `
		UPDATE workflows
		SET name = $2, description = $3, nodes = $4, connections = $5, config = $6, 
			status = $7, updated_at = $8, version = $9, owner_id = $10, team_id = $11, tags = $12
		WHERE id = $1
		RETURNING id, name, description, nodes, connections, config, status, created_at, updated_at, version, owner_id, team_id, tags
	`

	// Convert nodes, connections, and tags to JSON
	nodesJSON, err := json.Marshal(workflow.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nodes: %w", err)
	}

	connectionsJSON, err := json.Marshal(workflow.Connections)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal connections: %w", err)
	}

	configJSON, err := json.Marshal(workflow.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	tagsJSON, err := json.Marshal(workflow.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	var updatedWorkflow models.Workflow
	var nodesBytes, connectionsBytes, configBytes, tagsBytes []byte

	err = wr.db.QueryRowContext(ctx, query,
		id,
		workflow.Name,
		workflow.Description,
		nodesJSON,
		connectionsJSON,
		configJSON,
		workflow.Status,
		workflow.UpdatedAt,
		workflow.Version,
		workflow.OwnerID,
		workflow.TeamID,
		tagsJSON,
	).Scan(
		&updatedWorkflow.ID,
		&updatedWorkflow.Name,
		&updatedWorkflow.Description,
		&nodesBytes,
		&connectionsBytes,
		&configBytes,
		&updatedWorkflow.Status,
		&updatedWorkflow.CreatedAt,
		&updatedWorkflow.UpdatedAt,
		&updatedWorkflow.Version,
		&updatedWorkflow.OwnerID,
		&updatedWorkflow.TeamID,
		&tagsBytes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(nodesBytes, &updatedWorkflow.Nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	if err := json.Unmarshal(connectionsBytes, &updatedWorkflow.Connections); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connections: %w", err)
	}

	if err := json.Unmarshal(configBytes, &updatedWorkflow.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(tagsBytes, &updatedWorkflow.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &updatedWorkflow, nil
}

// Delete deletes a workflow by ID
func (wr *WorkflowRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM workflows WHERE id = $1`
	result, err := wr.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workflow with ID %s not found", id)
	}

	return nil
}

// List retrieves a list of workflows with pagination
func (wr *WorkflowRepository) List(ctx context.Context, page, limit int) ([]*models.Workflow, error) {
	if page < 0 {
		page = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := page * limit

	query := `
		SELECT id, name, description, nodes, connections, config, status, created_at, updated_at, version, owner_id, team_id, tags
		FROM workflows
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := wr.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*models.Workflow

	for rows.Next() {
		var workflow models.Workflow
		var nodesBytes, connectionsBytes, configBytes, tagsBytes []byte

		err := rows.Scan(
			&workflow.ID,
			&workflow.Name,
			&workflow.Description,
			&nodesBytes,
			&connectionsBytes,
			&configBytes,
			&workflow.Status,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
			&workflow.Version,
			&workflow.OwnerID,
			&workflow.TeamID,
			&tagsBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(nodesBytes, &workflow.Nodes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
		}

		if err := json.Unmarshal(connectionsBytes, &workflow.Connections); err != nil {
			return nil, fmt.Errorf("failed to unmarshal connections: %w", err)
		}

		if err := json.Unmarshal(configBytes, &workflow.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if err := json.Unmarshal(tagsBytes, &workflow.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}

// GetActiveExecutions retrieves active executions for a workflow
func (wr *WorkflowRepository) GetActiveExecutions(ctx context.Context, workflowID string) ([]*models.Execution, error) {
	query := `
		SELECT id, workflow_id, name, status, started_at, completed_at, variables, node_results, error, triggered_by, trigger_params, parent_id, retry_count, user_id, team_id
		FROM executions
		WHERE workflow_id = $1 AND status IN ('running', 'pending', 'retrying')
	`

	rows, err := wr.db.QueryContext(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active executions: %w", err)
	}
	defer rows.Close()

	var executions []*models.Execution

	for rows.Next() {
		var execution models.Execution
		var variablesBytes, nodeResultsBytes, triggerParamsBytes []byte
		var completedAt, parentID, errorStr *time.Time
		var errorPtr *string

		err := rows.Scan(
			&execution.ID,
			&execution.WorkflowID,
			&execution.Name,
			&execution.Status,
			&execution.StartedAt,
			&completedAt,
			&variablesBytes,
			&nodeResultsBytes,
			&errorPtr,
			&execution.TriggeredBy,
			&triggerParamsBytes,
			&parentID,
			&execution.RetryCount,
			&execution.UserID,
			&execution.TeamID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}

		// Handle nullable fields
		if completedAt != nil {
			execution.CompletedAt = completedAt
		}
		if errorPtr != nil {
			execution.Error = errorPtr
		}
		if parentID != nil {
			parentIDStr := parentID.String()
			execution.ParentID = &parentIDStr
		}

		// Unmarshal JSON fields
		if variablesBytes != nil {
			if err := json.Unmarshal(variablesBytes, &execution.Variables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
			}
		}

		if nodeResultsBytes != nil {
			if err := json.Unmarshal(nodeResultsBytes, &execution.NodeResults); err != nil {
				return nil, fmt.Errorf("failed to unmarshal node_results: %w", err)
			}
		}

		if triggerParamsBytes != nil {
			if err := json.Unmarshal(triggerParamsBytes, &execution.TriggerParams); err != nil {
				return nil, fmt.Errorf("failed to unmarshal trigger_params: %w", err)
			}
		}

		executions = append(executions, &execution)
	}

	return executions, nil
}

// LogExecution logs an execution
func (wr *WorkflowRepository) LogExecution(ctx context.Context, log *models.ExecutionLog) error {
	query := `
		INSERT INTO execution_logs (id, workflow_id, execution_id, node_id, status, action, message, timestamp, parameters, details, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	parametersJSON, err := json.Marshal(log.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	_, err = wr.db.ExecContext(ctx, query,
		uuid.New().String(),
		log.WorkflowID,
		log.ExecutionID,
		log.NodeID,
		log.Status,
		log.Action,
		log.Message,
		log.Timestamp,
		parametersJSON,
		detailsJSON,
		log.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to log execution: %w", err)
	}

	return nil
}

// GetUserWorkflows retrieves workflows for a specific user
func (wr *WorkflowRepository) GetUserWorkflows(ctx context.Context, userID string, page, limit int) ([]*models.Workflow, error) {
	if page < 0 {
		page = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := page * limit

	query := `
		SELECT id, name, description, nodes, connections, config, status, created_at, updated_at, version, owner_id, team_id, tags
		FROM workflows
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := wr.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*models.Workflow

	for rows.Next() {
		var workflow models.Workflow
		var nodesBytes, connectionsBytes, configBytes, tagsBytes []byte

		err := rows.Scan(
			&workflow.ID,
			&workflow.Name,
			&workflow.Description,
			&nodesBytes,
			&connectionsBytes,
			&configBytes,
			&workflow.Status,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
			&workflow.Version,
			&workflow.OwnerID,
			&workflow.TeamID,
			&tagsBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(nodesBytes, &workflow.Nodes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
		}

		if err := json.Unmarshal(connectionsBytes, &workflow.Connections); err != nil {
			return nil, fmt.Errorf("failed to unmarshal connections: %w", err)
		}

		if err := json.Unmarshal(configBytes, &workflow.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if err := json.Unmarshal(tagsBytes, &workflow.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}

// UpdateWorkflowStatus updates the status of a workflow
func (wr *WorkflowRepository) UpdateWorkflowStatus(ctx context.Context, id string, status models.WorkflowStatus) error {
	query := `
		UPDATE workflows
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := wr.db.ExecContext(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %w", err)
	}

	return nil
}

// CountWorkflows counts all workflows
func (wr *WorkflowRepository) CountWorkflows(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM workflows`
	var count int64
	err := wr.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workflows: %w", err)
	}
	return count, nil
}

// SearchWorkflows searches workflows by name or description
func (wr *WorkflowRepository) SearchWorkflows(ctx context.Context, searchTerm string, page, limit int) ([]*models.Workflow, error) {
	if page < 0 {
		page = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := page * limit

	query := `
		SELECT id, name, description, nodes, connections, config, status, created_at, updated_at, version, owner_id, team_id, tags
		FROM workflows
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := wr.db.QueryContext(ctx, query, "%"+searchTerm+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*models.Workflow

	for rows.Next() {
		var workflow models.Workflow
		var nodesBytes, connectionsBytes, configBytes, tagsBytes []byte

		err := rows.Scan(
			&workflow.ID,
			&workflow.Name,
			&workflow.Description,
			&nodesBytes,
			&connectionsBytes,
			&configBytes,
			&workflow.Status,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
			&workflow.Version,
			&workflow.OwnerID,
			&workflow.TeamID,
			&tagsBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(nodesBytes, &workflow.Nodes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
		}

		if err := json.Unmarshal(connectionsBytes, &workflow.Connections); err != nil {
			return nil, fmt.Errorf("failed to unmarshal connections: %w", err)
		}

		if err := json.Unmarshal(configBytes, &workflow.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if err := json.Unmarshal(tagsBytes, &workflow.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		workflows = append(workflows, &workflow)
	}

	return workflows, nil
}