package video

// import (
// 	"fmt"
// 	"regexp"
// 	"sort"
// 	"strings"

// 	"github.com/Arvind215271/yt2bulk/internal/youtube"
// )

// var mergeWindowSize = 300.0
// const mergeThreshold = 0.70
// const topKWords = 12


// type MergedWindow struct {
// 	Start float64
// 	End   float64
// 	Words []WindowWordStat
// }

// func topWords(words []WindowWordStat) map[string]bool {
// 	sort.Slice(words, func(i, j int) bool {
// 		return words[i].Duration > words[j].Duration
// 	})

// 	if len(words) > topKWords {
// 		words = words[:topKWords]
// 	}

// 	m := make(map[string]bool)
// 	for _, w := range words {
// 		m[w.Word] = true
// 	}
// 	return m
// }

// func jaccard(a, b map[string]bool) float64 {
// 	if len(a) == 0 && len(b) == 0 {
// 		return 1
// 	}

// 	inter := 0
// 	union := make(map[string]bool)

// 	for k := range a {
// 		union[k] = true
// 		if b[k] {
// 			inter++
// 		}
// 	}
// 	for k := range b {
// 		union[k] = true
// 	}

// 	return float64(inter) / float64(len(union))
// }

// func RunWindowMergeWordStatsTest(video youtube.Video) VideoTestResult {
// 	if video.Transcript == nil {
// 		return VideoTestResult{Name: "window_merge_word_stats"}
// 	}

// 	rawText := video.Transcript.ToPlainText()
// 	originalSize := len(rawText)

// 	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// 	windowMap := make(map[int]map[string]*WindowWordStat)

// 	// ---------------------------
// 	// STEP 1: build windows
// 	// ---------------------------
// 	for _, seg := range video.Transcript.Segments {
// 		idx := int(seg.Start / mergeWindowSize)

// 		if windowMap[idx] == nil {
// 			windowMap[idx] = make(map[string]*WindowWordStat)
// 		}

// 		clean := reg.ReplaceAllString(strings.ToLower(seg.Text), "")
// 		words := strings.Fields(clean)

// 		if len(words) == 0 {
// 			continue
// 		}

// 		duration := seg.End - seg.Start
// 		perWord := duration / float64(len(words))

// 		for _, w := range words {
// 			if IsHeavyStopWord(w) {
// 				continue
// 			}

// 			if windowMap[idx][w] == nil {
// 				windowMap[idx][w] = &WindowWordStat{Word: w}
// 			}

// 			windowMap[idx][w].Count++
// 			windowMap[idx][w].Duration += perWord
// 		}
// 	}

// 	// sort keys
// 	var keys []int
// 	for k := range windowMap {
// 		keys = append(keys, k)
// 	}
// 	sort.Ints(keys)

// 	// convert to slice
// 	type rawWindow struct {
// 		idx   int
// 		start float64
// 		end   float64
// 		words []WindowWordStat
// 	}

// 	var windows []rawWindow

// 	for _, idx := range keys {
// 		var ws []WindowWordStat
// 		for _, w := range windowMap[idx] {
// 			ws = append(ws, *w)
// 		}

// 		windows = append(windows, rawWindow{
// 			idx:   idx,
// 			start: float64(idx) * mergeWindowSize,
// 			end:   float64(idx+1) * mergeWindowSize,
// 			words: ws,
// 		})
// 	}

// 	// ---------------------------
// 	// STEP 2: MERGE WINDOWS
// 	// ---------------------------
// 	var merged []MergedWindow

// 	for i := 0; i < len(windows); i++ {

// 		curr := windows[i]

// 		currTop := topWords(curr.words)

// 		if len(merged) == 0 {
// 			merged = append(merged, MergedWindow{
// 				Start: curr.start,
// 				End:   curr.end,
// 				Words: curr.words,
// 			})
// 			continue
// 		}

// 		last := &merged[len(merged)-1]

// 		lastTop := topWords(last.Words)

// 		sim := jaccard(currTop, lastTop)

// 		if sim >= mergeThreshold {
// 			// merge
// 			last.End = curr.end
// 			last.Words = append(last.Words, curr.words...)
// 		} else {
// 			merged = append(merged, MergedWindow{
// 				Start: curr.start,
// 				End:   curr.end,
// 				Words: curr.words,
// 			})
// 		}
// 	}

// 	// ---------------------------
// 	// STEP 3: OUTPUT
// 	// ---------------------------
// 	var b strings.Builder

// 	for _, mw := range merged {

// 		b.WriteString(fmt.Sprintf(
// 			"\nMERGED WINDOW (%.0fs - %.0fs)\n",
// 			mw.Start, mw.End,
// 		))

// 		b.WriteString("word,count,duration\n")

// 		sort.Slice(mw.Words, func(i, j int) bool {
// 			return mw.Words[i].Duration > mw.Words[j].Duration
// 		})

// 		for _, w := range mw.Words {

// 			if w.Count == 1 && w.Duration < 2 {
// 				continue
// 			}

// 			b.WriteString(fmt.Sprintf(
// 				"%s,%d,%.0f\n",
// 				w.Word,
// 				w.Count,
// 				w.Duration,
// 			))
// 		}
// 	}

// 	output := b.String()
// 	resultSize := len(output)

// 	reduction := 0.0
// 	if originalSize > 0 {
// 		reduction = float64(originalSize-resultSize) / float64(originalSize) * 100
// 	}

// 	fmt.Printf("\n=== MERGED WINDOW WORD STATS ===\n")
// 	fmt.Printf("Video ID   : %s\n", video.ID)
// 	fmt.Printf("Windows    : %d → %d\n", len(windows), len(merged))
// 	fmt.Printf("Original   : %d\n", originalSize)
// 	fmt.Printf("Reduced    : %d\n", resultSize)
// 	fmt.Printf("Reduction  : %.2f%%\n", reduction)
// 	fmt.Printf("=================================\n")

// 	return VideoTestResult{
// 		Name:             "window_merge_word_stats",
// 		OriginalSize:     originalSize,
// 		ResultSize:       resultSize,
// 		ReductionPercent: reduction,
// 		Output:           output,
// 	}
// }