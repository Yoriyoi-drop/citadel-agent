package engine

import (
	"fmt"
)

// DependencyResolver handles complex node dependencies and execution order
type DependencyResolver struct {
	nodes map[string]*Node
	edges []Edge
}

// NewDependencyResolver creates a new dependency resolver instance
func NewDependencyResolver(nodes []Node, edges []Edge) *DependencyResolver {
	nodeMap := make(map[string]*Node)
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	return &DependencyResolver{
		nodes: nodeMap,
		edges: edges,
	}
}

// ResolveExecutionOrder returns the order in which nodes should be executed
// using topological sorting to handle dependencies
func (dr *DependencyResolver) ResolveExecutionOrder() ([]string, error) {
	// Build adjacency list for dependencies
	adjacencyList := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize all nodes in the maps
	for nodeID := range dr.nodes {
		adjacencyList[nodeID] = []string{}
		inDegree[nodeID] = 0
	}

	// Populate adjacency list and in-degree counts
	for _, edge := range dr.edges {
		source := edge.Source
		target := edge.Target

		// Validate that both nodes exist
		if _, exists := dr.nodes[source]; !exists {
			return nil, fmt.Errorf("source node %s does not exist in workflow", source)
		}
		if _, exists := dr.nodes[target]; !exists {
			return nil, fmt.Errorf("target node %s does not exist in workflow", target)
		}

		adjacencyList[source] = append(adjacencyList[source], target)
		inDegree[target]++
	}

	// Find all nodes with no incoming edges (in-degree = 0)
	var queue []string
	for nodeID := range dr.nodes {
		if inDegree[nodeID] == 0 {
			queue = append(queue, nodeID)
		}
	}

	var executionOrder []string

	// Process nodes in topological order
	for len(queue) > 0 {
		// Remove node from front of queue
		currentNode := queue[0]
		queue = queue[1:]

		executionOrder = append(executionOrder, currentNode)

		// Process all neighbors of the current node
		for _, neighbor := range adjacencyList[currentNode] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles - if not all nodes were processed, there's a cycle
	if len(executionOrder) != len(dr.nodes) {
		return nil, fmt.Errorf("workflow contains circular dependencies")
	}

	return executionOrder, nil
}

// GetNodeDependencies returns all direct dependencies (predecessor nodes) for a given node
func (dr *DependencyResolver) GetNodeDependencies(nodeID string) ([]string, error) {
	if _, exists := dr.nodes[nodeID]; !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	var dependencies []string
	for _, edge := range dr.edges {
		if edge.Target == nodeID {
			dependencies = append(dependencies, edge.Source)
		}
	}

	return dependencies, nil
}

// GetNodeDependents returns all direct dependents (successor nodes) for a given node
func (dr *DependencyResolver) GetNodeDependents(nodeID string) ([]string, error) {
	if _, exists := dr.nodes[nodeID]; !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	var dependents []string
	for _, edge := range dr.edges {
		if edge.Source == nodeID {
			dependents = append(dependents, edge.Target)
		}
	}

	return dependents, nil
}

// ValidateWorkflow checks if the workflow is valid (no cycles, all nodes referenced exist)
func (dr *DependencyResolver) ValidateWorkflow() error {
	// Check for cycles using topological sort
	_, err := dr.ResolveExecutionOrder()
	if err != nil {
		return fmt.Errorf("workflow validation failed: %w", err)
	}

	// Verify all edges connect existing nodes
	for _, edge := range dr.edges {
		if _, exists := dr.nodes[edge.Source]; !exists {
			return fmt.Errorf("edge references non-existent source node: %s", edge.Source)
		}
		if _, exists := dr.nodes[edge.Target]; !exists {
			return fmt.Errorf("edge references non-existent target node: %s", edge.Target)
		}
	}

	return nil
}

// GetExecutionLayers returns nodes grouped by execution layers where each layer
// can be executed in parallel since they don't depend on each other
func (dr *DependencyResolver) GetExecutionLayers() ([][]string, error) {
	// Validate the workflow before processing
	if err := dr.ValidateWorkflow(); err != nil {
		return nil, err
	}

	// Group nodes by their maximum distance from start nodes
	// This creates execution layers
	nodeLevels := make(map[string]int)
	for nodeID := range dr.nodes {
		nodeLevels[nodeID] = -1
	}

	// Find start nodes (nodes with no incoming edges)
	startNodes := []string{}
	for nodeID := range dr.nodes {
		hasIncomingEdge := false
		for _, edge := range dr.edges {
			if edge.Target == nodeID {
				hasIncomingEdge = true
				break
			}
		}
		if !hasIncomingEdge {
			startNodes = append(startNodes, nodeID)
			nodeLevels[nodeID] = 0
		}
	}

	// Calculate levels for all nodes using BFS
	for _, startNode := range startNodes {
		dr.calculateNodeLevels(startNode, nodeLevels)
	}

	// Group nodes by level
	maxLevel := 0
	for _, level := range nodeLevels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	layers := make([][]string, maxLevel+1)
	for nodeID, level := range nodeLevels {
		if level >= 0 {
			layers[level] = append(layers[level], nodeID)
		}
	}

	return layers, nil
}

// Helper function to calculate node levels using BFS
func (dr *DependencyResolver) calculateNodeLevels(startNode string, nodeLevels map[string]int) {
	queue := []string{startNode}

	for len(queue) > 0 {
		currentNode := queue[0]
		queue = queue[1:]

		dependents, _ := dr.GetNodeDependents(currentNode)
		currentLevel := nodeLevels[currentNode]

		for _, dependent := range dependents {
			// If this dependent hasn't been assigned a level yet, or its level is less than current+1
			if nodeLevels[dependent] == -1 || nodeLevels[dependent] < currentLevel+1 {
				nodeLevels[dependent] = currentLevel + 1
				queue = append(queue, dependent)
			}
		}
	}
}

// CanExecute returns true if a node can be executed based on its dependencies
// (i.e., all dependency nodes have been successfully executed)
func (dr *DependencyResolver) CanExecute(nodeID string, executedNodes map[string]bool) (bool, error) {
	if _, exists := dr.nodes[nodeID]; !exists {
		return false, fmt.Errorf("node %s does not exist", nodeID)
	}

	dependencies, err := dr.GetNodeDependencies(nodeID)
	if err != nil {
		return false, err
	}

	// Check if all dependencies have been executed
	for _, dependency := range dependencies {
		if !executedNodes[dependency] {
			return false, nil // Not all dependencies executed yet
		}
	}

	return true, nil
}