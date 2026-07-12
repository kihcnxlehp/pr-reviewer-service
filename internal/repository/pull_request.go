package repository

import (
	"context"
	"database/sql"
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

// GetFullPullRequest retrieves a pull request with all its details and assigned reviewers.
// Returns model.ErrNotFound if the PR does not exist.
func (r *PullRequestRepository) GetFullPullRequest(ctx context.Context, prID string) (model.PullRequest, error) {
	var pr model.PullRequest
	var createdAt, mergedAt sql.NullTime

	err := r.pool.QueryRow(ctx, `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
FROM pull_requests
WHERE pull_request_id = $1`, prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.PullRequest{}, model.ErrNotFound
		}
		return model.PullRequest{}, fmt.Errorf("get pull request: %w", err)
	}

	if createdAt.Valid {
		createdAtStr := createdAt.Time.Format(time.RFC3339)
		pr.CreatedAt = &createdAtStr
	}
	if mergedAt.Valid {
		mergedAtStr := mergedAt.Time.Format(time.RFC3339)
		pr.MergedAt = &mergedAtStr
	}

	// Fetch assigned reviewers.
	rows, err := r.pool.Query(ctx, `SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1`, prID)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("get reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return model.PullRequest{}, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, id)
	}
	if err := rows.Err(); err != nil {
		return model.PullRequest{}, fmt.Errorf("iterate reviewers: %w", err)
	}
	pr.AssignedReviewers = reviewers
	return pr, nil
}

// Merge marks a pull request as MERGED and sets merged_at.
// This operation is idempotent: calling it on an already-merged PR returns
// the current state without error.
// Returns the updated PullRequest with all fields and assigned reviewers.
// Returns model.ErrNotFound if the PR does not exist.
func (r *PullRequestRepository) Merge(ctx context.Context, prID string) (model.PullRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check current status.
	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM pull_requests WHERE pull_request_id = $1`,
		prID,
	).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.PullRequest{}, model.ErrNotFound
		}
		return model.PullRequest{}, fmt.Errorf("check status: %w", err)
	}

	//if already merged, just return the current state (idempotent behavior).
	if currentStatus != "MERGED" {
		// Update status and merged it.
		_, err = tx.Exec(ctx, `UPDATE pull_requests
SET status = 'MERGED', merged_at = now()
WHERE pull_request_id = $1
RETURNING merged_at`, prID)
		if err != nil {
			return model.PullRequest{}, fmt.Errorf("update status: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.PullRequest{}, fmt.Errorf("commit transaction: %w", err)
	}

	// Fetch full PR with reviewers.
	return r.GetFullPullRequest(ctx, prID)
}

// GetAuthorAndTeam returns the author_id and team_name for a giver PR.
func (r *PullRequestRepository) GetAuthorAndTeam(ctx context.Context, prID string) (authorID, teamName string, err error) {
	err = r.pool.QueryRow(ctx, `SELECT author_id, team_name
FROM pull_requests p
JOIN users u on u.user_id = p.author_id
WHERE p.pull_request_id = $1`,
		prID,
	).Scan(&authorID, &teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", model.ErrNotFound
		}
		return "", "", fmt.Errorf("get author and team: %w", err)
	}

	return authorID, teamName, nil
}

// GetActiveCandidates returns active users in the team who are not the author and not already reviewers.
func (r *PullRequestRepository) GetActiveCandidates(ctx context.Context, teamName, authorID, prID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT user_id FROM users
WHERE team_name = $1
AND user_id != $2
AND is_active = TRUE
AND user_id NOT IN (SELECT user_id FROM pr_reviewers WHERE pull_request_id = $3)
ORDER BY user_id`, teamName, authorID, prID)
	if err != nil {
		return nil, fmt.Errorf("get active candidates: %w", err)
	}
	defer rows.Close()

	var candidates []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan candidate: %w", err)
		}
		candidates = append(candidates, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate candidates: %w", err)
	}

	return candidates, nil
}

// ReassignReviewer atomically replaces oldReviewerID with newReviewerID.
// Performs all validations inside a transaction with row-level locking.
// Returns model.ErrPRMerged if PR is already merged.
// Returns model.ErrNotAssigned if oldReviewerID is not assigned to the PR.
func (r *PullRequestRepository) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the PR row and check status (prevents concurrent modifications)
	var status string
	err = tx.QueryRow(ctx,
		`SELECT status FROM pull_requests WHERE pull_request_id = $1 FOR UPDATE`,
		prID,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotFound
		}
		return fmt.Errorf("lock and check PR status: %w", err)
	}

	if status == "MERGED" {
		return model.ErrPRMerged
	}

	// Check if oldReviewerID is actually assigned
	var isAssigned bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2)`,
		prID, oldReviewerID,
	).Scan(&isAssigned)
	if err != nil {
		return fmt.Errorf("check reviewer assignment: %w", err)
	}
	if !isAssigned {
		return model.ErrNotAssigned
	}

	var isNewReviewerActive bool
	err = tx.QueryRow(ctx,
		`SELECT is_active FROM users WHERE user_id = $1`,
		newReviewerID,
	).Scan(&isNewReviewerActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNoCandidate // Пользователь не существует
		}
		return fmt.Errorf("check new reviewer active status: %w", err)
	}
	if !isNewReviewerActive {
		return model.ErrNoCandidate // Пользователь деактивирован
	}

	// Remove old reviewer
	_, err = tx.Exec(ctx,
		"DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2",
		prID, oldReviewerID)
	if err != nil {
		return fmt.Errorf("remove old reviewer: %w", err)
	}

	// Insert new reviewer
	_, err = tx.Exec(ctx,
		"INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)",
		prID, newReviewerID)
	if err != nil {
		return fmt.Errorf("insert new reviewer: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
