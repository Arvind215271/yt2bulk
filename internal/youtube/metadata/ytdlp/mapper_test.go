package ytdlp

import (
	"reflect"
	"testing"
)

func TestMapToPlaylistSurfaceMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected PlaylistSurfaceMetadata
	}{
		{
			name: "full fields",
			input: map[string]any{
				"id":             "PL123",
				"title":          "Test Playlist",
				"description":    "Test Description",
				"channel":        "Test Channel",
				"channel_id":     "UC123",
				"thumbnail":      "https://example.com/thumb.jpg",
				"tags":           []any{"tag1", "tag2"},
				"webpage_url":    "https://youtube.com/playlist?list=PL123",
				"uploader":       "Test Uploader",
				"playlist_count": float64(10),
			},
			expected: PlaylistSurfaceMetadata{
				ID:            "PL123",
				Title:         "Test Playlist",
				Description:   "Test Description",
				Channel:       "Test Channel",
				ChannelID:     "UC123",
				Thumbnail:     "https://example.com/thumb.jpg",
				Tags:          []string{"tag1", "tag2"},
				WebpageURL:    "https://youtube.com/playlist?list=PL123",
				Uploader:      "Test Uploader",
				PlaylistCount: 10,
			},
		},
		{
			name: "missing optional fields and different types",
			input: map[string]any{
				"id":             "PL456",
				"title":          "Another Playlist",
				"playlist_count": 5,
				"tags":           []string{"tag3"},
			},
			expected: PlaylistSurfaceMetadata{
				ID:            "PL456",
				Title:         "Another Playlist",
				Description:   "",
				Channel:       "",
				ChannelID:     "",
				Thumbnail:     "",
				Tags:          []string{"tag3"},
				WebpageURL:    "",
				Uploader:      "",
				PlaylistCount: 5,
			},
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: PlaylistSurfaceMetadata{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapToPlaylistSurfaceMetadata(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("MapToPlaylistSurfaceMetadata() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}
