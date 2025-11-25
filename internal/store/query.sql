-- name: CreateTeam :exec
INSERT INTO teams (name) VALUES ($1) ON CONFLICT (name) DO NOTHING;

-- name: GetTeam :one
SELECT name FROM teams WHERE name = $1;

-- name: GetUsersByTeam :many
SELECT id, username, team_name, is_active FROM users WHERE team_name = $1;

-- name: CreateUser :exec
INSERT INTO users (id, username, team_name, is_active)
VALUES ($1, $2, $3, $4)
    ON CONFLICT (id) DO UPDATE SET
    username = EXCLUDED.username,
                            team_name = EXCLUDED.team_name,
                            is_active = EXCLUDED.is_active;

-- name: GetUser :one
SELECT id, username, team_name, is_active FROM users WHERE id = $1;

-- name: GetActiveUsersInTeamExcluding :many
SELECT id FROM users
WHERE team_name = $1 AND is_active = true AND id != $2;

-- name: CreatePR :exec
INSERT INTO pull_requests (id, name, author_id, status, assigned_reviewers)
VALUES ($1, $2, $3, 'OPEN', $4);

-- name: GetPR :one
SELECT id, name, author_id, status, assigned_reviewers, created_at, merged_at
FROM pull_requests WHERE id = $1;

-- name: MergePR :exec
UPDATE pull_requests
SET status = 'MERGED', merged_at = NOW()
WHERE id = $1 AND status = 'OPEN';

-- name: UpdatePRReviewers :exec
UPDATE pull_requests
SET assigned_reviewers = $2
WHERE id = $1;

-- name: GetPRsByReviewer :many
SELECT id, name, author_id, status
FROM pull_requests
WHERE $1 = ANY(assigned_reviewers);

-- name: GetPRCountByReviewer :many
SELECT unnest(assigned_reviewers) as reviewer_id, COUNT(*) as cnt
FROM pull_requests
WHERE status = 'OPEN'
GROUP BY reviewer_id;

-- name: DeactivateUsers :exec
UPDATE users SET is_active = false WHERE id = ANY($1::text[]);

-- name: GetOpenPRsByReviewers :many
SELECT id, assigned_reviewers FROM pull_requests
WHERE status = 'OPEN' AND $1 && assigned_reviewers;

-- name: SetUserActive :exec
UPDATE users SET is_active = $2 WHERE id = $1;