package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

const baseURL = "http://localhost:8080"

func TestE2E(t *testing.T) {
	if os.Getenv("SKIP_E2E") == "1" {
		t.Skip("SKIP_E2E=1")
	}

	client := &http.Client{Timeout: 5 * time.Second}

	teamName := "backend"
	user1 := uuid.NewString()
	user2 := uuid.NewString()
	user3 := uuid.NewString()

	// Create team
	teamBody := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": user1, "username": "Alice", "is_active": true},
			{"user_id": user2, "username": "Bob", "is_active": true},
			{"user_id": user3, "username": "Carol", "is_active": true},
		},
	}
	resp := post(t, client, "/team/add", teamBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// Create PR
	prID := uuid.NewString()
	prBody := map[string]string{
		"pull_request_id":   prID,
		"pull_request_name": "feat: add X",
		"author_id":         user1,
	}
	resp = post(t, client, "/pullRequest/create", prBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var prResp struct {
		PR struct {
			AssignedReviewers []string `json:"assigned_reviewers"`
		} `json:"pr"`
	}
	json.NewDecoder(resp.Body).Decode(&prResp)
	if len(prResp.PR.AssignedReviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(prResp.PR.AssignedReviewers))
	}

	// Reassign
	reassignBody := map[string]string{
		"pull_request_id": prID,
		"old_reviewer_id": prResp.PR.AssignedReviewers[0],
	}
	resp = post(t, client, "/pullRequest/reassign", reassignBody)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on reassign, got %d", resp.StatusCode)
	}

	// Merge
	mergeBody := map[string]string{"pull_request_id": prID}
	resp = post(t, client, "/pullRequest/merge", mergeBody)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on merge, got %d", resp.StatusCode)
	}

	// Reassign after merge → should fail
	resp = post(t, client, "/pullRequest/reassign", reassignBody)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 after merge, got %d", resp.StatusCode)
	}

	// Stats
	resp = get(t, client, "/stats/reviewers")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stats failed")
	}
}

func post(t *testing.T, client *http.Client, path string, body interface{}) *http.Response {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func get(t *testing.T, client *http.Client, path string) *http.Response {
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
