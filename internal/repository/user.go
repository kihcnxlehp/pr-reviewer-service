package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// UpdateIsActive updates the is_active status of user.
// Returns model.ErrNotFound if the user doesn't exist.
func (r *UserRepository) UpdateIsActive(ctx context.Context, userID string, isActive bool) (model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `UPDATE users
SET is_active = $1
WHERE user_id = $2
RETURNING user_id, username, team_name, is_active`,
		isActive, userID,
	).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, model.ErrNotFound
		}
		return model.User{}, fmt.Errorf("update user is_active: %w", err)
	}
	return user, nil
}

// GetTeam returns the team name of the user.
// Returns model.ErrNotFound if the user doesn't exist.
func (r *UserRepository) GetTeam(ctx context.Context, userID string) (string, error) {
	var teamName string
	err := r.pool.QueryRow(ctx, "SELECT team_name FROM users WHERE user_id = $1", userID).Scan(&teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrNotFound
		}
		return "", fmt.Errorf("get user team: %w", err)
	}

	return teamName, nil
}
