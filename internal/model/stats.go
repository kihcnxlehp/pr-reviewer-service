package model

// StatsSummary contains aggregate counts for the system or a specific team.
type StatsSummary struct {
	TeamsCount         int `json:"teams_count"`
	UsersCount         int `json:"users_count"`
	ActiveUsersCount   int `json:"active_users_count"`
	PullRequestsTotal  int `json:"pull_requests_total"`
	PullRequestsOpen   int `json:"pull_requests_open"`
	PullRequestsMerged int `json:"pull_requests_merged"`
}

// UserStat represents a user with an aggregated metric (review assignment or authored PRs).
type UserStat struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	TeamName      string `json:"team_name"`
	AssignedCount int    `json:"assigned_count,omitempty"`
	PRCount       int    `json:"pr_count,omitempty"`
}

// Stats is the full response for GET /stats.
type Stats struct {
	Summary      StatsSummary `json:"summary"`
	TopReviewers []UserStat   `json:"top_reviewers"`
	TopAuthors   []UserStat   `json:"top_authors"`
}
