// Package service contains business logic of the application.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// TeamRepository defines the data access contract for teams.
type TeamRepository interface {
	CreateTeam(ctx context.Context, team model.Team) error
	GetTeam(ctx context.Context, teamName string) (model.Team, error)
}

// TeamService handles team-related business logic.
type TeamService struct {
	repo TeamRepository
}

// NewTeamService creates a new TeamService
func NewTeamService(repo TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

// CreateTeam validates input and creates a team with its members.
func (s *TeamService) CreateTeam(ctx context.Context, team model.Team) (model.Team, error) {
	if err := s.validateTeam(team); err != nil {
		return model.Team{}, err
	}

	if err := s.repo.CreateTeam(ctx, team); err != nil {
		if errors.Is(err, model.ErrTeamExists) {
			return model.Team{}, err
		}
		return model.Team{}, fmt.Errorf("create team: %w", err)
	}

	return team, nil
}

// GetTeam returns a team by name.
func (s *TeamService) GetTeam(ctx context.Context, teamName string) (model.Team, error) {
	if teamName == "" {
		return model.Team{}, fmt.Errorf("%w: team_name is required", model.ErrInvalidInput)
	}

	return s.repo.GetTeam(ctx, teamName)
}

// validateTeam enforces business rules for team creation.
func (s *TeamService) validateTeam(team model.Team) error {
	if team.TeamName == "" {
		return fmt.Errorf("%w: team_name is required", model.ErrInvalidInput)
	}
	if len(team.Members) == 0 {
		return fmt.Errorf("%w: members list cannot be empty", model.ErrInvalidInput)
	}
	for i, m := range team.Members {
		if m.UserID == "" {
			return fmt.Errorf("%w: members[%d].user_id is required", model.ErrInvalidInput, i)
		}
		if m.Username == "" {
			return fmt.Errorf("%w: members[%d].username is required", model.ErrInvalidInput, i)
		}
	}
	return nil
}
