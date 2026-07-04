package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// TeamService defines the contract for team-related business logic.
type TeamService interface {
	CreateTeam(ctx context.Context, team model.Team) (model.Team, error)
	GetTeam(ctx context.Context, teamName string) (model.Team, error)
}

type TeamHandler struct {
	service TeamService
}

func NewTeamHandler(service TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

// AddTeam handles POST /team/add
func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)
	defer r.Body.Close()

	var team model.Team
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&team); err != nil {
		writeError(w, err)
		return
	}

	created, err := h.service.CreateTeam(r.Context(), team)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"team": created})
}

// GetTeam handles GET /team/get
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, model.ErrInvalidInput)
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, team)
}
