package service

import (
	"context"
	"fmt"

	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

type StatsRepository interface {
	GetStatsSummary(ctx context.Context, teamName string) (model.StatsSummary, error)
	GetTopReviewers(ctx context.Context, teamName string) ([]model.UserStat, error)
	GetTopAuthors(ctx context.Context, teamName string) ([]model.UserStat, error)
}

type StatsService struct {
	statsRepo StatsRepository
	teamRepo  TeamRepository
}

func NewStatsService(statsRepo StatsRepository, teamRepo TeamRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo, teamRepo: teamRepo}
}

func (s *StatsService) GetStats(ctx context.Context, teamName string) (model.Stats, error) {
	if teamName != "" {
		exists, err := s.teamRepo.TeamExists(ctx, teamName)
		if err != nil {
			return model.Stats{}, fmt.Errorf("check team existence: %w", err)
		}
		if !exists {
			return model.Stats{}, model.ErrNotFound
		}
	}

	summary, err := s.statsRepo.GetStatsSummary(ctx, teamName)
	if err != nil {
		return model.Stats{}, fmt.Errorf("get summary: %w", err)
	}

	topReviewers, err := s.statsRepo.GetTopReviewers(ctx, teamName)
	if err != nil {
		return model.Stats{}, fmt.Errorf("get top reviewers: %w", err)
	}

	topAuthors, err := s.statsRepo.GetTopAuthors(ctx, teamName)
	if err != nil {
		return model.Stats{}, fmt.Errorf("get top authors: %w", err)
	}

	return model.Stats{Summary: summary, TopReviewers: topReviewers, TopAuthors: topAuthors}, nil
}
