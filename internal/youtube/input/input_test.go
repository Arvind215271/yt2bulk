package youtubeurl

import (
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/api"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		wantID         string
		wantType       InputType
		wantNormalized string
		wantErrorCode  string
	}{
		{
			name:           "mobile youtube domain",
			url:            "https://m.youtube.com/playlist?list=PL123",
			wantID:         "PL123",
			wantType:       InputTypePlaylist,
			wantNormalized: "https://www.youtube.com/playlist?list=PL123",
		},
		{
			name:           "youtube.com domain",
			url:            "https://youtube.com/playlist?list=PL123",
			wantID:         "PL123",
			wantType:       InputTypePlaylist,
			wantNormalized: "https://www.youtube.com/playlist?list=PL123",
		},
		{
			name:           "playlist with extra query params",
			url:            "https://www.youtube.com/playlist?list=PL123&si=abcdef",
			wantID:         "PL123",
			wantType:       InputTypePlaylist,
			wantNormalized: "https://www.youtube.com/playlist?list=PL123",
		},
		{
			name:           "playlist from watch url",
			url:            "https://www.youtube.com/watch?v=abc123&list=PL123",
			wantID:         "abc123",
			wantType:       InputTypeVideo,
			wantNormalized: "https://www.youtube.com/watch?v=abc123",
		},
		{
			name:           "standard video url",
			url:            "https://www.youtube.com/watch?v=abc123",
			wantID:         "abc123",
			wantType:       InputTypeVideo,
			wantNormalized: "https://www.youtube.com/watch?v=abc123",
		},
		{
			name:           "short url",
			url:            "https://youtu.be/abc123",
			wantID:         "abc123",
			wantType:       InputTypeVideo,
			wantNormalized: "https://www.youtube.com/watch?v=abc123",
		},
		{
			name:           "shorts url",
			url:            "https://www.youtube.com/shorts/abc123",
			wantID:         "abc123",
			wantType:       InputTypeVideo,
			wantNormalized: "https://www.youtube.com/watch?v=abc123",
		},
		{
			name:          "empty url",
			url:           "",
			wantErrorCode: "YOUTUBE_EMPTY_URL",
		},
		{
			name:          "whitespace url",
			url:           "   ",
			wantErrorCode: "YOUTUBE_EMPTY_URL",
		},
		{
			name:          "invalid url format",
			url:           "://broken-url",
			wantErrorCode: "YOUTUBE_INVALID_URL",
		},
		{
			name:          "missing list query value",
			url:           "https://www.youtube.com/playlist?list=",
			wantErrorCode: "YOUTUBE_MISSING_ID",
		},
		{
			name:          "random text",
			url:           "hello world",
			wantErrorCode: "YOUTUBE_INVALID_DOMAIN",
		},
		{
			name:          "subdomain not youtube",
			url:           "https://music.google.com/playlist?list=PL123",
			wantErrorCode: "YOUTUBE_INVALID_DOMAIN",
		},
		{
			name:          "fake youtube domain",
			url:           "https://youtube-fake.com/playlist?list=PL123",
			wantErrorCode: "YOUTUBE_INVALID_DOMAIN",
		},
		{
			name:          "notyoutube domain",
			url:           "https://notyoutube.com/playlist?list=PL123",
			wantErrorCode: "YOUTUBE_INVALID_DOMAIN",
		},
		{
			name:           "country domain uk",
			url:            "https://youtube.co.uk/playlist?list=PL123",
			wantID:         "PL123",
			wantType:       InputTypePlaylist,
			wantNormalized: "https://www.youtube.com/playlist?list=PL123",
		},
		{
			name:           "country domain de",
			url:            "https://youtube.de/playlist?list=PL123",
			wantID:         "PL123",
			wantType:       InputTypePlaylist,
			wantNormalized: "https://www.youtube.com/playlist?list=PL123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.url)

			if tt.wantErrorCode != "" {
				if err == nil {
					t.Fatalf("expected error %q", tt.wantErrorCode)
				}

				appErr, ok := err.(*api.AppError)
				if !ok {
					t.Fatalf("expected *api.AppError, got %T", err)
				}

				if appErr.Code != tt.wantErrorCode {
					t.Fatalf(
						"expected error code %q, got %q",
						tt.wantErrorCode,
						appErr.Code,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.ID != tt.wantID {
				t.Fatalf(
					"expected id %q, got %q",
					tt.wantID,
					got.ID,
				)
			}

			if got.InputType != tt.wantType {
				t.Fatalf(
					"expected type %q, got %q",
					tt.wantType,
					got.InputType,
				)
			}

			if got.NormalizedURL != tt.wantNormalized {
				t.Fatalf(
					"expected normalized url %q, got %q",
					tt.wantNormalized,
					got.NormalizedURL,
				)
			}
		})
	}
}


type wantItem struct {
	id    string
	err   bool
}

func TestParseMany(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []wantItem
	}{
		{
            name:  "mixed valid and invalid with comma + space + newline",
            input: "https://youtu.be/a1, invalid-url\nhttps://youtube.com/watch?v=b2 https://youtu.be/c3",
            expected: []wantItem{
                {id: "a1", err: false},
                {id: "b2", err: false},
                {id: "c3", err: false},
            },
        },
        {
            name:  "only valid urls with newlines",
            input: "https://youtu.be/a\nhttps://youtu.be/b",
            expected: []wantItem{
                {id: "a"},
                {id: "b"},
            },
        },
        {
             
            // In an extractor pattern, pure text noise is ignored entirely.
            name:     "all invalid noise text",
            input:    "bad1 bad2",
            expected: []wantItem{}, // Expected empty because no URL entities were found
        },
        {
            name:     "empty input",
            input:    "   \n   ",
            expected: []wantItem{},
        },
        {
            name:  "multiple separators together (comma + newline + spaces)",
            input: "https://youtu.be/a1,,\n   https://youtu.be/b2   https://youtu.be/c3",
            expected: []wantItem{
                {id: "a1"},
                {id: "b2"},
                {id: "c3"},
            },
        },
        {
            name:  "trailing commas and spaces",
            input: "https://youtu.be/a1, https://youtu.be/b2,   ",
            expected: []wantItem{
                {id: "a1"},
                {id: "b2"},
            },
        },
        {
            // Only the actual URL entities are extracted and parsed.
            name:  "mixed garbage text and valid urls",
            input: "hello world https://youtu.be/a1 random text https://youtu.be/b2",
            expected: []wantItem{
                {id: "a1"},
                {id: "b2"},
            },
        },
        {
            name:  "duplicate urls",
            input: "https://youtu.be/a https://youtu.be/a https://youtu.be/b",
            expected: []wantItem{
                {id: "a"},
                {id: "a"},
                {id: "b"},
            },
        },
        {
            name:  "empty tokens between separators",
            input: "https://youtu.be/a1,,,   , https://youtu.be/b2",
            expected: []wantItem{
                {id: "a1"},
                {id: "b2"},
            },
        },
        {
            name:  "mixed video and playlist inputs",
            input: "https://youtu.be/a1 https://youtube.com/playlist?list=PL123 https://youtu.be/b2",
            expected: []wantItem{
                {id: "a1"},
                {id: "PL123"},
                {id: "b2"},
            },
        },
        {
            name:  "shorts and playlist mixed",
            input: "https://youtube.com/shorts/a1 https://youtu.be/b2 https://youtube.com/playlist?list=PL999",
            expected: []wantItem{
                {id: "a1"},
                {id: "b2"},
                {id: "PL999"},
            },
        },
        {
            // CHANGED: Complete noise is ignored instead of filling up the error array with ghost items.
            name:     "completely invalid noise input",
            input:    "@@@ ### ??? hello world",
            expected: []wantItem{}, // Evaluates to nothing found
        },
        {
            name:  "urls with tracking params",
            input: "https://youtu.be/a1?si=123 https://youtube.com/watch?v=b2&ab_channel=test",
            expected: []wantItem{
                {id: "a1"},
                {id: "b2"},
            },
        },
        {
            // CHANGED: Conversational junk text around the URLs is completely stripped 
            // out by the regex, leaving only the 4 true YouTube entities to be parsed.
            name:  "real world messy paste scenario",
            input: `
        Hey check these videos:
        https://youtu.be/a1,     https://youtube.com/watch?v=b2

        also this playlist: https://youtube.com/playlist?list=PL123

        ignore this junk -> random text here https://youtu.be/c3
        `,
            expected: []wantItem{
                {id: "a1"},
                {id: "b2"},
                {id: "PL123"},
                {id: "c3"},
            },
        },
    }

		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := ParseMany(tt.input)

			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(got))
			}

			for i, exp := range tt.expected {
				item := got[i]

				if exp.err {
					if item.Error == nil {
						t.Fatalf("expected error at index %d", i)
					}
					continue
				}

				if item.Error != nil {
					t.Fatalf("unexpected error at index %d: %v", i, item.Error)
				}

				if item.Input == nil {
					t.Fatalf("expected input at index %d", i)
				}

				if item.Input.ID != exp.id {
					t.Fatalf(
						"expected id %s at index %d, got %s",
						exp.id, i, item.Input.ID,
					)
				}
			}
		})
	}
}