CREATE TABLE teams (
    name TEXT PRIMARY KEY
);

CREATE TABLE users (
     id TEXT PRIMARY KEY,
     username TEXT NOT NULL,
     team_name TEXT NOT NULL REFERENCES teams(name),
     is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE pull_requests (
     id TEXT PRIMARY KEY,
     name TEXT NOT NULL,
     author_id TEXT NOT NULL REFERENCES users(id),
     status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
     assigned_reviewers TEXT[] NOT NULL DEFAULT '{}',
     created_at TIMESTAMPTZ DEFAULT NOW(),
     merged_at TIMESTAMPTZ
);

CREATE INDEX idx_users_team_active ON users(team_name, is_active);
CREATE INDEX idx_pr_reviewers ON pull_requests USING GIN(assigned_reviewers);