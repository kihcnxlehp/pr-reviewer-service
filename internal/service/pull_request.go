package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// PullRequestRepository defines the data access for pull requests.
type PullRequestRepository interface {
	Create(ctx context.Context, prID, prName, authorID string, reviewersIDs []string) (model.PullRequest, error)
	GetActiveReviewers(ctx context.Context, authorID, teamName string) ([]string, error)
	Merge(ctx context.Context, prID string) (model.PullRequest, error)
	GetPRStatus(ctx context.Context, prID string) (string, error)
	GetAuthorAndTeam(ctx context.Context, prID string) (authorID, teamName string, err error)
	GetPRReviewers(ctx context.Context, prID string) ([]string, error)
	GetActiveCandidates(ctx context.Context, teamName, authorID, prID string) ([]string, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	GetFullPullRequest(ctx context.Context, prID string) (model.PullRequest, error)
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

	candidatesCopy := make([]string, len(candidates))
	copy(candidatesCopy, candidates)

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

// ReassignReviewers replaces a reviewer on a PR with another active team member.
// Returns the updated PR and the new reviewer's user_id.
func (s *PullRequestService) ReassignReviewers(ctx context.Context, prID, oldReviewerID string) (model.PullRequest, string, error) {
	if prID == "" || oldReviewerID == "" {
		return model.PullRequest{}, "", model.ErrInvalidInput
	}

	status, err := s.prRepo.GetPRStatus(ctx, prID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.PullRequest{}, "", err
		}
		return model.PullRequest{}, "", fmt.Errorf("get pull request status: %w", err)
	}

	if status == "MERGED" {
		return model.PullRequest{}, "", model.ErrPRMerged
	}

	reviewers, err := s.prRepo.GetPRReviewers(ctx, prID)
	if err != nil {
		return model.PullRequest{}, "", fmt.Errorf("get  reviewers: %w", err)
	}

	if !contains(reviewers, oldReviewerID) {
		return model.PullRequest{}, "", model.ErrNotAssigned
	}

	authorID, teamName, err := s.prRepo.GetAuthorAndTeam(ctx, prID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.PullRequest{}, "", err
		}
		return model.PullRequest{}, "", fmt.Errorf("get author and team: %w", err)
	}

	candidates, err := s.prRepo.GetActiveCandidates(ctx, teamName, authorID, prID)
	if err != nil {
		return model.PullRequest{}, "", fmt.Errorf("get candidates: %w", err)
	}
	if len(candidates) == 0 {
		return model.PullRequest{}, "", model.ErrNoCandidate
	}

	selected := selectReviewers(candidates, 1)
	newReviewerID := selected[0]

	if err = s.prRepo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		return model.PullRequest{}, "", fmt.Errorf("replace pull request: %w", err)
	}

	pr, err := s.prRepo.GetFullPullRequest(ctx, prID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.PullRequest{}, "", err
		}
		return model.PullRequest{}, "", fmt.Errorf("get pull request: %w", err)
	}

	return pr, newReviewerID, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
