package store

import (
	"avito-pr-reviewer/internal/model"
	"avito-pr-reviewer/internal/store/queries"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	q *queries.Queries
}

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return &PostgresStore{q: queries.New(pool)}, nil
}

func (s *PostgresStore) Close() {
	
}

func (s *PostgresStore) CreateTeam(ctx context.Context, name string) error {
	return s.q.CreateTeam(ctx, name)
}

func (s *PostgresStore) GetTeam(ctx context.Context, name string) (*model.Team, error) {
	_, err := s.q.GetTeam(ctx, name)
	if err != nil {
		return nil, err
	}
	users, err := s.q.GetUsersByTeam(ctx, name)
	if err != nil {
		return nil, err
	}
	members := make([]model.User, len(users))
	for i, u := range users {
		members[i] = model.User{
			ID:       u.ID,
			Username: u.Username,
			TeamName: u.TeamName,
			IsActive: u.IsActive,
		}
	}
	return &model.Team{Name: name, Members: members}, nil
}

func (s *PostgresStore) CreateUser(ctx context.Context, id, username, teamName string, isActive bool) error {
	return s.q.CreateUser(ctx, queries.CreateUserParams{
		ID:       id,
		Username: username,
		TeamName: teamName,
		IsActive: isActive,
	})
}

func (s *PostgresStore) GetUser(ctx context.Context, id string) (*model.User, error) {
	u, err := s.q.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.User{
		ID:       u.ID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}, nil
}

func (s *PostgresStore) GetActiveUsersInTeamExcluding(ctx context.Context, teamName, excludeUserID string) ([]string, error) {
	return s.q.GetActiveUsersInTeamExcluding(ctx, queries.GetActiveUsersInTeamExcludingParams{
		TeamName: teamName,
		ID:       excludeUserID,
	})
}

func (s *PostgresStore) CreatePR(ctx context.Context, id, name, authorID string, reviewers []string) error {
	return s.q.CreatePR(ctx, queries.CreatePRParams{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		AssignedReviewers: reviewers,
	})
}

func (s *PostgresStore) GetPR(ctx context.Context, id string) (*model.PullRequest, error) {
	pr, err := s.q.GetPR(ctx, id)
	if err != nil {
		return nil, err
	}
	var createdAt time.Time
	if pr.CreatedAt.Valid {
		createdAt = pr.CreatedAt.Time
	}
	var mergedAt *time.Time
	if pr.MergedAt.Valid {
		t := pr.MergedAt.Time
		mergedAt = &t
	}
	return &model.PullRequest{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            model.Status(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}, nil
}

func (s *PostgresStore) MergePR(ctx context.Context, id string) error {
	return s.q.MergePR(ctx, id)
}

func (s *PostgresStore) UpdatePRReviewers(ctx context.Context, id string, reviewers []string) error {
	return s.q.UpdatePRReviewers(ctx, queries.UpdatePRReviewersParams{
		ID:                id,
		AssignedReviewers: reviewers,
	})
}

func (s *PostgresStore) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]model.PullRequest, error) {
	rows, err := s.q.GetPRsByReviewer(ctx, []string{reviewerID})
	if err != nil {
		return nil, err
	}
	var res []model.PullRequest
	for _, r := range rows {
		res = append(res, model.PullRequest{
			ID:       r.ID,
			Name:     r.Name,
			AuthorID: r.AuthorID,
			Status:   model.Status(r.Status),
		})
	}
	return res, nil
}

func (s *PostgresStore) GetPRCountByReviewer(ctx context.Context) (map[string]int64, error) {
	rows, err := s.q.GetPRCountByReviewer(ctx)
	if err != nil {
		return nil, err
	}
	stats := make(map[string]int64)
	for _, r := range rows {
		if id, ok := r.ReviewerID.(string); ok {
			stats[id] = r.Cnt
		}
	}
	return stats, nil
}

func (s *PostgresStore) DeactivateUsers(ctx context.Context, ids []string) error {
	return s.q.DeactivateUsers(ctx, ids)
}

func (s *PostgresStore) GetOpenPRsByReviewers(ctx context.Context, reviewerIDs []string) ([]struct {
	ID                string
	AssignedReviewers []string
}, error) {
	rows, err := s.q.GetOpenPRsByReviewers(ctx, reviewerIDs)
	if err != nil {
		return nil, err
	}
	var res []struct {
		ID                string
		AssignedReviewers []string
	}
	for _, r := range rows {
		res = append(res, struct {
			ID                string
			AssignedReviewers []string
		}{
			ID:                r.ID,
			AssignedReviewers: r.AssignedReviewers,
		})
	}
	return res, nil
}

func (s *PostgresStore) SetUserActive(ctx context.Context, userID string, isActive bool) error {
	return s.q.SetUserActive(ctx, queries.SetUserActiveParams{
		ID:       userID,
		IsActive: isActive,
	})
}
