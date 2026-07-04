package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// UserRepository defines the data access contracts for users.
type UserRepository interface {
	UpdateIsActive(ctx context.Context, userID string, isActive bool) error
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
func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	if userID == "" {
		return fmt.Errorf("user id is empty: %w", model.ErrInvalidInput)
	}

	if err := s.repo.UpdateIsActive(ctx, userID, isActive); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return err
		}
		return fmt.Errorf("set user is_active: %w", err)
	}
	return nil
}
