// backend/internal/nodes/integrations/registry.go
package integrations

import (
	"citadel-agent/backend/internal/workflow/core/engine"
)

// RegisterAllIntegrations registers all integration nodes with the engine
func RegisterAllIntegrations(registry *engine.NodeRegistry) {
	RegisterGitHubNode(registry)
	RegisterSlackNode(registry)
	RegisterEmailNode(registry)
	
	// Future registrations will go here:
	// RegisterDiscordNode(registry)
	// RegisterTelegramNode(registry)
	// RegisterTwilioNode(registry)
	// RegisterStripeNode(registry)
	// etc.
}