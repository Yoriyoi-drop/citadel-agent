// backend/internal/repositories/team_repository.go
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TeamRepository handles team-related database operations
type TeamRepository struct {
	db *pgxpool.Pool
}

// Team represents a team in the database
type Team struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	OwnerID     string                 `json:"owner_id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Settings    map[string]interface{} `json:"settings"`
}

// TeamMember represents a team member in the database
type TeamMember struct {
	ID       string    `json:"id"`
	TeamID   string    `json:"team_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"` // 'admin', 'member', 'viewer'
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{
		db: db,
	}
}

// Create creates a new team
func (tr *TeamRepository) Create(ctx context.Context, team *Team) (*Team, error) {
	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(team.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO teams (
			id, name, description, owner_id, created_at, updated_at, settings
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, description, owner_id, created_at, updated_at
	`

	var createdTeam Team
	err = tr.db.QueryRow(ctx, query,
		team.ID,
		team.Name,
		team.Description,
		team.OwnerID,
		team.CreatedAt,
		team.UpdatedAt,
		settingsJSON,
	).Scan(
		&createdTeam.ID,
		&createdTeam.Name,
		&createdTeam.Description,
		&createdTeam.OwnerID,
		&createdTeam.CreatedAt,
		&createdTeam.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Set settings from original input since they're not returned by the query
	createdTeam.Settings = team.Settings

	return &createdTeam, nil
}

// GetByID retrieves a team by ID
func (tr *TeamRepository) GetByID(ctx context.Context, id string) (*Team, error) {
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at, settings
		FROM teams
		WHERE id = $1
	`

	var team Team
	var settingsJSON []byte

	err := tr.db.QueryRow(ctx, query, id).Scan(
		&team.ID,
		&team.Name,
		&team.Description,
		&team.OwnerID,
		&team.CreatedAt,
		&team.UpdatedAt,
		&settingsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	// Deserialize settings
	if settingsJSON != nil {
		if err := json.Unmarshal(settingsJSON, &team.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}

	return &team, nil
}

// GetByOwner retrieves teams by owner ID
func (tr *TeamRepository) GetByOwner(ctx context.Context, ownerID string) ([]*Team, error) {
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at, settings
		FROM teams
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := tr.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		var team Team
		var settingsJSON []byte

		err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&team.OwnerID,
			&team.CreatedAt,
			&team.UpdatedAt,
			&settingsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}

		// Deserialize settings
		if settingsJSON != nil {
			if err := json.Unmarshal(settingsJSON, &team.Settings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
			}
		}

		teams = append(teams, &team)
	}

	return teams, nil
}

// GetByUser retrieves teams by user ID (teams the user is a member of)
func (tr *TeamRepository) GetByUser(ctx context.Context, userID string) ([]*Team, error) {
	query := `
		SELECT t.id, t.name, t.description, t.owner_id, t.created_at, t.updated_at, t.settings
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1 AND tm.is_active = true
		ORDER BY t.created_at DESC
	`

	rows, err := tr.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams for user: %w", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		var team Team
		var settingsJSON []byte

		err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&team.OwnerID,
			&team.CreatedAt,
			&team.UpdatedAt,
			&settingsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}

		// Deserialize settings
		if settingsJSON != nil {
			if err := json.Unmarshal(settingsJSON, &team.Settings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
			}
		}

		teams = append(teams, &team)
	}

	return teams, nil
}

// Update updates an existing team
func (tr *TeamRepository) Update(ctx context.Context, team *Team) (*Team, error) {
	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(team.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE teams
		SET name = $2, description = $3, updated_at = $4, settings = $5
		WHERE id = $1
		RETURNING id, name, description, owner_id, created_at, updated_at
	`

	var updatedTeam Team
	err = tr.db.QueryRow(ctx, query,
		team.ID,
		team.Name,
		team.Description,
		time.Now(),
		settingsJSON,
	).Scan(
		&updatedTeam.ID,
		&updatedTeam.Name,
		&updatedTeam.Description,
		&updatedTeam.OwnerID,
		&updatedTeam.CreatedAt,
		&updatedTeam.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	// Set settings from original input
	updatedTeam.Settings = team.Settings

	return &updatedTeam, nil
}

// Delete removes a team by ID
func (tr *TeamRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM teams WHERE id = $1"

	result, err := tr.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("team not found: %s", id)
	}

	return nil
}

// AddMember adds a user to a team
func (tr *TeamRepository) AddMember(ctx context.Context, member *TeamMember) (*TeamMember, error) {
	query := `
		INSERT INTO team_members (
			id, team_id, user_id, role, joined_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, team_id, user_id, role, joined_at, is_active
	`

	var createdMember TeamMember
	err := tr.db.QueryRow(ctx, query,
		member.ID,
		member.TeamID,
		member.UserID,
		member.Role,
		member.JoinedAt,
		member.IsActive,
	).Scan(
		&createdMember.ID,
		&createdMember.TeamID,
		&createdMember.UserID,
		&createdMember.Role,
		&createdMember.JoinedAt,
		&createdMember.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to add member to team: %w", err)
	}

	return &createdMember, nil
}

// RemoveMember removes a user from a team
func (tr *TeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	query := "DELETE FROM team_members WHERE team_id = $1 AND user_id = $2"

	result, err := tr.db.Exec(ctx, query, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member from team: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found in team")
	}

	return nil
}

// UpdateMember updates a team member's role or active status
func (tr *TeamRepository) UpdateMember(ctx context.Context, teamID, userID, role string, isActive bool) (*TeamMember, error) {
	query := `
		UPDATE team_members
		SET role = $3, is_active = $4
		WHERE team_id = $1 AND user_id = $2
		RETURNING id, team_id, user_id, role, joined_at, is_active
	`

	var updatedMember TeamMember
	err := tr.db.QueryRow(ctx, query, teamID, userID, role, isActive).Scan(
		&updatedMember.ID,
		&updatedMember.TeamID,
		&updatedMember.UserID,
		&updatedMember.Role,
		&updatedMember.JoinedAt,
		&updatedMember.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update team member: %w", err)
	}

	return &updatedMember, nil
}

// GetMembers retrieves all members of a team
func (tr *TeamRepository) GetMembers(ctx context.Context, teamID string, limit, offset int) ([]*TeamMember, error) {
	query := `
		SELECT id, team_id, user_id, role, joined_at, is_active
		FROM team_members
		WHERE team_id = $1
		ORDER BY joined_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := tr.db.Query(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	var members []*TeamMember
	for rows.Next() {
		var member TeamMember

		err := rows.Scan(
			&member.ID,
			&member.TeamID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		members = append(members, &member)
	}

	return members, nil
}

// GetMember retrieves a specific member from a team
func (tr *TeamRepository) GetMember(ctx context.Context, teamID, userID string) (*TeamMember, error) {
	query := `
		SELECT id, team_id, user_id, role, joined_at, is_active
		FROM team_members
		WHERE team_id = $1 AND user_id = $2
	`

	var member TeamMember

	err := tr.db.QueryRow(ctx, query, teamID, userID).Scan(
		&member.ID,
		&member.TeamID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
		&member.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("member not found in team")
		}
		return nil, fmt.Errorf("failed to get team member: %w", err)
	}

	return &member, nil
}

// CountMembers counts the number of members in a team
func (tr *TeamRepository) CountMembers(ctx context.Context, teamID string) (int64, error) {
	query := "SELECT COUNT(*) FROM team_members WHERE team_id = $1 AND is_active = true"

	var count int64
	err := tr.db.QueryRow(ctx, query, teamID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count team members: %w", err)
	}

	return count, nil
}

// GetTeamCount retrieves the total number of teams
func (tr *TeamRepository) GetTeamCount(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM teams"

	var count int64
	err := tr.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count teams: %w", err)
	}

	return count, nil
}