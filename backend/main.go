package main

import (
	"log"
	"net/http"
	"os"

	"citadel-agent/backend/internal/api/handlers"
	httpnode "citadel-agent/backend/internal/nodes/http"
	"citadel-agent/backend/internal/nodes/utility"
	"citadel-agent/backend/internal/workflow/core/engine"
)

func main() {
	// Initialize the node registry
	registry := engine.NewNodeTypeRegistry()

	// Register node types
	registerNodes(registry)

	// Initialize workflow executor
	executor := engine.NewWorkflowExecutor(registry)

	// Initialize handlers
	workflowHandler := handlers.NewWorkflowHandler(executor)
	nodeHandler := handlers.NewNodeHandler(registry)

	// Set up routes
	setupRoutes(workflowHandler, nodeHandler)

	// Start server
	port := getPort()
	log.Printf("Starting Citadel Agent API server on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func registerNodes(registry *engine.NodeTypeRegistryImpl) {
	// Register the HTTP Request node
	httpMetadata := httpnode.NewHTTPRequestNode().GetMetadata()
	registry.RegisterNodeType(
		"http_request",
		httpnode.NewHTTPRequestNode,
		httpMetadata,
	)

	// Register the Logger node
	loggerMetadata := utility.NewLoggerNode().GetMetadata()
	registry.RegisterNodeType(
		"logger",
		utility.NewLoggerNode,
		loggerMetadata,
	)

	// Register the Data Transformer node
	transformerMetadata := utility.NewDataTransformerNode().GetMetadata()
	registry.RegisterNodeType(
		"data_transformer",
		utility.NewDataTransformerNode,
		transformerMetadata,
	)

	// Register the If/Else node
	ifelseMetadata := utility.NewIfElseNode().GetMetadata()
	registry.RegisterNodeType(
		"if_else",
		utility.NewIfElseNode,
		ifelseMetadata,
	)

	// Register the For Each node
	foreachMetadata := utility.NewForEachNode().GetMetadata()
	registry.RegisterNodeType(
		"for_each",
		utility.NewForEachNode,
		foreachMetadata,
	)

	log.Printf("Registered %d node types", len(registry.ListNodeTypes()))
}

func setupRoutes(workflowHandler *handlers.WorkflowHandler, nodeHandler *handlers.NodeHandler) {
	// CORS middleware
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Allow requests from frontend
			origin := r.Header.Get("Origin")
			allowedOrigins := []string{
				"http://localhost:3000",
				"http://localhost:5173",
				"http://localhost:8080",
			}

			for _, allowed := range allowedOrigins {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Workflow routes
	http.HandleFunc("/api/workflows/execute", corsMiddleware(workflowHandler.ExecuteWorkflowHandler))
	http.HandleFunc("/api/workflows/", corsMiddleware(workflowHandler.GetWorkflowHandler))
	http.HandleFunc("/api/workflows", corsMiddleware(workflowHandler.ListWorkflowsHandler))

	// Node routes
	http.HandleFunc("/api/nodes/", corsMiddleware(nodeHandler.GetNodeHandler))
	http.HandleFunc("/api/nodes", corsMiddleware(nodeHandler.ListNodesHandler))

	// Registry routes (for frontend node palette)
	http.HandleFunc("/api/v1/registry/nodes", corsMiddleware(nodeHandler.ListNodesHandler))

	// Root endpoint
	http.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to Citadel Agent API", "version": "0.1.0"}`))
	}))
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
