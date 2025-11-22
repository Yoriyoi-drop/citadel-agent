// backend/internal/workflow/nodes/node_registry.go
package nodes

import (
	"fmt"
	"sync"
)

// NodeInstance interface that all nodes must implement
type NodeInstance interface {
	Execute(input map[string]interface{}) (map[string]interface{}, error)
	GetDescription() string
	GetInputs() []string
	GetOutputs() []string
}

// NodeRegistry manages all available node types
type NodeRegistry struct {
	nodes map[string]NodeInstance
	mutex sync.RWMutex
}

var registryInstance *NodeRegistry
var once sync.Once

// GetNodeRegistry returns singleton instance of NodeRegistry
func GetNodeRegistry() *NodeRegistry {
	once.Do(func() {
		registryInstance = &NodeRegistry{
			nodes: make(map[string]NodeInstance),
		}
		registryInstance.registerDefaultNodes()
	})
	return registryInstance
}

// RegisterNode registers a new node type
func (nr *NodeRegistry) RegisterNode(name string, node NodeInstance) error {
	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	if _, exists := nr.nodes[name]; exists {
		return fmt.Errorf("node %s already registered", name)
	}
	
	nr.nodes[name] = node
	return nil
}

// GetNode returns a registered node instance
func (nr *NodeRegistry) GetNode(name string) (NodeInstance, error) {
	nr.mutex.RLock()
	defer nr.mutex.RUnlock()

	node, exists := nr.nodes[name]
	if !exists {
		return nil, fmt.Errorf("node %s not found", name)
	}
	
	return node, nil
}

// ListNodes returns all registered node types
func (nr *NodeRegistry) ListNodes() map[string]NodeInstance {
	nr.mutex.RLock()
	defer nr.mutex.RUnlock()

	nodes := make(map[string]NodeInstance)
	for name, node := range nr.nodes {
		nodes[name] = node
	}
	
	return nodes
}

// registerDefaultNodes registers built-in node types
func (nr *NodeRegistry) registerDefaultNodes() {
	// Register HTTP Request node
	nr.nodes["http_request"] = &HTTPRequestNode{
		Description: "Makes an HTTP request to a specified URL",
		Inputs:      []string{"url", "method", "headers", "body"},
		Outputs:     []string{"response", "status_code", "headers"},
	}
	
	// Register Delay node
	nr.nodes["delay"] = &DelayNode{
		Description: "Waits for a specified amount of time",
		Inputs:      []string{"duration"},
		Outputs:     []string{"completed"},
	}
	
	// Register Function node
	nr.nodes["function"] = &FunctionNode{
		Description: "Executes custom JavaScript code",
		Inputs:      []string{"code", "context"},
		Outputs:     []string{"result"},
	}
	
	// Register Data Processing node
	nr.nodes["data_process"] = &DataProcessNode{
		Description: "Processes data with transformations",
		Inputs:      []string{"data", "transformations"},
		Outputs:     []string{"result"},
	}
}

// HTTPRequestNode implements HTTP request functionality
type HTTPRequestNode struct {
	Description string
	Inputs      []string
	Outputs     []string
}

func (n *HTTPRequestNode) Execute(input map[string]interface{}) (map[string]interface{}, error) {
	// Implementation would make actual HTTP request
	// This is a simplified version
	return map[string]interface{}{
		"response":    "HTTP response would be here",
		"status_code": 200,
		"headers":     map[string]string{},
	}, nil
}

func (n *HTTPRequestNode) GetDescription() string {
	return n.Description
}

func (n *HTTPRequestNode) GetInputs() []string {
	return n.Inputs
}

func (n *HTTPRequestNode) GetOutputs() []string {
	return n.Outputs
}

// DelayNode implements delay/wait functionality
type DelayNode struct {
	Description string
	Inputs      []string
	Outputs     []string
}

func (n *DelayNode) Execute(input map[string]interface{}) (map[string]interface{}, error) {
	// Implementation would delay for specified duration
	return map[string]interface{}{
		"completed": true,
	}, nil
}

func (n *DelayNode) GetDescription() string {
	return n.Description
}

func (n *DelayNode) GetInputs() []string {
	return n.Inputs
}

func (n *DelayNode) GetOutputs() []string {
	return n.Outputs
}

// FunctionNode implements code execution functionality
type FunctionNode struct {
	Description string
	Inputs      []string
	Outputs     []string
}

func (n *FunctionNode) Execute(input map[string]interface{}) (map[string]interface{}, error) {
	// Implementation would execute JavaScript code
	return map[string]interface{}{
		"result": "Function result would be here",
	}, nil
}

func (n *FunctionNode) GetDescription() string {
	return n.Description
}

func (n *FunctionNode) GetInputs() []string {
	return n.Inputs
}

func (n *FunctionNode) GetOutputs() []string {
	return n.Outputs
}

// DataProcessNode implements data transformation functionality
type DataProcessNode struct {
	Description string
	Inputs      []string
	Outputs     []string
}

func (n *DataProcessNode) Execute(input map[string]interface{}) (map[string]interface{}, error) {
	// Implementation would process data transformations
	return map[string]interface{}{
		"result": "Processed data would be here",
	}, nil
}

func (n *DataProcessNode) GetDescription() string {
	return n.Description
}

func (n *DataProcessNode) GetInputs() []string {
	return n.Inputs
}

func (n *DataProcessNode) GetOutputs() []string {
	return n.Outputs
}