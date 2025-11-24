package loader

import (
	"log"

	"github.com/citadel-agent/backend/internal/nodes/base"
	"github.com/citadel-agent/backend/internal/nodes/registry"
	// Import all node packages
	// httpNodes "github.com/citadel-agent/backend/internal/nodes/http"
	// dbNodes "github.com/citadel-agent/backend/internal/nodes/database"
	// aiNodes "github.com/citadel-agent/backend/internal/nodes/ai"
)

// LoadAllNodes registers all available nodes
func LoadAllNodes() error {
	reg := registry.GetRegistry()

	// HTTP Nodes
	if err := registerHTTPNodes(reg); err != nil {
		return err
	}

	// Database Nodes
	// if err := registerDatabaseNodes(reg); err != nil {
	// 	return err
	// }

	// AI Nodes
	// if err := registerAINodes(reg); err != nil {
	// 	return err
	// }

	log.Printf("Loaded %d nodes successfully", reg.Count())
	return nil
}

// registerHTTPNodes registers all HTTP-related nodes
func registerHTTPNodes(reg *registry.Registry) error {
	// Note: HTTP Request node already exists with different signature
	// We'll skip it for now and add other HTTP nodes later

	// TODO: Add other HTTP nodes:
	// - Webhook
	// - GraphQL
	// - OAuth2
	// etc.

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
