package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

// AddTeam handles POST /team/add
func (h *Handler) AddTeam(w http.ResponseWriter, r *http.Request) {
	// Protect against oversized payloads.
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
func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
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

// Health handles GET /health.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
