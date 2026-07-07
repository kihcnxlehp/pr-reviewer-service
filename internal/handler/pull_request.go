package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// PullRequestService defines the contracts for pull request business logic.
type PullRequestService interface {
	CreatePullRequest(ctx context.Context, prID, prName, authorID string) (model.PullRequest, error)
}

// PullRequestHandler handles HTTP requests for pull request endpoints.
type PullRequestHandler struct {
	prService PullRequestService
}

// NewPullRequestHandler creates a new PullRequestHandler.
func NewPullRequestHandler(prService PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{prService: prService}
}

// createPullRequestReq represents the request body for POST /pullRequest/create.
type createPullRequestReq struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// CreatePullRequest handles POST /pullRequest/create.
func (h *PullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)
	defer r.Body.Close()

	var req createPullRequestReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		writeError(w, err)
		return
	}

	pr, err := h.prService.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"pr": pr})
}
