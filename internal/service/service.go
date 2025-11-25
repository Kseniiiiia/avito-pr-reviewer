package service

import (
	"avito-pr-reviewer/internal/model"
	"avito-pr-reviewer/internal/store"
	"avito-pr-reviewer/internal/util"
	"context"
)

type Service struct {
	store store.Repository
}

func New(store store.Repository) *Service {
	return &Service{store: store}
}

func (s *Service) GetUser(ctx context.Context, userID string) (*model.User, error) {
	return s.store.GetUser(ctx, userID)
}

func (s *Service) CreateTeam(ctx context.Context, name string, members []model.User) error {
	err := s.store.CreateTeam(ctx, name)
	if err != nil {
		return model.ErrTeamExists
	}

	for _, m := range members {
		err := s.store.CreateUser(ctx, m.ID, m.Username, name, m.IsActive)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetTeam(ctx context.Context, name string) (*model.Team, error) {
	return s.store.GetTeam(ctx, name)
}

func (s *Service) SetActive(ctx context.Context, userID string, isActive bool) error {
	_, err := s.store.GetUser(ctx, userID)
	if err != nil {
		return model.ErrNotFound
	}

	return s.store.SetUserActive(ctx, userID, isActive)
}

func (s *Service) CreatePR(ctx context.Context, id, name, authorID string) (*model.PullRequest, error) {
	author, err := s.store.GetUser(ctx, authorID)
	if err != nil {
		return nil, model.ErrNotFound
	}

	users, err := s.store.GetActiveUsersInTeamExcluding(ctx, author.TeamName, authorID)
	if err != nil {
		return nil, model.ErrNotFound
	}

	reviewers := make([]string, 0, 2)
	for _, u := range users {
		reviewers = append(reviewers, u)
	}

	if len(reviewers) > 2 {
		util.Shuffle(reviewers)
		reviewers = reviewers[:2]
	}

	err = s.store.CreatePR(ctx, id, name, authorID, reviewers)
	if err != nil {
		return nil, model.ErrPRExists
	}

	pr, err := s.store.GetPR(ctx, id)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (*model.PullRequest, error) {
	pr, err := s.store.GetPR(ctx, prID)
	if err != nil {
		return nil, model.ErrNotFound
	}
	if pr.Status == model.StatusMerged {
		return pr, nil
	}
	err = s.store.MergePR(ctx, prID)
	if err != nil {
		return nil, err
	}

	pr, err = s.store.GetPR(ctx, prID)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldUserID string) (newUserID string, prOut *model.PullRequest, err error) {
	pr, err := s.store.GetPR(ctx, prID)
	if err != nil {
		return "", nil, model.ErrNotFound
	}
	if pr.Status == model.StatusMerged {
		return "", nil, model.ErrPRMerged
	}

	found := false
	for _, r := range pr.AssignedReviewers {
		if r == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return "", nil, model.ErrNotAssigned
	}

	oldUser, err := s.store.GetUser(ctx, oldUserID)
	if err != nil {
		return "", nil, model.ErrNotFound
	}

	candidates, err := s.store.GetActiveUsersInTeamExcluding(ctx, oldUser.TeamName, oldUserID)
	if err != nil {
		return "", nil, err
	}

	available := make([]string, 0)
	for _, c := range candidates {
		isAssigned := false
		for _, r := range pr.AssignedReviewers {
			if c == r {
				isAssigned = true
				break
			}
		}
		if !isAssigned && c != oldUserID {
			available = append(available, c)
		}
	}

	if len(available) == 0 {
		return "", nil, model.ErrNoCandidate
	}

	newUserID = available[0]
	if len(available) > 1 {
		util.Shuffle(available)
		newUserID = available[0]
	}

	newReviewers := make([]string, len(pr.AssignedReviewers))
	for i, r := range pr.AssignedReviewers {
		if r == oldUserID {
			newReviewers[i] = newUserID
		} else {
			newReviewers[i] = r
		}
	}

	err = s.store.UpdatePRReviewers(ctx, prID, newReviewers)
	if err != nil {
		return "", nil, err
	}

	pr.AssignedReviewers = newReviewers
	return newUserID, pr, nil
}

func (s *Service) GetUserReviews(ctx context.Context, userID string) ([]model.PullRequest, error) {
	_, err := s.store.GetUser(ctx, userID)
	if err != nil {
		return nil, model.ErrNotFound
	}

	return s.store.GetPRsByReviewer(ctx, userID)
}

func (s *Service) GetStats(ctx context.Context) (map[string]int64, error) {
	return s.store.GetPRCountByReviewer(ctx)
}

func (s *Service) MassDeactivate(ctx context.Context, teamName string, userIDs []string) error {
	return s.store.DeactivateUsers(ctx, userIDs)
}
