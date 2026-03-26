package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockPinger struct{ err error }

func (m *mockPinger) Ping(_ context.Context) error { return m.err }

func TestHandler_ServeHTTP(t *testing.T) {
	pingErr := errors.New("connection refused")

	tests := []struct {
		name          string
		clickupErr    error
		githubErr     error
		wantStatus    int
		wantOverall   string
		wantClickupOK bool
		wantGithubOK  bool
	}{
		{
			name:          "all healthy",
			clickupErr:    nil,
			githubErr:     nil,
			wantStatus:    http.StatusOK,
			wantOverall:   "ok",
			wantClickupOK: true,
			wantGithubOK:  true,
		},
		{
			name:          "clickup down",
			clickupErr:    pingErr,
			githubErr:     nil,
			wantStatus:    http.StatusServiceUnavailable,
			wantOverall:   "degraded",
			wantClickupOK: false,
			wantGithubOK:  true,
		},
		{
			name:          "github down",
			clickupErr:    nil,
			githubErr:     pingErr,
			wantStatus:    http.StatusServiceUnavailable,
			wantOverall:   "degraded",
			wantClickupOK: true,
			wantGithubOK:  false,
		},
		{
			name:          "both down",
			clickupErr:    pingErr,
			githubErr:     pingErr,
			wantStatus:    http.StatusServiceUnavailable,
			wantOverall:   "degraded",
			wantClickupOK: false,
			wantGithubOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockPinger{tt.clickupErr}, &mockPinger{tt.githubErr})

			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %s, want application/json", got)
			}

			var resp healthResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Status != tt.wantOverall {
				t.Errorf("status = %q, want %q", resp.Status, tt.wantOverall)
			}

			clickup := resp.Services["clickup"]
			if tt.wantClickupOK {
				if clickup.Status != "ok" {
					t.Errorf("clickup.status = %q, want ok", clickup.Status)
				}
			} else {
				if clickup.Status != "error" {
					t.Errorf("clickup.status = %q, want error", clickup.Status)
				}
				if clickup.Message == "" {
					t.Error("clickup.message should not be empty on error")
				}
			}

			github := resp.Services["github"]
			if tt.wantGithubOK {
				if github.Status != "ok" {
					t.Errorf("github.status = %q, want ok", github.Status)
				}
			} else {
				if github.Status != "error" {
					t.Errorf("github.status = %q, want error", github.Status)
				}
				if github.Message == "" {
					t.Error("github.message should not be empty on error")
				}
			}
		})
	}
}
