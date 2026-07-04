package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// UserService defines the contract for user-related business logic.
type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) error
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

	err := h.UserService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
