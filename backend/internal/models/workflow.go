// backend/internal/models/workflow.go
package models

import "time"

// Workflow represents a workflow definition
type Workflow struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Nodes       []*Node                `json:"nodes" db:"nodes"`
	Connections []*Connection          `json:"connections" db:"connections"`
	Config      map[string]interface{} `json:"config" db:"config"`
	Status      WorkflowStatus         `json:"status" db:"status"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Version     int                    `json:"version" db:"version"`
	OwnerID     string                 `json:"owner_id" db:"owner_id"`
	TeamID      string                 `json:"team_id" db:"team_id"`
	Tags        []string               `json:"tags" db:"tags"`
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus string

const (
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusInactive  WorkflowStatus = "inactive"
	WorkflowStatusDraft     WorkflowStatus = "draft"
	WorkflowStatusArchived  WorkflowStatus = "archived"
	WorkflowStatusSuspended WorkflowStatus = "suspended"
)

// Node represents a single node in a workflow
type Node struct {
	ID          string                 `json:"id" db:"id"`
	Type        string                 `json:"type" db:"type"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Config      map[string]interface{} `json:"config" db:"config"`
	Dependencies []string              `json:"dependencies" db:"dependencies"`
	Inputs      map[string]interface{} `json:"inputs" db:"inputs"`
	Outputs     map[string]interface{} `json:"outputs" db:"outputs"`
	Status      NodeStatus             `json:"status" db:"status"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Position    NodePosition           `json:"position" db:"position"`
	Group       string                 `json:"group" db:"group"`
}

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusSuccess   NodeStatus = "success"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusSkipped   NodeStatus = "skipped"
	NodeStatusCancelled NodeStatus = "cancelled"
)

// NodePosition represents the position of a node in the UI
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Connection represents a connection between nodes
type Connection struct {
	ID           string `json:"id" db:"id"`
	WorkflowID   string `json:"workflow_id" db:"workflow_id"`
	SourceNodeID string `json:"source_node_id" db:"source_node_id"`
	TargetNodeID string `json:"target_node_id" db:"target_node_id"`
	SourceType   string `json:"source_type" db:"source_type"` // Output port type
	TargetType   string `json:"target_type" db:"target_type"` // Input port type
	Port         string `json:"port,omitempty" db:"port"`     // Specific port
	Condition    string `json:"condition,omitempty" db:"condition"` // Conditional execution
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Execution represents a running instance of a workflow
type Execution struct {
	ID            string                 `json:"id" db:"id"`
	WorkflowID    string                 `json:"workflow_id" db:"workflow_id"`
	Name          string                 `json:"name" db:"name"`
	Status        ExecutionStatus        `json:"status" db:"status"`
	StartedAt     time.Time              `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	Variables     map[string]interface{} `json:"variables" db:"variables"`
	NodeResults   map[string]*NodeResult `json:"node_results" db:"node_results"`
	Error         *string                `json:"error,omitempty" db:"error"`
	TriggeredBy   string                 `json:"triggered_by" db:"triggered_by"`
	TriggerParams map[string]interface{} `json:"trigger_params,omitempty" db:"trigger_params"`
	ParentID      *string                `json:"parent_id,omitempty" db:"parent_id"` // For sub-workflows
	RetryCount    int                    `json:"retry_count" db:"retry_count"`
	UserID        string                 `json:"user_id" db:"user_id"`
	TeamID        string                 `json:"team_id" db:"team_id"`
}

// ExecutionStatus represents the status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSuccess   ExecutionStatus = "success"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusPaused    ExecutionStatus = "paused"
	ExecutionStatusRetrying  ExecutionStatus = "retrying"
)

// NodeResult represents the result of a node execution
type NodeResult struct {
	ID            string                 `json:"id" db:"id"`
	ExecutionID   string                 `json:"execution_id" db:"execution_id"`
	NodeID        string                 `json:"node_id" db:"node_id"`
	Status        NodeStatus             `json:"status" db:"status"`
	Output        map[string]interface{} `json:"output" db:"output"`
	Error         *string                `json:"error,omitempty" db:"error"`
	StartedAt     time.Time              `json:"started_at" db:"started_at"`
	CompletedAt   time.Time              `json:"completed_at" db:"completed_at"`
	ExecutionTime int64                  `json:"execution_time" db:"execution_time"` // in milliseconds
	RetryCount    int                    `json:"retry_count" db:"retry_count"`
}

// ExecutionLog represents a log entry for an execution
type ExecutionLog struct {
	ID          string                 `json:"id" db:"id"`
	WorkflowID  string                 `json:"workflow_id" db:"workflow_id"`
	ExecutionID string                 `json:"execution_id" db:"execution_id"`
	NodeID      *string                `json:"node_id,omitempty" db:"node_id"`
	Status      ExecutionStatus        `json:"status" db:"status"`
	Action      string                 `json:"action" db:"action"`
	Message     string                 `json:"message" db:"message"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Parameters  map[string]interface{} `json:"parameters" db:"parameters"`
	Details     map[string]interface{} `json:"details" db:"details"`
	UserID      string                 `json:"user_id" db:"user_id"`
}

// User represents a user in the system
type User struct {
	ID            string    `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Name          string    `json:"name" db:"name"`
	PasswordHash  string    `json:"-" db:"password_hash"` // Never return this in JSON
	Role          UserRole  `json:"role" db:"role"`
	Status        UserStatus `json:"status" db:"status"`
	LastLoginAt   time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	Profile       UserProfile `json:"profile" db:"profile"`
	Preferences   map[string]interface{} `json:"preferences" db:"preferences"`
}

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRoleManager   UserRole = "manager"
	UserRoleDeveloper UserRole = "developer"
	UserRoleViewer    UserRole = "viewer"
	UserRoleGuest     UserRole = "guest"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending  UserStatus = "pending"
)

// UserProfile contains user profile information
type UserProfile struct {
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	AvatarURL   string    `json:"avatar_url" db:"avatar_url"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Timezone    string    `json:"timezone" db:"timezone"`
	Locale      string    `json:"locale" db:"locale"`
	LastAccess  time.Time `json:"last_access" db:"last_access"`
}

// Team represents a team of users
type Team struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	MemberCount int       `json:"member_count" db:"member_count"`
	Settings    map[string]interface{} `json:"settings" db:"settings"`
}

// TeamMember represents a member of a team
type TeamMember struct {
	ID        string    `json:"id" db:"id"`
	TeamID    string    `json:"team_id" db:"team_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      UserRole  `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	UserID      string    `json:"user_id" db:"user_id"`
	TeamID      *string   `json:"team_id,omitempty" db:"team_id"`
	KeyHash     string    `json:"-" db:"key_hash"`
	Prefix      string    `json:"prefix" db:"prefix"` // First few characters for identification
	Permissions []string  `json:"permissions" db:"permissions"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
}

// Plugin represents a plugin in the system
type Plugin struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Version     string    `json:"version" db:"version"`
	Author      string    `json:"author" db:"author"`
	URL         string    `json:"url" db:"url"`
	License     string    `json:"license" db:"license"`
	Category    string    `json:"category" db:"category"`
	Settings    map[string]interface{} `json:"settings" db:"settings"`
	Manifest    PluginManifest `json:"manifest" db:"manifest"`
	Status      PluginStatus `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	FileInfo    PluginFileInfo `json:"file_info" db:"file_info"`
}

// PluginStatus represents the status of a plugin
type PluginStatus string

const (
	PluginStatusActive    PluginStatus = "active"
	PluginStatusInactive  PluginStatus = "inactive"
	PluginStatusPending   PluginStatus = "pending"
	PluginStatusSuspended PluginStatus = "suspended"
	PluginStatusError     PluginStatus = "error"
)

// PluginManifest contains plugin metadata
type PluginManifest struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Repository  string                 `json:"repository"`
	Homepage    string                 `json:"homepage"`
	Category    string                 `json:"category"`
	Keywords    []string               `json:"keywords"`
	Dependencies map[string]string     `json:"dependencies"`
	Engines     map[string]string      `json:"engines"`
	Activation  string                 `json:"activation"`
	Config      map[string]interface{} `json:"config"`
	Nodes       []PluginNode           `json:"nodes"`
	Permissions []string               `json:"permissions"`
	APIs        []PluginAPI            `json:"apis"`
}

// PluginNode represents a node provided by a plugin
type PluginNode struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Type        string                 `json:"type"`
	Category    string                 `json:"category"`
	Inputs      []NodeInput            `json:"inputs"`
	Outputs     []NodeOutput           `json:"outputs"`
	Config      map[string]interface{} `json:"config"`
	Documentation string               `json:"documentation"`
}

// NodeInput represents an input for a node
type NodeInput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     interface{} `json:"default"`
}

// NodeOutput represents an output for a node
type NodeOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// PluginAPI represents an API endpoint provided by a plugin
type PluginAPI struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Handler     string `json:"handler"`
	Description string `json:"description"`
	Auth        bool   `json:"auth"`
}

// PluginFileInfo contains information about the plugin file
type PluginFileInfo struct {
	Size        int64     `json:"size"`
	Hash        string    `json:"hash"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string    `json:"id" db:"id"`
	UserID    *string   `json:"user_id,omitempty" db:"user_id"`
	TeamID    *string   `json:"team_id,omitempty" db:"team_id"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	ResourceID string   `json:"resource_id" db:"resource_id"`
	OldValues map[string]interface{} `json:"old_values" db:"old_values"`
	NewValues map[string]interface{} `json:"new_values" db:"new_values"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
}