package video

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

var windowSize = 300.0

type WindowWordStat struct {
	Word     string
	Count    int
	Duration float64
}

func RunWindowWordStatsTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{Name: "window_word_stats"}
	}

	rawText := video.Transcript.ToPlainText()
	originalSize := len(rawText)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	windowMap := make(map[int]map[string]*WindowWordStat)

	for _, seg := range video.Transcript.Segments {

		windowIdx := int(seg.Start / windowSize)

		if windowMap[windowIdx] == nil {
			windowMap[windowIdx] = make(map[string]*WindowWordStat)
		}

		clean := reg.ReplaceAllString(strings.ToLower(seg.Text), "")
		words := strings.Fields(clean)

		if len(words) == 0 {
			continue
		}

		duration := seg.End - seg.Start
		perWordDuration := duration / float64(len(words))

		for _, w := range words {

			if IsHeavyStopWord(w) {
				continue
			}

			if windowMap[windowIdx][w] == nil {
				windowMap[windowIdx][w] = &WindowWordStat{Word: w}
			}

			windowMap[windowIdx][w].Count++
			windowMap[windowIdx][w].Duration += perWordDuration
		}
	}

	// sort windows
	var windowKeys []int
	for k := range windowMap {
		windowKeys = append(windowKeys, k)
	}
	sort.Ints(windowKeys)

	var b strings.Builder

	for _, idx := range windowKeys {

		start := float64(idx) * windowSize
		end := start + windowSize

		b.WriteString(fmt.Sprintf(
			"\nWINDOW %d (%.0fs - %.0fs)\n",
			idx, start, end,
		))

		b.WriteString("word,count,duration\n")

		var words []WindowWordStat
		for _, w := range windowMap[idx] {
			words = append(words, *w)
		}

		sort.Slice(words, func(i, j int) bool {
			return words[i].Duration > words[j].Duration
		})

		for _, w := range words {

			// REMOVE ultra-noise words
			if w.Count <= 3 && w.Duration < 2 {
				continue
			}

			b.WriteString(fmt.Sprintf(
				"%s,%d,%.0f\n",
				w.Word,
				w.Count,
				w.Duration,
			))
		}
	}

	output := b.String()
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) / float64(originalSize) * 100
	}

	fmt.Printf("\n=== WINDOW WORD STATS (BLOCK FORMAT) ===\n")
	fmt.Printf("Video ID   : %s\n", video.ID)
	fmt.Printf("Windows    : %d\n", len(windowKeys))
	fmt.Printf("Original   : %d\n", originalSize)
	fmt.Printf("Reduced    : %d\n", resultSize)
	fmt.Printf("Reduction  : %.2f%%\n", reduction)
	fmt.Printf("=========================================\n")

	return VideoTestResult{
		Name:             "window_word_stats",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}