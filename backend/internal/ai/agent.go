// backend/internal/ai/agent.go
package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/citadel-agent/backend/internal/engine"
	"github.com/google/uuid"
)

// Agent represents an AI agent with memory and tools
type Agent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Model       string                 `json:"model"`
	SystemPrompt string               `json:"system_prompt"`
	Tools       []Tool                `json:"tools"`
	Memory      *MemoryManager        `json:"-"` // Don't serialize memory with agent
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// Tool represents an available tool for the AI agent
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Executor    ToolExecutor          `json:"-"`
}

// ToolExecutor defines the interface for executing tools
type ToolExecutor interface {
	Execute(params map[string]interface{}) (interface{}, error)
}

// MemoryItem represents a single memory entry
type MemoryItem struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// MemoryManager handles memory operations for AI agents
type MemoryManager struct {
	ID        string        `json:"id"`
	AgentID   string        `json:"agent_id"`
	Messages  []MemoryItem  `json:"messages"`
	SizeLimit int          `json:"size_limit"` // Max number of messages to keep
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(agentID string, sizeLimit int) *MemoryManager {
	return &MemoryManager{
		ID:        uuid.New().String(),
		AgentID:   agentID,
		SizeLimit: sizeLimit,
		Messages:  make([]MemoryItem, 0),
	}
}

// AddMessage adds a message to the agent's memory
func (mm *MemoryManager) AddMessage(role, content string, metadata map[string]interface{}) {
	item := MemoryItem{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	mm.Messages = append(mm.Messages, item)

	// Keep only the most recent messages based on size limit
	if len(mm.Messages) > mm.SizeLimit {
		mm.Messages = mm.Messages[len(mm.Messages)-mm.SizeLimit:]
	}
}

// GetRecentMessages returns the most recent messages
func (mm *MemoryManager) GetRecentMessages(count int) []MemoryItem {
	if count > len(mm.Messages) {
		count = len(mm.Messages)
	}

	start := len(mm.Messages) - count
	if start < 0 {
		start = 0
	}

	return mm.Messages[start:]
}

// Clear clears all memory
func (mm *MemoryManager) Clear() {
	mm.Messages = make([]MemoryItem, 0)
}

// AIAgentNode represents a node that runs an AI agent
type AIAgentNode struct {
	AgentID    string                 `json:"agent_id"`
	Input      string                 `json:"input"`
	Parameters map[string]interface{} `json:"parameters"`
}

// Execute executes the AI agent node
func (ai *AIAgentNode) Execute(ctx context.Context, input map[string]interface{}) (*engine.ExecutionResult, error) {
	agentID, exists := input["agent_id"].(string)
	if !exists {
		return &engine.ExecutionResult{
			Status: "error",
			Error:  "agent_id is required for AI agent node",
			Timestamp: time.Now(),
		}, nil
	}

	agentInput, exists := input["input"].(string)
	if !exists {
		// If input is not a string, convert it to string
		agentInput = fmt.Sprintf("%v", input)
	}

	// In a real implementation, this would use an actual AI service
	// For now, simulate the AI agent response
	result := map[string]interface{}{
		"agent_id": agentID,
		"input":    agentInput,
		"response": fmt.Sprintf("AI Agent processed: %s", agentInput),
		"timestamp": time.Now().Unix(),
	}

	return &engine.ExecutionResult{
		Status: "success",
		Data:   result,
		Timestamp: time.Now(),
	}, nil
}