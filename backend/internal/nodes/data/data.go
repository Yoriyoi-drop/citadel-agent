package data

import (
	"citadel-agent/backend/internal/workflow/core/engine"
)

// RegisterDataTransformNode registers all available data transformation nodes with the engine
func RegisterDataTransformNode(registry *engine.NodeRegistry) {
	// Register data transformation nodes
	RegisterDataTransformerNode(registry)
	
	// In a complete implementation, we would also register:
	// - Data validation nodes
	// - Data cleaning nodes
	// - Data enrichment nodes
	// - Data normalization nodes
	// - Data aggregation nodes
	// - Data filtering nodes
	// - Data merging nodes
	// - And other data processing nodes
}