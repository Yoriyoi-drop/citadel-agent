// backend/internal/ai/manager.go
package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AIManager manages AI agents within the engine
type AIManager struct {
	agents map[string]*Agent
	mutex  sync.RWMutex
}

// NewAIManager creates a new AI manager
func NewAIManager() *AIManager {
	return &AIManager{
		agents: make(map[string]*Agent),
	}
}

// RegisterAgent registers an AI agent
func (am *AIManager) RegisterAgent(agent *Agent) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}

	if agent.Name == "" {
		return errors.New("agent name is required")
	}

	// Initialize memory manager if not provided
	if agent.Memory == nil {
		agent.Memory = NewMemoryManager(agent.ID, 100) // Keep last 100 messages
	}

	am.agents[agent.ID] = agent
	return nil
}

// ExecuteAgent executes an AI agent with input 
func (am *AIManager) ExecuteAgent(ctx context.Context, agentID string, input map[string]interface{}) (interface{}, error) {
	am.mutex.RLock()
	agent, exists := am.agents[agentID]
	am.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	// Prepare input context and add to memory
	var inputStr string
	if inputParam, exists := input["input"]; exists {
		if inputStrVal, ok := inputParam.(string); ok {
			inputStr = inputStrVal
		} else {
			inputStr = fmt.Sprintf("%v", inputParam)
		}
	} else {
		inputStr = fmt.Sprintf("%v", input)
	}

	agent.Memory.AddMessage("user", inputStr, map[string]interface{}{
		"context":   input,
		"timestamp": time.Now().Unix(),
	})

	// In a real implementation, this would call an LLM API
	// For now, simulate the AI response
	response := fmt.Sprintf("AI Agent %s processed: %s", agent.Name, inputStr)

	// Add AI response to memory
	agent.Memory.AddMessage("assistant", response, map[string]interface{}{
		"timestamp": time.Now().Unix(),
	})

	result := map[string]interface{}{
		"agent_id":    agent.ID,
		"agent_name":  agent.Name,
		"input":       input,
		"response":    response,
		"timestamp":   time.Now().Unix(),
		"memory_size": len(agent.Memory.Messages),
	}

	return result, nil
}

// GetAgent returns an agent by ID
func (am *AIManager) GetAgent(agentID string) (*Agent, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agent, exists := am.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}
	return agent, nil
}

// ListAgents returns all registered agents
func (am *AIManager) ListAgents() []*Agent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	agents := make([]*Agent, 0, len(am.agents))
	for _, agent := range am.agents {
		agents = append(agents, agent)
	}
	return agents
}

// HasAgent checks if an agent exists
func (am *AIManager) HasAgent(agentID string) bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	_, exists := am.agents[agentID]
	return exists
}

// ExecuteAgentWithCtx executes an agent with context and cancellation support
func (am *AIManager) ExecuteAgentWithCtx(ctx context.Context, agentID string, input map[string]interface{}) (interface{}, error) {
	// Check if context is cancelled first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result, err := am.ExecuteAgent(ctx, agentID, input)
	if err != nil {
		return nil, err
	}

	// Check context again after execution
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return result, nil
	}
}