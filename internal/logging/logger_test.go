package logging

import (
	"context"
	"log/slog"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
	}{
		{name: "debug level", level: slog.LevelDebug},
		{name: "info level", level: slog.LevelInfo},
		{name: "warn level", level: slog.LevelWarn},
		{name: "error level", level: slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			if logger == nil {
				t.Fatal("expected non-nil logger")
			}
		})
	}
}

func TestWithTaskContext(t *testing.T) {
	tests := []struct {
		name   string
		taskID string
		phase  string
	}{
		{name: "spec phase", taskID: "task-001", phase: "SPEC"},
		{name: "code phase", taskID: "task-002", phase: "CODE"},
		{name: "empty values", taskID: "", phase: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := WithTaskContext(ctx, tt.taskID, tt.phase)

			if newCtx == nil {
				t.Fatal("expected non-nil context")
			}

			attrs := TaskAttrsFromContext(newCtx)
			if attrs == nil {
				t.Fatal("expected non-nil attrs from context")
			}
			if len(attrs) != 2 {
				t.Fatalf("expected 2 attrs, got %d", len(attrs))
			}
			if attrs[0].Key != "task_id" || attrs[0].Value.String() != tt.taskID {
				t.Errorf("task_id attr = %q, want %q", attrs[0].Value.String(), tt.taskID)
			}
			if attrs[1].Key != "phase" || attrs[1].Value.String() != tt.phase {
				t.Errorf("phase attr = %q, want %q", attrs[1].Value.String(), tt.phase)
			}
		})
	}
}

func TestTaskAttrsFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	attrs := TaskAttrsFromContext(ctx)
	if attrs != nil {
		t.Errorf("expected nil attrs from empty context, got %v", attrs)
	}
}
