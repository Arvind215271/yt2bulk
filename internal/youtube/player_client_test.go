package youtube

import (
	"testing"
)

func TestPlayerClient_IsValid(t *testing.T) {
	tests := []struct {
		client   PlayerClient
		expected bool
	}{
		{PlayerClientDefault, true},
		{PlayerClientAndroid, true},
		{PlayerClientAndroidVR, true},
		{PlayerClientVisionOS, true},
		{PlayerClientTVSamsung, true},
		{PlayerClientAll, true},
		{PlayerClientWeb, false},
		{PlayerClient("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.client), func(t *testing.T) {
			if got := tt.client.IsValid(); got != tt.expected {
				t.Errorf("PlayerClient(%s).IsValid() = %v, want %v", tt.client, got, tt.expected)
			}
		})
	}
}
