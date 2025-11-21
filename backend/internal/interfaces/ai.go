// backend/internal/interfaces/ai.go
package interfaces

import "context"

// AIManagerInterface defines the interface for AI operations that engine can use
// This breaks the circular dependency between engine and ai packages
type AIManagerInterface interface {
	ExecuteAgent(ctx context.Context, agentID string, input map[string]interface{}) (interface{}, error)
	RegisterAgent(agent interface{}) error
	GetAgent(agentID string) (interface{}, error)
	HasAgent(agentID string) bool
	ExecuteAgentWithCtx(ctx context.Context, agentID string, input map[string]interface{}) (interface{}, error)
}