package clickup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v2/list/list123/task" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "test-token" {
			t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
		}

		resp := map[string]any{
			"tasks": []map[string]any{
				{
					"id":          "task1",
					"name":        "Test Task",
					"description": "desc",
					"status":      map[string]any{"status": "Ready For Spec"},
					"custom_fields": []map[string]any{
						{"name": "GitHub PR URL", "value": "https://github.com/pr/1"},
					},
					"date_created": "1234567890",
					"date_updated": "1234567891",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-token", "list123")
	client.httpClient = server.Client()
	tasks, err := getTasksWithBaseURL(client, context.Background(), server.URL+"/api/v2")
	if err != nil {
		t.Fatalf("GetTasks() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	task := tasks[0]
	if task.ID != "task1" {
		t.Errorf("expected ID task1, got %s", task.ID)
	}
	if task.Name != "Test Task" {
		t.Errorf("expected name 'Test Task', got %s", task.Name)
	}
	if task.Status != "ready for spec" {
		t.Errorf("expected status 'ready for spec', got %s", task.Status)
	}
	if task.CustomFields["github_pr_url"] != "https://github.com/pr/1" {
		t.Errorf("expected custom field github_pr_url, got %v", task.CustomFields)
	}
}

func TestGetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v2/task/task1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := map[string]any{
			"id":            "task1",
			"name":          "Single Task",
			"description":   "single desc",
			"status":        map[string]any{"status": "Implementing"},
			"custom_fields": []map[string]any{},
			"date_created":  "1234567890",
			"date_updated":  "1234567891",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-token", "list123")
	client.httpClient = server.Client()

	task, err := getTaskWithBaseURL(client, context.Background(), server.URL+"/api/v2", "task1")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if task.ID != "task1" {
		t.Errorf("expected ID task1, got %s", task.ID)
	}
	if task.Status != "implementing" {
		t.Errorf("expected status 'implementing', got %s", task.Status)
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/v2/task/task1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]string
		_ = json.Unmarshal(body, &payload)
		if payload["status"] != "implementing" {
			t.Errorf("expected status 'implementing', got %s", payload["status"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", "list123")
	client.httpClient = server.Client()

	err := updateTaskStatusWithBaseURL(client, context.Background(), server.URL+"/api/v2", "task1", "implementing")
	if err != nil {
		t.Fatalf("UpdateTaskStatus() error = %v", err)
	}
}

func TestGetTasksErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient("test-token", "list123")
	client.httpClient = server.Client()

	_, err := getTasksWithBaseURL(client, context.Background(), server.URL+"/api/v2")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestGetTaskErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient("test-token", "list123")
	client.httpClient = server.Client()

	_, err := getTaskWithBaseURL(client, context.Background(), server.URL+"/api/v2", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestUpdateTaskStatusErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient("test-token", "list123")
	client.httpClient = server.Client()

	err := updateTaskStatusWithBaseURL(client, context.Background(), server.URL+"/api/v2", "task1", "closed")
	if err == nil {
		t.Fatal("expected error for 403 response, got nil")
	}
}

// テスト用ヘルパー: baseURLを差し替え可能にする
func getTasksWithBaseURL(c *Client, ctx context.Context, base string) ([]Task, error) {
	url := base + "/list/" + c.listID + "/task"
	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result apiTasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	tasks := make([]Task, len(result.Tasks))
	for i, t := range result.Tasks {
		tasks[i] = t.toTask()
	}
	return tasks, nil
}

func getTaskWithBaseURL(c *Client, ctx context.Context, base string, taskID string) (*Task, error) {
	url := base + "/task/" + taskID
	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var t apiTask
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}

	task := t.toTask()
	return &task, nil
}

func updateTaskStatusWithBaseURL(c *Client, ctx context.Context, base string, taskID string, status string) error {
	url := base + "/task/" + taskID
	body := `{"status":"` + status + `"}`
	resp, err := c.doRequest(ctx, http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
