CREATE INDEX idx_users_team_active ON users (team_name) WHERE is_active = TRUE;
CREATE INDEX idx_pr_reviewers_user ON pr_reviewers (user_id);
CREATE INDEX idx_pr_author ON pull_requests (author_id);