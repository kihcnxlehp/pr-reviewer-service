package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// PullRequestRepository defines the data access for pull requests.
type PullRequestRepository interface {
	Create(ctx context.Context, prID, prName, authorID string, reviewersIDs []string) (model.PullRequest, error)
	GetActiveReviewers(ctx context.Context, authorID, teamName string) ([]string, error)
	Merge(ctx context.Context, prID string) (model.PullRequest, error)
}

// PullRequestService handles pull request business logic.
type PullRequestService struct {
	prRepo   PullRequestRepository
	userRepo UserRepository
}

// NewPullRequestService creates a new PullRequestService.
func NewPullRequestService(prRepo PullRequestRepository, userRepo UserRepository) *PullRequestService {
	return &PullRequestService{prRepo: prRepo, userRepo: userRepo}
}

// CreatePullRequest validates input, resolves the author's team, selects up to 2 random
// reviewers from the team (excluding the author), and creates the PR in a transaction.
func (s *PullRequestService) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (model.PullRequest, error) {
	// Validation.
	if prID == "" || prName == "" || authorID == "" {
		return model.PullRequest{}, fmt.Errorf("%w: pull_request_id, pull_request_name and author_id are required", model.ErrInvalidInput)
	}

	// 1. Resolve author's team (also verifies the author exists).
	teamName, err := s.userRepo.GetTeam(ctx, authorID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.PullRequest{}, model.ErrNotFound
		}
		return model.PullRequest{}, fmt.Errorf("get author team: %w", err)
	}

	// 2. Fetch active candidates in the same team, excluding the author.
	candidates, err := s.prRepo.GetActiveReviewers(ctx, authorID, teamName)
	if err != nil {
		return model.PullRequest{}, fmt.Errorf("get active reviewers: %w", err)
	}

	// 3. Randomly select up to 2 reviewers.
	selected := selectReviewers(candidates, 2)

	// 4. Crate PR + reviewers in a single transaction.
	pr, err := s.prRepo.Create(ctx, prID, prName, authorID, selected)
	if err != nil {
		if errors.Is(err, model.ErrPRExists) || errors.Is(err, model.ErrNotFound) {
			return model.PullRequest{}, err
		}
		return model.PullRequest{}, fmt.Errorf("create pull request: %w", err)
	}

	return pr, nil
}

// selectReviewers randomly picks up to max reviewers from the candidates slice.
// If there are fewer candidates than max, returns all of them (possibly empty).
func selectReviewers(candidates []string, max int) []string {
	if len(candidates) <= max {
		result := make([]string, len(candidates))
		copy(result, candidates)
		return result
	}
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	result := make([]string, max)
	copy(result, candidates[:max])
	return result
}

// MergePullRequest marks a pull request as merged.
// Returns the updated PullRequest with merged_at timestamp.
func (s *PullRequestService) MergePullRequest(ctx context.Context, prID string) (model.PullRequest, error) {
	if prID == "" {
		return model.PullRequest{}, fmt.Errorf("%w: pull_request_id is required", model.ErrInvalidInput)
	}

	pr, err := s.prRepo.Merge(ctx, prID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.PullRequest{}, err
		}
		return model.PullRequest{}, fmt.Errorf("merge pull request: %w", err)
	}

	return pr, nil
}
