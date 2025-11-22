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





// TeamMember represents a member of a team
type TeamMember struct {
	ID        string    `json:"id" db:"id"`
	TeamID    string    `json:"team_id" db:"team_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"` // Using string instead of UserRole since it's defined in user.go
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