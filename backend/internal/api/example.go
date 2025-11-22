package api

import (
	"context"
	"fmt"
	"time"
)

// Example demonstrates how to use the API server
func Example() {
	fmt.Println("Citadel Agent API Server Example")
	
	// This would show how to:
	// 1. Initialize all required components
	// 2. Create the server
	// 3. Register workflows and plugins
	// 4. Start the server
	
	// The actual implementation is in cmd/api/main.go
	// This example shows the concepts
}

// ExampleAPICalls demonstrates API usage patterns
func ExampleAPICalls() {
	fmt.Println("Example API calls for Citadel Agent:")
	fmt.Println("")
	fmt.Println("1. Health check:")
	fmt.Println("   GET http://localhost:3000/health")
	fmt.Println("")
	fmt.Println("2. List workflows:")
	fmt.Println("   GET http://localhost:3000/api/v1/workflows")
	fmt.Println("")
	fmt.Println("3. Create a workflow:")
	fmt.Println("   POST http://localhost:3000/api/v1/workflows")
	fmt.Println("   {")
	fmt.Println("     \"id\": \"my-workflow\",")
	fmt.Println("     \"name\": \"My Workflow\",")
	fmt.Println("     \"description\": \"A sample workflow\",")
	fmt.Println("     \"nodes\": [...],")
	fmt.Println("     \"connections\": [...],")
	fmt.Println("     \"options\": {...}")
	fmt.Println("   }")
	fmt.Println("")
	fmt.Println("4. Execute a workflow:")
	fmt.Println("   POST http://localhost:3000/api/v1/workflows/my-workflow/execute")
	fmt.Println("   {")
	fmt.Println("     \"parameters\": {")
	fmt.Println("       \"input1\": \"value1\",")
	fmt.Println("       \"input2\": \"value2\"")
	fmt.Println("     }")
	fmt.Println("   }")
	fmt.Println("")
	fmt.Println("5. Get workflow status:")
	fmt.Println("   GET http://localhost:3000/api/v1/workflows/my-workflow/status")
	fmt.Println("")
	fmt.Println("6. List plugins:")
	fmt.Println("   GET http://localhost:3000/api/v1/plugins")
	fmt.Println("")
	fmt.Println("7. Register a plugin:")
	fmt.Println("   POST http://localhost:3000/api/v1/plugins/register")
	fmt.Println("   {")
	fmt.Println("     \"id\": \"my-plugin\",")
	fmt.Println("     \"path\": \"/path/to/plugin\",")
	fmt.Println("     \"name\": \"My Plugin\",")
	fmt.Println("     \"description\": \"A sample plugin\"")
	fmt.Println("   }")
	fmt.Println("")
	fmt.Println("8. Execute a plugin directly:")
	fmt.Println("   POST http://localhost:3000/api/v1/plugins/my-plugin/execute")
	fmt.Println("   {")
	fmt.Println("     \"param1\": \"value1\",")
	fmt.Println("     \"param2\": \"value2\"")
	fmt.Println("   }")
}