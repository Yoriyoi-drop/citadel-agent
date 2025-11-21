// backend/internal/api/server.go
package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/citadel-agent/backend/internal/auth"
	"github.com/citadel-agent/backend/internal/ai"
	"github.com/citadel-agent/backend/internal/runtimes"
	"github.com/citadel-agent/backend/internal/engine"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Server represents the API server
type Server struct {
	app         *fiber.App
	db          *pgxpool.Pool
	authService *auth.AuthService
	aiService   *ai.AIService
	runtimeMgr  *runtimes.MultiRuntimeManager
	nodeRegistry *engine.NodeRegistry
	executor     *engine.Executor
	runner       *engine.Runner
}

// NewServer creates a new API server instance
func NewServer(
	db *pgxpool.Pool,
	authSvc *auth.AuthService,
	aiSvc *ai.AIService,
	nodeRegistry *engine.NodeRegistry,
	executor *engine.Executor,
	runner *engine.Runner,
	runtimeMgr *runtimes.MultiRuntimeManager,
) *Server {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log the error
			log.Printf("Error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Create server instance
	server := &Server{
		app:          app,
		db:           db,
		authService:  authSvc,
		aiService:    aiSvc,
		runtimeMgr:   runtimeMgr,
		nodeRegistry: nodeRegistry,
		executor:     executor,
		runner:       runner,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// OAuthHandler handles OAuth authentication flows
type OAuthHandler struct {
	githubConfig *oauth2.Config
	googleConfig *oauth2.Config
	authService  *auth.AuthService
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(authService *auth.AuthService) *OAuthHandler {
	// Initialize GitHub OAuth config
	githubConfig := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_CALLBACK_URL"), // e.g., http://localhost:5001/api/v1/auth/github/callback
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Initialize Google OAuth config
	googleConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"), // e.g., http://localhost:5001/api/v1/auth/google/callback
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	return &OAuthHandler{
		githubConfig: githubConfig,
		googleConfig: googleConfig,
		authService:  authService,
	}
}

// GithubAuth redirects to GitHub OAuth
func (h *OAuthHandler) GithubAuth(c *fiber.Ctx) error {
	url := h.githubConfig.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

// GithubCallback handles GitHub OAuth callback
func (h *OAuthHandler) GithubCallback(c *fiber.Ctx) error {
	// Get the authorization code from the callback
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No authorization code provided",
		})
	}

	// Exchange code for token
	token, err := h.githubConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	// Get user info from GitHub API
	user, err := h.getGithubUser(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get GitHub user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info from GitHub",
		})
	}

	// In a real implementation, you would create or update the user in your database
	// and generate a JWT token for the user

	// For now, return the user info as a simple response
	return c.JSON(fiber.Map{
		"message": "GitHub authentication successful",
		"user":    user,
		"token":   token.AccessToken, // In a real implementation, this would be your JWT
	})
}

// GoogleAuth redirects to Google OAuth
func (h *OAuthHandler) GoogleAuth(c *fiber.Ctx) error {
	url := h.googleConfig.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

// GoogleCallback handles Google OAuth callback
func (h *OAuthHandler) GoogleCallback(c *fiber.Ctx) error {
	// Get the authorization code from the callback
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No authorization code provided",
		})
	}

	// Exchange code for token
	token, err := h.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange code for token",
		})
	}

	// Get user info from Google API
	user, err := h.getGoogleUser(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get Google user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info from Google",
		})
	}

	// In a real implementation, you would create or update the user in your database
	// and generate a JWT token for the user

	// For now, return the user info as a simple response
	return c.JSON(fiber.Map{
		"message": "Google authentication successful",
		"user":    user,
		"token":   token.AccessToken, // In a real implementation, this would be your JWT
	})
}

// getGithubUser retrieves user information from GitHub API
func (h *OAuthHandler) getGithubUser(accessToken string) (map[string]interface{}, error) {
	// In a real implementation, this would make an HTTP request to GitHub API
	// For example: GET https://api.github.com/user with Authorization: Bearer {token}
	
	// Mock implementation
	user := map[string]interface{}{
		"id":    "github_user_id",
		"name":  "GitHub User",
		"email": "githubuser@example.com",
		"avatar_url": "https://avatars.githubusercontent.com/u/123456789",
	}
	
	fmt.Println("Retrieving user info from GitHub with token:", accessToken[:10]+"...")
	return user, nil
}

// getGoogleUser retrieves user information from Google API
func (h *OAuthHandler) getGoogleUser(accessToken string) (map[string]interface{}, error) {
	// In a real implementation, this would make an HTTP request to Google API
	// For example: GET https://www.googleapis.com/oauth2/v2/userinfo with Authorization: Bearer {token}
	
	// Mock implementation
	user := map[string]interface{}{
		"id":    "google_user_id",
		"name":  "Google User",
		"email": "googleuser@example.com",
		"avatar_url": "https://lh3.googleusercontent.com/a-/123456789",
	}
	
	fmt.Println("Retrieving user info from Google with token:", accessToken[:10]+"...")
	return user, nil
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Create OAuth handler
	oauthHandler := NewOAuthHandler(s.authService)

	// API v1 group
	v1 := s.app.Group("/api/v1")

	// Public routes
	v1.Post("/auth/login", s.handleLogin)
	v1.Post("/auth/register", s.handleRegister)
	
	// OAuth routes
	v1.Get("/auth/github", oauthHandler.GithubAuth)
	v1.Get("/auth/github/callback", oauthHandler.GithubCallback)
	v1.Get("/auth/google", oauthHandler.GoogleAuth)
	v1.Get("/auth/google/callback", oauthHandler.GoogleCallback)

	// Protected routes
	v1.Use(s.authService.AuthMiddleware())
	{
		// Workflow routes
		v1.Post("/workflows", s.handleCreateWorkflow)
		v1.Get("/workflows", s.handleGetWorkflows)
		v1.Get("/workflows/:id", s.handleGetWorkflow)
		v1.Put("/workflows/:id", s.handleUpdateWorkflow)
		v1.Delete("/workflows/:id", s.handleDeleteWorkflow)
		v1.Post("/workflows/:id/run", s.handleRunWorkflow)

		// Node routes
		v1.Get("/nodes", s.handleGetNodes)
		v1.Get("/nodes/types", s.handleGetNodeTypes)

		// AI agent routes
		v1.Post("/ai-agents", s.handleCreateAgent)
		v1.Get("/ai-agents", s.handleGetAgents)
		v1.Get("/ai-agents/:id", s.handleGetAgent)
		v1.Put("/ai-agents/:id", s.handleUpdateAgent)
		v1.Delete("/ai-agents/:id", s.handleDeleteAgent)
		v1.Post("/ai-agents/:id/execute", s.handleExecuteAgent)
		v1.Post("/ai-agents/:id/tools", s.handleAddToolToAgent)

		// Multi-runtime routes
		v1.Post("/runtime/execute", s.handleExecuteRuntime)

		// User routes
		v1.Get("/users/me", s.handleGetCurrentUser)
		v1.Put("/users/me", s.handleUpdateUser)
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	return s.app.Listen(addr)
}

// Handler functions (these will be implemented properly)
func (s *Server) handleLogin(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Login endpoint"})
}

func (s *Server) handleRegister(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Register endpoint"})
}

func (s *Server) handleCreateWorkflow(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Create workflow endpoint"})
}

func (s *Server) handleGetWorkflows(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get workflows endpoint"})
}

func (s *Server) handleGetWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Get workflow endpoint", "id": id})
}

func (s *Server) handleUpdateWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Update workflow endpoint", "id": id})
}

func (s *Server) handleDeleteWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Delete workflow endpoint", "id": id})
}

func (s *Server) handleRunWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Run workflow endpoint", "id": id})
}

func (s *Server) handleGetNodes(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get nodes endpoint"})
}

func (s *Server) handleGetNodeTypes(c *fiber.Ctx) error {
	// Return all registered node types
	nodeTypes := []string{
		"http_request",
		"delay",
		"function",
		"trigger",
		"data_process",
		"go_code",
		"javascript_code",
		"python_code",
		"java_code",
		"ruby_code",
		"php_code",
		"rust_code",
		"csharp_code",
		"shell_script",
		"ai_agent",
		"multi_runtime",
	}

	return c.JSON(fiber.Map{
		"node_types": nodeTypes,
	})
}

func (s *Server) handleCreateAgent(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Create AI agent endpoint"})
}

func (s *Server) handleGetAgents(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get AI agents endpoint"})
}

func (s *Server) handleGetAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Get AI agent endpoint", "id": id})
}

func (s *Server) handleUpdateAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Update AI agent endpoint", "id": id})
}

func (s *Server) handleDeleteAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Delete AI agent endpoint", "id": id})
}

func (s *Server) handleExecuteAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Execute AI agent endpoint", "id": id})
}

func (s *Server) handleAddToolToAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"message": "Add tool to AI agent endpoint", "id": id})
}

func (s *Server) handleExecuteRuntime(c *fiber.Ctx) error {
	var req struct {
		RuntimeType string                 `json:"runtime_type" validate:"required"`
		Code        string                 `json:"code" validate:"required"`
		Input       map[string]interface{} `json:"input"`
		Timeout     *int                   `json:"timeout"` // in seconds
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Use the provided timeout or default to 30 seconds
	timeoutDuration := 30 * time.Second
	if req.Timeout != nil {
		timeoutDuration = time.Duration(*req.Timeout) * time.Second
	}

	// Execute the code in the appropriate runtime
	runtimeType := runtimes.RuntimeType(req.RuntimeType)
	result, err := s.runtimeMgr.ExecuteCode(context.Background(), runtimeType, req.Code, req.Input, timeoutDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"result": result,
	})
}

func (s *Server) handleGetCurrentUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Get current user endpoint"})
}

func (s *Server) handleUpdateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Update user endpoint"})
}