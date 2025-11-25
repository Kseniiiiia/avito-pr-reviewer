package store

import (
	"avito-pr-reviewer/internal/model"
	"context"
)

type Repository interface {
	CreateTeam(ctx context.Context, name string) error
	GetTeam(ctx context.Context, name string) (*model.Team, error)
	CreateUser(ctx context.Context, id, username, teamName string, isActive bool) error
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetActiveUsersInTeamExcluding(ctx context.Context, teamName, excludeUserID string) ([]string, error)
	CreatePR(ctx context.Context, id, name, authorID string, reviewers []string) error
	GetPR(ctx context.Context, id string) (*model.PullRequest, error)
	MergePR(ctx context.Context, id string) error
	UpdatePRReviewers(ctx context.Context, id string, reviewers []string) error
	GetPRsByReviewer(ctx context.Context, reviewerID string) ([]model.PullRequest, error)
	GetPRCountByReviewer(ctx context.Context) (map[string]int64, error)
	DeactivateUsers(ctx context.Context, ids []string) error
	GetOpenPRsByReviewers(ctx context.Context, reviewerIDs []string) ([]struct {
		ID                string
		AssignedReviewers []string
	}, error)
	SetUserActive(ctx context.Context, userID string, isActive bool) error
}
