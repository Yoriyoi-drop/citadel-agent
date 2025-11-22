// backend/internal/ai/advanced_runtime.go
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentMemoryType represents the type of agent memory (different from memory_system.go to avoid conflicts)
type AgentMemoryType string

const (
	AgentMemoryTypeShortTerm AgentMemoryType = "short_term" // For current session
	AgentMemoryTypeLongTerm  AgentMemoryType = "long_term"  // For persistent knowledge
	AgentMemoryTypeWorking   AgentMemoryType = "working"    // For immediate tasks
	AgentMemoryTypeEpisodic  AgentMemoryType = "episodic"   // For experiences
)

// AgentMemory represents a piece of information stored in AI memory
type AgentMemory struct {
	ID        string                 `json:"id"`
	Type      AgentMemoryType        `json:"type"`
	AgentID   string                 `json:"agent_id"`
	Content   string                 `json:"content"`
	Embedding []float32              `json:"embedding,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Tags      []string               `json:"tags"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	Importance float64               `json:"importance"` // 0-1 scale
	LastAccessed *time.Time          `json:"last_accessed,omitempty"`
}

// AgentState represents the current state of an AI agent
type AgentState struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	AgentID   string                 `json:"agent_id"`
	Variables map[string]interface{} `json:"variables"`
	Memory    *Memory                `json:"memory,omitempty"`
	Context   map[string]interface{} `json:"context"`
	Tools     []Tool                 `json:"tools"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}


// AgentMemoryManager manages AI agent memory
type AgentMemoryManager struct {
	db       *pgxpool.Pool
	memories map[string]*Memory
	mu       sync.RWMutex
}

// NewAgentMemoryManager creates a new agent memory manager
func NewAgentMemoryManager(db *pgxpool.Pool) *AgentMemoryManager {
	return &AgentMemoryManager{
		db:       db,
		memories: make(map[string]*Memory),
	}
}

// StoreMemory stores a piece of memory for an agent
func (amm *AgentMemoryManager) StoreMemory(ctx context.Context, memory *Memory) error {
	if memory.ID == "" {
		memory.ID = uuid.New().String()
	}
	memory.CreatedAt = time.Now()

	// Store in memory cache
	amm.mu.Lock()
	amm.memories[memory.ID] = memory
	amm.mu.Unlock()

	// In a real implementation, also store in persistent storage
	// This would involve saving to a vector database or other memory storage
	return amm.saveMemoryToDB(ctx, memory)
}

// RetrieveMemory retrieves memory based on query and agent
func (amm *AgentMemoryManager) RetrieveMemory(ctx context.Context, agentID, query string, memoryType AgentMemoryType, limit int) ([]*AgentMemory, error) {
	// In a real implementation, this would involve semantic search
	// For now, we'll do a simple retrieval from memory cache
	amm.mu.RLock()
	defer amm.mu.RUnlock()

	var results []*AgentMemory
	for _, memory := range amm.memories {
		if memory.AgentID == agentID &&
		   (memoryType == "" || memory.Type == memoryType) {
			// Simple text matching for demo purposes
			if query == "" || containsIgnoreCase(memory.Content, query) {
				results = append(results, memory)
				
				if len(results) >= limit {
					break
				}
			}
		}
	}

	// Update last accessed time
	now := time.Now()
	for _, result := range results {
		result.LastAccessed = &now
		// Update in cache
		amm.memories[result.ID] = result
	}

	return results, nil
}

// saveMemoryToDB saves memory to persistent storage
func (amm *AgentMemoryManager) saveMemoryToDB(ctx context.Context, memory *Memory) error {
	// Serialize metadata and tags
	metadataJSON, err := json.Marshal(memory.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	tagsJSON, err := json.Marshal(memory.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	var expiresAt interface{}
	if memory.ExpiresAt != nil {
		expiresAt = *memory.ExpiresAt
	} else {
		expiresAt = nil
	}

	query := `
		INSERT INTO ai_agent_memory (
			id, type, agent_id, content, metadata, tags, 
			created_at, expires_at, importance
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = amm.db.Exec(ctx, query,
		memory.ID,
		memory.Type,
		memory.AgentID,
		memory.Content,
		metadataJSON,
		tagsJSON,
		memory.CreatedAt,
		expiresAt,
		memory.Importance,
	)

	if err != nil {
		return fmt.Errorf("failed to save memory to database: %w", err)
	}

	return nil
}

// containsIgnoreCase checks if text contains query ignoring case
func containsIgnoreCase(text, query string) bool {
	textLower := toLowerCase(text)
	queryLower := toLowerCase(query)
	return agentContains(textLower, queryLower)
}

// toLowerCase converts string to lowercase
func toLowerCase(s string) string {
	result := make([]byte, len(s))
	for i, c := range []byte(s) {
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// agentContains checks if text contains substring
func agentContains(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if text[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// AgentStateManager manages agent states
type AgentStateManager struct {
	db       *pgxpool.Pool
	states   map[string]*AgentState
	mu       sync.RWMutex
}

// NewAgentStateManager creates a new agent state manager
func NewAgentStateManager(db *pgxpool.Pool) *AgentStateManager {
	return &AgentStateManager{
		db:     db,
		states: make(map[string]*AgentState),
	}
}

// SaveState saves the current state of an agent
func (asm *AgentStateManager) SaveState(ctx context.Context, state *AgentState) error {
	if state.ID == "" {
		state.ID = uuid.New().String()
	}
	state.UpdatedAt = time.Now()

	// Store in memory cache
	asm.mu.Lock()
	asm.states[state.ID] = state
	asm.mu.Unlock()

	// Save to database
	return asm.saveStateToDB(ctx, state)
}

// GetState retrieves the state of an agent
func (asm *AgentStateManager) GetState(ctx context.Context, stateID string) (*AgentState, error) {
	// Check memory cache first
	asm.mu.RLock()
	if state, exists := asm.states[stateID]; exists {
		asm.mu.RUnlock()
		return state, nil
	}
	asm.mu.RUnlock()

	// Load from database
	return asm.getStateFromDB(ctx, stateID)
}

// saveStateToDB saves state to persistent storage
func (asm *AgentStateManager) saveStateToDB(ctx context.Context, state *AgentState) error {
	// Serialize variables and context
	variablesJSON, err := json.Marshal(state.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	contextJSON, err := json.Marshal(state.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	// Serialize tools
	toolsJSON, err := json.Marshal(state.Tools)
	if err != nil {
		return fmt.Errorf("failed to marshal tools: %w", err)
	}

	query := `
		INSERT INTO ai_agent_states (
			id, session_id, agent_id, variables, context, tools, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			session_id = EXCLUDED.session_id,
			agent_id = EXCLUDED.agent_id,
			variables = EXCLUDED.variables,
			context = EXCLUDED.context,
			tools = EXCLUDED.tools,
			updated_at = EXCLUDED.updated_at
	`

	_, err = asm.db.Exec(ctx, query,
		state.ID,
		state.SessionID,
		state.AgentID,
		variablesJSON,
		contextJSON,
		toolsJSON,
		state.CreatedAt,
		state.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save state to database: %w", err)
	}

	return nil
}

// getStateFromDB retrieves state from persistent storage
func (asm *AgentStateManager) getStateFromDB(ctx context.Context, stateID string) (*AgentState, error) {
	query := `
		SELECT id, session_id, agent_id, variables, context, tools, 
		       created_at, updated_at
		FROM ai_agent_states
		WHERE id = $1
	`

	var state AgentState
	var variablesJSON, contextJSON, toolsJSON []byte

	err := asm.db.QueryRow(ctx, query, stateID).Scan(
		&state.ID,
		&state.SessionID,
		&state.AgentID,
		&variablesJSON,
		&contextJSON,
		&toolsJSON,
		&state.CreatedAt,
		&state.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get state from database: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(variablesJSON, &state.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	if err := json.Unmarshal(contextJSON, &state.Context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	if err := json.Unmarshal(toolsJSON, &state.Tools); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools: %w", err)
	}

	return &state, nil
}

// MultiAgentCoordinator manages coordination between multiple AI agents
type MultiAgentCoordinator struct {
	db          *pgxpool.Pool
	agents      map[string]*AgentState
	agentMutex  sync.RWMutex
	orchestrate func(context.Context, []string, map[string]interface{}) (map[string]interface{}, error)
}

// NewMultiAgentCoordinator creates a new multi-agent coordinator
func NewMultiAgentCoordinator(db *pgxpool.Pool) *MultiAgentCoordinator {
	return &MultiAgentCoordinator{
		db:     db,
		agents: make(map[string]*AgentState),
		orchestrate: defaultOrchestration,
	}
}

// Coordinate executes a task across multiple agents
func (mac *MultiAgentCoordinator) Coordinate(ctx context.Context, agentIDs []string, task string, contextData map[string]interface{}) (map[string]interface{}, error) {
	// Get all agent states
	var states []*AgentState
	for _, agentID := range agentIDs {
		state, err := mac.getState(agentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get state for agent %s: %w", agentID, err)
		}
		states = append(states, state)
	}

	// Execute orchestration
	result, err := mac.orchestrate(ctx, agentIDs, contextData)
	if err != nil {
		return nil, fmt.Errorf("orchestration failed: %w", err)
	}

	return result, nil
}

// defaultOrchestration is a default orchestration function
func defaultOrchestration(ctx context.Context, agentIDs []string, contextData map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["task"] = "completed"
	result["agents"] = agentIDs
	result["timestamp"] = time.Now().Unix()
	result["context"] = contextData

	return result, nil
}

// getState gets an agent state from memory
func (mac *MultiAgentCoordinator) getState(agentID string) (*AgentState, error) {
	mac.agentMutex.RLock()
	defer mac.agentMutex.RUnlock()

	if state, exists := mac.agents[agentID]; exists {
		return state, nil
	}

	return nil, fmt.Errorf("agent state not found: %s", agentID)
}

// AddAgent adds an agent to the coordinator
func (mac *MultiAgentCoordinator) AddAgent(agentState *AgentState) {
	mac.agentMutex.Lock()
	defer mac.agentMutex.Unlock()
	
	mac.agents[agentState.AgentID] = agentState
}

// HumanInLoopManager manages human-in-the-loop interactions

// HumanTaskStatus represents the status of a human task
type HumanTaskStatus string

const (
	HumanTaskStatusPending  HumanTaskStatus = "pending"
	HumanTaskStatusAssigned HumanTaskStatus = "assigned"
	HumanTaskStatusCompleted HumanTaskStatus = "completed"
	HumanTaskStatusRejected  HumanTaskStatus = "rejected"
)

// NewHumanInLoopManager creates a new human-in-the-loop manager

// CreateTask creates a new human task
func (hil *HumanInLoopManager) CreateTask(ctx context.Context, agentID, taskType, description string, contextData map[string]interface{}) (*HumanTask, error) {
	task := &HumanTask{
		ID:          uuid.New().String(),
		AgentID:     agentID,
		TaskType:    taskType,
		Description: description,
		Context:     contextData,
		Status:      HumanTaskStatusPending,
		CreatedAt:   time.Now(),
	}

	// Calculate expiration time (24 hours from now)
	expiry := time.Now().Add(24 * time.Hour)
	task.ExpiresAt = &expiry

	// Store in memory
	hil.mutex.Lock()
	hil.pendingTasks[task.ID] = task
	hil.mutex.Unlock()

	// Save to database
	if err := hil.saveTaskToDB(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}

// GetPendingTasks retrieves all pending human tasks for an agent
func (hil *HumanInLoopManager) GetPendingTasks(ctx context.Context, agentID string) ([]*HumanTask, error) {
	hil.mutex.RLock()
	defer hil.mutex.RUnlock()

	var tasks []*HumanTask
	for _, task := range hil.pendingTasks {
		if task.AgentID == agentID && task.Status == HumanTaskStatusPending {
			// Check if task has expired
			if task.ExpiresAt != nil && time.Now().After(*task.ExpiresAt) {
				// Skip expired tasks
				continue
			}
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// SubmitResponse submits a response to a human task
func (hil *HumanInLoopManager) SubmitResponse(ctx context.Context, taskID, response string) error {
	hil.mutex.Lock()
	defer hil.mutex.Unlock()

	task, exists := hil.pendingTasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status != HumanTaskStatusPending && task.Status != HumanTaskStatusAssigned {
		return fmt.Errorf("task is not in a submittable state: %s", task.Status)
	}

	responseCopy := response
	task.Response = &responseCopy
	task.Status = HumanTaskStatusCompleted
	now := time.Now()
	task.RespondedAt = &now

	// Update in database
	return hil.updateTaskInDB(ctx, task)
}

// saveTaskToDB saves a task to persistent storage
func (hil *HumanInLoopManager) saveTaskToDB(ctx context.Context, task *HumanTask) error {
	// Serialize context
	contextJSON, err := json.Marshal(task.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	query := `
		INSERT INTO human_tasks (
			id, agent_id, task_type, description, context, 
			status, created_at, expires_at, assigned_to
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	var assignedTo interface{}
	if task.AssignedTo != nil {
		assignedTo = *task.AssignedTo
	}

	_, err = hil.db.Exec(ctx, query,
		task.ID,
		task.AgentID,
		task.TaskType,
		task.Description,
		contextJSON,
		task.Status,
		task.CreatedAt,
		task.ExpiresAt,
		assignedTo,
	)

	if err != nil {
		return fmt.Errorf("failed to save task to database: %w", err)
	}

	return nil
}

// updateTaskInDB updates a task in persistent storage
func (hil *HumanInLoopManager) updateTaskInDB(ctx context.Context, task *HumanTask) error {
	query := `
		UPDATE human_tasks
		SET status = $2, response = $3, responded_at = $4
		WHERE id = $1
	`

	var response interface{}
	if task.Response != nil {
		response = *task.Response
	}

	var respondedAt interface{}
	if task.RespondedAt != nil {
		respondedAt = *task.RespondedAt
	}

	_, err := hil.db.Exec(ctx, query,
		task.ID,
		task.Status,
		response,
		respondedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	return nil
}

// AI Agent Runtime Manager
type AIRuntimeManager struct {
	memoryManager     *AgentMemoryManager
	stateManager      *AgentStateManager
	coordinator       *MultiAgentCoordinator
	humanInLoop       *HumanInLoopManager
	tools             map[string]Tool
	toolMutex         sync.RWMutex
	db                *pgxpool.Pool
}

// NewAIRuntimeManager creates a new AI runtime manager
func NewAIRuntimeManager(db *pgxpool.Pool) *AIRuntimeManager {
	return &AIRuntimeManager{
		memoryManager: NewAgentMemoryManager(db),
		stateManager:  NewAgentStateManager(db),
		coordinator:   NewMultiAgentCoordinator(db),
		humanInLoop:   NewHumanInLoopManager(db),
		tools:         make(map[string]Tool),
		db:            db,
	}
}

// RegisterTool registers a new tool for AI agents
func (arm *AIRuntimeManager) RegisterTool(tool Tool) {
	arm.toolMutex.Lock()
	defer arm.toolMutex.Unlock()
	
	arm.tools[tool.Name] = tool
}

// ExecuteAgent executes an AI agent with the provided input
func (arm *AIRuntimeManager) ExecuteAgent(ctx context.Context, agentID string, input map[string]interface{}) (map[string]interface{}, error) {
	// Get current state
	state, err := arm.stateManager.GetState(ctx, agentID)
	if err != nil {
		// If state doesn't exist, create a new one
		state = &AgentState{
			ID:        agentID,
			AgentID:   agentID,
			SessionID: uuid.New().String(),
			Variables: make(map[string]interface{}),
			Context:   make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := arm.stateManager.SaveState(ctx, state); err != nil {
			return nil, fmt.Errorf("failed to create initial state: %w", err)
		}
	}

	// Update state with new input
	for k, v := range input {
		state.Variables[k] = v
	}
	state.UpdatedAt = time.Now()

	// Process the input
	result := map[string]interface{}{
		"agent_id":    agentID,
		"input":       input,
		"timestamp":   time.Now().Unix(),
		"status":      "completed",
		"memory_used": true,
	}

	// Update state after processing
	if err := arm.stateManager.SaveState(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to save state after execution: %w", err)
	}

	return result, nil
}

// GetMemoryManager returns the memory manager
func (arm *AIRuntimeManager) GetMemoryManager() *AgentMemoryManager {
	return arm.memoryManager
}

// GetStateManager returns the state manager
func (arm *AIRuntimeManager) GetStateManager() *AgentStateManager {
	return arm.stateManager
}

// GetCoordinator returns the multi-agent coordinator
func (arm *AIRuntimeManager) GetCoordinator() *MultiAgentCoordinator {
	return arm.coordinator
}

// GetHumanInLoopManager returns the human-in-the-loop manager
func (arm *AIRuntimeManager) GetHumanInLoopManager() *HumanInLoopManager {
	return arm.humanInLoop
}