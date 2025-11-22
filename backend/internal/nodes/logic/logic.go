package logic

import (
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// RegisterLogicNode registers all available logic nodes with the engine
func RegisterLogicNode(registry *engine.NodeRegistry) {
	// Register condition-based logic nodes
	RegisterConditionProcessorNode(registry)
	
	// Register loop-based logic nodes
	RegisterLoopProcessorNode(registry)
	
	// In a complete implementation, we would also register:
	// - Switch/Case nodes
	// - Decision tree nodes
	// - Boolean operation nodes
	// - Expression evaluator nodes
	// - And other logic operation nodes
}