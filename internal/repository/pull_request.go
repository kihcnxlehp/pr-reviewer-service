package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// PullRequestRepository handles database operations for pull requests and reviewers.
type PullRequestRepository struct {
	pool *pgxpool.Pool
}

// NewPullRequestRepository creates a new PullRequestRepository.
func NewPullRequestRepository(pool *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{pool: pool}
}

// GetActiveReviewers return user_ids of active users in the team, excluding the author.
func (r *PullRequestRepository) GetActiveReviewers(ctx context.Context, authorID, teamName string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT user_id FROM users 
               WHERE team_name = $1 AND user_id != $2 AND is_active = TRUE 
               ORDER BY user_id`,
		teamName, authorID)
	if err != nil {
		return nil, fmt.Errorf("get active reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan reviewer id: %w", err)
		}
		reviewers = append(reviewers, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewers: %w", err)
	}

	return reviewers, nil
}

// Create inserts a pull request and its reviewers in a single transaction.
// Verifies that the author is active inside the transaction to prevent race conditions.
// Returns the created PullRequest with all fields populated from the database.
// Returns model.ErrPRExists if PR ID already exists.
// Returns model.ErrNotFound if the author does not exist or is not active.
func (r *PullRequestRepository) Create(ctx context.Context, prID, prName, authorID string, reviewersIDs []string) (model.PullRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verify author is active (inside transaction to prevent race conditions).
	var isActive bool
	err = tx.QueryRow(ctx, "SELECT is_active FROM users WHERE user_id=$1", authorID).Scan(&isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.PullRequest{}, model.ErrNotFound
		}
		return model.PullRequest{}, fmt.Errorf("verify author: %w", err)
	}
	if !isActive {
		return model.PullRequest{}, model.ErrNotFound
	}

	// Insert pull request and return created_at.
	var createdAt time.Time
	err = tx.QueryRow(ctx, `INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
VALUES ($1, $2, $3, 'OPEN') RETURNING created_at`,
		prID, prName, authorID,
	).Scan(&createdAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.PullRequest{}, model.ErrPRExists
		}
		return model.PullRequest{}, fmt.Errorf("insert pull request: %w", err)
	}

	// Insert assigned reviewers.
	for _, reviewerID := range reviewersIDs {
		_, err = tx.Exec(ctx, "INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)",
			prID, reviewerID)
		if err != nil {
			return model.PullRequest{}, fmt.Errorf("insert reviewer %s: %w", reviewerID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return model.PullRequest{}, fmt.Errorf("commit transaction: %w", err)
	}

	// Build the response.
	createdAtStr := createdAt.Format(time.RFC3339)
	return model.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: reviewersIDs,
		CreatedAt:         &createdAtStr,
	}, nil
}
