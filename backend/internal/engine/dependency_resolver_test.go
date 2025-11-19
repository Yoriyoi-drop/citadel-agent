package engine

import (
	"testing"
)

func TestDependencyResolver_ResolveExecutionOrder(t *testing.T) {
	// Test case 1: Simple linear workflow A -> B -> C
	nodes := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
		{ID: "C", Type: "test", Name: "Node C"},
	}
	edges := []Edge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "B", Target: "C"},
	}

	resolver := NewDependencyResolver(nodes, edges)
	order, err := resolver.ResolveExecutionOrder()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedOrder := []string{"A", "B", "C"}
	if len(order) != len(expectedOrder) {
		t.Errorf("Expected order length %d, got %d", len(expectedOrder), len(order))
	}

	for i, nodeID := range order {
		if nodeID != expectedOrder[i] {
			t.Errorf("Expected node %s at position %d, got %s", expectedOrder[i], i, nodeID)
		}
	}

	// Test case 2: Parallel workflow A -> B, A -> C
	nodes2 := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
		{ID: "C", Type: "test", Name: "Node C"},
	}
	edges2 := []Edge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "A", Target: "C"},
	}

	resolver2 := NewDependencyResolver(nodes2, edges2)
	order2, err := resolver2.ResolveExecutionOrder()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// The order should have A first, then B and C in any order
	if order2[0] != "A" {
		t.Errorf("Expected A as first element, got %s", order2[0])
	}

	// B and C should both be present after A
	if (order2[1] != "B" || order2[2] != "C") && (order2[1] != "C" || order2[2] != "B") {
		t.Errorf("Expected B and C after A in any order, got %v", order2)
	}

	// Test case 3: Cycle detection
	nodes3 := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
	}
	edges3 := []Edge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "B", Target: "A"}, // Creates a cycle
	}

	resolver3 := NewDependencyResolver(nodes3, edges3)
	_, err = resolver3.ResolveExecutionOrder()
	if err == nil {
		t.Error("Expected error for cyclic graph, got nil")
	}
}

func TestDependencyResolver_GetNodeDependencies(t *testing.T) {
	nodes := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
		{ID: "C", Type: "test", Name: "Node C"},
	}
	edges := []Edge{
		{ID: "e1", Source: "A", Target: "C"},
		{ID: "e2", Source: "B", Target: "C"},
	}

	resolver := NewDependencyResolver(nodes, edges)

	deps, err := resolver.GetNodeDependencies("C")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies for node C, got %d", len(deps))
	}

	// Check that both A and B are dependencies
	hasA := false
	hasB := false
	for _, dep := range deps {
		if dep == "A" {
			hasA = true
		}
		if dep == "B" {
			hasB = true
		}
	}

	if !hasA || !hasB {
		t.Errorf("Expected both A and B to be dependencies of C, got %v", deps)
	}
}

func TestDependencyResolver_GetNodeDependents(t *testing.T) {
	nodes := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
		{ID: "C", Type: "test", Name: "Node C"},
	}
	edges := []Edge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "A", Target: "C"},
	}

	resolver := NewDependencyResolver(nodes, edges)

	deps, err := resolver.GetNodeDependents("A")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 dependents for node A, got %d", len(deps))
	}

	// Check that both B and C are dependents
	hasB := false
	hasC := false
	for _, dep := range deps {
		if dep == "B" {
			hasB = true
		}
		if dep == "C" {
			hasC = true
		}
	}

	if !hasB || !hasC {
		t.Errorf("Expected both B and C to be dependents of A, got %v", deps)
	}
}

func TestDependencyResolver_ValidateWorkflow(t *testing.T) {
	// Test valid workflow
	nodes := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
	}
	edges := []Edge{
		{ID: "e1", Source: "A", Target: "B"},
	}

	resolver := NewDependencyResolver(nodes, edges)
	err := resolver.ValidateWorkflow()
	if err != nil {
		t.Errorf("Expected no error for valid workflow, got: %v", err)
	}

	// Test invalid workflow (non-existent node reference)
	edgesInvalid := []Edge{
		{ID: "e1", Source: "A", Target: "D"}, // D doesn't exist
	}

	resolverInvalid := NewDependencyResolver(nodes, edgesInvalid)
	err = resolverInvalid.ValidateWorkflow()
	if err == nil {
		t.Error("Expected error for invalid node reference, got nil")
	}

	// Test cyclic workflow
	nodesCycle := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
	}
	edgesCycle := []Edge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "B", Target: "A"},
	}

	resolverCycle := NewDependencyResolver(nodesCycle, edgesCycle)
	err = resolverCycle.ValidateWorkflow()
	if err == nil {
		t.Error("Expected error for cyclic workflow, got nil")
	}
}

func TestDependencyResolver_GetExecutionLayers(t *testing.T) {
	// Test case: A -> B -> D and A -> C -> D
	nodes := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
		{ID: "C", Type: "test", Name: "Node C"},
		{ID: "D", Type: "test", Name: "Node D"},
	}
	edges := []Edge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "A", Target: "C"},
		{ID: "e3", Source: "B", Target: "D"},
		{ID: "e4", Source: "C", Target: "D"},
	}

	resolver := NewDependencyResolver(nodes, edges)
	layers, err := resolver.GetExecutionLayers()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Expected layers: [A], [B, C], [D]
	if len(layers) != 3 {
		t.Errorf("Expected 3 layers, got %d", len(layers))
	}

	if len(layers[0]) != 1 || layers[0][0] != "A" {
		t.Errorf("Expected [A] in first layer, got %v", layers[0])
	}

	if len(layers[2]) != 1 || layers[2][0] != "D" {
		t.Errorf("Expected [D] in last layer, got %v", layers[2])
	}

	// Check that B and C are in the middle layer (order doesn't matter)
	if len(layers[1]) != 2 {
		t.Errorf("Expected 2 nodes in middle layer, got %d", len(layers[1]))
	}
}

func TestDependencyResolver_CanExecute(t *testing.T) {
	nodes := []Node{
		{ID: "A", Type: "test", Name: "Node A"},
		{ID: "B", Type: "test", Name: "Node B"},
		{ID: "C", Type: "test", Name: "Node C"},
	}
	edges := []Edge{
		{ID: "e1", Source: "A", Target: "C"},
		{ID: "e2", Source: "B", Target: "C"},
	}

	resolver := NewDependencyResolver(nodes, edges)

	// Initially, A and B can execute (no dependencies)
	executedNodes := make(map[string]bool)

	canA, err := resolver.CanExecute("A", executedNodes)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !canA {
		t.Error("Expected A to be executable (no dependencies)")
	}

	canB, err := resolver.CanExecute("B", executedNodes)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !canB {
		t.Error("Expected B to be executable (no dependencies)")
	}

	// C cannot execute yet (both A and B are dependencies)
	canC, err := resolver.CanExecute("C", executedNodes)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if canC {
		t.Error("Expected C to not be executable (dependencies not met)")
	}

	// Execute A, C still can't execute
	executedNodes["A"] = true
	canC, _ = resolver.CanExecute("C", executedNodes)
	if canC {
		t.Error("Expected C to not be executable (B not executed yet)")
	}

	// Execute B, now C can execute
	executedNodes["B"] = true
	canC, _ = resolver.CanExecute("C", executedNodes)
	if !canC {
		t.Error("Expected C to be executable (all dependencies met)")
	}
}