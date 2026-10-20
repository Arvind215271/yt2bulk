package subtitle

import (
	"testing"

	subdomain "github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
)

func TestBuildPreferences(t *testing.T) {
	t.Run("Empty preferences returns defaults", func(t *testing.T) {
		prefs, err := BuildPreferences(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prefs) == 0 {
			t.Fatal("expected default preferences, got empty")
		}
	})

	t.Run("Valid preference mapping", func(t *testing.T) {
		reqs := []SubtitlePreferenceRequest{
			{Language: "en", Type: "manual>automatic"},
		}
		prefs, err := BuildPreferences(reqs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(prefs) != 1 || prefs[0].Type != subdomain.ManualPreferred {
			t.Errorf("unexpected mapped preference: %+v", prefs[0])
		}
	})

	t.Run("Invalid preference type", func(t *testing.T) {
		reqs := []SubtitlePreferenceRequest{
			{Language: "en", Type: "invalid-type"},
		}
		_, err := BuildPreferences(reqs)
		if err == nil {
			t.Fatal("expected error for invalid type, got nil")
		}
	})
}
