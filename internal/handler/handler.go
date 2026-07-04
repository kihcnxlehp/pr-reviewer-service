// Package handler provides HTTP handlers for the API.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// Service defines the contract for team-related business logic.
type Service interface {
	CreateTeam(ctx context.Context, team model.Team) (model.Team, error)
	GetTeam(ctx context.Context, teamName string) (model.Team, error)
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	service Service
}

// NewHandler creates a new Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// MaxRequestBodySize limits payload size to 1 MB.
const MaxRequestBodySize = 1 << 20

// Register adds all routes to the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /team/add", h.AddTeam)
	mux.HandleFunc("GET /team/get", h.GetTeam)
	mux.HandleFunc("GET /health", h.Health)
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
		return "invalid request payload"
	case status >= http.StatusInternalServerError:
		// Hide all internal errors.
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
