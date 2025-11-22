package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
	"go.temporal.io/sdk/client"
)

// TemporalWorkflowService provides integration between Citadel Agent and Temporal
type TemporalWorkflowService struct {
	temporalClient  *TemporalClient
	engine          *engine.Engine
	workflowDefs    map[string]*WorkflowDefinition
	definitionMutex sync.RWMutex
}

// NewTemporalWorkflowService creates a new Temporal workflow service
func NewTemporalWorkflowService(temporalClient *TemporalClient, baseEngine *engine.Engine) *TemporalWorkflowService {
	return &TemporalWorkflowService{
		temporalClient: temporalClient,
		engine:         baseEngine,
		workflowDefs:   make(map[string]*WorkflowDefinition),
	}
}

// RegisterWorkflowDefinition registers a workflow definition
func (s *TemporalWorkflowService) RegisterWorkflowDefinition(definition *WorkflowDefinition) {
	s.definitionMutex.Lock()
	defer s.definitionMutex.Unlock()
	
	s.workflowDefs[definition.ID] = definition
}

// GetWorkflowDefinition retrieves a workflow definition by ID
func (s *TemporalWorkflowService) GetWorkflowDefinition(workflowID string) (*WorkflowDefinition, error) {
	s.definitionMutex.RLock()
	defer s.definitionMutex.RUnlock()
	
	def, exists := s.workflowDefs[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow definition %s not found", workflowID)
	}
	
	return def, nil
}

// ExecuteWorkflow executes a workflow using Temporal
func (s *TemporalWorkflowService) ExecuteWorkflow(ctx context.Context, workflowID string, parameters map[string]interface{}) (string, error) {
	// Get the workflow definition
	def, err := s.GetWorkflowDefinition(workflowID)
	if err != nil {
		return "", fmt.Errorf("failed to get workflow definition: %w", err)
	}

	// Prepare input for Temporal workflow
	input := WorkflowInput{
		ID:          def.ID,
		Name:        def.Name,
		Description: def.Description,
		TriggeredBy: "api", // This can be from scheduler, API, etc.
		Parameters:  parameters,
	}

	// Create Temporal workflow options
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID + "-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		TaskQueue: "citadel-agent-workflows",
	}

	// Execute the workflow
	workflowRun, err := s.temporalClient.ExecuteWorkflow(ctx, CitadelAgentWorkflow, input)
	if err != nil {
		return "", fmt.Errorf("failed to execute workflow: %w", err)
	}

	return workflowRun.GetID(), nil
}

// ExecuteWorkflowWithDefinition executes a workflow with inline definition
func (s *TemporalWorkflowService) ExecuteWorkflowWithDefinition(ctx context.Context, definition *WorkflowDefinition, parameters map[string]interface{}) (string, error) {
	// Register the definition temporarily
	tempID := fmt.Sprintf("temp_%s_%d", definition.ID, time.Now().UnixNano())
	definition.ID = tempID
	s.RegisterWorkflowDefinition(definition)

	// Prepare input for Temporal workflow
	input := WorkflowInput{
		ID:          tempID,
		Name:        definition.Name,
		Description: definition.Description,
		TriggeredBy: "api",
		Parameters:  parameters,
	}

	// Create Temporal workflow options
	workflowOptions := client.StartWorkflowOptions{
		ID:        tempID,
		TaskQueue: "citadel-agent-workflows",
	}

	// Execute the workflow
	workflowRun, err := s.temporalClient.ExecuteWorkflow(ctx, CitadelAgentWorkflow, input)
	if err != nil {
		return "", fmt.Errorf("failed to execute workflow: %w", err)
	}

	return workflowRun.GetID(), nil
}

// GetWorkflowResult retrieves the result of a workflow execution
func (s *TemporalWorkflowService) GetWorkflowResult(ctx context.Context, workflowID, runID string) (*WorkflowOutput, error) {
	workflowRun := s.temporalClient.GetWorkflow(ctx, workflowID, runID)
	
	var result WorkflowOutput
	err := workflowRun.Get(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow result: %w", err)
	}
	
	return &result, nil
}

// CancelWorkflow cancels a running workflow
func (s *TemporalWorkflowService) CancelWorkflow(ctx context.Context, workflowID, runID string) error {
	return s.temporalClient.GetClient().CancelWorkflow(ctx, workflowID, runID)
}

// TerminateWorkflow terminates a running workflow
func (s *TemporalWorkflowService) TerminateWorkflow(ctx context.Context, workflowID, runID, reason string) error {
	return s.temporalClient.GetClient().TerminateWorkflow(ctx, workflowID, runID, reason)
}

// SignalWorkflow sends a signal to a running workflow
func (s *TemporalWorkflowService) SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, arg interface{}) error {
	return s.temporalClient.SignalWorkflow(ctx, workflowID, runID, signalName, arg)
}

// ListWorkflows lists active workflows
func (s *TemporalWorkflowService) ListWorkflows(ctx context.Context, pageSize int, nextPageToken []byte) (*client.WorkflowListIterator, error) {
	request := &client.WorkflowListRequest{
		PageSize:      int32(pageSize),
		NextPageToken: nextPageToken,
		Query:         "WorkflowType = 'CitadelAgentWorkflow'",
	}

	return s.temporalClient.GetClient().ListWorkflow(ctx, request), nil
}

// QueryWorkflowHistory queries the history of a workflow
func (s *TemporalWorkflowService) QueryWorkflowHistory(ctx context.Context, workflowID, runID string) ([]byte, error) {
	iter := s.temporalClient.GetClient().GetWorkflowHistory(ctx, workflowID, runID, false, 0)
	
	var history []byte
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, err
		}
		
		eventBytes, _ := json.Marshal(event)
		history = append(history, eventBytes...)
		history = append(history, []byte("\n")...)
	}
	
	return history, nil
}

// RegisterNodeTypes registers node types with the base engine for compatibility
func (s *TemporalWorkflowService) RegisterNodeTypes() {
	// Register basic node types that can be used both in Temporal and local engine
	s.engine.GetNodeRegistry().RegisterNodeType("http_request", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &HTTPRequestNode{Config: config}, nil
	})
	
	s.engine.GetNodeRegistry().RegisterNodeType("condition", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &ConditionNode{Config: config}, nil
	})
	
	s.engine.GetNodeRegistry().RegisterNodeType("delay", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &DelayNode{Config: config}, nil
	})
	
	s.engine.GetNodeRegistry().RegisterNodeType("database_query", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &DatabaseNode{Config: config}, nil
	})
	
	s.engine.GetNodeRegistry().RegisterNodeType("script_execution", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &ScriptNode{Config: config}, nil
	})
	
	// Register advanced node types
	s.engine.GetNodeRegistry().RegisterNodeType("ai_agent", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &AIAgentNode{Config: config}, nil
	})
	
	s.engine.GetNodeRegistry().RegisterNodeType("data_transformer", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &DataTransformerNode{Config: config}, nil
	})
	
	s.engine.GetNodeRegistry().RegisterNodeType("notification", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &NotificationNode{Config: config}, nil
	})
	
	s.engine.GetNodeRegistry().RegisterNodeType("loop", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &LoopNode{Config: config}, nil
	})
	
	s.engine.GetNodeRegistry().RegisterNodeType("error_handler", func(config map[string]interface{}) (interfaces.NodeInstance, error) {
		return &ErrorHandlerNode{Config: config}, nil
	})
}