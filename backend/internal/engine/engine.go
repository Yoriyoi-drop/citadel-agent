package engine

import (
	"context"
	"errors"
)

// Node represents a node in the workflow
type Node struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Input    map[string]interface{} `json:"input"`
	Output   map[string]interface{} `json:"output"`
	Status   string                 `json:"status"`
	Children []*Node               `json:"children,omitempty"`
}

// Workflow represents a workflow
type Workflow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Nodes       []*Node `json:"nodes"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// NodeDefinition defines the structure of a node type
type NodeDefinition struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	Config      map[string]interface{} `json:"config"`
	Icon        string                 `json:"icon"`
	Category    string                 `json:"category"`
}

// NodeRegistry manages node definitions
type NodeRegistry struct {
	nodes map[string]*NodeDefinition
}

// NewNodeRegistry creates a new node registry
func NewNodeRegistry() *NodeRegistry {
	registry := &NodeRegistry{
		nodes: make(map[string]*NodeDefinition),
	}
	
	// Register default node types
	registry.RegisterNode(&NodeDefinition{
		Type:        "http_request",
		Name:        "HTTP Request",
		Description: "Makes an HTTP request to a specified URL",
		Inputs: map[string]interface{}{
			"url":        "string",
			"method":     "string",
			"headers":    "object",
			"body":       "object",
		},
		Outputs: map[string]interface{}{
			"response": "object",
			"status":   "number",
		},
		Config: map[string]interface{}{
			"timeout": 30,
		},
		Category: "Communication",
		Icon:     "🌐",
	})
	
	registry.RegisterNode(&NodeDefinition{
		Type:        "delay",
		Name:        "Delay",
		Description: "Waits for a specified amount of time",
		Inputs: map[string]interface{}{
			"duration": "number",
		},
		Outputs: map[string]interface{}{
			"completed": "boolean",
		},
		Config:   map[string]interface{}{},
		Category: "Utility",
		Icon:     "⏱️",
	})
	
	return registry
}

// RegisterNode registers a new node type
func (r *NodeRegistry) RegisterNode(nodeDef *NodeDefinition) {
	r.nodes[nodeDef.Type] = nodeDef
}

// GetNodeDefinition returns a node definition by type
func (r *NodeRegistry) GetNodeDefinition(nodeType string) (*NodeDefinition, error) {
	nodeDef, exists := r.nodes[nodeType]
	if !exists {
		return nil, errors.New("node type not found: " + nodeType)
	}
	return nodeDef, nil
}

// GetAllNodeDefinitions returns all registered node definitions
func (r *NodeRegistry) GetAllNodeDefinitions() []*NodeDefinition {
	var definitions []*NodeDefinition
	for _, def := range r.nodes {
		definitions = append(definitions, def)
	}
	return definitions
}

// Executor executes workflows
type Executor struct {
	registry *NodeRegistry
}

// NewExecutor creates a new executor
func NewExecutor(registry *NodeRegistry) *Executor {
	return &Executor{
		registry: registry,
	}
}

// ExecuteWorkflow executes a workflow
func (e *Executor) ExecuteWorkflow(ctx context.Context, workflow *Workflow) error {
	// In a real implementation, this would execute the workflow nodes
	// according to their dependencies and configurations
	
	// For now, this is a placeholder implementation
	for _, node := range workflow.Nodes {
		if err := e.executeNode(ctx, node); err != nil {
			return err
		}
	}
	
	return nil
}

// executeNode executes a single node
func (e *Executor) executeNode(ctx context.Context, node *Node) error {
	// In a real implementation, this would execute the specific node
	// based on its type and configuration
	
	// For now, this is a placeholder implementation
	node.Status = "completed"
	node.Output = map[string]interface{}{
		"message": "Node executed successfully",
		"node_id": node.ID,
	}
	
	return nil
}

// Runner runs workflows
type Runner struct {
	executor *Executor
}

// NewRunner creates a new runner
func NewRunner(executor *Executor) *Runner {
	return &Runner{
		executor: executor,
	}
}

// RunWorkflow runs a workflow
func (r *Runner) RunWorkflow(ctx context.Context, workflow *Workflow) error {
	return r.executor.ExecuteWorkflow(ctx, workflow)
}