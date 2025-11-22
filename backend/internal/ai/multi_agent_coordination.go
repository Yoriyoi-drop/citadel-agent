// backend/internal/ai/multi_agent_coordination.go
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// AgentRole defines the role of an agent in a multi-agent system
type AgentRole string

const (
	AgentRoleManager    AgentRole = "manager"
	AgentRoleWorker     AgentRole = "worker"
	AgentRoleCritic     AgentRole = "critic"
	AgentRolePlanner    AgentRole = "planner"
	AgentRoleExecutor   AgentRole = "executor"
	AgentRoleObserver   AgentRole = "observer"
	AgentRoleCoordinator AgentRole = "coordinator"
	AgentRoleCommunicator AgentRole = "communicator"
)

// CoordinationProtocol defines how agents communicate
type CoordinationProtocol string

const (
	ProtocolDirect    CoordinationProtocol = "direct"
	ProtocolMessageQueue CoordinationProtocol = "message_queue"
	ProtocolEventStream CoordinationProtocol = "event_stream"
	ProtocolBroadcast CoordinationProtocol = "broadcast"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusAssigned  TaskStatus = "assigned"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// MultiAgentConfig represents configuration for multi-agent coordination
type MultiAgentConfig struct {
	CoordinationProtocol CoordinationProtocol `json:"coordination_protocol"`
	MaxAgents          int                  `json:"max_agents"`
	MaxTasks           int                  `json:"max_tasks"`
	CommunicationTimeout time.Duration        `json:"communication_timeout"`
	EnableLoadBalancing bool                `json:"enable_load_balancing"`
	EnableFaultTolerance bool               `json:"enable_fault_tolerance"`
	EnableLeaderElection bool               `json:"enable_leader_election"`
	TaskAssignmentStrategy string            `json:"task_assignment_strategy"` // "round_robin", "least_loaded", "specialized"
	MaxRetriesPerTask    int                 `json:"max_retries_per_task"`
	RetryDelay           time.Duration       `json:"retry_delay"`
}

// Task represents a task that can be assigned to agents
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Instructions string               `json:"instructions"`
	Input       map[string]interface{} `json:"input"`
	Priority    int                   `json:"priority"` // Higher number means higher priority
	Deadline    *time.Time            `json:"deadline,omitempty"`
	AssignedTo  *string               `json:"assigned_to,omitempty"`
	Status      TaskStatus            `json:"status"`
	Result      interface{}           `json:"result,omitempty"`
	Error       *string               `json:"error,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Dependencies []string             `json:"dependencies"`
	AgentRequirements map[string]interface{} `json:"agent_requirements"`
	TaskGroup   *string               `json:"task_group,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CoordinatedAgent represents an individual AI agent in the multi-agent system
type CoordinatedAgent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Role        AgentRole `json:"role"`
	Status      AgentStatus `json:"status"`
	Capabilities []string `json:"capabilities"`
	CurrentTask *string   `json:"current_task,omitempty"`
	LastSeen    time.Time `json:"last_seen"`
	Load        int       `json:"load"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusReady    AgentStatus = "ready"
	AgentStatusWorking  AgentStatus = "working"
	AgentStatusOffline  AgentStatus = "offline"
	AgentStatusError    AgentStatus = "error"
	AgentStatusDraining AgentStatus = "draining"
)

// Message represents a communication between agents
type Message struct {
	ID            string                 `json:"id"`
	From          string                 `json:"from"`
	To            string                 `json:"to"`
	Type          MessageType            `json:"type"`
	Content       interface{}            `json:"content"`
	TaskID        *string               `json:"task_id,omitempty"`
	CorrelationID *string               `json:"correlation_id,omitempty"`
	Timestamp     time.Time             `json:"timestamp"`
	Priority      int                   `json:"priority"`
	Timeout       *time.Duration        `json:"timeout,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeRequest      MessageType = "request"
	MessageTypeResponse     MessageType = "response"
	MessageTypeNotification MessageType = "notification"
	MessageTypeHeartbeat    MessageType = "heartbeat"
	MessageTypeError       MessageType = "error"
	MessageTypeTaskAssignment MessageType = "task_assignment"
	MessageTypeTaskResult   MessageType = "task_result"
	MessageTypeCoordination MessageType = "coordination"
)

// MultiAgentCoordinator manages coordination between multiple AI agents
type MultiAgentCoordinator struct {
	config      *MultiAgentConfig
	agents      map[string]*CoordinatedAgent
	tasks       map[string]*Task
	messages    chan *Message
	agentMutex  sync.RWMutex
	taskMutex   sync.RWMutex
	messageMutex sync.Mutex
	taskQueue   chan *Task
	workers     sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	taskAssigner TaskAssigner
	leaderAgent  *string
	leaderMutex  sync.RWMutex
}

// TaskAssigner interface for different task assignment strategies
type TaskAssigner interface {
	AssignTask(agents []*CoordinatedAgent, task *Task) (*CoordinatedAgent, error)
}

// NewMultiAgentCoordinator creates a new multi-agent coordinator
func NewMultiAgentCoordinator(config *MultiAgentConfig) *MultiAgentCoordinator {
	if config.CommunicationTimeout == 0 {
		config.CommunicationTimeout = 30 * time.Second
	}
	if config.MaxRetriesPerTask == 0 {
		config.MaxRetriesPerTask = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	coordinator := &MultiAgentCoordinator{
		config:      config,
		agents:      make(map[string]*CoordinatedAgent),
		tasks:       make(map[string]*Task),
		messages:    make(chan *Message, 1000), // Buffered channel for messages
		taskQueue:   make(chan *Task, 100),     // Buffered channel for tasks
		ctx:         ctx,
		cancel:      cancel,
		taskAssigner: NewTaskAssigner(config.TaskAssignmentStrategy),
	}

	// Start message processor
	go coordinator.processMessages()

	// Start task processor
	go coordinator.processTasks()

	// Start heartbeat monitor if fault tolerance is enabled
	if config.EnableFaultTolerance {
		go coordinator.monitorAgentHealth()
	}

	// Start leader election if enabled
	if config.EnableLeaderElection {
		go coordinator.electLeader()
	}

	return coordinator
}

// AddAgent adds an agent to the coordination system
func (mc *MultiAgentCoordinator) AddAgent(agent *CoordinatedAgent) error {
	mc.agentMutex.Lock()
	defer mc.agentMutex.Unlock()

	if len(mc.agents) >= mc.config.MaxAgents {
		return fmt.Errorf("maximum agent limit reached: %d", mc.config.MaxAgents)
	}

	if _, exists := mc.agents[agent.ID]; exists {
		return fmt.Errorf("agent with ID %s already exists", agent.ID)
	}

	agent.Status = AgentStatusReady
	agent.LastSeen = time.Now()
	
	mc.agents[agent.ID] = agent

	// Notify other agents about the new agent
	mc.broadcastMessage(&Message{
		Type:    MessageTypeNotification,
		Content: map[string]interface{}{
			"event": "agent_added",
			"agent": agent,
		},
	})

	return nil
}

// RemoveAgent removes an agent from the coordination system
func (mc *MultiAgentCoordinator) RemoveAgent(agentID string) error {
	mc.agentMutex.Lock()
	defer mc.agentMutex.Unlock()

	agent, exists := mc.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Reassign tasks assigned to this agent
	if agent.CurrentTask != nil {
		taskID := *agent.CurrentTask
		if task, exists := mc.tasks[taskID]; exists {
			// Mark task as unassigned
			task.Status = TaskStatusPending
			task.AssignedTo = nil
		}
	}

	// Remove agent
	delete(mc.agents, agentID)

	// Notify other agents about the removal
	mc.broadcastMessage(&Message{
		Type: MessageTypeNotification,
		Content: map[string]interface{}{
			"event": "agent_removed",
			"agent_id": agentID,
		},
	})

	return nil
}

// SubmitTask submits a new task for execution by agents
func (mc *MultiAgentCoordinator) SubmitTask(ctx context.Context, task *Task) error {
	if mc.config.MaxTasks > 0 && len(mc.tasks) >= mc.config.MaxTasks {
		return fmt.Errorf("maximum task limit reached: %d", mc.config.MaxTasks)
	}

	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	task.Status = TaskStatusPending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// Store the task
	mc.taskMutex.Lock()
	mc.tasks[task.ID] = task
	mc.taskMutex.Unlock()

	// Add to queue for processing
	select {
	case mc.taskQueue <- task:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// processTasks processes tasks from the queue
func (mc *MultiAgentCoordinator) processTasks() {
	for {
		select {
		case task := <-mc.taskQueue:
			mc.assignTaskToAgent(task)
		case <-mc.ctx.Done():
			return
		}
	}
}

// assignTaskToAgent assigns a task to an available agent
func (mc *MultiAgentCoordinator) assignTaskToAgent(task *Task) {
	// Find suitable agents
	availableAgents := mc.getAvailableAgentsForTask(task)

	if len(availableAgents) == 0 {
		// No agents available, keep task in pending status
		task.UpdatedAt = time.Now()
		return
	}

	// Use task assigner to select the best agent
	selectedAgent, err := mc.taskAssigner.AssignTask(availableAgents, task)
	if err != nil {
		// Failed to assign task, update status and return
		mc.taskMutex.Lock()
		if t, exists := mc.tasks[task.ID]; exists {
			t.Status = TaskStatusFailed
			t.Error = &err.Error()
			t.UpdatedAt = time.Now()
		}
		mc.taskMutex.Unlock()
		return
	}

	// Assign task to agent
	mc.taskMutex.Lock()
	if t, exists := mc.tasks[task.ID]; exists {
		t.Status = TaskStatusAssigned
		t.AssignedTo = &selectedAgent.ID
		t.UpdatedAt = time.Now()
	}
	mc.taskMutex.Unlock()

	mc.agentMutex.Lock()
	if agent, exists := mc.agents[selectedAgent.ID]; exists {
		agent.Status = AgentStatusWorking
		agent.CurrentTask = &task.ID
		agent.Load++
		agent.LastSeen = time.Now()
	}
	mc.agentMutex.Unlock()

	// Send task assignment message to agent
	message := &Message{
		Type:      MessageTypeTaskAssignment,
		To:        selectedAgent.ID,
		TaskID:    &task.ID,
		Content:   task,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"assignment_timestamp": time.Now().Unix(),
			"task_priority":        task.Priority,
		},
	}

	mc.sendMessage(message)
}

// getAvailableAgentsForTask returns agents that can handle the given task
func (mc *MultiAgentCoordinator) getAvailableAgentsForTask(task *Task) []*Agent {
	mc.agentMutex.RLock()
	defer mc.agentMutex.RUnlock()

	var availableAgents []*Agent

	for _, agent := range mc.agents {
		if agent.Status == AgentStatusReady {
			// Check if agent has required capabilities
			if mc.agentHasRequiredCapabilities(agent, task) {
				availableAgents = append(availableAgents, agent)
			}
		}
	}

	return availableAgents
}

// agentHasRequiredCapabilities checks if an agent has required capabilities for a task
func (mc *MultiAgentCoordinator) agentHasRequiredCapabilities(agent *Agent, task *Task) bool {
	// If no specific requirements, any agent can handle it
	if task.AgentRequirements == nil || len(task.AgentRequirements) == 0 {
		return true
	}

	// Check capabilities if specified in requirements
	if reqCaps, exists := task.AgentRequirements["capabilities"]; exists {
		if reqCapsList, ok := reqCaps.([]interface{}); ok {
			for _, reqCap := range reqCapsList {
				reqCapStr := fmt.Sprintf("%v", reqCap)
				found := false
				for _, cap := range agent.Capabilities {
					if cap == reqCapStr {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}

	return true
}

// processMessages processes incoming messages
func (mc *MultiAgentCoordinator) processMessages() {
	for {
		select {
		case message := <-mc.messages:
			mc.handleMessage(message)
		case <-mc.ctx.Done():
			return
		}
	}
}

// handleMessage processes a single message
func (mc *MultiAgentCoordinator) handleMessage(message *Message) {
	switch message.Type {
	case MessageTypeTaskResult:
		mc.handleTaskResult(message)
	case MessageTypeHeartbeat:
		mc.handleHeartbeat(message)
	case MessageTypeError:
		mc.handleErrorMessage(message)
	case MessageTypeCoordination:
		mc.handleCoordinationMessage(message)
	default:
		// For other message types, we might broadcast or handle differently
		mc.handleGenericMessage(message)
	}
}

// handleTaskResult handles task completion results
func (mc *MultiAgentCoordinator) handleTaskResult(message *Message) {
	if message.TaskID == nil {
		return // Invalid message
	}

	taskID := *message.TaskID

	mc.taskMutex.Lock()
	defer mc.taskMutex.Unlock()

	task, exists := mc.tasks[taskID]
	if !exists {
		return // Task no longer exists
	}

	// Update task status based on result
	resultMap, ok := message.Content.(map[string]interface{})
	if !ok {
		task.Status = TaskStatusFailed
		errorMsg := "invalid result format"
		task.Error = &errorMsg
	} else {
		task.Status = TaskStatusCompleted
		task.Result = resultMap["result"]
		task.Error = nil // Clear any previous error
		
		if errorMsg, exists := resultMap["error"]; exists {
			if errorMsgStr, ok := errorMsg.(string); ok {
				task.Status = TaskStatusFailed
				task.Error = &errorMsgStr
			}
		}
	}

	task.UpdatedAt = time.Now()

	// Update agent status
	if task.AssignedTo != nil {
		mc.agentMutex.Lock()
		if agent, exists := mc.agents[*task.AssignedTo]; exists {
			agent.Status = AgentStatusReady
			agent.CurrentTask = nil
			agent.Load--
			agent.LastSeen = time.Now()
		}
		mc.agentMutex.Unlock()
	}
}

// handleHeartbeat processes agent heartbeats
func (mc *MultiAgentCoordinator) handleHeartbeat(message *Message) {
	mc.agentMutex.Lock()
	defer mc.agentMutex.Unlock()

	// Update agent last seen time
	if agent, exists := mc.agents[message.From]; exists {
		agent.LastSeen = time.Now()
		// Update status based on load or other factors
		if agent.Status == AgentStatusOffline || agent.Status == AgentStatusError {
			agent.Status = AgentStatusReady
		}
	}
}

// handleErrorMessage processes error messages
func (mc *MultiAgentCoordinator) handleErrorMessage(message *Message) {
	// Log error and potentially retry or reassign task
	if taskID, exists := message.Metadata["task_id"]; exists {
		if taskIDStr, ok := taskID.(string); ok {
			mc.taskMutex.Lock()
			if task, exists := mc.tasks[taskIDStr]; exists {
				task.Status = TaskStatusFailed
				if errMsg, ok := message.Content.(string); ok {
					task.Error = &errMsg
				}
				task.UpdatedAt = time.Now()
			}
			mc.taskMutex.Unlock()
		}
	}
}

// handleCoordinationMessage handles coordination-related messages
func (mc *MultiAgentCoordinator) handleCoordinationMessage(message *Message) {
	// Process coordination messages between agents
	// This could include synchronization, resource sharing, etc.
	
	coordinationData, ok := message.Content.(map[string]interface{})
	if !ok {
		return
	}
	
	command, exists := coordinationData["command"]
	if !exists {
		return
	}
	
	commandStr, ok := command.(string)
	if !ok {
		return
	}
	
	switch commandStr {
	case "resource_request":
		mc.handleResourceRequest(message, coordinationData)
	case "resource_response":
		mc.handleResourceResponse(message, coordinationData)
	case "synchronization":
		mc.handleSynchronization(message, coordinationData)
	}
}

// handleResourceRequest handles resource sharing requests
func (mc *MultiAgentCoordinator) handleResourceRequest(message *Message, data map[string]interface{}) {
	// Implementation for resource sharing between agents
	// This would coordinate agents sharing resources like data, computation, etc.
}

// handleResourceResponse handles resource sharing responses
func (mc *MultiAgentCoordinator) handleResourceResponse(message *Message, data map[string]interface{}) {
	// Implementation for handling resource responses
}

// handleSynchronization handles synchronization requests
func (mc *MultiAgentCoordinator) handleSynchronization(message *Message, data map[string]interface{}) {
	// Implementation for agent synchronization
}

// handleGenericMessage handles other message types
func (mc *MultiAgentCoordinator) handleGenericMessage(message *Message) {
	// Depending on the protocol, might forward to specific agent or broadcast
	switch mc.config.CoordinationProtocol {
	case ProtocolMessageQueue, ProtocolEventStream:
		// Forward to specific agent if To is specified
		if message.To != "" {
			mc.forwardMessageToAgent(message)
		} else {
			// Broadcast if no specific recipient
			mc.broadcastMessage(message)
		}
	case ProtocolBroadcast:
		// Always broadcast
		mc.broadcastMessage(message)
	}
}

// forwardMessageToAgent sends a message to a specific agent
func (mc *MultiAgentCoordinator) forwardMessageToAgent(message *Message) {
	// In a real implementation, this would send to the specific agent
	// For now, we'll just add it back to the messages channel
	select {
	case mc.messages <- message:
	default:
		// If channel is full, we would need a different approach
		// In a real system, this might involve retries or persistent queues
	}
}

// broadcastMessage sends a message to all agents
func (mc *MultiAgentCoordinator) broadcastMessage(message *Message) {
	mc.agentMutex.RLock()
	defer mc.agentMutex.RUnlock()

	// Send to all agents (or all except sender depending on need)
	for agentID := range mc.agents {
		broadcastMsg := *message // Copy message
		broadcastMsg.To = agentID
		select {
		case mc.messages <- &broadcastMsg:
		default:
			// Handle case where message channel is full
		}
	}
}

// sendMessage adds a message to the message queue
func (mc *MultiAgentCoordinator) sendMessage(message *Message) {
	select {
	case mc.messages <- message:
	default:
		// Message channel is full
		// In a real system, we might want to handle this differently
	}
}

// getAgentByID returns an agent by ID
func (mc *MultiAgentCoordinator) getAgentByID(agentID string) (*Agent, error) {
	mc.agentMutex.RLock()
	defer mc.agentMutex.RUnlock()

	agent, exists := mc.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	return agent, nil
}

// getTaskByID returns a task by ID
func (mc *MultiAgentCoordinator) getTaskByID(taskID string) (*Task, error) {
	mc.taskMutex.RLock()
	defer mc.taskMutex.RUnlock()

	task, exists := mc.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return task, nil
}

// GetAllAgents returns all registered agents
func (mc *MultiAgentCoordinator) GetAllAgents() []*Agent {
	mc.agentMutex.RLock()
	defer mc.agentMutex.RUnlock()

	agents := make([]*Agent, 0, len(mc.agents))
	for _, agent := range mc.agents {
		// Create a copy to avoid race conditions
		agentCopy := *agent
		agents = append(agents, &agentCopy)
	}

	return agents
}

// GetAllTasks returns all registered tasks
func (mc *MultiAgentCoordinator) GetAllTasks() []*Task {
	mc.taskMutex.RLock()
	defer mc.taskMutex.RUnlock()

	tasks := make([]*Task, 0, len(mc.tasks))
	for _, task := range mc.tasks {
		// Create a copy to avoid race conditions
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
	}

	return tasks
}

// GetTaskStatus returns the status of a specific task
func (mc *MultiAgentCoordinator) GetTaskStatus(taskID string) (TaskStatus, error) {
	task, err := mc.getTaskByID(taskID)
	if err != nil {
		return "", err
	}

	return task.Status, nil
}

// CancelTask cancels a running task
func (mc *MultiAgentCoordinator) CancelTask(taskID string) error {
	mc.taskMutex.Lock()
	defer mc.taskMutex.Unlock()

	task, exists := mc.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	if task.Status == TaskStatusRunning || task.Status == TaskStatusAssigned {
		task.Status = TaskStatusCancelled
		task.UpdatedAt = time.Now()
		
		// Notify assigned agent to cancel task
		if task.AssignedTo != nil {
			cancelMessage := &Message{
				Type:      MessageTypeCoordination,
				To:        *task.AssignedTo,
				TaskID:    &taskID,
				Content:   map[string]interface{}{"command": "cancel_task"},
				Timestamp: time.Now(),
			}
			
			mc.sendMessage(cancelMessage)
		}
	}

	return nil
}

// monitorAgentHealth monitors agent health and detects offline agents
func (mc *MultiAgentCoordinator) monitorAgentHealth() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.checkAgentHealth()
		case <-mc.ctx.Done():
			return
		}
	}
}

// checkAgentHealth checks if agents are still responsive
func (mc *MultiAgentCoordinator) checkAgentHealth() {
	mc.agentMutex.Lock()
	defer mc.agentMutex.Unlock()

	now := time.Now()
	for agentID, agent := range mc.agents {
		// If last seen was more than 2 minutes ago, mark as offline
		if now.Sub(agent.LastSeen) > 2*time.Minute {
			agent.Status = AgentStatusOffline
			
			// Reassign any current task
			if agent.CurrentTask != nil {
				taskID := *agent.CurrentTask
				if task, exists := mc.tasks[taskID]; exists {
					task.Status = TaskStatusPending
					task.AssignedTo = nil
					task.UpdatedAt = time.Now()
				}
			}
		}
	}
}

// electLeader performs leader election among agents
func (mc *MultiAgentCoordinator) electLeader() {
	ticker := time.NewTicker(1 * time.Minute) // Check election every minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if mc.config.EnableLeaderElection {
				mc.performLeaderElection()
			}
		case <-mc.ctx.Done():
			return
		}
	}
}

// performLeaderElection selects a leader among available agents
func (mc *MultiAgentCoordinator) performLeaderElection() {
	mc.agentMutex.Lock()
	defer mc.agentMutex.Unlock()

	var candidates []*Agent
	for _, agent := range mc.agents {
		if agent.Status == AgentStatusReady {
			candidates = append(candidates, agent)
		}
	}

	if len(candidates) == 0 {
		// Clear leader if no candidates
		mc.leaderAgent = nil
		return
	}

	// Simple leader election: pick the agent with highest priority based on some metrics
	var leader *Agent
	highestScore := -1

	for _, agent := range candidates {
		score := mc.calculateLeaderScore(agent)
		if score > highestScore {
			highestScore = score
			leader = agent
		}
	}

	if leader != nil {
		leaderID := leader.ID
		mc.leaderAgent = &leaderID
		
		// Notify agents about new leader
		newLeaderMsg := &Message{
			Type: MessageTypeNotification,
			Content: map[string]interface{}{
				"event": "leader_elected",
				"leader_agent_id": leader.ID,
			},
			Timestamp: time.Now(),
		}
		
		mc.broadcastMessage(newLeaderMsg)
	}
}

// calculateLeaderScore calculates a score for leader election
func (mc *MultiAgentCoordinator) calculateLeaderScore(agent *Agent) int {
	score := 0
	
	// Base score on capabilities count
	score += len(agent.Capabilities) * 10
	
	// Bonus for certain roles
	switch agent.Role {
	case AgentRoleManager, AgentRoleCoordinator:
		score += 20
	case AgentRolePlanner:
		score += 15
	}
	
	// Adjust for current load (lower load = higher score)
	score -= agent.Load * 5
	
	// Bonus for last seen recency (more recently active = better)
	timeFactor := time.Since(agent.LastSeen).Minutes()
	if timeFactor < 10 { // Active in last 10 minutes
		score += 10
	} else if timeFactor < 30 { // Active in last 30 minutes
		score += 5
	}
	
	return score
}

// GetLeader returns the current leader agent
func (mc *MultiAgentCoordinator) GetLeader() (*Agent, error) {
	mc.leaderMutex.RLock()
	defer mc.leaderMutex.RUnlock()
	
	if mc.leaderAgent == nil {
		return nil, fmt.Errorf("no leader currently elected")
	}
	
	return mc.getAgentByID(*mc.leaderAgent)
}

// waitForTaskCompletion waits for a task to complete with timeout
func (mc *MultiAgentCoordinator) waitForTaskCompletion(ctx context.Context, taskID string, timeout time.Duration) (*Task, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := mc.GetTaskStatus(taskID)
			if err != nil {
				return nil, fmt.Errorf("error getting task status: %w", err)
			}
			
			if status == TaskStatusCompleted || status == TaskStatusFailed || status == TaskStatusCancelled {
				return mc.getTaskByID(taskID)
			}
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for task completion")
		}
	}
}

// Close shuts down the coordinator
func (mc *MultiAgentCoordinator) Close() {
	mc.cancel()
}

// RoundRobinAssigner implements round-robin task assignment
type RoundRobinAssigner struct {
	lastAssigned int
	mutex        sync.Mutex
}

func NewRoundRobinAssigner() *RoundRobinAssigner {
	return &RoundRobinAssigner{
		lastAssigned: -1,
	}
}

func (r *RoundRobinAssigner) AssignTask(agents []*Agent, task *Task) (*Agent, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if len(agents) == 0 {
		return nil, fmt.Errorf("no available agents")
	}
	
	// Move to next agent in rotation
	r.lastAssigned = (r.lastAssigned + 1) % len(agents)
	
	return agents[r.lastAssigned], nil
}

// LeastLoadedAssigner assigns tasks to least loaded agents
type LeastLoadedAssigner struct{}

func NewLeastLoadedAssigner() *LeastLoadedAssigner {
	return &LeastLoadedAssigner{}
}

func (l *LeastLoadedAssigner) AssignTask(agents []*Agent, task *Task) (*Agent, error) {
	if len(agents) == 0 {
		return nil, fmt.Errorf("no available agents")
	}
	
	var selectedAgent *Agent
	minLoad := int(^uint(0) >> 1) // Max int value
	
	for _, agent := range agents {
		if agent.Load < minLoad {
			minLoad = agent.Load
			selectedAgent = agent
		}
	}
	
	if selectedAgent == nil {
		return nil, fmt.Errorf("no agent found")
	}
	
	return selectedAgent, nil
}

// SpecializedAssigner assigns tasks based on agent capabilities
type SpecializedAssigner struct{}

func NewSpecializedAssigner() *SpecializedAssigner {
	return &SpecializedAssigner{}
}

func (s *SpecializedAssigner) AssignTask(agents []*Agent, task *Task) (*Agent, error) {
	if len(agents) == 0 {
		return nil, fmt.Errorf("no available agents")
	}
	
	// First, try to find an agent with the exact role needed
	for _, agent := range agents {
		if agent.Role == getRequiredRoleForTask(task) {
			return agent, nil
		}
	}
	
	// If no exact role match, find an agent with required capabilities
	for _, agent := range agents {
		if hasRequiredCapabilities(agent, task) {
			return agent, nil
		}
	}
	
	// If no specialized agent found, use least loaded
	return NewLeastLoadedAssigner().AssignTask(agents, task)
}

// getRequiredRoleForTask determines the required role for a task
func getRequiredRoleForTask(task *Task) AgentRole {
	switch task.Type {
	case "planning":
		return AgentRolePlanner
	case "execution":
		return AgentRoleExecutor
	case "review":
		return AgentRoleCritic
	case "coordination":
		return AgentRoleCoordinator
	default:
		return AgentRoleWorker
	}
}

// hasRequiredCapabilities checks if agent has required capabilities for task
func hasRequiredCapabilities(agent *Agent, task *Task) bool {
	if task.AgentRequirements == nil {
		return true
	}
	
	if reqCaps, exists := task.AgentRequirements["capabilities"]; exists {
		if reqCapsList, ok := reqCaps.([]interface{}); ok {
			for _, reqCap := range reqCapsList {
				reqStr := fmt.Sprintf("%v", reqCap)
				found := false
				for _, agentCap := range agent.Capabilities {
					if agentCap == reqStr {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}
	
	return true
}

// NewTaskAssigner creates a task assigner based on strategy
func NewTaskAssigner(strategy string) TaskAssigner {
	switch strategy {
	case "round_robin":
		return NewRoundRobinAssigner()
	case "least_loaded":
		return NewLeastLoadedAssigner()
	case "specialized":
		return NewSpecializedAssigner()
	default:
		return NewRoundRobinAssigner() // Default to round-robin
	}
}