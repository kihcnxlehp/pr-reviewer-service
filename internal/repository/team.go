// Package repository provides data access layer for the application
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// TeamRepository handles database operations for teams and their members.
type TeamRepository struct {
	pool *pgxpool.Pool
}

// NewTeamRepository creates a new TeamRepository.
func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

// CreateTeam creates a new team and inserts its members in a single transaction.
// Returns model.ErrTeamExists if the team name is already taken.
func (r *TeamRepository) CreateTeam(ctx context.Context, team model.Team) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// Rollback is a no-op after Commit, so defer is safe.
	defer tx.Rollback(ctx)

	// Insert the team.
	_, err = tx.Exec(ctx, "INSERT INTO teams(team_name) VALUES($1)", team.TeamName)
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrTeamExists
		}
		return fmt.Errorf("insert team: %w", err)
	}

	// Insert team members.
	for _, member := range team.Members {
		_, err = tx.Exec(ctx,
			`INSERT INTO users (user_id, username, team_name, is_active)
				VALUES ($1, $2, $3, $4)`, member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// GetTeam returns a team with all its members.
// Returns model.ErrNotFound if the team does not exist.
func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (model.Team, error) {
	// Verify team exists.
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	if err != nil {
		return model.Team{}, fmt.Errorf("check team existence: %w", err)
	}
	if !exists {
		return model.Team{}, model.ErrNotFound
	}

	// Fetch members.
	rows, err := r.pool.Query(ctx,
		`SELECT user_id, username, is_active
		 FROM users
		 WHERE team_name = $1
		 ORDER BY user_id`,
		teamName)
	if err != nil {
		return model.Team{}, fmt.Errorf("query members: %w", err)
	}
	defer rows.Close()

	team := model.Team{
		TeamName: teamName,
		Members:  []model.TeamMember{},
	}
	for rows.Next() {
		var member model.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return model.Team{}, fmt.Errorf("scan member row: %w", err)
		}
		team.Members = append(team.Members, member)
	}
	if err := rows.Err(); err != nil {
		return model.Team{}, fmt.Errorf("iterate members: %w", err)
	}

	return team, nil
}

func isUniqueViolation(err error) bool {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code == "23505"
	}
	return false
}
