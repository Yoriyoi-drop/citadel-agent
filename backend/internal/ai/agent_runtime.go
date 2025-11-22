// backend/internal/ai/agent_runtime.go
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/citadel-agent/backend/internal/interfaces"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

// AIModelProvider represents the AI model provider
type AIModelProvider string

const (
	ProviderOpenAI     AIModelProvider = "openai"
	ProviderAnthropic  AIModelProvider = "anthropic"
	ProviderHuggingFace AIModelProvider = "huggingface"
	ProviderLocal      AIModelProvider = "local"
)

// AIAgentType represents the type of AI agent
type AIAgentType string

const (
	AgentTypeSimple      AIAgentType = "simple"
	AgentTypeConversational AIAgentType = "conversational"
	AgentTypeToolUsing      AIAgentType = "tool_using"
	AgentTypeMultiModal     AIAgentType = "multimodal"
	AgentTypeReAct          AIAgentType = "react" // Reason + Act
	AgentTypePlanAndExecute AIAgentType = "plan_and_execute"
)

// AIAgentConfig represents the configuration for an AI agent
type AIAgentConfig struct {
	Provider         AIModelProvider    `json:"provider"`
	ModelName        string             `json:"model_name"`
	APIKey          string             `json:"api_key"`
	Temperature     float64            `json:"temperature"`
	MaxTokens       int                `json:"max_tokens"`
	EnableMemory    bool               `json:"enable_memory"`
	EnableTools     bool               `json:"enable_tools"`
	EnableVision    bool               `json:"enable_vision"`
	EnableAudio     bool               `json:"enable_audio"`
	SystemPrompt    string             `json:"system_prompt"`
	ExecutionTimeout time.Duration      `json:"execution_timeout"`
	MaxRetries      int               `json:"max_retries"`
	EnableReasoning bool               `json:"enable_reasoning"`
	EnableReflection bool              `json:"enable_reflection"`
	HumanInLoop     bool               `json:"human_in_loop"`
	AgentType       AIAgentType        `json:"agent_type"`
	MemoryConfig    *AgentMemoryConfig      `json:"memory_config"`
	ToolConfig      *ToolConfig        `json:"tool_config"`
	ThreadingConfig *ThreadingConfig   `json:"threading_config"`
}

// AgentMemoryConfig represents memory system configuration for agent runtime
type AgentMemoryConfig struct {
	EnableShortTerm  bool          `json:"enable_short_term"`
	EnableLongTerm   bool          `json:"enable_long_term"`
	ShortTermLimit   int           `json:"short_term_limit"`  // Max conversations to keep in ST memory
	LongTermStorage  string        `json:"long_term_storage"` // "vector_db", "postgres", "redis", etc.
	ContextWindowSize int          `json:"context_window_size"`
	Summarization    bool          `json:"enable_summarization"`
	Persistence      bool          `json:"enable_persistence"`
	CollectionName   string        `json:"collection_name"`
	EmbeddingModel   string        `json:"embedding_model"`
	MaxHistoryLength int           `json:"max_history_length"`
	Compression      bool          `json:"enable_compression"`
	CompressionRatio float64       `json:"compression_ratio"`
}

// ToolConfig represents tool configuration
type ToolConfig struct {
	AvailableTools   []AIAgentTool     `json:"available_tools"`
	EnableDynamicTools bool            `json:"enable_dynamic_tools"`
	MaxToolCalls     int               `json:"max_tool_calls"`
	ToolCallTimeout  time.Duration     `json:"tool_call_timeout"`
	ToolExecutionConcurrency int        `json:"tool_execution_concurrency"`
}

// AIAgentTool represents a tool available to an AI agent
type AIAgentTool struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Parameters   map[string]ToolParameter `json:"parameters"`
	Implementation func(context.Context, map[string]interface{}) (interface{}, error) `json:"-"`
}

// ToolParameter represents a parameter for an AI agent tool
type ToolParameter struct {
	Type        string `json:"type"`        // "string", "number", "boolean", "object", "array"
	Description string `json:"description"`
	Required    bool   `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Enum        []interface{} `json:"enum,omitempty"`
}

// ThreadingConfig represents multi-agent threading configuration
type ThreadingConfig struct {
	EnableThreading   bool              `json:"enable_threading"`
	MaxThreads        int               `json:"max_threads"`
	EnableCoordination bool             `json:"enable_coordination"`
	OrchestrationModel string           `json:"orchestration_model"`
	ThreadTimeout     time.Duration     `json:"thread_timeout"`
	CommunicationProtocol string        `json:"communication_protocol"` // "message_queue", "direct_call", "event_stream"
}

// AIAgentRuntime manages AI agent execution
type AIAgentRuntime struct {
	config          *AIAgentConfig
	llm             llms.Model
	agentMemory     *AIAgentMemory
	availableTools  map[string]AIAgentTool
	threadManager   *ThreadManager
	coordinationMgr *CoordinationManager
	humanInLoopMgr  *HumanInteractionManager
	modelCache      map[string]llms.Model
	mutex           sync.RWMutex
}

// AIAgentMemory manages agent memory systems
type AIAgentMemory struct {
	shortTerm *ShortTermMemory
	longTerm  *LongTermMemory
	storage   MemoryStorage
	config    *AgentMemoryConfig
}

// ShortTermMemory stores recent conversation history
type ShortTermMemory struct {
	history     []schema.ChatMessage
	maxHistory  int
	summary     string
	needsUpdate bool
}

// LongTermMemory stores persistent memories
type LongTermMemory struct {
	memories    map[string]*MemoryEntry
	storage     MemoryStorage
	embeddingFn func(string) ([]float32, error)
}

// AgentMemoryEntry represents a single memory entry for agent runtime
type AgentMemoryEntry struct {
	ID          string     `json:"id"`
	Content     string     `json:"content"`
	Source      string     `json:"source"`
	Timestamp   time.Time  `json:"timestamp"`
	Importance  float64    `json:"importance"` // 0.0-1.0, higher is more important
	Tags        []string   `json:"tags"`
	Embedding   []float32  `json:"embedding,omitempty"`
	LastAccessed time.Time `json:"last_accessed"`
	AccessCount int       `json:"access_count"`
}

// AgentMemoryStorage interface for persistent memory storage for agent runtime
type AgentMemoryStorage interface {
	Save(ctx context.Context, memory *AgentMemoryEntry) error
	Retrieve(ctx context.Context, id string) (*AgentMemoryEntry, error)
	Search(ctx context.Context, query string, limit int) ([]*AgentMemoryEntry, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*AgentMemoryEntry, error)
	Close() error
}

// ThreadManager manages multi-agent threads
type ThreadManager struct {
	threads   map[string]*AgentThread
	maxThreads int
	mutex     sync.RWMutex
}

// AgentThread represents a single agent thread
type AgentThread struct {
	ID        string
	Agents    []AIAgentInstance
	State     AgentThreadState
	StartTime time.Time
	EndTime   *time.Time
	Results   map[string]interface{}
	Error     error
	CancelFunc context.CancelFunc
}

// AgentThreadState represents the state of an agent thread
type AgentThreadState string

const (
	ThreadStateReady    AgentThreadState = "ready"
	ThreadStateRunning  AgentThreadState = "running"
	ThreadStatePaused   AgentThreadState = "paused"
	ThreadStateComplete AgentThreadState = "complete"
	ThreadStateError    AgentThreadState = "error"
	ThreadStateCancelled AgentThreadState = "cancelled"
)

// CoordinationManager manages multi-agent coordination
type CoordinationManager struct {
	protocol     string
	communication chan *CoordinationMessage
	agents       map[string]*AIAgentInstance
	mutex        sync.RWMutex
}

// CoordinationMessage represents a message between agents
type CoordinationMessage struct {
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Type        CoordinationMessageType `json:"type"`
	Content     interface{}            `json:"content"`
	Timestamp   time.Time              `json:"timestamp"`
	CorrelationID string              `json:"correlation_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CoordinationMessageType represents the type of coordination message
type CoordinationMessageType string

const (
	CoordinationMessageRequest  CoordinationMessageType = "request"
	CoordinationMessageResponse CoordinationMessageType = "response"
	CoordinationMessageNotice   CoordinationMessageType = "notice"
	CoordinationMessageError    CoordinationMessageType = "error"
)

// HumanInteractionManager manages human interaction in AI workflows
type HumanInteractionManager struct {
	enabled bool
	requests map[string]*HumanInteractionRequest
	mutex   sync.RWMutex
	callback func(string, string) error
}

// HumanInteractionRequest represents a request for human input
type HumanInteractionRequest struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agent_id"`
	RequestType string    `json:"request_type"` // "approval", "information", "confirmation", "correction"
	Message     string    `json:"message"`
	Context     interface{} `json:"context"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      HumanInteractionStatus `json:"status"`
	Response    *string   `json:"response,omitempty"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// HumanInteractionStatus represents the status of human interaction
type HumanInteractionStatus string

const (
	HumanInteractionPending   HumanInteractionStatus = "pending"
	HumanInteractionApproved  HumanInteractionStatus = "approved"
	HumanInteractionRejected  HumanInteractionStatus = "rejected"
	HumanInteractionTimeout   HumanInteractionStatus = "timeout"
	HumanInteractionCancelled HumanInteractionStatus = "cancelled"
)

// NewAIAgentRuntime creates a new AI agent runtime
func NewAIAgentRuntime(config *AIAgentConfig) (*AIAgentRuntime, error) {
	if config.ExecutionTimeout == 0 {
		config.ExecutionTimeout = 300 * time.Second // 5 minutes default
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1024
	}

	// Initialize LLM based on provider
	llm, err := initLLM(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM: %w", err)
	}

	// Initialize memory system
	agentMemory := &AIAgentMemory{
		shortTerm: &ShortTermMemory{
			history:    make([]schema.ChatMessage, 0),
			maxHistory: config.MemoryConfig.ShortTermLimit,
		},
		config: config.MemoryConfig,
	}

	if config.MemoryConfig.EnableLongTerm {
		agentMemory.longTerm = &LongTermMemory{
			memories: make(map[string]*MemoryEntry),
			// Initialize with appropriate storage based on config
		}
	}

	// Initialize coordination manager if threading is enabled
	var coordMgr *CoordinationManager
	if config.ThreadingConfig.EnableCoordination {
		coordMgr = &CoordinationManager{
			protocol:     config.ThreadingConfig.CommunicationProtocol,
			communication: make(chan *CoordinationMessage, 100),
			agents:       make(map[string]*AIAgentInstance),
		}
	}

	// Initialize human-in-loop manager
	humanInLoopMgr := &HumanInteractionManager{
		enabled: config.HumanInLoop,
		requests: make(map[string]*HumanInteractionRequest),
	}

	// Start coordination message processor if enabled
	if coordMgr != nil {
		go coordMgr.processMessages()
	}

	runtime := &AIAgentRuntime{
		config:          config,
		llm:             llm,
		agentMemory:     agentMemory,
		availableTools:  make(map[string]AIAgentTool),
		threadManager:   NewThreadManager(config.ThreadingConfig.MaxThreads),
		coordinationMgr: coordMgr,
		humanInLoopMgr:  humanInLoopMgr,
		modelCache:      make(map[string]llms.Model),
	}

	// Register default tools if enabled
	if config.EnableTools {
		runtime.registerDefaultTools()
	}

	return runtime, nil
}

// initLLM initializes the appropriate LLM based on provider
func initLLM(config *AIAgentConfig) (llms.Model, error) {
	switch config.Provider {
	case ProviderOpenAI:
		return initOpenAI(config)
	case ProviderAnthropic:
		return initAnthropic(config)
	case ProviderHuggingFace:
		return initHuggingFace(config)
	case ProviderLocal:
		return initLocalModel(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// AIAgentInstance represents a single AI agent instance
type AIAgentInstance struct {
	runtime    *AIAgentRuntime
	config     *AIAgentConfig
	agentType  AIAgentType
	memory     *AIAgentMemory
	tools      []AIAgentTool
	state      map[string]interface{} // Agent-specific state
	mutex      sync.RWMutex
}

// NewAIAgentInstance creates a new AI agent instance
func (runtime *AIAgentRuntime) NewAIAgentInstance(config *AIAgentConfig) *AIAgentInstance {
	agent := &AIAgentInstance{
		runtime:   runtime,
		config:    config,
		agentType: config.AgentType,
		memory:    runtime.agentMemory,
		tools:     make([]AIAgentTool, 0),
		state:     make(map[string]interface{}),
	}

	// Add available tools to this agent instance
	for _, tool := range runtime.availableTools {
		agent.tools = append(agent.tools, tool)
	}

	return agent
}

// Execute executes the AI agent with the given input
func (agent *AIAgentInstance) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Apply execution timeout
	ctx, cancel := context.WithTimeout(ctx, agent.config.ExecutionTimeout)
	defer cancel()

	var result map[string]interface{}
	var err error
	
	for attempt := 0; attempt <= agent.config.MaxRetries; attempt++ {
		result, err = agent.executeInternal(ctx, input)
		if err == nil {
			break // Success
		}
		
		if attempt < agent.config.MaxRetries {
			// Wait before retry with exponential backoff
			waitTime := time.Duration(attempt+1) * 2 * time.Second
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("AI agent execution failed after %d attempts: %w", agent.config.MaxRetries+1, err)
	}

	return result, nil
}

// executeInternal executes the internal AI logic
func (agent *AIAgentInstance) executeInternal(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Add input to memory if memory is enabled
	if agent.config.EnableMemory {
		agent.addToMemory(ctx, "input", input)
	}

	var response string
	var toolCalls []llms.ToolCall

	switch agent.agentType {
	case AgentTypeSimple:
		response = agent.executeSimple(ctx, input)
	case AgentTypeConversational:
		response = agent.executeConversational(ctx, input)
	case AgentTypeToolUsing:
		response, toolCalls = agent.executeToolUsing(ctx, input)
	case AgentTypeMultiModal:
		response = agent.executeMultiModal(ctx, input)
	case AgentTypeReAct:
		response = agent.executeReAct(ctx, input)
	case AgentTypePlanAndExecute:
		response = agent.executePlanAndExecute(ctx, input)
	default:
		response = agent.executeSimple(ctx, input)
	}

	// Process tool calls if any
	if len(toolCalls) > 0 {
		toolResults, err := agent.executeToolCalls(ctx, toolCalls)
		if err != nil {
			return nil, fmt.Errorf("failed to execute tool calls: %w", err)
		}

		// If we have tool results, we might need to run the LLM again
		if agent.config.EnableReasoning {
			response = agent.refineResponse(ctx, input, response, toolResults)
		}
	}

	// Add response to memory if memory is enabled
	if agent.config.EnableMemory {
		agent.addToMemory(ctx, "output", response)
	}

	result := map[string]interface{}{
		"success":      true,
		"response":     response,
		"agent_type":   string(agent.agentType),
		"input":        input,
		"timestamp":    time.Now().Unix(),
		"execution_id": uuid.New().String(),
	}

	if len(toolCalls) > 0 {
		result["tool_calls"] = toolCalls
	}

	if agent.config.EnableMemory {
		result["memory_updated"] = true
	}

	return result, nil
}

// executeSimple executes a simple AI agent
func (agent *AIAgentInstance) executeSimple(ctx context.Context, input map[string]interface{}) string {
	// Prepare the prompt
	prompt := agent.prepareSimplePrompt(input)

	// Generate response with LLM
	completion, err := llms.GenerateFromSinglePrompt(ctx, agent.runtime.llm, prompt, 
		llms.WithTemperature(agent.config.Temperature),
		llms.WithMaxTokens(agent.config.MaxTokens),
	)
	if err != nil {
		return fmt.Sprintf("Error generating response: %v", err)
	}

	return completion
}

// executeConversational executes a conversational AI agent
func (agent *AIAgentInstance) executeConversational(ctx context.Context, input map[string]interface{}) string {
	// Get conversation history if memory is enabled
	var history []schema.ChatMessage
	if agent.config.EnableMemory {
		history = agent.getConversationHistory()
	}

	// Prepare the messages for conversational model
	messages := make([]schema.ChatMessage, 0)
	
	if agent.config.SystemPrompt != "" {
		messages = append(messages, schema.SystemMessage{Content: agent.config.SystemPrompt})
	}
	
	// Add history if available
	messages = append(messages, history...)
	
	// Add the current user message
	inputStr := fmt.Sprintf("%v", input)
	messages = append(messages, schema.HumanMessage{Content: inputStr})

	// Generate response with chat model
	completion, err := agent.runtime.llm.GenerateContent(ctx, messages,
		llms.WithTemperature(agent.config.Temperature),
		llms.WithMaxTokens(agent.config.MaxTokens),
	)
	if err != nil {
		return fmt.Sprintf("Error generating conversational response: %v", err)
	}

	if len(completion.Choices) > 0 {
		return completion.Choices[0].Content
	}

	return "No response generated"
}

// executeToolUsing executes a tool-using AI agent
func (agent *AIAgentInstance) executeToolUsing(ctx context.Context, input map[string]interface{}) (string, []llms.ToolCall) {
	// Prepare the messages
	messages := make([]schema.ChatMessage, 0)
	
	if agent.config.SystemPrompt != "" {
		messages = append(messages, schema.SystemMessage{Content: agent.config.SystemPrompt})
	}
	
	inputStr := fmt.Sprintf("%v", input)
	messages = append(messages, schema.HumanMessage{Content: inputStr})

	// Prepare tools for the LLM
	tools := agent.prepareToolsForLLM()

	// Generate response with tool calling capability
	completion, err := agent.runtime.llm.GenerateContent(ctx, messages,
		llms.WithTemperature(agent.config.Temperature),
		llms.WithMaxTokens(agent.config.MaxTokens),
		llms.WithTools(tools),
	)
	if err != nil {
		return fmt.Sprintf("Error generating response with tools: %v", err), nil
	}

	var toolCalls []llms.ToolCall
	response := ""

	if len(completion.Choices) > 0 {
		choice := completion.Choices[0]
		response = choice.Content
		
		if choice.ToolCalls != nil {
			toolCalls = choice.ToolCalls
		}
	}

	return response, toolCalls
}

// executeMultiModal executes a multi-modal AI agent
func (agent *AIAgentInstance) executeMultiModal(ctx context.Context, input map[string]interface{}) string {
	// Extract text and media from input
	text := ""
	images := []string{}
	audio := []string{}
	video := []string{}

	if textVal, exists := input["text"]; exists {
		if textStr, ok := textVal.(string); ok {
			text = textStr
		}
	}

	if imagesVal, exists := input["images"]; exists {
		if imagesSlice, ok := imagesVal.([]interface{}); ok {
			for _, img := range imagesSlice {
				if imgStr, ok := img.(string); ok {
					images = append(images, imgStr)
				}
			}
		}
	}

	if audioVal, exists := input["audio"]; exists {
		if audioSlice, ok := audioVal.([]interface{}); ok {
			for _, aud := range audioSlice {
				if audStr, ok := aud.(string); ok {
					audio = append(audio, audStr)
				}
			}
		}
	}

	if videoVal, exists := input["video"]; exists {
		if videoSlice, ok := videoVal.([]interface{}); ok {
			for _, vid := range videoSlice {
				if vidStr, ok := vid.(string); ok {
					video = append(video, vidStr)
				}
			}
		}
	}

	// For now, we'll process just the text part
	// In a real implementation, we would handle multi-modal inputs

	parts := make([]llms.ContentPart, 0)
	
	// Add text part
	if text != "" {
		parts = append(parts, llms.TextPart(text))
	}

	// Add image parts if supported
	for _, img := range images {
		if agent.config.EnableVision {
			parts = append(parts, llms.ImageURLPart(img))
		}
	}

	// Create content
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	// Generate multi-modal response
	completion, err := agent.runtime.llm.GenerateContent(ctx, content,
		llms.WithTemperature(agent.config.Temperature),
		llms.WithMaxTokens(agent.config.MaxTokens),
	)
	if err != nil {
		return fmt.Sprintf("Error generating multi-modal response: %v", err)
	}

	if len(completion.Choices) > 0 {
		return completion.Choices[0].Content
	}

	return "No multi-modal response generated"
}

// executeReAct executes a ReAct (Reason + Act) AI agent
func (agent *AIAgentInstance) executeReAct(ctx context.Context, input map[string]interface{}) string {
	// ReAct agents alternate between reasoning and acting
	textInput := fmt.Sprintf("%v", input)
	
	// Iteratively reason and act
	currentThought := textInput
	maxIterations := 5 // Prevent infinite loops
	
	for i := 0; i < maxIterations; i++ {
		// First, reason about the current state
		reasonPrompt := fmt.Sprintf("Reason about: %s. What should be the next step?", currentThought)
		reasonResponse, err := llms.GenerateFromSinglePrompt(ctx, agent.runtime.llm, reasonPrompt,
			llms.WithTemperature(agent.config.Temperature),
			llms.WithMaxTokens(agent.config.MaxTokens),
		)
		if err != nil {
			return fmt.Sprintf("Error in reasoning: %v", err)
		}

		// Then, act based on the reasoning
		actPrompt := fmt.Sprintf("Based on your reasoning: '%s', what should be your action?", reasonResponse)
		actResponse, toolCalls := agent.executeToolUsing(ctx, map[string]interface{}{
			"text": actPrompt,
		})

		// If there are tool calls, execute them and update the thought
		if len(toolCalls) > 0 {
			toolResults, err := agent.executeToolCalls(ctx, toolCalls)
			if err != nil {
				return fmt.Sprintf("Error executing action: %v", err)
			}

			// Update thought based on tool results
			currentThought = fmt.Sprintf("Previous thought: %s\nAction taken: %s\nResults: %v", 
				currentThought, actResponse, toolResults)
		} else {
			// If no more actions, return the final response
			return fmt.Sprintf("Final response after ReAct cycle:\nReasoning: %s\nAction: %s", 
				reasonResponse, actResponse)
		}
	}

	return currentThought
}

// executePlanAndExecute executes a plan-and-execute AI agent
func (agent *AIAgentInstance) executePlanAndExecute(ctx context.Context, input map[string]interface{}) string {
	textInput := fmt.Sprintf("%v", input)

	// First, create a plan
	planPrompt := fmt.Sprintf("Create a detailed execution plan for: %s. Return the plan as a numbered list of steps.", textInput)
	planResponse, err := llms.GenerateFromSinglePrompt(ctx, agent.runtime.llm, planPrompt,
		llms.WithTemperature(agent.config.Temperature),
		llms.WithMaxTokens(agent.config.MaxTokens),
	)
	if err != nil {
		return fmt.Sprintf("Error creating plan: %v", err)
	}

	// Parse the plan (simplified - in reality would use more sophisticated parsing)
	steps := parseSteps(planResponse)

	// Execute each step
	executionResults := make([]map[string]interface{}, 0)
	
	for stepNum, step := range steps {
		// Check if step requires human approval
		if agent.requiresHumanApproval(step) && agent.config.HumanInLoop {
			approval, err := agent.requestHumanApproval(ctx, step)
			if err != nil {
				return fmt.Sprintf("Error requesting human approval: %v", err)
			}
			if !approval {
				return fmt.Sprintf("Step %d was rejected by human", stepNum+1)
			}
		}

		// Execute the step
		stepResult, err := agent.executeStep(ctx, step)
		if err != nil {
			return fmt.Sprintf("Error executing step %d: %v", stepNum+1, err)
		}

		executionResults = append(executionResults, map[string]interface{}{
			"step":   stepNum + 1,
			"action": step,
			"result": stepResult,
			"status": "completed",
		})
	}

	finalResult := fmt.Sprintf("Plan completed with %d steps. Results: %v", len(steps), executionResults)
	return finalResult
}

// requiresHumanApproval checks if a step requires human approval
func (agent *AIAgentInstance) requiresHumanApproval(step string) bool {
	// Define patterns that require human approval
	highRiskPatterns := []string{
		"delete", "remove", "destroy", "shutdown", "terminate", 
		"transfer", "pay", "send money", "execute payment",
		"modify permissions", "grant access", "firewall", "security",
		"administrative", "sudo", "root", "privileged",
	}
	
	stepLower := strings.ToLower(step)
	for _, pattern := range highRiskPatterns {
		if strings.Contains(stepLower, pattern) {
			return true
		}
	}
	
	return false
}

// requestHumanApproval requests approval for a step
func (agent *AIAgentInstance) requestHumanApproval(ctx context.Context, step string) (bool, error) {
	if !agent.config.HumanInLoop {
		// If human-in-the-loop is not enabled, auto-approve
		return true, nil
	}

	// Create a human interaction request
	request := &HumanInteractionRequest{
		ID:          uuid.New().String(),
		AgentID:     agent.getID(),
		RequestType: "approval",
		Message:     fmt.Sprintf("Please approve the following step: %s", step),
		Context:     map[string]interface{}{"step": step},
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // Expires in 10 minutes
		Status:      HumanInteractionPending,
	}

	// Add to human-in-the-loop manager
	agent.runtime.humanInLoopMgr.requests[request.ID] = request

	// In a real implementation, this would trigger a notification to humans
	// For now, we'll just return true to avoid hanging

	return true, nil
}

// executeStep executes a single step in the plan
func (agent *AIAgentInstance) executeStep(ctx context.Context, step string) (interface{}, error) {
	// For now, we'll treat each step as a simple AI query
	// In a real implementation, this would involve more sophisticated action parsing
	// and would potentially involve tool calls
	
	// Check if the step looks like it requires a tool
	if agent.config.EnableTools {
		// Try to identify if this step requires a specific tool
		identifiedTool, args, found := agent.identifyToolFromStep(step)
		if found {
			// Execute the identified tool
			result, err := agent.executeSpecificTool(ctx, identifiedTool, args)
			return result, err
		}
	}

	// Treat as a simple query
	response, err := llms.GenerateFromSinglePrompt(ctx, agent.runtime.llm, step,
		llms.WithTemperature(agent.config.Temperature),
		llms.WithMaxTokens(agent.config.MaxTokens),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing step: %w", err)
	}

	return response, nil
}

// identifyToolFromStep tries to identify a tool from a step description
func (agent *AIAgentInstance) identifyToolFromStep(step string) (string, map[string]interface{}, bool) {
	// A simple heuristic approach - in reality, this would be more sophisticated
	// and might use NLP or pattern matching
	
	stepLower := strings.ToLower(step)
	
	for _, tool := range agent.tools {
		toolLower := strings.ToLower(tool.Name)
		descLower := strings.ToLower(tool.Description)
		
		if strings.Contains(stepLower, toolLower) || strings.Contains(stepLower, descLower) {
			// Try to extract arguments from the step
			args := agent.extractArgumentsFromStep(step, tool.Parameters)
			return tool.Name, args, true
		}
	}
	
	return "", nil, false
}

// extractArgumentsFromStep extracts arguments from a step based on tool parameters
func (agent *AIAgentInstance) extractArgumentsFromStep(step string, params map[string]ToolParameter) map[string]interface{} {
	args := make(map[string]interface{})
	
	stepWords := strings.Fields(step)
	
	for paramName, paramSpec := range params {
		for i, word := range stepWords {
			// Simple approach: look for patterns like "paramName=paramValue"
			if strings.Contains(word, paramName) && i+1 < len(stepWords) {
				args[paramName] = stepWords[i+1]
				break
			}
		}
	}
	
	return args
}

// executeSpecificTool executes a specific tool with arguments
func (agent *AIAgentInstance) executeSpecificTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	tool, exists := agent.runtime.availableTools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", toolName)
	}

	result, err := tool.Implementation(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("error executing tool '%s': %w", toolName, err)
	}

	return result, nil
}

// prepareSimplePrompt prepares a simple prompt for the AI model
func (agent *AIAgentInstance) prepareSimplePrompt(input map[string]interface{}) string {
	prompt := ""
	if agent.config.SystemPrompt != "" {
		prompt += fmt.Sprintf("System: %s\n", agent.config.SystemPrompt)
	}
	
	inputJSON, _ := json.Marshal(input)
	prompt += fmt.Sprintf("Input: %s\nOutput:", string(inputJSON))
	
	return prompt
}

// prepareToolsForLLM prepares tools in the format expected by the LLM
func (agent *AIAgentInstance) prepareToolsForLLM() []llms.Tool {
	tools := make([]llms.Tool, 0, len(agent.tools))
	
	for _, tool := range agent.tools {
		function := llms.FunctionDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  agent.convertToolParamsToSchema(tool.Parameters),
		}
		
		tools = append(tools, llms.Tool{
			Type:     "function",
			Function: &function,
		})
	}
	
	return tools
}

// convertToolParamsToSchema converts tool parameters to the schema expected by the LLM
func (agent *AIAgentInstance) convertToolParamsToSchema(params map[string]ToolParameter) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0)
	
	for name, param := range params {
		property := make(map[string]interface{})
		property["type"] = param.Type
		property["description"] = param.Description
		
		if param.Enum != nil {
			property["enum"] = param.Enum
		}
		
		if param.DefaultValue != nil {
			property["default"] = param.DefaultValue
		}
		
		properties[name] = property
		
		if param.Required {
			required = append(required, name)
		}
	}
	
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// executeToolCalls executes the tool calls returned by the LLM
func (agent *AIAgentInstance) executeToolCalls(ctx context.Context, toolCalls []llms.ToolCall) ([]interface{}, error) {
	if len(toolCalls) > agent.config.ToolConfig.MaxToolCalls {
		return nil, fmt.Errorf("too many tool calls: %d (max: %d)", len(toolCalls), agent.config.ToolConfig.MaxToolCalls)
	}
	
	results := make([]interface{}, 0, len(toolCalls))
	
	for _, toolCall := range toolCalls {
		// Find the tool implementation
		tool, exists := agent.runtime.availableTools[toolCall.Function.Name]
		if !exists {
			return nil, fmt.Errorf("tool '%s' not available", toolCall.Function.Name)
		}
		
		// Parse the arguments
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
		}
		
		// Execute the tool with timeout
		toolCtx, cancel := context.WithTimeout(ctx, agent.config.ToolConfig.ToolCallTimeout)
		result, err := tool.Implementation(toolCtx, args)
		cancel()
		
		if err != nil {
			return nil, fmt.Errorf("tool '%s' execution failed: %w", toolCall.Function.Name, err)
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// refineResponse refines the response using tool results
func (agent *AIAgentInstance) refineResponse(ctx context.Context, originalInput map[string]interface{}, initialResponse string, toolResults []interface{}) string {
	refinementPrompt := fmt.Sprintf(
		"Original input: %v\nInitial response: %s\nTool results: %v\n\n"+
		"Based on the tool results, provide a refined response that incorporates the new information.",
		originalInput, initialResponse, toolResults,
	)
	
	refinedResponse, err := llms.GenerateFromSinglePrompt(ctx, agent.runtime.llm, refinementPrompt,
		llms.WithTemperature(agent.config.Temperature),
		llms.WithMaxTokens(agent.config.MaxTokens),
	)
	
	if err != nil {
		return initialResponse // Return initial response if refinement fails
	}
	
	return refinedResponse
}

// getID returns a unique ID for the agent instance
func (agent *AIAgentInstance) getID() string {
	// In a real implementation, this would return a more persistent ID
	// For now, we'll return a hash of the configuration
	configJSON, _ := json.Marshal(agent.config)
	return fmt.Sprintf("agent_%x", len(configJSON))
}

// addToMemory adds information to the agent's memory
func (agent *AIAgentInstance) addToMemory(ctx context.Context, entryType string, content interface{}) error {
	if !agent.config.EnableMemory {
		return nil
	}

	contentStr := fmt.Sprintf("%v", content)

	memoryEntry := &MemoryEntry{
		ID:          uuid.New().String(),
		Content:     contentStr,
		Source:      entryType,
		Timestamp:   time.Now(),
		Importance:  0.5, // Default importance
		Tags:        []string{entryType},
		LastAccessed: time.Now(),
		AccessCount: 1,
	}

	// Add to short-term memory
	if agent.config.MemoryConfig.EnableShortTerm {
		agent.memory.shortTerm.history = append(agent.memory.shortTerm.history, 
			schema.ChatMessage{
				Content: contentStr,
				Type:    schema.ChatMessageTypeHuman, // or appropriate type
			})
		
		// Trim history if needed
		if len(agent.memory.shortTerm.history) > agent.config.MemoryConfig.ShortTermLimit {
			agent.memory.shortTerm.history = agent.memory.shortTerm.history[len(agent.memory.shortTerm.history)-agent.config.MemoryConfig.ShortTermLimit:]
		}
	}

	// Add to long-term memory if enabled
	if agent.config.MemoryConfig.EnableLongTerm {
		// In a real implementation, this would add to persistent storage
		// For now, we'll keep it in memory
		agent.memory.longTerm.memories[memoryEntry.ID] = memoryEntry
	}

	return nil
}

// getConversationHistory retrieves conversation history from memory
func (agent *AIAgentInstance) getConversationHistory() []schema.ChatMessage {
	if !agent.config.EnableMemory || !agent.config.MemoryConfig.EnableShortTerm {
		return nil
	}
	
	// Return a copy of the history
	history := make([]schema.ChatMessage, len(agent.memory.shortTerm.history))
	copy(history, agent.memory.shortTerm.history)
	
	return history
}

// registerDefaultTools registers default tools for the AI agent
func (runtime *AIAgentRuntime) registerDefaultTools() {
	// Web search tool
	runtime.availableTools["web_search"] = AIAgentTool{
		Name:        "web_search",
		Description: "Search the web for current information",
		Parameters: map[string]ToolParameter{
			"query": {
				Type:        "string",
				Description: "Search query",
				Required:    true,
			},
			"num_results": {
				Type:        "number",
				Description: "Number of results to return",
				Required:    false,
				DefaultValue: 5.0,
			},
		},
		Implementation: runtime.webSearchTool,
	}

	// Calculator tool
	runtime.availableTools["calculator"] = AIAgentTool{
		Name:        "calculator",
		Description: "Perform mathematical calculations",
		Parameters: map[string]ToolParameter{
			"expression": {
				Type:        "string",
				Description: "Mathematical expression to evaluate",
				Required:    true,
			},
		},
		Implementation: runtime.calculatorTool,
	}

	// Current datetime tool
	runtime.availableTools["datetime"] = AIAgentTool{
		Name:        "datetime",
		Description: "Get current date and time",
		Parameters:  map[string]ToolParameter{},
		Implementation: runtime.datetimeTool,
	}
}

// webSearchTool implementation
func (runtime *AIAgentRuntime) webSearchTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// In a real implementation, this would call a web search API
	// For now, we'll return a mock result
	
	query, exists := args["query"]
	if !exists {
		return nil, fmt.Errorf("'query' argument is required")
	}
	
	numResults := 5.0
	if n, exists := args["num_results"]; exists {
		if nFloat, ok := n.(float64); ok {
			numResults = nFloat
		}
	}
	
	result := map[string]interface{}{
		"query": query,
		"results": []interface{}{
			map[string]interface{}{
				"title": "Mock Search Result",
				"url":   "https://mock.example.com",
				"snippet": "This is a mock search result for demonstration purposes",
			},
		},
		"num_results_returned": 1,
		"total_results":        1,
	}
	
	return result, nil
}

// calculatorTool implementation
func (runtime *AIAgentRuntime) calculatorTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// In a real implementation, this would evaluate the expression safely
	// For now, we'll return a mock result
	
	expression, exists := args["expression"]
	if !exists {
		return nil, fmt.Errorf("'expression' argument is required")
	}
	
	result := map[string]interface{}{
		"expression": expression,
		"result":     "calculation_result_placeholder",
		"operation":  "mathematical_calculation",
	}
	
	return result, nil
}

// datetimeTool implementation
func (runtime *AIAgentRuntime) datetimeTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	currentTime := time.Now()
	
	result := map[string]interface{}{
		"current_datetime": currentTime.Format(time.RFC3339),
		"timestamp":        currentTime.Unix(),
		"timezone":         currentTime.Location().String(),
		"iso_format":       currentTime.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	return result, nil
}

// parseSteps parses a plan response into individual steps
func parseSteps(plan string) []string {
	// Very simple parsing - in reality, this would be more sophisticated
	lines := strings.Split(plan, "\n")
	trimmedLines := make([]string, 0)
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Look for lines that start with a number (indicating a step)
		if strings.HasPrefix(trimmed, "1.") || 
		   strings.HasPrefix(trimmed, "2.") || 
		   strings.HasPrefix(trimmed, "3.") ||
		   strings.HasPrefix(trimmed, "4.") ||
		   strings.HasPrefix(trimmed, "5.") ||
		   strings.HasPrefix(trimmed, "6.") ||
		   strings.HasPrefix(trimmed, "7.") ||
		   strings.HasPrefix(trimmed, "8.") ||
		   strings.HasPrefix(trimmed, "9.") {
			// Extract step without the number prefix
			parts := strings.SplitN(trimmed, ".", 2)
			if len(parts) == 2 {
				trimmedLine := strings.TrimSpace(parts[1])
				if trimmedLine != "" {
					trimmedLines = append(trimmedLines, trimmedLine)
				}
			}
		} else if strings.Contains(trimmed, "Step") || strings.Contains(trimmed, "step") {
			// Alternative format: "Step 1: do something"
			trimmedLines = append(trimmedLines, trimmed)
		}
	}
	
	return trimmedLines
}

// NewThreadManager creates a new thread manager
func NewThreadManager(maxThreads int) *ThreadManager {
	if maxThreads == 0 {
		maxThreads = 10
	}
	
	return &ThreadManager{
		threads:    make(map[string]*AgentThread),
		maxThreads: maxThreads,
	}
}

// ExecuteThread executes a thread with multiple AI agents
func (tm *ThreadManager) ExecuteThread(ctx context.Context, agents []*AIAgentInstance, initialInput map[string]interface{}) (map[string]interface{}, error) {
	tm.mutex.Lock()
	if len(tm.threads) >= tm.maxThreads {
		tm.mutex.Unlock()
		return nil, fmt.Errorf("maximum number of threads reached: %d", tm.maxThreads)
	}
	
	threadID := uuid.New().String()
	threadCtx, cancel := context.WithCancel(ctx)
	
	thread := &AgentThread{
		ID:        threadID,
		Agents:    agents,
		State:     ThreadStateReady,
		StartTime: time.Now(),
		CancelFunc: cancel,
		Results:   make(map[string]interface{}),
	}
	
	tm.threads[threadID] = thread
	tm.mutex.Unlock()
	
	// Update thread state
	tm.updateThreadState(threadID, ThreadStateRunning)
	
	// Execute agents in the thread
	for i, agent := range agents {
		agentInput := initialInput
		if i > 0 {
			// For subsequent agents, use previous results
			agentInput = thread.Results
		}
		
		result, err := agent.Execute(threadCtx, agentInput)
		if err != nil {
			thread.Error = err
			tm.updateThreadState(threadID, ThreadStateError)
			cancel() // Cancel other agents in the thread
			break
		}
		
		thread.Results[fmt.Sprintf("agent_%d_result", i)] = result
	}
	
	// Complete thread
	endTime := time.Now()
	thread.EndTime = &endTime
	
	if thread.Error == nil {
		tm.updateThreadState(threadID, ThreadStateComplete)
	} else {
		tm.updateThreadState(threadID, ThreadStateError)
	}
	
	result := map[string]interface{}{
		"thread_id": threadID,
		"success":   thread.Error == nil,
		"results":   thread.Results,
		"start_time": thread.StartTime,
		"end_time":   endTime,
		"duration":   endTime.Sub(thread.StartTime),
		"error":      thread.Error,
	}
	
	return result, nil
}

// updateThreadState updates the state of a thread
func (tm *ThreadManager) updateThreadState(threadID string, state AgentThreadState) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	if thread, exists := tm.threads[threadID]; exists {
		thread.State = state
	}
}

// processMessages processes coordination messages
func (cm *CoordinationManager) processMessages() {
	for msg := range cm.communication {
		cm.handleMessage(msg)
	}
}

// handleMessage handles a coordination message
func (cm *CoordinationManager) handleMessage(msg *CoordinationMessage) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Route message to the appropriate agent
	if agent, exists := cm.agents[msg.To]; exists {
		// In a real implementation, this would pass the message to the agent
		// For now, we'll just log it
		fmt.Printf("Message routed to agent %s: %s\n", msg.To, msg.Content)
	}
}

// RegisterAIAgentNode registers the AI agent node with the engine
func RegisterAIAgentNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("ai_agent_runtime", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var provider AIModelProvider
		if prov, exists := config["provider"]; exists {
			if provStr, ok := prov.(string); ok {
				provider = AIModelProvider(provStr)
			}
		}

		var modelName string
		if model, exists := config["model_name"]; exists {
			if modelStr, ok := model.(string); ok {
				modelName = modelStr
			}
		}

		var apiKey string
		if key, exists := config["api_key"]; exists {
			if keyStr, ok := key.(string); ok {
				apiKey = keyStr
			}
		}

		var temperature float64
		if temp, exists := config["temperature"]; exists {
			if tempFloat, ok := temp.(float64); ok {
				temperature = tempFloat
			}
		}

		var maxTokens float64
		if tokens, exists := config["max_tokens"]; exists {
			if tokensFloat, ok := tokens.(float64); ok {
				maxTokens = tokensFloat
			}
		}

		var enableMemory bool
		if mem, exists := config["enable_memory"]; exists {
			if memBool, ok := mem.(bool); ok {
				enableMemory = memBool
			}
		}

		var enableTools bool
		if tools, exists := config["enable_tools"]; exists {
			if toolsBool, ok := tools.(bool); ok {
				enableTools = toolsBool
			}
		}

		var systemPrompt string
		if prompt, exists := config["system_prompt"]; exists {
			if promptStr, ok := prompt.(string); ok {
				systemPrompt = promptStr
			}
		}

		var timeout float64
		if t, exists := config["execution_timeout_seconds"]; exists {
			if tFloat, ok := t.(float64); ok {
				timeout = tFloat
			}
		}

		var maxRetries float64
		if retries, exists := config["max_retries"]; exists {
			if retriesFloat, ok := retries.(float64); ok {
				maxRetries = retriesFloat
			}
		}

		var agentType AIAgentType
		if aType, exists := config["agent_type"]; exists {
			if aTypeStr, ok := aType.(string); ok {
				agentType = AIAgentType(aTypeStr)
			}
		}

		var enableReasoning bool
		if reason, exists := config["enable_reasoning"]; exists {
			if reasonBool, ok := reason.(bool); ok {
				enableReasoning = reasonBool
			}
		}

		var humanInLoop bool
		if hilo, exists := config["human_in_loop"]; exists {
			if hiloBool, ok := hilo.(bool); ok {
				humanInLoop = hiloBool
			}
		}

		var memoryConfig *AgentMemoryConfig
		if memConf, exists := config["memory_config"]; exists {
			if memConfMap, ok := memConf.(map[string]interface{}); ok {
				var shortTermLimit float64
				if stl, exists := memConfMap["short_term_limit"]; exists {
					if stlFloat, ok := stl.(float64); ok {
						shortTermLimit = stlFloat
					}
				}

				var contextWindowSize float64
				if cws, exists := memConfMap["context_window_size"]; exists {
					if cwsFloat, ok := cws.(float64); ok {
						contextWindowSize = cwsFloat
					}
				}

				var maxHistoryLength float64
				if mhl, exists := memConfMap["max_history_length"]; exists {
					if mhlFloat, ok := mhl.(float64); ok {
						maxHistoryLength = mhlFloat
					}
				}

				var compressionRatio float64
				if cr, exists := memConfMap["compression_ratio"]; exists {
					if crFloat, ok := cr.(float64); ok {
						compressionRatio = crFloat
					}
				}

				var enableST, enableLT, enableSumm, enablePersist, enableCompr bool
				if est, exists := memConfMap["enable_short_term"]; exists {
					if estBool, ok := est.(bool); ok {
						enableST = estBool
					}
				}
				if elt, exists := memConfMap["enable_long_term"]; exists {
					if eltBool, ok := elt.(bool); ok {
						enableLT = eltBool
					}
				}
				if es, exists := memConfMap["enable_summarization"]; exists {
					if esBool, ok := es.(bool); ok {
						enableSumm = esBool
					}
				}
				if ep, exists := memConfMap["enable_persistence"]; exists {
					if epBool, ok := ep.(bool); ok {
						enablePersist = epBool
					}
				}
				if ec, exists := memConfMap["enable_compression"]; exists {
					if ecBool, ok := ec.(bool); ok {
						enableCompr = ecBool
					}
				}

				var storage, collection, embeddingModel string
				if st, exists := memConfMap["long_term_storage"]; exists {
					if stStr, ok := st.(string); ok {
						storage = stStr
					}
				}
				if coll, exists := memConfMap["collection_name"]; exists {
					if collStr, ok := coll.(string); ok {
						collection = collStr
					}
				}
				if emb, exists := memConfMap["embedding_model"]; exists {
					if embStr, ok := emb.(string); ok {
						embeddingModel = embStr
					}
				}

				memoryConfig = &AgentMemoryConfig{
					EnableShortTerm:  enableST,
					EnableLongTerm:   enableLT,
					ShortTermLimit:   int(shortTermLimit),
					LongTermStorage:  storage,
					ContextWindowSize: int(contextWindowSize),
					Summarization:    enableSumm,
					Persistence:      enablePersist,
					CollectionName:   collection,
					EmbeddingModel:   embeddingModel,
					MaxHistoryLength: int(maxHistoryLength),
					Compression:      enableCompr,
					CompressionRatio: compressionRatio,
				}
			}
		}

		// Similar processing for toolConfig, threadingConfig if needed

		nodeConfig := &AIAgentConfig{
			Provider:         provider,
			ModelName:        modelName,
			APIKey:           apiKey,
			Temperature:      temperature,
			MaxTokens:        int(maxTokens),
			EnableMemory:     enableMemory,
			EnableTools:      enableTools,
			SystemPrompt:     systemPrompt,
			ExecutionTimeout: time.Duration(timeout) * time.Second,
			MaxRetries:       int(maxRetries),
			AgentType:        agentType,
			EnableReasoning:  enableReasoning,
			HumanInLoop:      humanInLoop,
			MemoryConfig:     memoryConfig,
		}

		runtime, err := NewAIAgentRuntime(nodeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create AI agent runtime: %w", err)
		}

		return runtime.NewAIAgentInstance(nodeConfig), nil
	})
}