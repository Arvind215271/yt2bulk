package video

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

type WindowConcept struct {
	Word     string
	Count    int
	Duration float64
	Score    float64
}
func RunWindowCompressedConceptStatsTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{Name: "window_concept_stats"}
	}

	rawText := video.Transcript.ToPlainText()
	originalSize := len(rawText)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	// -----------------------------
	// BUILD WINDOWS
	// -----------------------------
	windowMap := make(map[int]map[string]*WindowConcept)
	windowWordCount := make(map[int]int)

	for _, seg := range video.Transcript.Segments {
		widx := int(seg.Start / windowSize)

		if windowMap[widx] == nil {
			windowMap[widx] = map[string]*WindowConcept{}
		}

		clean := reg.ReplaceAllString(strings.ToLower(seg.Text), "")
		words := strings.Fields(clean)

		if len(words) == 0 {
			continue
		}

		windowWordCount[widx] += len(words)

		dur := seg.End - seg.Start
		perWord := dur / float64(len(words))

		for _, w := range words {
			if IsHeavyStopWord(w) {
				continue
			}

			if windowMap[widx][w] == nil {
				windowMap[widx][w] = &WindowConcept{Word: w}
			}

			windowMap[widx][w].Count++
			windowMap[widx][w].Duration += perWord
		}
	}

	// -----------------------------
	// OUTPUT HEADER (UNCHANGED FORMAT)
	// -----------------------------
	var b strings.Builder

	fmt.Fprintf(&b,
		"WS=%.0f\n"+
			"FORMAT:\n"+
			"#<window_id>\n"+
			"line1 = freq>=8 (core)\n"+
			"line2 = freq 4-7 (mid)\n"+
			"line3 = freq 2-3 (low)\n"+
			"NOTE: comma-separated words only\n\n",
		windowSize,
	)

	// -----------------------------
	// SORT WINDOWS
	// -----------------------------
	keys := make([]int, 0, len(windowMap))
	for k := range windowMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	maxSize := 18000
	totalOut := 0

	// -----------------------------
	// PROCESS WINDOWS (NO EARLY EXIT)
	// -----------------------------
	for wi, idx := range keys {

		m := windowMap[idx]
		// wordCount := windowWordCount[idx]

		header := fmt.Sprintf("#%d\n", wi+1)

		var used int
		b.WriteString(header)
		used += len(header)

		// -----------------------------
		// adaptive thresholding
		// -----------------------------
		var minFreq int
		minFreq = 0
		// var capG1, capG2, capG3 int

		// switch {
		// case wordCount < 40:
		// 	minFreq = 2
		// 	capG1, capG2, capG3 = 12, 18, 18
		// case wordCount < 120:
		// 	minFreq = 3
		// 	capG1, capG2, capG3 = 8, 12, 12
		// default:
		// 	minFreq = 4
		// 	capG1, capG2, capG3 = 6, 10, 10
		// }

		type item struct {
			word  string
			count int
			score float64
		}

		var list []item
		var maxScore float64

		for _, w := range m {
			if w.Count < minFreq {
				continue
			}

			score := math.Log1p(float64(w.Count))*2 +
				math.Log1p(w.Duration)

			list = append(list, item{
				word:  w.Word,
				count: w.Count,
				score: score,
			})

			if score > maxScore {
				maxScore = score
			}
		}

		if len(list) == 0 {
			continue
		}

		for i := range list {
			list[i].score /= maxScore
		}

		sort.Slice(list, func(i, j int) bool {
			return list[i].score > list[j].score
		})

		// ------------------------------------
		// Dynamic character budget per window
		// ------------------------------------
		targetTotalChars := 30000

		// reserve some space for the header and metadata
		budgetPerWindow := targetTotalChars / len(keys)
		if budgetPerWindow < 80 {
			budgetPerWindow = 80
		}

		remaining := budgetPerWindow - used

		g1 := []string{}
		g2 := []string{}
		g3 := []string{}

		// ----------------------------
		// Bucket words
		// ----------------------------
		for _, it := range list {

			switch {
			case it.count >= 8:
				g1 = append(g1, it.word)

			case it.count >= 4:
				g2 = append(g2, it.word)

			case it.count >= 2:
				g3 = append(g3, it.word)
			}
		}

		// ----------------------------
		// Writer using remaining budget
		// ----------------------------
		writeBudget := func(words []string) {

			if remaining <= 0 {
				b.WriteString("\n")
				return
			}

			if len(words) == 0 {
				b.WriteString("\n")
				return
			}

			var out []string
			chars := 0

			for _, w := range words {

				add := len(w)
				if len(out) > 0 {
					add++ // comma
				}

				if chars+add > remaining {
					break
				}

				out = append(out, w)
				chars += add
			}

			line := strings.Join(out, ",")
			b.WriteString(line)
			b.WriteByte('\n')

			used += chars + 1
			remaining -= chars + 1
		}

		// ----------------------------------------
		// 1. Keep ALL core concepts (no cap)
		// ----------------------------------------
		writeBudget(g1)

		// ----------------------------------------
		// 2. Mid concepts use most remaining space
		// ----------------------------------------
		midBudget := remaining * 7 / 10
		old := remaining
		remaining = midBudget
		writeBudget(g2)
		remaining = old - (midBudget - remaining)

		// ----------------------------------------
		// 3. Low concepts consume whatever remains
		// ----------------------------------------
		writeBudget(g3)

		totalOut += used

		// ❗ NO early exit anymore (important fix)
		if totalOut > maxSize {
			// we just stop adding new content, but KEEP processing
			continue
		}
	}

	output := b.String()
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) / float64(originalSize) * 100
	}

	fmt.Printf("\n=== WINDOW CONCEPT FINGERPRINT v5 (FIXED FULL PASS) ===\n")
	fmt.Printf("Video ID   : %s\n", video.ID)
	fmt.Printf("Windows    : %d\n", len(keys))
	fmt.Printf("Original   : %d\n", originalSize)
	fmt.Printf("Reduced    : %d\n", resultSize)
	fmt.Printf("Reduction  : %.2f%%\n", reduction)
	fmt.Printf("=====================================================\n")

	return VideoTestResult{
		Name:             "window_concept_fingerprint_v5",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}