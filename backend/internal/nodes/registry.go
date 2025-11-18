package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sync"

	"citadel-agent/backend/internal/engine"
)

// NodeType defines the grade level of a node
type NodeType string

const (
	Basic       NodeType = "basic"
	Intermediate NodeType = "intermediate"
	Advanced    NodeType = "advanced"
	Elite       NodeType = "elite"
)

// NodeCategory defines the functional category of a node
type NodeCategory string

const (
	AI          NodeCategory = "ai"
	HTTP        NodeCategory = "http"
	Database    NodeCategory = "database"
	File        NodeCategory = "file"
	Security    NodeCategory = "security"
	Utility     NodeCategory = "utility"
	Messaging   NodeCategory = "messaging"
	System      NodeCategory = "system"
	Schedule    NodeCategory = "schedule"
	Text        NodeCategory = "text"
	Image       NodeCategory = "image"
	Network     NodeCategory = "network"
	Workflow    NodeCategory = "workflow"
	Data        NodeCategory = "data"
	Testing     NodeCategory = "testing"
	Logging     NodeCategory = "logging"
	Media       NodeCategory = "media"
	Email       NodeCategory = "email"
	Finance     NodeCategory = "finance"
	Analytics   NodeCategory = "analytics"
	Deployment  NodeCategory = "deployment"
	Cache       NodeCategory = "cache"
)

// NodeDefinition represents the schema for a node
type NodeDefinition struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         NodeType    `json:"type"`
	Category     NodeCategory `json:"category"`
	Icon         string      `json:"icon"`
	SettingsSchema interface{} `json:"settings_schema"`
}

// NodeRegistry manages all available nodes
type NodeRegistry struct {
	nodes map[string]engine.NodeExecutor
	definitions map[string]*NodeDefinition
	mutex sync.RWMutex
}

// NewNodeRegistry creates a new node registry
func NewNodeRegistry() *NodeRegistry {
	registry := &NodeRegistry{
		nodes:       make(map[string]engine.NodeExecutor),
		definitions: make(map[string]*NodeDefinition),
	}

	// Load node definitions from JSON
	if err := registry.loadNodeDefinitions(); err != nil {
		// Log error but don't fail completely - we can work with programmatic registrations
		fmt.Printf("Warning: Could not load node definitions: %v\n", err)
	}

	// Register built-in nodes
	registry.registerBuiltInNodes()

	return registry
}

// RegisterNode registers a node executor with its definition
func (nr *NodeRegistry) RegisterNode(nodeID string, executor engine.NodeExecutor, definition *NodeDefinition) {
	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	nr.nodes[nodeID] = executor
	if definition != nil {
		definition.ID = nodeID
		nr.definitions[nodeID] = definition
	}
}

// GetNodeExecutor retrieves a node executor by ID
func (nr *NodeRegistry) GetNodeExecutor(nodeID string) (engine.NodeExecutor, bool) {
	nr.mutex.RLock()
	defer nr.mutex.RUnlock()

	executor, exists := nr.nodes[nodeID]
	return executor, exists
}

// GetNodeDefinition retrieves a node definition by ID
func (nr *NodeRegistry) GetNodeDefinition(nodeID string) (*NodeDefinition, bool) {
	nr.mutex.RLock()
	defer nr.mutex.RUnlock()

	definition, exists := nr.definitions[nodeID]
	return definition, exists
}

// GetAllNodeDefinitions returns all registered node definitions
func (nr *NodeRegistry) GetAllNodeDefinitions() []*NodeDefinition {
	nr.mutex.RLock()
	defer nr.mutex.RUnlock()

	definitions := make([]*NodeDefinition, 0, len(nr.definitions))
	for _, definition := range nr.definitions {
		definitions = append(definitions, definition)
	}

	return definitions
}

// ExecuteNode executes a specific node with the given input
func (nr *NodeRegistry) ExecuteNode(ctx context.Context, nodeID string, input map[string]interface{}) (*engine.ExecutionResult, error) {
	executor, exists := nr.GetNodeExecutor(nodeID)
	if !exists {
		return &engine.ExecutionResult{
			Status: "error",
			Error:  fmt.Sprintf("Node with ID %s not found", nodeID),
		}, nil
	}

	return executor.Execute(ctx, input)
}

// loadNodeDefinitions loads node definitions from the JSON file
func (nr *NodeRegistry) loadNodeDefinitions() error {
	data, err := ioutil.ReadFile("nodes.json") // Assuming nodes.json is in the working directory
	if err != nil {
		// Try alternative path relative to the executable
		exePath := filepath.Dir("")
		data, err = ioutil.ReadFile(filepath.Join(exePath, "nodes.json"))
		if err != nil {
			return fmt.Errorf("could not read nodes.json file: %w", err)
		}
	}

	var nodeData map[string]interface{}
	if err := json.Unmarshal(data, &nodeData); err != nil {
		return fmt.Errorf("could not parse nodes.json: %w", err)
	}

	nodesArray, ok := nodeData["nodes"].([]interface{})
	if !ok {
		return fmt.Errorf("nodes.json does not contain a 'nodes' array")
	}

	for _, nodeInterface := range nodesArray {
		nodeMap, ok := nodeInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Convert to NodeDefinition
		definition := &NodeDefinition{}
		
		if id, ok := nodeMap["id"].(string); ok {
			definition.ID = id
		} else {
			continue // Skip if no ID
		}
		
		if name, ok := nodeMap["name"].(string); ok {
			definition.Name = name
		}
		
		if desc, ok := nodeMap["description"].(string); ok {
			definition.Description = desc
		}
		
		if typeStr, ok := nodeMap["type"].(string); ok {
			definition.Type = NodeType(typeStr)
		}
		
		if categoryStr, ok := nodeMap["category"].(string); ok {
			definition.Category = NodeCategory(categoryStr)
		}
		
		if icon, ok := nodeMap["icon"].(string); ok {
			definition.Icon = icon
		}
		
		if schema, ok := nodeMap["settings_schema"]; ok {
			definition.SettingsSchema = schema
		}

		nr.definitions[definition.ID] = definition
	}

	return nil
}

// registerBuiltInNodes registers all built-in node executors
func (nr *NodeRegistry) registerBuiltInNodes() {
	// Elite AI Nodes
	nr.RegisterNode("ai_auto_repair", &AIAutoRepairNode{
		modelProvider:  "local",
		repairStrategy: "code_fix",
	}, &NodeDefinition{
		ID:          "ai_auto_repair",
		Name:        "AI Auto Repair Node",
		Description: "Perbaiki node lain otomatis",
		Type:        Elite,
		Category:    AI,
		Icon:        "ai",
	})

	nr.RegisterNode("ai_prompt_generator", &AIPromptGeneratorNode{
		defaultLanguage: "en",
		defaultStyle:    "direct",
	}, &NodeDefinition{
		ID:          "ai_prompt_generator",
		Name:        "AI Prompt Generator",
		Description: "Buat prompt otomatis",
		Type:        Elite,
		Category:    AI,
		Icon:        "ai",
	})

	// Elite Workflow Nodes
	nr.RegisterNode("workflow_time_machine", &WorkflowTimeMachineNode{
		storagePath: "/tmp/workflow_snapshots",
	}, &NodeDefinition{
		ID:          "workflow_time_machine",
		Name:        "Workflow Time Machine",
		Description: "Rollback versi lama",
		Type:        Elite,
		Category:    Workflow,
		Icon:        "time-machine",
	})

	// Elite Testing Nodes
	nr.RegisterNode("load_test", &LoadTestNode{
		defaultConcurrentUsers: 100,
		defaultTestDuration:    60 * time.Second,
		defaultRampUpTime:      10 * time.Second,
	}, &NodeDefinition{
		ID:          "load_test",
		Name:        "Load Test Node",
		Description: "Stress test API",
		Type:        Elite,
		Category:    Testing,
		Icon:        "load-test",
	})

	// Register other nodes here as they are implemented
	// Basic nodes
	nr.RegisterNode("string_replace", &StringReplaceNode{}, &NodeDefinition{
		ID:          "string_replace",
		Name:        "String Replace",
		Description: "Ganti teks otomatis",
		Type:        Basic,
		Category:    Text,
		Icon:        "text",
	})

	nr.RegisterNode("json_path_extractor", &JSONPathExtractorNode{}, &NodeDefinition{
		ID:          "json_path_extractor",
		Name:        "JSON Path Extractor",
		Description: "Ambil data pakai JSONPath",
		Type:        Intermediate,
		Category:    Data,
		Icon:        "json",
	})

	// Add more node registrations as needed
}

// StringReplaceNode is a basic node for replacing strings
type StringReplaceNode struct{}

func (s *StringReplaceNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	inputStr, ok := input["input"].(string)
	if !ok {
		return &engine.ExecutionResult{
			Status: "error",
			Error:  "input is required and must be a string",
		}, nil
	}

	find, ok := input["find"].(string)
	if !ok {
		return &engine.ExecutionResult{
			Status: "error",
			Error:  "find is required and must be a string",
		}, nil
	}

	replace, ok := input["replace"].(string)
	if !ok {
		replace = ""
	}

	global, _ := input["global"].(bool)
	if global {
		// Perform global replacement
		result := ""
		start := 0
		for i := 0; i <= len(inputStr)-len(find); i++ {
			if inputStr[i:i+len(find)] == find {
				result += inputStr[start:i] + replace
				start = i + len(find)
				if !global {
					break // Only replace first occurrence if not global
				}
			}
		}
		result += inputStr[start:]
		
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"original": inputStr,
				"result":   result,
				"replacements": countReplacements(inputStr, find, global),
			},
		}, nil
	} else {
		// Replace first occurrence only
		if idx := indexOf(inputStr, find); idx != -1 {
			result := inputStr[:idx] + replace + inputStr[idx+len(find):]
			return &engine.ExecutionResult{
				Status: "success",
				Data: map[string]interface{}{
					"original": inputStr,
					"result":   result,
				},
			}, nil
		}
		
		// If no match found, return original string
		return &engine.ExecutionResult{
			Status: "success",
			Data: map[string]interface{}{
				"original": inputStr,
				"result":   inputStr,
			},
		}, nil
	}
}

// Helper functions for string operations
func indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func countReplacements(str, substr string, global bool) int {
	if !global {
		if indexOf(str, substr) != -1 {
			return 1
		}
		return 0
	}
	
	count := 0
	start := 0
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			count++
			start = i + len(substr)
		}
	}
	return count
}

// JSONPathExtractorNode extracts data using JSONPath
type JSONPathExtractorNode struct{}

func (j *JSONPathExtractorNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	jsonData, ok := input["json_data"]
	if !ok {
		return &engine.ExecutionResult{
			Status: "error",
			Error:  "json_data is required",
		}, nil
	}

	jsonPath, ok := input["json_path"].(string)
	if !ok {
		return &engine.ExecutionResult{
			Status: "error",
			Error:  "json_path is required and must be a string",
		}, nil
	}

	// In a real implementation, we would parse the JSONPath and extract the value
	// For this example, we'll just return the full data
	result := map[string]interface{}{
		"extracted_data": jsonData,
		"json_path":      jsonPath,
		"full_data":      jsonData,
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data:   result,
	}, nil
}

// GetNodesByType returns all nodes of a specific type (grade)
func (nr *NodeRegistry) GetNodesByType(nodeType NodeType) []*NodeDefinition {
	nr.mutex.RLock()
	defer nr.mutex.RUnlock()

	var nodes []*NodeDefinition
	for _, definition := range nr.definitions {
		if definition.Type == nodeType {
			nodes = append(nodes, definition)
		}
	}

	return nodes
}

// GetNodesByCategory returns all nodes of a specific category
func (nr *NodeRegistry) GetNodesByCategory(category NodeCategory) []*NodeDefinition {
	nr.mutex.RLock()
	defer nr.mutex.RUnlock()

	var nodes []*NodeDefinition
	for _, definition := range nr.definitions {
		if definition.Category == category {
			nodes = append(nodes, definition)
		}
	}

	return nodes
}

// IsNodeRegistered checks if a node is registered
func (nr *NodeRegistry) IsNodeRegistered(nodeID string) bool {
	_, exists := nr.GetNodeExecutor(nodeID)
	return exists
}