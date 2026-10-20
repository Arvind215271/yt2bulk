package video

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

type WordStat struct {
	Word     string  `json:"word"`
	Count    int     `json:"count"`
	Duration float64 `json:"duration"`
}

func RunWordStatsTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{
			Name: "word_stats",
		}
	}

	rawText := video.Transcript.ToPlainText()
	originalSize := len(rawText)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	stats := make(map[string]*WordStat)

	for _, seg := range video.Transcript.Segments {
		duration := seg.End - seg.Start

		clean := reg.ReplaceAllString(
			strings.ToLower(seg.Text),
			"",
		)

		words := strings.Fields(clean)

		for _, word := range words {
			if _, ok := stats[word]; !ok {
				stats[word] = &WordStat{
					Word: word,
				}
			}

			stats[word].Count++
			stats[word].Duration += duration
		}
	}

	var result []WordStat

	for _, stat := range stats {
		result = append(result, *stat)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Duration > result[j].Duration
	})

	var b strings.Builder

	limit := len(result)

	fmt.Fprintf(&b, "Word, Count, Duration\n")
	for _, stat := range result[:limit] {
		fmt.Fprintf(
			&b,
			"%s, %d, %.0fs\n",
			stat.Word,
			stat.Count,
			stat.Duration,
		)
	}

	output := b.String()
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) /
			float64(originalSize) * 100
	}

	fmt.Printf("\n=== WORD STATS ===\n")
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
	fmt.Printf("Words        : %d\n", len(result))
	fmt.Printf("Top Exported : %d\n", limit)
	fmt.Printf("Reduction    : %.2f%%\n", reduction)
	fmt.Printf("================================\n")

	return VideoTestResult{
		Name:             "word_stats",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}