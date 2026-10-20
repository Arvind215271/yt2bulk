package subtitle

import (
	"fmt"

	subdomain "github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
)

func BuildPreferences(reqs []SubtitlePreferenceRequest) ([]subdomain.SubtitlePreference, error) {
	if len(reqs) == 0 {
		return subdomain.DefaultPreferences(), nil
	}

	prefs := make([]subdomain.SubtitlePreference, len(reqs))
	for i, r := range reqs {
		pt := subdomain.PreferenceType(r.Type)
		switch pt {
		case subdomain.ManualOnly, subdomain.AutomaticOnly, subdomain.ManualPreferred, subdomain.AutomaticPreferred:
		default:
			return nil, fmt.Errorf("invalid subtitle preference type: %s", r.Type)
		}
		prefs[i] = subdomain.SubtitlePreference{
			Language: r.Language,
			Type:     pt,
		}
	}
	return prefs, nil
}
