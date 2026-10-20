package yt_bench

// import (
// 	"bufio"
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"testing"

// 	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp"
// )

// func TestFlatPlaylistCount(t *testing.T) {
// 	playlistID := "PLgUwDviBIf0oF6QL8m22w1hIDC1vJ_BHz"
// 	flatOutput, err := FetchPlaylistFlatMetadata(context.Background(), playlistID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	scanner := bufio.NewScanner(bytes.NewReader(flatOutput))
// 	count := 0
// 	for scanner.Scan() {
// 		var entry ytdlp.YTPlaylistEntry
// 		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
// 			count++
// 		}
// 	}
// 	t.Logf("DEBUG: Found %d videos in flat playlist", count)
// }
