package video

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

func RunHeavyStopWordUniqueWordTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{
			Name: "heavy_stopword_unique_words",
		}
	}

	rawText := video.Transcript.ToPlainText()
	originalSize := len(rawText)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	cleanText := reg.ReplaceAllString(
		strings.ToLower(rawText),
		"",
	)

	words := strings.Fields(cleanText)

	filtered := make([]string, 0, len(words))

	// Apply heavy stopword filtering.
	for _, word := range words {
		if IsHeavyStopWord(word) {
			continue
		}

		filtered = append(filtered, word)
	}

	// Keep only unique words.
	seen := make(map[string]struct{})
	unique := make([]string, 0, len(filtered))

	for _, word := range filtered {
		if _, exists := seen[word]; exists {
			continue
		}

		seen[word] = struct{}{}
		unique = append(unique, word)
	}

	output := strings.Join(unique, " ")
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) /
			float64(originalSize) * 100
	}

	fmt.Printf("\n=== HEAVY STOPWORD + UNIQUE WORDS ===\n")
	fmt.Printf("Video ID      : %s\n", video.ID)
	fmt.Printf("Title         : %s\n", video.Title)
	fmt.Printf("Original      : %d chars (%.2f KB)\n",
		originalSize,
		float64(originalSize)/1024,
	)
	fmt.Printf("Reduced       : %d chars (%.2f KB)\n",
		resultSize,
		float64(resultSize)/1024,
	)
	fmt.Printf("Unique Words  : %d\n", len(unique))
	fmt.Printf("Reduction     : %.2f%%\n", reduction)
	fmt.Printf("====================================\n")

	return VideoTestResult{
		Name:             "heavy_stopword_unique_words",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}