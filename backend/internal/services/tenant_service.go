// backend/internal/services/tenant_service.go
package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/citadel-agent/backend/internal/models"
	"github.com/citadel-agent/backend/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	OwnerID     string                 `json:"owner_id"`
	Status      TenantStatus           `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Settings    map[string]interface{} `json:"settings"`
	Usage       *TenantUsage           `json:"usage"`
	Plan        *TenantPlan            `json:"plan"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusPending   TenantStatus = "pending"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// TenantUsage represents the usage statistics for a tenant
type TenantUsage struct {
	WorkflowExecutions int64 `json:"workflow_executions"`
	ActiveUsers       int   `json:"active_users"`
	StorageUsed       int64 `json:"storage_used"` // in bytes
	APIRequests       int64 `json:"api_requests"`
	MaxWorkflows      int   `json:"max_workflows"`
	MaxUsers          int   `json:"max_users"`
}

// TenantPlan represents the plan for a tenant
type TenantPlan struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxUsers    int    `json:"max_users"`
	MaxStorage  int64  `json:"max_storage"` // in bytes
	MaxAPIReq   int64  `json:"max_api_requests"`
	Features    []string `json:"features"`
	Price       string `json:"price"` // In a currency format
}

// TenantService handles multi-tenant operations
type TenantService struct {
	db           *pgxpool.Pool
	tenantRepo   *repositories.TenantRepository
	userRepo     *repositories.UserRepository
	teamRepo     *repositories.TeamRepository
	workflowRepo *repositories.WorkflowRepository
	settings     *TenantSettings
}

// TenantSettings holds tenant configuration
type TenantSettings struct {
	DefaultTenantLimit int
	IsolationLevel     string // "database", "schema", or "row"
	StorageLimit       int64  // Default storage limit per tenant
	EnableMultiTenant  bool
}

// NewTenantService creates a new tenant service
func NewTenantService(
	db *pgxpool.Pool,
	tenantRepo *repositories.TenantRepository,
	userRepo *repositories.UserRepository,
	teamRepo *repositories.TeamRepository,
	workflowRepo *repositories.WorkflowRepository,
) *TenantService {
	return &TenantService{
		db:           db,
		tenantRepo:   tenantRepo,
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		workflowRepo: workflowRepo,
		settings: &TenantSettings{
			DefaultTenantLimit: 50,
			IsolationLevel:     "row", // Default to row-level isolation
			StorageLimit:       10 * 1024 * 1024 * 1024, // 10GB default
			EnableMultiTenant:  true,
		},
	}
}

// CreateTenant creates a new tenant
func (ts *TenantService) CreateTenant(ctx context.Context, name, description, ownerID string) (*Tenant, error) {
	// Validate inputs
	if name == "" {
		return nil, fmt.Errorf("tenant name is required")
	}

	if ownerID == "" {
		return nil, fmt.Errorf("owner ID is required")
	}

	// Check if user exists
	_, err := ts.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("owner user does not exist: %s", ownerID)
		}
		return nil, fmt.Errorf("failed to verify owner user: %w", err)
	}

	// Check if tenant name already exists
	existingTenant, err := ts.tenantRepo.GetByName(ctx, name)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing tenant: %w", err)
	}
	if existingTenant != nil {
		return nil, fmt.Errorf("tenant with name '%s' already exists", name)
	}

	// Create the tenant
	tenant := &models.Tenant{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		Status:      TenantStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Settings:    map[string]interface{}{},
	}

	createdTenant, err := ts.tenantRepo.Create(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create initial team for the tenant (tenant owner's team)
	team := &models.Team{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("%s Team", name),
		Description: fmt.Sprintf("Default team for %s", name),
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = ts.teamRepo.Create(ctx, team)
	if err != nil {
		// Rollback tenant creation if team creation fails
		rollbackErr := ts.tenantRepo.UpdateStatus(ctx, createdTenant.ID, TenantStatusDeleted)
		if rollbackErr != nil {
			return nil, fmt.Errorf("failed to create team and failed to rollback tenant creation: %w, %v", err, rollbackErr)
		}
		return nil, fmt.Errorf("failed to create initial team: %w", err)
	}

	// Add owner to the team
	teamMember := &models.TeamMember{
		ID:       uuid.New().String(),
		TeamID:   team.ID,
		UserID:   ownerID,
		Role:     "admin", // Owner gets admin role by default
		JoinedAt: time.Now(),
	}

	_, err = ts.teamRepo.AddMember(ctx, teamMember)
	if err != nil {
		return nil, fmt.Errorf("failed to add owner to team: %w", err)
	}

	// Convert to response type
	responseTenant := &Tenant{
		ID:          createdTenant.ID,
		Name:        createdTenant.Name,
		Description: createdTenant.Description,
		OwnerID:     createdTenant.OwnerID,
		Status:      TenantStatus(createdTenant.Status),
		CreatedAt:   createdTenant.CreatedAt,
		UpdatedAt:   createdTenant.UpdatedAt,
		Settings:    createdTenant.Settings,
		Usage: &TenantUsage{
			WorkflowExecutions: 0,
			ActiveUsers:       1, // Owner counts as the first user
			StorageUsed:       0,
			APIRequests:       0,
			MaxWorkflows:      100, // Default limit
			MaxUsers:          10,  // Default limit
		},
		Plan: &TenantPlan{
			Name:        "free",
			Description: "Free plan with basic features",
			MaxUsers:    10,
			MaxStorage:  ts.settings.StorageLimit,
			MaxAPIReq:   10000,
			Features:    []string{"basic-workflows", "basic-security"},
			Price:       "0 USD",
		},
	}

	return responseTenant, nil
}

// GetTenant retrieves a tenant by ID
func (ts *TenantService) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	dbTenant, err := ts.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Get tenant usage
	usage, err := ts.getTenantUsage(ctx, tenantID)
	if err != nil {
		// Just log the error, don't fail the entire operation
		fmt.Printf("Warning: failed to get tenant usage: %v\n", err)
	}

	tenant := &Tenant{
		ID:          dbTenant.ID,
		Name:        dbTenant.Name,
		Description: dbTenant.Description,
		OwnerID:     dbTenant.OwnerID,
		Status:      TenantStatus(dbTenant.Status),
		CreatedAt:   dbTenant.CreatedAt,
		UpdatedAt:   dbTenant.UpdatedAt,
		Settings:    dbTenant.Settings,
		Usage:       usage,
	}

	return tenant, nil
}

// UpdateTenant updates a tenant
func (ts *TenantService) UpdateTenant(ctx context.Context, tenantID, name, description string, status *TenantStatus) (*Tenant, error) {
	dbTenant, err := ts.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Update fields if provided
	if name != "" {
		dbTenant.Name = name
	}
	if description != "" {
		dbTenant.Description = description
	}
	if status != nil {
		dbTenant.Status = string(*status)
	}
	dbTenant.UpdatedAt = time.Now()

	updatedTenant, err := ts.tenantRepo.Update(ctx, dbTenant)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Get updated tenant usage
	usage, err := ts.getTenantUsage(ctx, tenantID)
	if err != nil {
		// Just log the error, don't fail the entire operation
		fmt.Printf("Warning: failed to get tenant usage: %v\n", err)
	}

	tenant := &Tenant{
		ID:          updatedTenant.ID,
		Name:        updatedTenant.Name,
		Description: updatedTenant.Description,
		OwnerID:     updatedTenant.OwnerID,
		Status:      TenantStatus(updatedTenant.Status),
		CreatedAt:   updatedTenant.CreatedAt,
		UpdatedAt:   updatedTenant.UpdatedAt,
		Settings:    updatedTenant.Settings,
		Usage:       usage,
	}

	return tenant, nil
}

// DeleteTenant deletes a tenant (soft delete)
func (ts *TenantService) DeleteTenant(ctx context.Context, tenantID string) error {
	dbTenant, err := ts.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("tenant not found: %s", tenantID)
		}
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Check if tenant has active resources that prevent deletion
	usage, err := ts.getTenantUsage(ctx, tenantID)
	if err != nil {
		// Just log and continue
		fmt.Printf("Warning: failed to check tenant usage before deletion: %v\n", err)
	}

	// For now, we'll just update the status to deleted
	// In a real implementation, you might want to:
	// 1. Disable all users in the tenant
	// 2. Cancel all active workflows
	// 3. Delete or archive all tenant data
	dbTenant.Status = string(TenantStatusDeleted)
	dbTenant.UpdatedAt = time.Now()

	_, err = ts.tenantRepo.Update(ctx, dbTenant)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

// ListTenants lists tenants with optional filters
func (ts *TenantService) ListTenants(ctx context.Context, status *TenantStatus, limit, offset int) ([]*Tenant, error) {
	filters := repositories.TenantFilters{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}

	dbTenants, err := ts.tenantRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	tenants := make([]*Tenant, len(dbTenants))
	for i, dbTenant := range dbTenants {
		usage, err := ts.getTenantUsage(ctx, dbTenant.ID)
		if err != nil {
			// Just log the error, don't fail the entire operation
			fmt.Printf("Warning: failed to get tenant usage for %s: %v\n", dbTenant.ID, err)
		}

		tenants[i] = &Tenant{
			ID:          dbTenant.ID,
			Name:        dbTenant.Name,
			Description: dbTenant.Description,
			OwnerID:     dbTenant.OwnerID,
			Status:      TenantStatus(dbTenant.Status),
			CreatedAt:   dbTenant.CreatedAt,
			UpdatedAt:   dbTenant.UpdatedAt,
			Settings:    dbTenant.Settings,
			Usage:       usage,
		}
	}

	return tenants, nil
}

// GetTenantByUser retrieves the tenant for a specific user
func (ts *TenantService) GetTenantByUser(ctx context.Context, userID string) (*Tenant, error) {
	// First, find which teams the user belongs to
	teams, err := ts.teamRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for user: %w", err)
	}

	if len(teams) == 0 {
		return nil, fmt.Errorf("user does not belong to any teams")
	}

	// For simplicity, we'll assume the user belongs to one tenant
	// In a real implementation, a user might belong to multiple tenants
	// and you'd need to handle this appropriately
	team := teams[0]

	// Get the team owner (who created the team, likely the tenant owner)
	teamDetails, err := ts.teamRepo.GetByID(ctx, team.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team details: %w", err)
	}

	// Find the tenant associated with this team
	// This assumes there's a relationship between teams and tenants
	// In a real implementation, you might have a direct tenant_id in the team table
	// or need to join with another table
	
	// For now, we'll assume the team owner is also the tenant owner
	tenants, err := ts.tenantRepo.GetByOwner(ctx, teamDetails.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by owner: %w", err)
	}

	if len(tenants) == 0 {
		return nil, fmt.Errorf("no tenant found for user")
	}

	// Return the first tenant (in a real system, users might belong to multiple tenants)
	dbTenant := tenants[0]
	
	usage, err := ts.getTenantUsage(ctx, dbTenant.ID)
	if err != nil {
		// Just log the error, don't fail the entire operation
		fmt.Printf("Warning: failed to get tenant usage: %v\n", err)
	}

	tenant := &Tenant{
		ID:          dbTenant.ID,
		Name:        dbTenant.Name,
		Description: dbTenant.Description,
		OwnerID:     dbTenant.OwnerID,
		Status:      TenantStatus(dbTenant.Status),
		CreatedAt:   dbTenant.CreatedAt,
		UpdatedAt:   dbTenant.UpdatedAt,
		Settings:    dbTenant.Settings,
		Usage:       usage,
	}

	return tenant, nil
}

// InviteUser invites a user to a tenant
func (ts *TenantService) InviteUser(ctx context.Context, tenantID, email, role string) error {
	// Find the team associated with the tenant
	// In a real implementation, you'd likely have a team_id in the tenants table
	// or a more direct association

	// First, check if user exists by email
	user, err := ts.userRepo.GetByEmail(ctx, email)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	var userID string
	if user != nil {
		// User already exists
		userID = user.ID
	} else {
		// User doesn't exist, we might want to create a pending invitation
		// For now, we'll just return an error
		return fmt.Errorf("user with email %s does not exist", email)
	}

	// Find the default team for this tenant (usually the owner's team)
	teams, err := ts.teamRepo.GetByOwner(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find team for tenant: %w", err)
	}

	if len(teams) == 0 {
		return fmt.Errorf("no team found for tenant")
	}

	// Add user to the team
	teamMember := &models.TeamMember{
		ID:       uuid.New().String(),
		TeamID:   teams[0].ID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}

	_, err = ts.teamRepo.AddMember(ctx, teamMember)
	if err != nil {
		return fmt.Errorf("failed to add user to team: %w", err)
	}

	return nil
}

// RemoveUser removes a user from a tenant
func (ts *TenantService) RemoveUser(ctx context.Context, tenantID, userID string) error {
	// Find all teams in the tenant that the user belongs to
	teams, err := ts.teamRepo.GetByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user's teams: %w", err)
	}

	// Remove user from all teams in the tenant
	for _, team := range teams {
		// Check if this team belongs to the specified tenant
		// In a real implementation, you'd verify the tenant-team relationship
		err = ts.teamRepo.RemoveMember(ctx, team.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to remove user from team %s: %w", team.ID, err)
		}
	}

	return nil
}

// GetTenantUsers retrieves all users in a tenant
func (ts *TenantService) GetTenantUsers(ctx context.Context, tenantID string, limit, offset int) ([]*models.User, error) {
	// Find all teams in the tenant
	// In a real implementation, this would be more direct
	// For now, we'll assume we need to go through teams
	
	// First, find the owner of the tenant to identify associated teams
	tenant, err := ts.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Get all teams owned by the tenant owner
	teams, err := ts.teamRepo.GetByOwner(ctx, tenant.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for tenant owner: %w", err)
	}

	// Collect all user IDs from all teams
	var allUserIDs []string
	for _, team := range teams {
		members, err := ts.teamRepo.GetMembers(ctx, team.ID, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get team members: %w", err)
		}
		
		for _, member := range members {
			allUserIDs = append(allUserIDs, member.UserID)
		}
	}

	// Get user details
	var users []*models.User
	for _, userID := range allUserIDs {
		user, err := ts.userRepo.GetByID(ctx, userID)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Warning: failed to get user %s: %v\n", userID, err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// getTenantUsage calculates the usage for a tenant
func (ts *TenantService) getTenantUsage(ctx context.Context, tenantID string) (*TenantUsage, error) {
	// This is a simplified implementation
	// In a real implementation, you would query actual usage data
	
	// For now, we'll return a placeholder with mock data
	// This would typically involve:
	// - Counting workflow executions in the last month
	// - Counting active users
	// - Calculating storage used by the tenant
	// - Counting API requests
	
	usage := &TenantUsage{
		WorkflowExecutions: 1250, // Mock data
		ActiveUsers:       5,     // Mock data
		StorageUsed:       1024 * 1024 * 500, // 500MB mock data
		APIRequests:       8500,  // Mock data
		MaxWorkflows:      100,   // Default limit
		MaxUsers:          10,    // Default limit
	}

	return usage, nil
}

// SwitchTenant allows a user to switch between tenants they belong to
func (ts *TenantService) SwitchTenant(ctx context.Context, userID, tenantID string) error {
	// Verify that the user belongs to the specified tenant
	userTenants, err := ts.getUserTenants(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user tenants: %w", err)
	}

	found := false
	for _, tenant := range userTenants {
		if tenant.ID == tenantID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("user does not belong to tenant %s", tenantID)
	}

	// In a real implementation, you might store the active tenant in the session
	// or update a user's preferences table
	// For now, we'll just verify that the switch is valid

	return nil
}

// getUserTenants gets all tenants a user belongs to
func (ts *TenantService) getUserTenants(ctx context.Context, userID string) ([]*Tenant, error) {
	// Find all teams the user belongs to
	teams, err := ts.teamRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's teams: %w", err)
	}

	var tenants []*Tenant
	processedTenantIDs := make(map[string]bool)

	for _, team := range teams {
		// Get team details to find the owner
		teamDetails, err := ts.teamRepo.GetByID(ctx, team.ID)
		if err != nil {
			continue // Skip if we can't get team details
		}

		// Find the tenant for this team (using the owner)
		tenantList, err := ts.tenantRepo.GetByOwner(ctx, teamDetails.OwnerID)
		if err != nil {
			continue // Skip if we can't find tenant for this team
		}

		for _, tenant := range tenantList {
			if _, exists := processedTenantIDs[tenant.ID]; !exists {
				// Get full tenant details
				fullTenant, err := ts.GetTenant(ctx, tenant.ID)
				if err != nil {
					continue // Skip if we can't get full details
				}
				tenants = append(tenants, fullTenant)
				processedTenantIDs[tenant.ID] = true
			}
		}
	}

	return tenants, nil
}

// UpdateTenantSettings updates the settings for a tenant
func (ts *TenantService) UpdateTenantSettings(ctx context.Context, tenantID string, settings map[string]interface{}) error {
	dbTenant, err := ts.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("tenant not found: %s", tenantID)
		}
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Update settings
	for key, value := range settings {
		dbTenant.Settings[key] = value
	}
	dbTenant.UpdatedAt = time.Now()

	_, err = ts.tenantRepo.Update(ctx, dbTenant)
	if err != nil {
		return fmt.Errorf("failed to update tenant settings: %w", err)
	}

	return nil
}

// GetTenantSettings retrieves the settings for a tenant
func (ts *TenantService) GetTenantSettings(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	dbTenant, err := ts.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return dbTenant.Settings, nil
}

// ValidateTenantAccess checks if a user has access to a tenant
func (ts *TenantService) ValidateTenantAccess(ctx context.Context, userID, tenantID string) (bool, error) {
	// Check if user belongs to any team in the tenant
	teams, err := ts.teamRepo.GetByUser(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user's teams: %w", err)
	}

	for _, team := range teams {
		// Check if this team belongs to the tenant
		// In a real implementation, you'd have a more direct way to check this
		teamDetails, err := ts.teamRepo.GetByID(ctx, team.ID)
		if err != nil {
			continue // Skip if we can't get team details
		}

		// Find tenants for the team owner
		tenants, err := ts.tenantRepo.GetByOwner(ctx, teamDetails.OwnerID)
		if err != nil {
			continue // Skip if we can't find tenants for this team
		}

		for _, tenant := range tenants {
			if tenant.ID == tenantID {
				return true, nil
			}
		}
	}

	return false, nil
}