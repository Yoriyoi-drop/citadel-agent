package engine

import (
	"context"
	"fmt"
	"log"
	"sync"

	"citadel-agent/backend/internal/workflow/core/types"
)

// Workflow represents a workflow with nodes and connections
type Workflow struct {
	ID    string                    `json:"id"`
	Name  string                    `json:"name"`
	Nodes map[string]*WorkflowNode  `json:"nodes"`
	Edges []WorkflowEdge           `json:"edges"`
}

// WorkflowNode represents a node in the workflow
type WorkflowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Position map[string]float64     `json:"position"`
}

// WorkflowEdge represents a connection between nodes
type WorkflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// WorkflowExecutor executes workflows
type WorkflowExecutor struct {
	registry *NodeTypeRegistryImpl
	mu       sync.Mutex
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(registry *NodeTypeRegistryImpl) *WorkflowExecutor {
	if registry == nil {
		registry = globalRegistry
	}
	return &WorkflowExecutor{
		registry: registry,
	}
}

// ExecuteWorkflow executes a workflow with the given inputs
func (we *WorkflowExecutor) ExecuteWorkflow(ctx context.Context, workflow *Workflow, inputs map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Executing workflow: %s", workflow.ID)

	// Initialize all nodes
	nodeInstances := make(map[string]types.NodeInstance)
	for nodeID, node := range workflow.Nodes {
		creator, exists := we.registry.GetNodeType(node.Type)
		if !exists {
			return nil, fmt.Errorf("unknown node type: %s", node.Type)
		}

		instance := creator()
		if err := instance.Initialize(node.Config); err != nil {
			return nil, fmt.Errorf("failed to initialize node %s: %v", nodeID, err)
		}

		if err := instance.Validate(); err != nil {
			return nil, fmt.Errorf("invalid configuration for node %s: %v", nodeID, err)
		}

		nodeInstances[nodeID] = instance
		defer func(nodeID string, instance types.NodeInstance) {
			if err := instance.Close(); err != nil {
				log.Printf("Error closing node %s: %v", nodeID, err)
			}
		}(nodeID, instance)
	}

	// Execute the workflow - for now, execute in a simple order
	// TODO: Implement proper DAG execution with parallel execution
	results := make(map[string]interface{})
	
	// Execute nodes in order - this is a simplified approach
	// In a real implementation, we would need to build a dependency graph
	for nodeID := range workflow.Nodes {
		instance := nodeInstances[nodeID]

		// Prepare input for this node
		input := types.NodeInput{Data: make(map[string]interface{})}

		// Find edges that point to this node and collect their results
		for _, edge := range workflow.Edges {
			if edge.Target == nodeID {
				// Get result from source node
				sourceResult := results[edge.Source]
				if sourceResult != nil {
					// Merge the results from source nodes
					if sourceMap, ok := sourceResult.(map[string]interface{}); ok {
						for k, v := range sourceMap {
							input.Data[k] = v
						}
					} else {
						// If source result is not a map, store under a default key
						input.Data["result"] = sourceResult
					}
				}
			}
		}

		// If this is a starting node, use provided inputs
		if len(input.Data) == 0 {
			input.Data = inputs
		}

		// Execute the node
		output := instance.Execute(ctx, input)
		if output.Error != nil {
			return nil, fmt.Errorf("error executing node %s: %v", nodeID, output.Error)
		}

		results[nodeID] = output.Data
	}

	return results, nil
}