package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/citadel-agent/backend/internal/workflow/core/engine"
	"github.com/citadel-agent/backend/internal/workflow/core/types"
)

// NodeHandler handles node-related API requests
type NodeHandler struct {
	registry *engine.NodeTypeRegistryImpl
}

// NewNodeHandler creates a new node handler
func NewNodeHandler(registry *engine.NodeTypeRegistryImpl) *NodeHandler {
	if registry == nil {
		registry = engine.NewNodeTypeRegistry()
	}
	return &NodeHandler{
		registry: registry,
	}
}

// ListNodesHandler returns all available node types
func (nh *NodeHandler) ListNodesHandler(w http.ResponseWriter, r *http.Request) {
	nodeTypes := nh.registry.ListNodeTypes()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nodes": nodeTypes,
		"count": len(nodeTypes),
	})
}

// GetNodeHandler returns details for a specific node type
func (nh *NodeHandler) GetNodeHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Path[len("/api/nodes/"):] // Extract node ID from URL path

	metadata, exists := nh.registry.GetNodeMetadata(nodeID)
	if !exists {
		http.Error(w, "Node type not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"node": metadata,
	})
}

// RegisterNodeHandler allows registering new node types via API (for development)
func (nh *NodeHandler) RegisterNodeHandler(w http.ResponseWriter, r *http.Request) {
	// This would typically only be available in development mode
	// For security reasons, this should be limited in production

	// TODO: Add security checks to ensure only authorized users can register nodes

	var config struct {
		ID          string                 `json:"id"`
		Name        string                 `json:"name"`
		Category    string                 `json:"category"`
		Description string                 `json:"description"`
		Inputs      map[string]interface{} `json:"inputs"`
		Outputs     map[string]interface{} `json:"outputs"`
		Icon        string                 `json:"icon"`
	}

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid node configuration", http.StatusBadRequest)
		return
	}

	// For now, we just register a mock node
	// In a real implementation, we would need to load and register an actual node implementation
	metadata := types.NodeMetadata{
		ID:          config.ID,
		Name:        config.Name,
		Category:    config.Category,
		Description: config.Description,
		Inputs:      config.Inputs,
		Outputs:     config.Outputs,
		Icon:        config.Icon,
	}

	// Register the node type (with a mock creator function)
	// Note: This won't actually work without the real implementation
	err := nh.registry.RegisterNodeType(config.ID, nil, metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Node type registered",
		"node_id": config.ID,
	})
}
