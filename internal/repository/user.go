package repository

import (
	"context"
	"fmt"

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
func (repo *UserRepository) UpdateIsActive(ctx context.Context, userID string, isActive bool) error {
	tag, err := repo.pool.Exec(ctx, "UPDATE users SET is_active = $1 WHERE user_id = $2", isActive, userID)
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
