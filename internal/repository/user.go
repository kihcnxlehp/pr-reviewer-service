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

// GetPRsByReviewer returns brief information about pull requests where the giver user is assigned as a reviewer.
func (r *UserRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]model.PullRequestShort, error) {
	rows, err := r.pool.Query(ctx, `SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status
FROM pr_reviewers pr
JOIN pull_requests p ON p.pull_request_id = pr.pull_request_id
WHERE pr.user_id = $1
ORDER BY p.created_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("get PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []model.PullRequestShort
	for rows.Next() {
		var pr model.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("scan PR: %w", err)
		}
		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate PRs: %w", err)
	}

	return prs, nil
}
