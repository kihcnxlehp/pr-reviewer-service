package handler

import (
	"context"
	"net/http"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

type StatsService interface {
	GetStats(ctx context.Context, teamName string) (model.Stats, error)
}

type StatsHandler struct {
	service StatsService
}

func NewStatsHandler(service StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")

	stats, err := h.service.GetStats(r.Context(), teamName)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
