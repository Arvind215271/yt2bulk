package video

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

func RunUniqueWordReductionTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{
			Name: "unique_word_reduction",
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

	seen := make(map[string]struct{})
	unique := make([]string, 0, len(words))

	for _, word := range words {
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

	fmt.Printf("\n=== UNIQUE WORD REDUCTION ===\n")
	fmt.Printf("Video ID     : %s\n", video.ID)
	fmt.Printf("Title        : %s\n", video.Title)
	fmt.Printf("Original     : %d chars (%.2f KB)\n",
		originalSize,
		float64(originalSize)/1024,
	)
	fmt.Printf("Reduced      : %d chars (%.2f KB)\n",
		resultSize,
		float64(resultSize)/1024,
	)
	fmt.Printf("Unique Words : %d\n", len(unique))
	fmt.Printf("Reduction    : %.2f%%\n", reduction)
	fmt.Printf("================================\n")

	return VideoTestResult{
		Name:             "unique_word_reduction",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}