package model

import (
	"errors"
	"time"
)

var (
	ErrTeamExists  = errors.New("team already exists")
	ErrPRExists    = errors.New("PR already exists")
	ErrPRMerged    = errors.New("cannot reassign on merged PR")
	ErrNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate = errors.New("no active replacement candidate in team")
	ErrNotFound    = errors.New("resource not found")
)

type Status string

const (
	StatusOpen   Status = "OPEN"
	StatusMerged Status = "MERGED"
)

type PullRequest struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            Status     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"created_at"`
	MergedAt          *time.Time `json:"merged_at,omitempty"`
}

type Team struct {
	Name    string `json:"team_name"`
	Members []User `json:"members"`
}

type User struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}
