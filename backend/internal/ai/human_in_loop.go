// backend/internal/ai/human_in_loop.go
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// HumanInLoopRequest represents a request for human intervention
type HumanInLoopRequest struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	TaskID      string                 `json:"task_id"`
	NodeID      string                 `json:"node_id"`
	RequestType HumanInLoopType       `json:"request_type"`
	Query       string                 `json:"query"`
	Context     map[string]interface{} `json:"context"`
	Options     []string               `json:"options,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Status      HumanInLoopStatus      `json:"status"`
	Response    *HumanInLoopResponse   `json:"response,omitempty"`
	RespondedAt *time.Time             `json:"responded_at,omitempty"`
	UserID      *string                `json:"user_id,omitempty"` // Specific user to ask
	Priority    int                    `json:"priority"` // Higher number = higher priority
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HumanInLoopType represents the type of human intervention needed
type HumanInLoopType string

const (
	RequestTypeApproval     HumanInLoopType = "approval"
	RequestTypeConfirmation HumanInLoopType = "confirmation"
	RequestTypeInformation  HumanInLoopType = "information"
	RequestTypeCorrection   HumanInLoopType = "correction"
	RequestTypeValidation   HumanInLoopType = "validation"
	RequestTypeDecision     HumanInLoopType = "decision"
)

// HumanInLoopStatus represents the status of a human-in-the-loop request
type HumanInLoopStatus string

const (
	StatusPending   HumanInLoopStatus = "pending"
	StatusApproved  HumanInLoopStatus = "approved"
	StatusRejected  HumanInLoopStatus = "rejected"
	StatusConfirmed HumanInLoopStatus = "confirmed"
	StatusCancelled HumanInLoopStatus = "cancelled"
	StatusTimeout   HumanInLoopStatus = "timeout"
	StatusError     HumanInLoopStatus = "error"
)

// HumanInLoopResponse represents a response from human intervention
type HumanInLoopResponse struct {
	RequestID   string                 `json:"request_id"`
	Response    string                 `json:"response"`
	SelectedOption *string            `json:"selected_option,omitempty"`
	Comment     *string                `json:"comment,omitempty"`
	Confidence  *float64              `json:"confidence,omitempty"` // 0.0-1.0
	Timestamp   time.Time              `json:"timestamp"`
	ResponderID *string               `json:"responder_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HumanInLoopManager manages human-in-the-loop requests
type HumanInLoopManager struct {
	requests   map[string]*HumanInLoopRequest
	requestsMu sync.RWMutex
	callback   func(*HumanInLoopResponse) error
	timeout    time.Duration
	notificationService NotificationService
	storage    HumanInLoopStorage
	config     *HumanInLoopConfig
	active     bool
}

// HumanInLoopConfig represents configuration for human-in-the-loop
type HumanInLoopConfig struct {
	EnableHumanInLoop bool          `json:"enable_human_in_loop"`
	DefaultTimeout    time.Duration `json:"default_timeout"`
	MaxConcurrentRequests int       `json:"max_concurrent_requests"`
	EnableNotifications bool       `json:"enable_notifications"`
	NotificationChannels []string   `json:"notification_channels"`
	RequiredForCriticalOperations bool `json:"required_for_critical_operations"`
	PriorityThreshold int           `json:"priority_threshold"`
	AutoApprovalThreshold float64   `json:"auto_approval_threshold"`
	AutoApprovalEnabled bool       `json:"auto_approval_enabled"`
}

// HumanInLoopStorage interface for storing human-in-the-loop requests
type HumanInLoopStorage interface {
	SaveRequest(ctx context.Context, request *HumanInLoopRequest) error
	GetRequest(ctx context.Context, id string) (*HumanInLoopRequest, error)
	UpdateRequest(ctx context.Context, request *HumanInLoopRequest) error
	GetPendingRequests(ctx context.Context, userID *string) ([]*HumanInLoopRequest, error)
	GetRequestsByWorkflow(ctx context.Context, workflowID string) ([]*HumanInLoopRequest, error)
	DeleteRequest(ctx context.Context, id string) error
	Close() error
}

// InMemoryHumanInLoopStorage implements in-memory storage for requests
type InMemoryHumanInLoopStorage struct {
	requests map[string]*HumanInLoopRequest
	mutex    sync.RWMutex
}

// NewInMemoryHumanInLoopStorage creates a new in-memory storage
func NewInMemoryHumanInLoopStorage() *InMemoryHumanInLoopStorage {
	return &InMemoryHumanInLoopStorage{
		requests: make(map[string]*HumanInLoopRequest),
	}
}

func (s *InMemoryHumanInLoopStorage) SaveRequest(ctx context.Context, request *HumanInLoopRequest) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.requests[request.ID] = request
	return nil
}

func (s *InMemoryHumanInLoopStorage) GetRequest(ctx context.Context, id string) (*HumanInLoopRequest, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	request, exists := s.requests[id]
	if !exists {
		return nil, fmt.Errorf("request %s not found", id)
	}
	
	return request, nil
}

func (s *InMemoryHumanInLoopStorage) UpdateRequest(ctx context.Context, request *HumanInLoopRequest) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.requests[request.ID]; !exists {
		return fmt.Errorf("request %s not found", request.ID)
	}
	
	s.requests[request.ID] = request
	return nil
}

func (s *InMemoryHumanInLoopStorage) GetPendingRequests(ctx context.Context, userID *string) ([]*HumanInLoopRequest, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var pendingRequests []*HumanInLoopRequest
	for _, request := range s.requests {
		if request.Status == StatusPending && 
		   (userID == nil || (request.UserID != nil && *request.UserID == *userID)) {
			pendingRequests = append(pendingRequests, request)
		}
	}
	
	return pendingRequests, nil
}

func (s *InMemoryHumanInLoopStorage) GetRequestsByWorkflow(ctx context.Context, workflowID string) ([]*HumanInLoopRequest, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var workflowRequests []*HumanInLoopRequest
	for _, request := range s.requests {
		if request.WorkflowID == workflowID {
			workflowRequests = append(workflowRequests, request)
		}
	}
	
	return workflowRequests, nil
}

func (s *InMemoryHumanInLoopStorage) DeleteRequest(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.requests[id]; !exists {
		return fmt.Errorf("request %s not found", id)
	}
	
	delete(s.requests, id)
	return nil
}

func (s *InMemoryHumanInLoopStorage) Close() error {
	return nil
}

// NewHumanInLoopManager creates a new human-in-the-loop manager
func NewHumanInLoopManager(config *HumanInLoopConfig, storage HumanInLoopStorage) *HumanInLoopManager {
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 24 * time.Hour
	}
	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = 100
	}

	manager := &HumanInLoopManager{
		requests:            make(map[string]*HumanInLoopRequest),
		timeout:             config.DefaultTimeout,
		notificationService: NewNotificationService(), // Assuming this exists
		storage:             storage,
		config:              config,
		active:              true,
	}

	// Start cleanup goroutine to remove expired requests
	go manager.cleanupExpiredRequests()

	return manager
}

// CreateRequest creates a new human-in-the-loop request
func (hilm *HumanInLoopManager) CreateRequest(ctx context.Context, requestType HumanInLoopType, workflowID, taskID, nodeID, query string, contextData map[string]interface{}, options []string) (*HumanInLoopRequest, error) {
	if !hilm.config.EnableHumanInLoop {
		return nil, fmt.Errorf("human-in-the-loop is disabled")
	}

	// Validate input
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	// Create request
	request := &HumanInLoopRequest{
		ID:          uuid.New().String(),
		WorkflowID:  workflowID,
		TaskID:      taskID,
		NodeID:      nodeID,
		RequestType: requestType,
		Query:       query,
		Context:     contextData,
		Options:     options,
		Timeout:     hilm.config.DefaultTimeout,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(hilm.config.DefaultTimeout),
		Status:      StatusPending,
		Priority:    1, // Default priority
		Metadata:    make(map[string]interface{}),
	}

	// Save to storage
	if err := hilm.storage.SaveRequest(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to save human-in-the-loop request: %w", err)
	}

	// Send notification if enabled
	if hilm.config.EnableNotifications {
		hilm.notifyRequest(ctx, request)
	}

	return request, nil
}

// GetRequest retrieves a human-in-the-loop request
func (hilm *HumanInLoopManager) GetRequest(ctx context.Context, id string) (*HumanInLoopRequest, error) {
	return hilm.storage.GetRequest(ctx, id)
}

// Respond responds to a human-in-the-loop request
func (hilm *HumanInLoopManager) Respond(ctx context.Context, requestID, response string, selectedOption *string, comment *string, confidence *float64, responderID *string) error {
	request, err := hilm.storage.GetRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	if request.Status != StatusPending {
		return fmt.Errorf("request is not in pending status, current status: %s", request.Status)
	}

	// Update request with response
	responseObj := &HumanInLoopResponse{
		RequestID:    requestID,
		Response:     response,
		SelectedOption: selectedOption,
		Comment:      comment,
		Confidence:   confidence,
		Timestamp:    time.Now(),
		ResponderID:  responderID,
	}

	request.Status = getStatusForResponse(response, selectedOption)
	request.Response = responseObj
	request.RespondedAt = &responseObj.Timestamp

	// Update in storage
	if err := hilm.storage.UpdateRequest(ctx, request); err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	// Trigger callback if configured
	if hilm.callback != nil {
		go func() {
			if err := hilm.callback(responseObj); err != nil {
				fmt.Printf("Warning: human-in-the-loop callback failed: %v\n", err)
			}
		}()
	}

	return nil
}

// getStatusForResponse determines the status based on the response
func getStatusForResponse(response string, selectedOption *string) HumanInLoopStatus {
	responseLower := strings.ToLower(response)
	
	if responseLower == "approved" || responseLower == "yes" || 
	   (selectedOption != nil && strings.ToLower(*selectedOption) == "approve") {
		return StatusApproved
	} else if responseLower == "rejected" || responseLower == "no" ||
	         (selectedOption != nil && strings.ToLower(*selectedOption) == "reject") {
		return StatusRejected
	} else if responseLower == "confirmed" ||
	         (selectedOption != nil && strings.ToLower(*selectedOption) == "confirm") {
		return StatusConfirmed
	} else {
		return StatusConfirmed // Default to confirmed for other responses
	}
}

// ApproveRequest approves a human-in-the-loop request
func (hilm *HumanInLoopManager) ApproveRequest(ctx context.Context, requestID, comment *string, responderID *string) error {
	return hilm.respondToRequest(ctx, requestID, "approved", nil, comment, &approvedConfidence, responderID)
}

// RejectRequest rejects a human-in-the-loop request
func (hilm *HumanInLoopManager) RejectRequest(ctx context.Context, requestID, comment *string, responderID *string) error {
	return hilm.respondToRequest(ctx, requestID, "rejected", nil, comment, &rejectedConfidence, responderID)
}

// ConfirmRequest confirms a human-in-the-loop request
func (hilm *HumanInLoopManager) ConfirmRequest(ctx context.Context, requestID, comment *string, responderID *string) error {
	return hilm.respondToRequest(ctx, requestID, "confirmed", nil, comment, &confirmedConfidence, responderID)
}

// respondToRequest is a helper for responding to requests
func (hilm *HumanInLoopManager) respondToRequest(ctx context.Context, requestID, response string, selectedOption *string, comment *string, confidence *float64, responderID *string) error {
	return hilm.Respond(ctx, requestID, response, selectedOption, comment, confidence, responderID)
}

// GetPendingRequests retrieves all pending human-in-the-loop requests
func (hilm *HumanInLoopManager) GetPendingRequests(ctx context.Context, userID *string) ([]*HumanInLoopRequest, error) {
	return hilm.storage.GetPendingRequests(ctx, userID)
}

// GetRequestsByWorkflow retrieves all human-in-the-loop requests for a workflow
func (hilm *HumanInLoopManager) GetRequestsByWorkflow(ctx context.Context, workflowID string) ([]*HumanInLoopRequest, error) {
	return hilm.storage.GetRequestsByWorkflow(ctx, workflowID)
}

// notifyRequest sends notifications about new requests
func (hilm *HumanInLoopManager) notifyRequest(ctx context.Context, request *HumanInLoopRequest) {
	// In a real implementation, this would send notifications via multiple channels
	// such as email, Slack, web push notifications, etc.
	
	notification := map[string]interface{}{
		"workflow_id": request.WorkflowID,
		"node_id":     request.NodeID,
		"request_type": string(request.RequestType),
		"query":       request.Query,
		"request_id":  request.ID,
		"created_at":  request.CreatedAt,
		"expires_at":  request.ExpiresAt,
		"priority":    request.Priority,
	}
	
	// Send notification based on configured channels
	if contains(hilm.config.NotificationChannels, "email") {
		hilm.notificationService.SendEmail(ctx, 
			getEmailForRequest(request), 
			"Human Intervention Required", 
			fmt.Sprintf("A human decision is required for workflow %s: %s", 
				request.WorkflowID, request.Query), 
			notification)
	}
	
	if contains(hilm.config.NotificationChannels, "slack") {
		hilm.notificationService.SendSlack(ctx,
			getSlackChannelForRequest(request),
			fmt.Sprintf("*Workflow %s requires human intervention*\n> %s", 
				request.WorkflowID, request.Query),
			notification)
	}
	
	if contains(hilm.config.NotificationChannels, "webhook") {
		hilm.notificationService.SendWebhook(ctx,
			getWebhookURLForRequest(request),
			notification)
	}
}

// get appropriate notification destination
func getEmailForRequest(request *HumanInLoopRequest) string {
	// In a real implementation, this would determine the appropriate email address
	// based on the request configuration, workflow owners, etc.
	return "default-notification@citadel-agent.com"
}

func getSlackChannelForRequest(request *HumanInLoopRequest) string {
	// In a real implementation, this would determine the appropriate Slack channel
	return "#workflow-approvals"
}

func getWebhookURLForRequest(request *HumanInLoopRequest) string {
	// In a real implementation, this would determine the appropriate webhook URL
	return "https://example.com/webhook/human-in-loop"
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// cleanupExpiredRequests cleans up expired requests
func (hilm *HumanInLoopManager) cleanupExpiredRequests() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hilm.cleanupExpired()
		case <-time.After(24 * time.Hour): // Also run once every 24 hours just in case
			hilm.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired requests
func (hilm *HumanInLoopManager) cleanupExpired() {
	ctx := context.Background()
	
	// In a real implementation, we would:
	// 1. Query storage for expired requests
	// 2. Update their status to "timeout"
	// 3. Potentially trigger alternative workflow paths
	
	// For now, we'll just log that we performed a cleanup
	fmt.Println("Performed cleanup of expired human-in-the-loop requests")
}

// SetCallback sets a callback to be called when a response is received
func (hilm *HumanInLoopManager) SetCallback(callback func(*HumanInLoopResponse) error) {
	hilm.callback = callback
}

// WithContext adds human-in-the-loop context to a context
func WithContext(ctx context.Context, manager *HumanInLoopManager) context.Context {
	return context.WithValue(ctx, "human_in_loop_manager", manager)
}

// FromContext retrieves the human-in-the-loop manager from context
func FromContext(ctx context.Context) (*HumanInLoopManager, bool) {
	manager, exists := ctx.Value("human_in_loop_manager").(*HumanInLoopManager)
	return manager, exists
}

// AutoApproveBasedOnRules automatically approves requests based on predefined rules
func (hilm *HumanInLoopManager) AutoApproveBasedOnRules(ctx context.Context, request *HumanInLoopRequest) (bool, error) {
	if !hilm.config.AutoApprovalEnabled || request.Priority > hilm.config.PriorityThreshold {
		return false, nil
	}

	// In a real implementation, this would evaluate approval rules
	// For now, we'll just return based on a simple heuristic
	if request.RequestType == RequestTypeConfirmation && request.Priority < 3 {
		// Automatically approve low-priority confirmations
		response := &HumanInLoopResponse{
			RequestID: request.ID,
			Response:  "approved",
			Timestamp: time.Now(),
		}

		request.Status = StatusApproved
		request.Response = response
		request.RespondedAt = &response.Timestamp

		if err := hilm.storage.UpdateRequest(ctx, request); err != nil {
			return false, fmt.Errorf("failed to auto-approve request: %w", err)
		}

		return true, nil
	}

	return false, nil
}

// BatchRespond handles multiple responses at once
func (hilm *HumanInLoopManager) BatchRespond(ctx context.Context, responses []*HumanInLoopResponse) error {
	var errors []error
	
	for _, response := range responses {
		err := hilm.Respond(ctx, response.RequestID, response.Response, response.SelectedOption, response.Comment, response.Confidence, response.ResponderID)
		if err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to process %d out of %d responses: %v", len(errors), len(responses), errors)
	}
	
	return nil
}

// GetStats returns statistics about human-in-the-loop requests
func (hilm *HumanInLoopManager) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// In a real implementation, this would aggregate statistics from storage
	stats := map[string]interface{}{
		"total_requests":          0,
		"pending_requests":        0,
		"approved_requests":       0,
		"rejected_requests":       0,
		"timeout_requests":        0,
		"average_response_time":   0.0,
		"auto_approved_ratio":     0.0,
		"most_common_request_types": []string{},
		"response_time_distribution": map[string]interface{}{},
	}
	
	return stats, nil
}

// CancelRequest cancels a pending human-in-the-loop request
func (hilm *HumanInLoopManager) CancelRequest(ctx context.Context, requestID string) error {
	request, err := hilm.storage.GetRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	if request.Status != StatusPending {
		return fmt.Errorf("can only cancel pending requests, current status: %s", request.Status)
	}

	request.Status = StatusCancelled
	request.RespondedAt = &time.Now()

	if err := hilm.storage.UpdateRequest(ctx, request); err != nil {
		return fmt.Errorf("failed to cancel request: %w", err)
	}

	return nil
}

// Close shuts down the human-in-the-loop manager
func (hilm *HumanInLoopManager) Close() error {
	hilm.active = false
	return hilm.storage.Close()
}