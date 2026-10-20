package video

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

func RunHeavyWordStatsBucketFormatTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{
			Name: "word_stats_bucket_format",
		}
	}

	rawText := video.Transcript.ToPlainText()
	originalSize := len(rawText)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	type agg struct {
		Word     string
		Count    int
		Duration float64
	}

	stats := make(map[string]*agg)

	totalWords := 0

	// -----------------------------
	// BUILD STATS FIRST
	// -----------------------------
	for _, seg := range video.Transcript.Segments {
		duration := seg.End - seg.Start

		clean := reg.ReplaceAllString(strings.ToLower(seg.Text), "")
		words := strings.Fields(clean)

		for _, w := range words {
			if _, ok := heavyStopWords[w]; ok {
				continue
			}

			if _, ok := stats[w]; !ok {
				stats[w] = &agg{Word: w}
			}

			stats[w].Count++
			stats[w].Duration += duration
			totalWords++
		}
	}

	// -----------------------------
	// ADAPTIVE FREQUENCY FILTER
	// -----------------------------
	avgFreq := float64(totalWords) / float64(len(stats)+1)

	minFreq := 1
	switch {
	case avgFreq > 8:
		minFreq = 3
	case avgFreq > 4:
		minFreq = 2
	default:
		minFreq = 1
	}

	// prune low-frequency words (IMPORTANT FIX)
	for w, s := range stats {
		if s.Count < minFreq {
			delete(stats, w)
		}
	}

	// recompute size after pruning
	if len(stats) == 0 {
		return VideoTestResult{
			Name:             "word_stats_bucket_format_fixed",
			OriginalSize:     originalSize,
			ResultSize:       0,
			ReductionPercent: 100,
			Output:           "",
		}
	}

	// -----------------------------
	// SCORE + BUCKETING
	// -----------------------------
	type scoredWord struct {
		Word     string
		Count    int
		Duration float64
		Score    float64
		Bucket   int
	}

	words := make([]scoredWord, 0, len(stats))

	var maxScore float64

	for _, s := range stats {
		score := (math.Log1p(float64(s.Count)) * 2.0) +
			(math.Log1p(s.Duration) * 1.0)

		if score > maxScore {
			maxScore = score
		}

		words = append(words, scoredWord{
			Word:     s.Word,
			Count:    s.Count,
			Duration: s.Duration,
			Score:    score,
		})
	}

	const buckets = 32

	for i := range words {
		if maxScore > 0 {
			words[i].Bucket = int((words[i].Score / maxScore) * float64(buckets-1))
		}
	}

	// -----------------------------
	// BUCKET AGGREGATION
	// -----------------------------
	type bucket struct {
		countSum    int
		countMin    int
		countMax    int
		durSum      float64
		durMin      float64
		durMax      float64
		wordCount   int
		words       []string
		initialized bool
	}

	bs := make([]bucket, buckets)

	for _, w := range words {
		b := &bs[w.Bucket]

		if !b.initialized {
			b.countMin = w.Count
			b.countMax = w.Count
			b.durMin = w.Duration
			b.durMax = w.Duration
			b.initialized = true
		}

		b.countSum += w.Count
		b.durSum += w.Duration
		b.wordCount++

		if w.Count < b.countMin {
			b.countMin = w.Count
		}
		if w.Count > b.countMax {
			b.countMax = w.Count
		}
		if w.Duration < b.durMin {
			b.durMin = w.Duration
		}
		if w.Duration > b.durMax {
			b.durMax = w.Duration
		}

		b.words = append(b.words, w.Word)
	}

	// -----------------------------
	// OUTPUT
	// -----------------------------
	var out strings.Builder

	fmt.Fprintf(&out,
		"# BUCKETED TRANSCRIPT WORD STATISTICS (HIGH → LOW IMPORTANCE)\n"+
			"# FORMAT: bucketId,countMin-countMax,durationMin-durationMax,words(space separated)\n"+
			"# NOTE: bucket 31 = highest signal, bucket 0 = lowest signal\n\n",
	)

	for i := buckets - 1; i >= 0; i-- {
		b := bs[i]
		if b.wordCount == 0 {
			continue
		}

		fmt.Fprintf(
			&out,
			"%d,%d-%d,%.0f-%.0f,%s\n",
			i,
			b.countMin, b.countMax,
			b.durMin, b.durMax,
			strings.Join(b.words, " "),
		)
	}

	output := out.String()
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) / float64(originalSize) * 100
	}

	// -----------------------------
	// DEBUG PRINT
	// -----------------------------
	fmt.Printf("\n=== WORD STATS (BUCKET FORMAT FIXED) ===\n")
	fmt.Printf("Video ID       : %s\n", video.ID)
	fmt.Printf("Original Chars : %d\n", originalSize)
	fmt.Printf("Output Chars   : %d\n", resultSize)
	fmt.Printf("Buckets        : %d\n", buckets)
	fmt.Printf("Words Kept     : %d\n", len(stats))
	fmt.Printf("Reduction      : %.2f%%\n", reduction)
	fmt.Printf("=========================================\n")

	return VideoTestResult{
		Name:             "word_stats_bucket_format_fixed",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}