CREATE TABLE teams
(
    team_name VARCHAR(100) PRIMARY KEY
);

CREATE TABLE users
(
    user_id   VARCHAR(100) PRIMARY KEY,
    username  VARCHAR(100) NOT NULL,
    team_name VARCHAR(100) NOT NULL REFERENCES teams (team_name),
    is_active BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE TABLE pull_requests
(
    pull_request_id   VARCHAR(100) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id         VARCHAR(100) NOT NULL REFERENCES users (user_id),
    status            VARCHAR(10)  NOT NULL DEFAULT 'OPEN'
        CHECK (status IN ('OPEN', 'MERGED')),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    merged_at         TIMESTAMPTZ
);

CREATE TABLE pr_reviewers
(
    pull_request_id VARCHAR(100) REFERENCES pull_requests (pull_request_id) ON DELETE CASCADE,
    user_id         VARCHAR(100) REFERENCES users (user_id),
    PRIMARY KEY (pull_request_id, user_id)
);