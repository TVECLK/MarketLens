package ingestion

import (
	"testing"

	"marketlens-go-backend/internal/models"
)

func TestNormalizeWorkMode(t *testing.T) {
	tests := []struct {
		in   string
		want models.WorkMode
	}{
		{"remote", models.WorkModeRemote},
		{"Remote", models.WorkModeRemote},
		{" HYBRID ", models.WorkModeHybrid},
		{"onsite", models.WorkModeOnsite},
		{"on-site", models.WorkModeOnsite},
		{"", models.WorkModeOnsite},
		{"work from home", models.WorkModeOnsite},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeWorkMode(tt.in); got != tt.want {
				t.Errorf("normalizeWorkMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
