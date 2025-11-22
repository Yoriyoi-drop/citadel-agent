// backend/internal/nodes/nodes.go
package nodes

import (
	"github.com/citadel-agent/backend/internal/nodes/ai"
	"github.com/citadel-agent/backend/internal/nodes/data"
	"github.com/citadel-agent/backend/internal/nodes/file"
	"github.com/citadel-agent/backend/internal/nodes/integrations"
	"github.com/citadel-agent/backend/internal/nodes/logic"
	"github.com/citadel-agent/backend/internal/nodes/logging"
	"github.com/citadel-agent/backend/internal/nodes/security"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
)

// RegisterAllNodes registers all available node types with the engine
func RegisterAllNodes(registry *engine.NodeRegistry) {
	// Register file operation nodes
	file.RegisterFileNode(registry)

	// Register logging nodes
	logging.RegisterLogNode(registry)

	// Register AI agent nodes
	ai.RegisterAINode(registry)
	ai.RegisterAIAgentOrchestratorNode(registry)
	ai.RegisterMLModelTrainingNode(registry)
	ai.RegisterAdvancedMLInferenceNode(registry)
	ai.RegisterMultiModalAIProcessorNode(registry)
	ai.RegisterAdvancedNLPProcessorNode(registry)
	ai.RegisterRealTimeMLTrainingNode(registry)
	ai.RegisterAdvancedRecommendationEngineNode(registry)
	ai.RegisterAdvancedAIAgentManagerNode(registry)
	ai.RegisterAdvancedDecisionEngineNode(registry)
	ai.RegisterAdvancedPredictiveAnalyticsNode(registry)
	ai.RegisterAdvancedContentIntelligenceNode(registry)
	ai.RegisterAdvancedDataIntelligenceNode(registry)

	// Register logic operation nodes
	logic.RegisterLogicNode(registry)

	// Register data transformation nodes
	data.RegisterDataTransformNode(registry)

	// Register integration nodes (GitHub, Slack, Email, etc.)
	integrations.RegisterAllIntegrations(registry)

	// Register security nodes
	security.RegisterSecurityNode(registry)
	security.RegisterFirewallManagerNode(registry)
	security.RegisterEncryptionNode(registry)
	security.RegisterAccessControlNode(registry)
	security.RegisterAPIKeyManagerNode(registry)
	security.RegisterJWTHandlerNode(registry)
	security.RegisterOAuth2ProviderNode(registry)

	// In a complete implementation, we would also register:
	// - Database nodes
	// - HTTP request nodes
	// - Loop and iteration nodes
	// - Error handling nodes
	// - Notification nodes
	// - Cache nodes
	// - Event nodes
	// - Schedule nodes
	// - And many more...
}