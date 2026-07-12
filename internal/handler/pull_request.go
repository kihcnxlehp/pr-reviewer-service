package handler

import (
	"context"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// PullRequestService defines the contracts for pull request business logic.
type PullRequestService interface {
	CreatePullRequest(ctx context.Context, prID, prName, authorID string) (model.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (model.PullRequest, error)
	ReassignReviewers(ctx context.Context, prID, oldReviewerID string) (model.PullRequest, string, error)
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
	var req createPullRequestReq
	if err := decodeJSON(w, r, &req); err != nil {
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

// mergePullRequest represents the request body for POST /pullRequest/merge.
type mergePullRequestReq struct {
	PullRequestID string `json:"pull_request_id"`
}

// MergePullRequest handles POST /pullRequest/merge.
func (h *PullRequestHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	var req mergePullRequestReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}

	pr, err := h.prService.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"pr": pr})
}

// ReassignPullRequestReq represents the request body for POST /pullRequest/reassign.
type ReassignPullRequestReq struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserId     string `json:"old_user_id"`
}

// ReassignPullRequest handles POST /pullRequest/reassign.
func (h *PullRequestHandler) ReassignPullRequest(w http.ResponseWriter, r *http.Request) {
	var req ReassignPullRequestReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, err)
	}

	pr, replacedBy, err := h.prService.ReassignReviewers(r.Context(), req.PullRequestID, req.OldUserId)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"pr": pr, "replaced_by": replacedBy})
}
