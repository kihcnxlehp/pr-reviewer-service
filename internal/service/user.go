package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// UserRepository defines the data access contracts for users.
type UserRepository interface {
	UpdateIsActive(ctx context.Context, userID string, isActive bool) (model.User, error)
	GetTeam(ctx context.Context, userID string) (string, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]model.PullRequestShort, error)
}

// UserService handles user-related business logic.
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// SetIsActive updates the active status of a user.
func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (model.User, error) {
	if userID == "" {
		return model.User{}, fmt.Errorf("%w: user_id is required", model.ErrInvalidInput)
	}

	user, err := s.repo.UpdateIsActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("set user is_active: %w", err)
	}

	return user, nil
}

// GetReview return pull requests where the user is assigned as a reviewer.
func (s *UserService) GetReview(ctx context.Context, userID string) ([]model.PullRequestShort, error) {
	prs, err := s.repo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get prs by reviewer: %w", err)
	}

	return prs, nil
}
