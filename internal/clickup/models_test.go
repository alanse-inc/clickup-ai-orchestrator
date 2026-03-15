package clickup

import "testing"

func TestIsTriggerStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{StatusReadyForSpec, true},
		{StatusReadyForCode, true},
		{StatusIdeaDraft, false},
		{StatusGeneratingSpec, false},
		{StatusImplementing, false},
		{StatusClosed, false},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := IsTriggerStatus(tt.status); got != tt.want {
				t.Errorf("IsTriggerStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIsProcessingStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{StatusGeneratingSpec, true},
		{StatusImplementing, true},
		{StatusReadyForSpec, false},
		{StatusClosed, false},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := IsProcessingStatus(tt.status); got != tt.want {
				t.Errorf("IsProcessingStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIsTerminalStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{StatusClosed, true},
		{StatusReadyForSpec, false},
		{StatusImplementing, false},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := IsTerminalStatus(tt.status); got != tt.want {
				t.Errorf("IsTerminalStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestPhaseFromStatus(t *testing.T) {
	tests := []struct {
		status  string
		want    Phase
		wantErr bool
	}{
		{StatusReadyForSpec, PhaseSpec, false},
		{StatusGeneratingSpec, PhaseSpec, false},
		{StatusSpecReview, PhaseSpec, false},
		{StatusReadyForCode, PhaseCode, false},
		{StatusImplementing, PhaseCode, false},
		{StatusPRReview, PhaseCode, false},
		{StatusIdeaDraft, "", true},
		{StatusClosed, "", true},
		{"unknown", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got, err := PhaseFromStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("PhaseFromStatus(%q) error = %v, wantErr %v", tt.status, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PhaseFromStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestProcessingStatusFor(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{PhaseSpec, StatusGeneratingSpec},
		{PhaseCode, StatusImplementing},
		{Phase("UNKNOWN"), ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			if got := ProcessingStatusFor(tt.phase); got != tt.want {
				t.Errorf("ProcessingStatusFor(%q) = %q, want %q", tt.phase, got, tt.want)
			}
		})
	}
}

func TestSuccessStatusFor(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{PhaseSpec, StatusSpecReview},
		{PhaseCode, StatusPRReview},
		{Phase("UNKNOWN"), ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			if got := SuccessStatusFor(tt.phase); got != tt.want {
				t.Errorf("SuccessStatusFor(%q) = %q, want %q", tt.phase, got, tt.want)
			}
		})
	}
}

func TestErrorStatusFor(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{PhaseSpec, StatusReadyForSpec},
		{PhaseCode, StatusReadyForCode},
		{Phase("UNKNOWN"), ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			if got := ErrorStatusFor(tt.phase); got != tt.want {
				t.Errorf("ErrorStatusFor(%q) = %q, want %q", tt.phase, got, tt.want)
			}
		})
	}
}
