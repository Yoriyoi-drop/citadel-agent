package messaging

import (
	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// RegisterNotificationNode registers all available notification and messaging nodes with the engine
func RegisterNotificationNode(registry *engine.NodeRegistry) {
	// Register notification/messaging nodes
	RegisterNotificationNode(registry)
	
	// In a complete implementation, we would also register:
	// - SMS nodes
	// - Push notification nodes
	// - Email nodes
	// - Slack/Discord integration nodes
	// - In-app notification nodes
	// - Webhook nodes
	// - And other messaging nodes
}