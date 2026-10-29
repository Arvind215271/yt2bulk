package cache

import (
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/logger"
)

func TestManager_SafeID(t *testing.T) {
	m := NewManager(Config{}, &logger.Logger{})

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://www.youtube.com/@krissk3b",
			expected: "_krissk3b",
		},
		{
			input:    "https://www.youtube.com/channel/UC123",
			expected: "UC123",
		},
		{
			input:    "UC123",
			expected: "UC123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := m.safeID(tt.input)
			if got != tt.expected {
				t.Errorf("safeID(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
