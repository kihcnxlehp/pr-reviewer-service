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
		return model.User{}, fmt.Errorf("user id is empty: %w", model.ErrInvalidInput)
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
