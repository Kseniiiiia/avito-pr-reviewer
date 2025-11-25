package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"avito-pr-reviewer/internal/model"
	"avito-pr-reviewer/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/team/add", h.CreateTeam)
	r.Get("/team/get", h.GetTeam)
	r.Post("/users/setIsActive", h.SetActive)
	r.Post("/pullRequest/create", h.CreatePR)
	r.Post("/pullRequest/merge", h.MergePR)
	r.Post("/pullRequest/reassign", h.Reassign)
	r.Get("/users/getReview", h.GetUserReviews)
	r.Get("/stats/reviewers", h.GetStats)
	r.Post("/users/massDeactivate", h.MassDeactivate)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, map[string]string{"status": "OK"})
	})
	return r
}

func writeError(w http.ResponseWriter, r *http.Request, code string, msg string, status int) {
	render.Status(r, status)
	render.JSON(w, r, map[string]map[string]string{
		"error": {"code": code, "message": msg},
	})
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName string       `json:"team_name"`
		Members  []model.User `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, "BAD_REQUEST", "invalid json", http.StatusBadRequest)
		return
	}

	err := h.svc.CreateTeam(r.Context(), req.TeamName, req.Members)
	if err != nil {
		if errors.Is(err, model.ErrTeamExists) {
			writeError(w, r, "TEAM_EXISTS", "team_name already exists", http.StatusBadRequest)
			return
		}
		writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"team": map[string]interface{}{
			"team_name": req.TeamName,
			"members":   req.Members,
		},
	})
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, r, "BAD_REQUEST", "team_name query param required", http.StatusBadRequest)
		return
	}

	team, err := h.svc.GetTeam(r.Context(), teamName)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, r, "NOT_FOUND", "team not found", http.StatusNotFound)
			return
		}
		writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]interface{}{"team": team})
}

func (h *Handler) SetActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, "BAD_REQUEST", "invalid json", http.StatusBadRequest)
		return
	}

	err := h.svc.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, r, "NOT_FOUND", "user not found", http.StatusNotFound)
			return
		}
		writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.svc.GetUser(r.Context(), req.UserID)
	if err != nil {
		writeError(w, r, "INTERNAL_ERROR", "failed to get user", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"user": user,
	})
}

func (h *Handler) MassDeactivate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName string   `json:"team_name"`
		UserIDs  []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, "BAD_REQUEST", "invalid json", http.StatusBadRequest)
		return
	}

	err := h.svc.MassDeactivate(r.Context(), req.TeamName, req.UserIDs)
	if err != nil {
		writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"message":           "users deactivated successfully",
		"deactivated_count": len(req.UserIDs),
	})
}

func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, "BAD_REQUEST", "invalid json", http.StatusBadRequest)
		return
	}

	pr, err := h.svc.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		if errors.Is(err, model.ErrPRExists) {
			writeError(w, r, "PR_EXISTS", "PR id already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, r, "NOT_FOUND", "author/team not found", http.StatusNotFound)
			return
		}
		writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{"pr": pr})
}

func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, "BAD_REQUEST", "invalid json", http.StatusBadRequest)
		return
	}

	pr, err := h.svc.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, r, "NOT_FOUND", "PR not found", http.StatusNotFound)
			return
		}
		writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]interface{}{"pr": pr})
}

func (h *Handler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, "BAD_REQUEST", "invalid json", http.StatusBadRequest)
		return
	}

	newID, pr, err := h.svc.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrNotFound):
			writeError(w, r, "NOT_FOUND", "PR or user not found", http.StatusNotFound)
		case errors.Is(err, model.ErrPRMerged):
			writeError(w, r, "PR_MERGED", "cannot reassign on merged PR", http.StatusConflict)
		case errors.Is(err, model.ErrNotAssigned):
			writeError(w, r, "NOT_ASSIGNED", "reviewer is not assigned to this PR", http.StatusConflict)
		case errors.Is(err, model.ErrNoCandidate):
			writeError(w, r, "NO_CANDIDATE", "no active replacement candidate in team", http.StatusConflict)
		default:
			writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		}
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"pr":          pr,
		"replaced_by": newID,
	})
}

func (h *Handler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, r, "BAD_REQUEST", "user_id query param required", http.StatusBadRequest)
		return
	}

	prs, err := h.svc.GetUserReviews(r.Context(), userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, r, "NOT_FOUND", "user not found", http.StatusNotFound)
			return
		}
		writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		writeError(w, r, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, map[string]interface{}{
		"stats": stats,
	})
}
