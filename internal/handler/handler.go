// Package handler provides HTTP handlers for the API.
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	team        *TeamHandler
	user        *UserHandler
	pullRequest *PullRequestHandler
	stats       *StatsHandler
}

// New creates a new Handler.
func New(team *TeamHandler, user *UserHandler, pr *PullRequestHandler, stats *StatsHandler) *Handler {
	return &Handler{team: team, user: user, pullRequest: pr, stats: stats}
}

// MaxRequestBodySize limits payload size to 1 MB.
const MaxRequestBodySize = 1 << 20

// Register adds all routes to the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// Team routes
	mux.HandleFunc("POST /team/add", h.team.AddTeam)
	mux.HandleFunc("GET /team/get", h.team.GetTeam)

	// User routes
	mux.HandleFunc("POST /users/setIsActive", h.user.SetIsActive)
	mux.HandleFunc("GET /users/getReview", h.user.GetReview)

	// PullRequest routes
	mux.HandleFunc("POST /pullRequest/create", h.pullRequest.CreatePullRequest)
	mux.HandleFunc("POST /pullRequest/merge", h.pullRequest.MergePullRequest)
	mux.HandleFunc("POST /pullRequest/reassign", h.pullRequest.ReassignPullRequest)

	// Stats route
	mux.HandleFunc("GET /stats", h.stats.GetStats)

	// Common routes
	mux.HandleFunc("GET /health", h.Health)
}

// Health handles GET /health.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeError writes a JSON error response in OpenAPI format.
func writeError(w http.ResponseWriter, err error) {
	status := model.HTTPStatus(err)
	code := model.ErrorCode(err)
	message := safeMessage(err, status)

	response := ErrorResponse{
		Code:    code,
		Message: message,
	}

	writeJSON(w, status, map[string]any{"error": response})
}

// writeJSON writes a successful JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

// safeMessage returns a client-safe error message based on the error type and HTTP status.
func safeMessage(err error, status int) string {
	switch {
	case isJSONParseError(err):
		return "invalid JSON payload"
	case errors.Is(err, model.ErrInvalidInput):
		// Strip internal details (e.g. "team_name is required"), keep only the generic message.
		log.Printf("invalid input: %v", err)
		return "invalid request payload"
	case status >= http.StatusInternalServerError:
		// Hide all internal errors.
		log.Printf("internal error: %v", err)
		return "internal server error"
	default:
		// Domain errors (ErrTeamExists, ErrNotFound, etc.) are safe to expose.
		return err.Error()
	}
}

// isJSONParseError checks whether the error is a JSON parsing error.
func isJSONParseError(err error) bool {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	return errors.As(err, &syntaxErr) || errors.As(err, &typeErr)
}

// decodeJSON reads and validates JSON request body.
// It applies MaxBytesReader limit and disallows unknown fields.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return err
	}

	return nil
}
