package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/citadel-agent/backend/internal/interfaces"
)

// RegisterNodeTypeRequest represents the request to register a node type
type RegisterNodeTypeRequest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ConfigSchema map[string]interface{} `json:"config_schema"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
}

// ListNodeTypesResponse represents the response for listing node types
type ListNodeTypesResponse struct {
	NodeTypes []NodeTypeInfo `json:"node_types"`
	Count     int            `json:"count"`
}

// NodeTypeInfo represents information about a node type
type NodeTypeInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Plugin      bool                   `json:"plugin"`
	ConfigSchema map[string]interface{} `json:"config_schema"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	CreatedAt   int64                  `json:"created_at"`
}

// listNodeTypes returns a list of available node types
func (s *Server) listNodeTypes(c *fiber.Ctx) error {
	// Get list of built-in node types
	builtinNodes := []string{
		"http_request",
		"condition",
		"delay",
		"database_query",
		"script_execution",
		"ai_agent",
		"data_transformer",
		"notification",
		"loop",
		"error_handler",
	}

	// Get list of plugin node types
	pluginNodes := s.pluginManager.ListAvailablePlugins()

	// Combine both lists
	nodeTypes := make([]NodeTypeInfo, 0, len(builtinNodes)+len(pluginNodes))

	// Add built-in nodes
	for _, nodeType := range builtinNodes {
		nodeTypes = append(nodeTypes, NodeTypeInfo{
			ID:          nodeType,
			Name:        fmt.Sprintf("%s Node", capitalizeFirst(nodeType)),
			Description: fmt.Sprintf("Built-in %s node", nodeType),
			Category:    "builtin",
			Plugin:      false,
			ConfigSchema: map[string]interface{}{},
			Inputs:      map[string]interface{}{},
			Outputs:     map[string]interface{}{},
		})
	}

	// Add plugin nodes
	for _, pluginID := range pluginNodes {
		// Get metadata from plugin
		metadata, err := s.pluginManager.GetNodeMetadata(pluginID)
		if err != nil {
			// If we can't get plugin metadata, use default values
			nodeTypes = append(nodeTypes, NodeTypeInfo{
				ID:          pluginID,
				Name:        fmt.Sprintf("%s Plugin Node", pluginID),
				Description: fmt.Sprintf("Plugin node: %s", pluginID),
				Category:    "plugin",
				Plugin:      true,
				ConfigSchema: map[string]interface{}{},
				Inputs:      map[string]interface{}{},
				Outputs:     map[string]interface{}{},
			})
			continue
		}

		nodeTypes = append(nodeTypes, NodeTypeInfo{
			ID:          metadata.ID,
			Name:        metadata.Name,
			Description: metadata.Description,
			Category:    metadata.Category,
			Plugin:      true,
			ConfigSchema: map[string]interface{}{}, // This would come from plugin
			Inputs:      map[string]interface{}{},
			Outputs:     map[string]interface{}{},
		})
	}

	response := ListNodeTypesResponse{
		NodeTypes: nodeTypes,
		Count:     len(nodeTypes),
	}

	return c.JSON(response)
}

// registerNodeType registers a new node type
func (s *Server) registerNodeType(c *fiber.Ctx) error {
	var req RegisterNodeTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// For now, return an error since node registration via API is not supported
	// Nodes should be registered through code or plugin system
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Dynamic node registration via API is not supported. Use plugin system instead.",
	})
}

// listPluginNodes returns a list of available plugin nodes
func (s *Server) listPluginNodes(c *fiber.Ctx) error {
	pluginNodes := s.pluginManager.ListAvailablePlugins()

	pluginList := make([]NodeTypeInfo, 0, len(pluginNodes))

	for _, pluginID := range pluginNodes {
		metadata, err := s.pluginManager.GetNodeMetadata(pluginID)
		if err != nil {
			// If we can't get plugin metadata, use default values
			pluginList = append(pluginList, NodeTypeInfo{
				ID:          pluginID,
				Name:        fmt.Sprintf("%s Plugin Node", pluginID),
				Description: fmt.Sprintf("Plugin node: %s", pluginID),
				Category:    "plugin",
				Plugin:      true,
				ConfigSchema: map[string]interface{}{},
				Inputs:      map[string]interface{}{},
				Outputs:     map[string]interface{}{},
			})
			continue
		}

		pluginList = append(pluginList, NodeTypeInfo{
			ID:          metadata.ID,
			Name:        metadata.Name,
			Description: metadata.Description,
			Category:    metadata.Category,
			Plugin:      true,
			ConfigSchema: map[string]interface{}{}, // This would come from plugin
			Inputs:      map[string]interface{}{},
			Outputs:     map[string]interface{}{},
		})
	}

	response := ListNodeTypesResponse{
		NodeTypes: pluginList,
		Count:     len(pluginList),
	}

	return c.JSON(response)
}

// Helper function to capitalize the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}