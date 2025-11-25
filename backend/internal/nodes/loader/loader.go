package loader

import (
	"log"

	"citadel-agent/backend/internal/nodes/base"
	"citadel-agent/backend/internal/nodes/registry"

	// Import all node packages
	"citadel-agent/backend/internal/nodes/ai"
	"citadel-agent/backend/internal/nodes/communication"
	"citadel-agent/backend/internal/nodes/database"
	"citadel-agent/backend/internal/nodes/flow"
	"citadel-agent/backend/internal/nodes/http"
	"citadel-agent/backend/internal/nodes/security"
	"citadel-agent/backend/internal/nodes/transform"
	"citadel-agent/backend/internal/nodes/utility"
	"citadel-agent/backend/internal/nodes/validation"
)

// LoadAllNodes registers all available nodes
func LoadAllNodes() error {
	reg := registry.GetRegistry()

	// Helper to register node
	register := func(creator func() base.Node) error {
		node := creator()
		return reg.Register(node.GetMetadata().ID, creator, node.GetMetadata())
	}

	// 1. HTTP Nodes
	if err := register(http.NewHTTPRequestNodeWrapper); err != nil {
		return err
	}
	if err := register(http.NewWebhookNode); err != nil {
		return err
	}

	// 2. Database Nodes
	if err := register(database.NewDatabaseQueryNode); err != nil {
		return err
	}
	if err := register(database.NewMongoDBNode); err != nil {
		return err
	}
	if err := register(database.NewRedisGetNode); err != nil {
		return err
	}
	if err := register(database.NewRedisSetNode); err != nil {
		return err
	}

	// 3. Transform Nodes
	if err := register(transform.NewJSONParserNode); err != nil {
		return err
	}
	if err := register(transform.NewXMLParserNode); err != nil {
		return err
	}
	if err := register(transform.NewCSVParserNode); err != nil {
		return err
	}
	if err := register(transform.NewDataMapperNode); err != nil {
		return err
	}

	// 4. Flow Control Nodes
	if err := register(flow.NewIfElseNode); err != nil {
		return err
	}
	if err := register(flow.NewForEachNode); err != nil {
		return err
	}
	if err := register(flow.NewDelayNode); err != nil {
		return err
	}

	// 5. AI Nodes
	if err := register(ai.NewOpenAIGPT4Node); err != nil {
		return err
	}
	if err := register(ai.NewOpenAIGPT35Node); err != nil {
		return err
	}

	// 6. Validation Nodes
	if err := register(validation.NewEmailValidatorNode); err != nil {
		return err
	}
	if err := register(validation.NewURLValidatorNode); err != nil {
		return err
	}
	if err := register(validation.NewRegexValidatorNode); err != nil {
		return err
	}

	// 7. Communication Nodes
	if err := register(communication.NewEmailNode); err != nil {
		return err
	}

	// 8. Security Nodes
	if err := register(security.NewAESEncryptNode); err != nil {
		return err
	}
	if err := register(security.NewJWTSignNode); err != nil {
		return err
	}
	if err := register(security.NewHashSHA256Node); err != nil {
		return err
	}

	// 9. Utility Nodes
	if err := register(utility.NewSetVariableNode); err != nil {
		return err
	}
	if err := register(utility.NewUUIDNode); err != nil {
		return err
	}
	if err := register(utility.NewRandomNumberNode); err != nil {
		return err
	}
	if err := register(utility.NewDateTimeNode); err != nil {
		return err
	}

	log.Printf("Loaded %d nodes successfully", reg.Count())
	return nil
}

// GetNodeCount returns the number of loaded nodes
func GetNodeCount() int {
	return registry.GetRegistry().Count()
}

// GetNodesByCategory returns nodes filtered by category
func GetNodesByCategory(category string) []base.NodeMetadata {
	return registry.GetRegistry().ListByCategory(category)
}

// SearchNodes searches for nodes
func SearchNodes(query string) []base.NodeMetadata {
	return registry.GetRegistry().Search(query)
}
