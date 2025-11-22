package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/citadel-agent/backend/internal/plugins"
)

// RegisterPluginRequest represents the request to register a plugin
type RegisterPluginRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Type        string `json:"type"` // javascript, python, builtin, custom
}

// RegisterPluginResponse represents the response for registering a plugin
type RegisterPluginResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ListPluginsResponse represents the response for listing plugins
type ListPluginsResponse struct {
	Plugins []PluginInfo `json:"plugins"`
	Count   int          `json:"count"`
}

// PluginInfo represents information about a plugin
type PluginInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Category    string                 `json:"category"`
	Type        plugins.PluginType     `json:"type"`
	Schema      map[string]interface{} `json:"schema"`
	Status      string                 `json:"status"`
	CreatedAt   int64                  `json:"created_at"`
}

// listPlugins returns a list of available plugins
func (s *Server) listPlugins(c *fiber.Ctx) error {
	pluginIDs := s.pluginManager.ListAvailablePlugins()

	pluginList := make([]PluginInfo, 0, len(pluginIDs))

	for _, pluginID := range pluginIDs {
		metadata, err := s.pluginManager.GetNodeMetadata(pluginID)
		if err != nil {
			// If we can't get plugin metadata, use default values
			pluginList = append(pluginList, PluginInfo{
				ID:          pluginID,
				Name:        fmt.Sprintf("%s Plugin", pluginID),
				Description: fmt.Sprintf("Plugin: %s", pluginID),
				Version:     "1.0.0",
				Author:      "Unknown",
				Category:    "utility",
				Type:        plugins.BuiltinPlugin, // Default to builtin
				Schema:      map[string]interface{}{},
				Status:      "available",
			})
			continue
		}

		pluginList = append(pluginList, PluginInfo{
			ID:          metadata.ID,
			Name:        metadata.Name,
			Description: metadata.Description,
			Version:     metadata.Version,
			Author:      metadata.Author,
			Category:    metadata.Category,
			Type:        plugins.BuiltinPlugin, // This would be determined by the plugin
			Schema:      map[string]interface{}{}, // Schema would come from plugin
			Status:      "available",
		})
	}

	response := ListPluginsResponse{
		Plugins: pluginList,
		Count:   len(pluginList),
	}

	return c.JSON(response)
}

// registerPlugin registers a new plugin from a file path
func (s *Server) registerPlugin(c *fiber.Ctx) error {
	var req RegisterPluginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.ID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Plugin ID is required",
		})
	}

	if req.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Plugin path is required",
		})
	}

	// Register the plugin with the plugin manager
	err := s.pluginManager.RegisterPluginAtPath(req.ID, req.Path)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to register plugin: %v", err),
		})
	}

	// If successful, also register it with the temporal service if needed
	// This allows the plugin to be used in Temporal workflows

	return c.JSON(RegisterPluginResponse{
		ID:      req.ID,
		Message: fmt.Sprintf("Plugin %s registered successfully", req.ID),
	})
}

// getPlugin returns details of a specific plugin
func (s *Server) getPlugin(c *fiber.Ctx) error {
	pluginID := c.Params("id")
	if pluginID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Plugin ID is required",
		})
	}

	metadata, err := s.pluginManager.GetNodeMetadata(pluginID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Plugin %s not found", pluginID),
		})
	}

	pluginInfo := PluginInfo{
		ID:          metadata.ID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Version:     metadata.Version,
		Author:      metadata.Author,
		Category:    metadata.Category,
		Type:        plugins.BuiltinPlugin, // This would be determined by implementation
		Schema:      map[string]interface{}{}, // Schema would come from plugin
		Status:      "available",
	}

	return c.JSON(pluginInfo)
}

// unregisterPlugin removes a plugin
func (s *Server) unregisterPlugin(c *fiber.Ctx) error {
	pluginID := c.Params("id")
	if pluginID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Plugin ID is required",
		})
	}

	// Check if plugin exists
	_, err := s.pluginManager.GetNodeMetadata(pluginID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Plugin %s not found", pluginID),
		})
	}

	// Unregister the plugin
	s.pluginManager.UnregisterPlugin(pluginID)

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Plugin %s unregistered successfully", pluginID),
	})
}

// executePlugin executes a plugin with given parameters
func (s *Server) executePlugin(c *fiber.Ctx) error {
	pluginID := c.Params("id")
	if pluginID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Plugin ID is required",
		})
	}

	var params map[string]interface{}
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Execute the plugin
	result, err := s.pluginManager.ExecuteNode(c.Context(), pluginID, params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to execute plugin: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"plugin_id": pluginID,
		"result":    result,
		"success":   true,
	})
}