package utility

import (
	"context"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"
)

// ForEachNode implements a node that iterates over a collection
type ForEachNode struct {
	id            string
	nodeType      string
	collectionKey string
	itemKey       string
	config        map[string]interface{}
}

// Initialize sets up the for each node with configuration
func (fe *ForEachNode) Initialize(config map[string]interface{}) error {
	fe.config = config

	if collectionKey, ok := config["collection_key"]; ok {
		if key, ok := collectionKey.(string); ok {
			fe.collectionKey = key
		} else {
			return fmt.Errorf("collection_key must be a string")
		}
	} else {
		// Default to using the entire input as the collection
		fe.collectionKey = "" // empty means use entire input
	}

	if itemKey, ok := config["item_key"]; ok {
		if key, ok := itemKey.(string); ok {
			fe.itemKey = key
		} else {
			return fmt.Errorf("item_key must be a string")
		}
	} else {
		fe.itemKey = "item" // default item key
	}

	return nil
}

// Execute iterates over the collection and processes each item
func (fe *ForEachNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	var collection []interface{}

	// Determine the collection to iterate over
	if fe.collectionKey != "" {
		// Get collection from specific key in input data
		if rawCollection, exists := inputs[fe.collectionKey]; exists {
			if col, ok := rawCollection.([]interface{}); ok {
				collection = col
			} else {
				// If it's not already a slice, try to treat it as a single item
				collection = []interface{}{rawCollection}
			}
		} else {
			// Key doesn't exist, create empty collection
			collection = []interface{}{}
		}
	} else {
		// Use the entire input data as a single item
		collection = []interface{}{inputs}
	}

	// Process each item in the collection
	results := make([]interface{}, 0, len(collection))

	for i, item := range collection {
		// Create a context for this iteration
		itemContext := map[string]interface{}{
			fe.itemKey: item,
			"index":    i,
			"total":    len(collection),
			"first":    i == 0,
			"last":     i == len(collection)-1,
		}

		// Add the original input data as well
		itemContext["original_input"] = inputs

		results = append(results, itemContext)
	}

	// Prepare output
	output := map[string]interface{}{
		"results":         results,
		"processed_count": len(results),
		"original_input":  inputs,
		"collection_key":  fe.collectionKey,
		"item_key":        fe.itemKey,
	}

	return output, nil
}

// GetType returns the type of the node
func (fe *ForEachNode) GetType() string {
	return fe.nodeType
}

// GetID returns the unique identifier for this node instance
func (fe *ForEachNode) GetID() string {
	return fe.id
}

// NewForEachNode creates a new for each node constructor for the registry
func NewForEachNode(config map[string]interface{}) (interfaces.NodeInstance, error) {
	node := &ForEachNode{
		id:       fmt.Sprintf("foreach_%d", time.Now().UnixNano()),
		nodeType: "for_each",
	}

	if err := node.Initialize(config); err != nil {
		return nil, err
	}

	return node, nil
}