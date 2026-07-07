package model

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Domain errors returned by the service layer.
var (
	ErrInvalidInput = errors.New("invalid request payload")
	ErrTeamExists   = errors.New("team_name already exists")
	ErrNotFound     = errors.New("resource not found")
	ErrPRExists     = errors.New("PR id already exists")
	ErrPRMerged     = errors.New("cannot reassign on merged PR")
	ErrNotAssigned  = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate  = errors.New("no active replacement candidate in team")
)

// ErrorCode maps domain errors to API error codes.
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return "INVALID_PAYLOAD"
	case errors.Is(err, ErrTeamExists):
		return "TEAM_EXISTS"
	case errors.Is(err, ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, ErrPRExists):
		return "PR_EXISTS"
	case errors.Is(err, ErrPRMerged):
		return "PR_MERGED"
	case errors.Is(err, ErrNotAssigned):
		return "NOT_ASSIGNED"
	case errors.Is(err, ErrNoCandidate):
		return "NO_CANDIDATE"
	default:
		// For JSON parsing errors
		var typeErr *json.UnmarshalTypeError
		if _, ok := errors.AsType[*json.SyntaxError](err); ok || errors.As(err, &typeErr) {
			return "INVALID_JSON"
		}
		return "INTERNAL_ERROR"
	}
}

// HTTPStatus maps domain errors to HTTP status codes.
func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrInvalidInput),
		errors.Is(err, ErrTeamExists):
		return http.StatusBadRequest
	case errors.Is(err, ErrPRExists),
		errors.Is(err, ErrPRMerged),
		errors.Is(err, ErrNotAssigned),
		errors.Is(err, ErrNoCandidate):
		return http.StatusConflict
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	default:
		var typeErr *json.UnmarshalTypeError
		if _, ok := errors.AsType[*json.SyntaxError](err); ok || errors.As(err, &typeErr) {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError
	}
}
