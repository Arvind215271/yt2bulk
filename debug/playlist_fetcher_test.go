package debug

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/config"
	"github.com/joho/godotenv"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

// TestPlaylistFetcher executes the playlist fetching process and prints the URLs.
func TestPlaylistFetcher(t *testing.T) {
	_ = godotenv.Load("../.env")
	cfg := config.Load()
	apiKey := cfg.YouTubeAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("YOUTUBE_API_KEY")
	}
	if apiKey == "" {
		t.Fatal("YOUTUBE_API_KEY environment variable or .env entry is required")
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating YouTube service: %v", err)
	}

	queries := []string{
		"full album playlist", "archive collection", "podcast full episodes",
		"gaming walkthrough full series", "music compilation massive", "history documentary series",
		"complete university lectures", "audiobook collection complete", "top hits collection",
		"live concert archive", "tutorial course complete bundle", "anime full series archive",
	}

	targetCount := 100
	foundCount := 0
	seenPlaylists := make(map[string]bool)

	for _, q := range queries {
		if foundCount >= targetCount {
			break
		}

		var nextPageToken string
		for {
			if foundCount >= targetCount {
				break
			}

			searchCall := service.Search.List([]string{"id"}).
				Q(q).
				Type("playlist").
				MaxResults(50)

			if nextPageToken != "" {
				searchCall = searchCall.PageToken(nextPageToken)
			}

			searchRes, err := searchCall.Do()
			if err != nil {
				log.Printf("Error searching for query '%s': %v", q, err)
				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 429 {
					t.Logf("YouTube API Quota exceeded (429). Stopping fetcher gracefully.")
					return
				}
				break
			}

			var playlistIds []string
			for _, item := range searchRes.Items {
				if item.Id.PlaylistId != "" && !seenPlaylists[item.Id.PlaylistId] {
					seenPlaylists[item.Id.PlaylistId] = true
					playlistIds = append(playlistIds, item.Id.PlaylistId)
				}
			}

			if len(playlistIds) > 0 {
				playlistsRes, err := service.Playlists.List([]string{"contentDetails"}).
					Id(playlistIds...).
					Do()
				if err != nil {
					log.Printf("Error fetching playlist details: %v", err)
					if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 429 {
						t.Logf("YouTube API Quota exceeded (429) on playlist details. Stopping fetcher gracefully.")
						return
					}
					continue
				}

				for _, pl := range playlistsRes.Items {
					if pl.ContentDetails != nil {
						count := pl.ContentDetails.ItemCount
						if count >= 1000 && count <= 2000 {
							fmt.Printf("https://www.youtube.com/playlist?list=%s /* %d videos */\n", pl.Id, count)
							foundCount++
							if foundCount >= targetCount {
								break
							}
						}
					}
				}
			}

			nextPageToken = searchRes.NextPageToken
			if nextPageToken == "" {
				break
			}
		}
	}
}
