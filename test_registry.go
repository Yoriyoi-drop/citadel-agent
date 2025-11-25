package main

import (
	"fmt"
	"citadel-agent/backend/internal/nodes"
)

func main() {
	factory := nodes.GetNodeFactory()
	types := factory.ListNodeTypes()
	
	fmt.Printf("Successfully loaded %d node types\n", len(types))
	
	// Check if some of the elite AI nodes are registered
	eliteAICount := 0
	for _, nodeType := range types {
		if string(nodeType) == "vision_ai_processor" ||
			string(nodeType) == "speech_to_text" ||
			string(nodeType) == "text_to_speech" ||
			string(nodeType) == "contextual_reasoning" ||
			string(nodeType) == "anomaly_detection_ai" {
			eliteAICount++
		}
	}
	
	fmt.Printf("Found %d Elite AI node types in registry\n", eliteAICount)
}