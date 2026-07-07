package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// UserService defines the contract for user-related business logic.
type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (model.User, error)
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
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)
	defer r.Body.Close()

	var req SetIsActiveRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
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
