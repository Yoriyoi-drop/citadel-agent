package workflow

import (
	"citadel-agent/backend/internal/workflow/core/engine"
)

// RegisterWorkflowControlNode registers all available workflow control nodes with the engine
func RegisterWorkflowControlNode(registry *engine.NodeRegistry) {
	// Register workflow control nodes
	RegisterWorkflowControlNode(registry)
	
	// In a complete implementation, we would also register:
	// - Workflow branch/split nodes
	// - Workflow join/merge nodes
	// - Workflow synchronization nodes
	// - Workflow timer/delay nodes
	// - Workflow conditional nodes
	// - Workflow loop/iteration nodes
	// - Workflow error handling nodes
	// - And other workflow control nodes
}