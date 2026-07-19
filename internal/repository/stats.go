package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kihcnxlehp/pr-reviewer-service/internal/model"
)

type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

func (r *StatsRepository) GetStatsSummary(ctx context.Context, teamName string) (model.StatsSummary, error) {
	var summary model.StatsSummary

	if teamName == "" {
		err := r.pool.QueryRow(ctx, `
			SELECT 
				(SELECT COUNT(*) FROM teams)::int,
				(SELECT COUNT(*) FROM users)::int,
				(SELECT COUNT(*) FROM users WHERE is_active = TRUE)::int,
				COUNT(*)::int,
				COUNT(*) FILTER (WHERE status = 'OPEN')::int,
				COUNT(*) FILTER (WHERE status = 'MERGED')::int
			FROM pull_requests
		`).Scan(
			&summary.TeamsCount, &summary.UsersCount, &summary.ActiveUsersCount,
			&summary.PullRequestsTotal, &summary.PullRequestsOpen, &summary.PullRequestsMerged,
		)
		if err != nil {
			return model.StatsSummary{}, fmt.Errorf("get global summary: %w", err)
		}

		return summary, nil
	}

	err := r.pool.QueryRow(ctx, `SELECT
1,
(SELECT COUNT(*) FROM users WHERE team_name = $1),
(SELECT COUNT(*) FROM users WHERE is_active = TRUE AND team_name = $1),
(SELECT COUNT(*) FROM pull_requests AS pr JOIN users AS u ON u.user_id = pr.author_id WHERE u.team_name = $1),
(SELECT COUNT(*) FROM pull_requests AS pr JOIN users AS u ON u.user_id = pr.author_id WHERE u.team_name = $1 AND pr.status = 'OPEN'),
(SELECT COUNT(*) FROM pull_requests AS pr JOIN users AS u ON u.user_id = pr.author_id WHERE u.team_name = $1 AND pr.status = 'MERGED')
`, teamName).Scan(
		&summary.TeamsCount,
		&summary.UsersCount,
		&summary.ActiveUsersCount,
		&summary.PullRequestsTotal,
		&summary.PullRequestsOpen,
		&summary.PullRequestsMerged,
	)
	if err != nil {
		return model.StatsSummary{}, fmt.Errorf("get team summary: %w", err)
	}

	return summary, nil
}

func (r *StatsRepository) GetTopReviewers(ctx context.Context, teamName string) ([]model.UserStat, error) {
	query := `
		SELECT u.user_id, u.username, u.team_name, COUNT(*) AS assigned_count
		FROM pr_reviewers AS p
		JOIN users AS u ON u.user_id = p.user_id
	`
	var args []any
	if teamName != "" {
		query += " WHERE u.team_name = $1"
		args = append(args, teamName)
	}
	query += `
		GROUP BY u.user_id, u.username, u.team_name
		ORDER BY assigned_count DESC, u.user_id ASC
		LIMIT 10`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get top reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []model.UserStat
	for rows.Next() {
		var reviewer model.UserStat
		if err := rows.Scan(&reviewer.UserID, &reviewer.Username, &reviewer.TeamName, &reviewer.AssignedCount); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewer)
	}

	if reviewers == nil {
		reviewers = []model.UserStat{}
	}
	return reviewers, nil
}

func (r *StatsRepository) GetTopAuthors(ctx context.Context, teamName string) ([]model.UserStat, error) {
	query := `
		SELECT u.user_id, u.username, u.team_name, COUNT(*) AS pr_count
		FROM pull_requests AS pr
		JOIN users AS u ON u.user_id = pr.author_id
	`
	var args []any
	if teamName != "" {
		query += " WHERE u.team_name = $1"
		args = append(args, teamName)
	}
	query += `
		GROUP BY u.user_id, u.username, u.team_name
		ORDER BY pr_count DESC, u.user_id ASC
		LIMIT 10`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get top authors: %w", err)
	}
	defer rows.Close()

	var authors []model.UserStat
	for rows.Next() {
		var author model.UserStat
		if err := rows.Scan(&author.UserID, &author.Username, &author.TeamName, &author.PRCount); err != nil {
			return nil, fmt.Errorf("scan author: %w", err)
		}
		authors = append(authors, author)
	}

	return authors, nil
}
