package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// ServicePinger は依存サービスへの疎通確認インターフェース。
type ServicePinger interface {
	Ping(ctx context.Context) error
}

// Handler はヘルスチェックエンドポイントのハンドラ。
type Handler struct {
	clickup ServicePinger
	github  ServicePinger
}

// NewHandler は新しい Handler を生成する。
func NewHandler(clickup, github ServicePinger) *Handler {
	return &Handler{clickup: clickup, github: github}
}

type serviceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type healthResponse struct {
	Status   string                   `json:"status"`
	Services map[string]serviceStatus `json:"services"`
}

type result struct {
	name string
	err  error
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	results := make(chan result, 2)
	pingers := map[string]ServicePinger{
		"clickup": h.clickup,
		"github":  h.github,
	}
	for name, pinger := range pingers {
		go func(name string, p ServicePinger) {
			pingCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			results <- result{name: name, err: p.Ping(pingCtx)}
		}(name, pinger)
	}

	services := make(map[string]serviceStatus, 2)
	degraded := false
	for range pingers {
		res := <-results
		if res.err != nil {
			degraded = true
			services[res.name] = serviceStatus{Status: "error", Message: res.err.Error()}
		} else {
			services[res.name] = serviceStatus{Status: "ok"}
		}
	}

	resp := healthResponse{Services: services}
	statusCode := http.StatusOK
	if degraded {
		resp.Status = "degraded"
		statusCode = http.StatusServiceUnavailable
	} else {
		resp.Status = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}
