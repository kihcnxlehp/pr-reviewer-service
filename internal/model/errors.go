package model

import "errors"

// Domain errors returned by the service layer.
var (
	ErrTeamExists  = errors.New("team already exists")
	ErrNotFound    = errors.New("resource not found")
	ErrPRExists    = errors.New("pull request already exists")
	ErrPRMerged    = errors.New("pull request is already merged")
	ErrNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate = errors.New("no active replacement candidate in team")
)

// ErrorCode maps domain errors to API error codes.
func ErrorCode(err error) string {
	switch {
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
		return "INTERNAL_ERROR"
	}
}

// HTTPStatus maps domain errors to HTTP status codes.
func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrTeamExists),
		errors.Is(err, ErrPRExists),
		errors.Is(err, ErrPRMerged),
		errors.Is(err, ErrNotAssigned),
		errors.Is(err, ErrNoCandidate):
		return 409 // Conflict
	case errors.Is(err, ErrNotFound):
		return 404 // Not Found
	default:
		return 500 // Internal Server Error
	}
}
