package handler

import (
	"context"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// UserService defines the contract for user-related business logic.
type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (model.User, error)
	GetReview(ctx context.Context, userID string) ([]model.PullRequestShort, error)
}

// UserHandler handles HTTP requests for user endpoints.
type UserHandler struct {
	UserService UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

// SetIsActiveRequest represents the request body for POST /users/setIsActive.
type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// SetIsActive handles POST /users/setIsActive.
func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	user, err := h.UserService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// GetReview handles GET /users/getReview?user_id=...
func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, model.ErrInvalidInput)
		return
	}

	prs, err := h.UserService.GetReview(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user_id": userID, "pull_requests": prs})
}
