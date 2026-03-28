package status

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rikeda71/clickup-ai-orchestrator/internal/orchestrator"
)

// StatusProvider はオーケストレータのスナップショットを提供するインターフェース
type StatusProvider interface {
	Status() orchestrator.OrchestratorStatus
}

// LimiterStatus はリミッタの状態を提供するインターフェース
type LimiterStatus interface {
	ActiveCount() int
	MaxConcurrent() int
}

// Handler は /status エンドポイントのハンドラ
type Handler struct {
	limiter   LimiterStatus
	providers []StatusProvider
}

// NewHandler は新しい Handler を生成する
func NewHandler(limiter LimiterStatus, providers []StatusProvider) *Handler {
	return &Handler{limiter: limiter, providers: providers}
}

type runningTaskJSON struct {
	TaskID    string    `json:"task_id"`
	StartedAt time.Time `json:"started_at"`
}

type retryPendingJSON struct {
	TaskID     string    `json:"task_id"`
	Phase      string    `json:"phase"`
	Attempt    int       `json:"attempt"`
	RetryAfter time.Time `json:"retry_after"`
}

type projectStatusJSON struct {
	Project      string             `json:"project"`
	RunningTasks []runningTaskJSON  `json:"running_tasks"`
	RetryPending []retryPendingJSON `json:"retry_pending"`
}

type statusResponse struct {
	ActiveTasks        int                 `json:"active_tasks"`
	MaxConcurrentTasks int                 `json:"max_concurrent_tasks"`
	Projects           []projectStatusJSON `json:"projects"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	projects := make([]projectStatusJSON, 0, len(h.providers))
	for _, p := range h.providers {
		snap := p.Status()

		running := make([]runningTaskJSON, 0, len(snap.RunningTasks))
		for _, rt := range snap.RunningTasks {
			running = append(running, runningTaskJSON{
				TaskID:    rt.TaskID,
				StartedAt: rt.StartedAt,
			})
		}

		retry := make([]retryPendingJSON, 0, len(snap.RetryPending))
		for _, rp := range snap.RetryPending {
			retry = append(retry, retryPendingJSON{
				TaskID:     rp.TaskID,
				Phase:      rp.Phase,
				Attempt:    rp.Attempt,
				RetryAfter: rp.RetryAfter,
			})
		}

		projects = append(projects, projectStatusJSON{
			Project:      snap.Project,
			RunningTasks: running,
			RetryPending: retry,
		})
	}

	var activeTasks, maxConcurrent int
	if h.limiter != nil {
		activeTasks = h.limiter.ActiveCount()
		maxConcurrent = h.limiter.MaxConcurrent()
	}

	resp := statusResponse{
		ActiveTasks:        activeTasks,
		MaxConcurrentTasks: maxConcurrent,
		Projects:           projects,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
