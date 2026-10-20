package video

// import (
// 	"fmt"
// 	"math"
// 	"regexp"
// 	"sort"
// 	"strings"

// 	"github.com/Arvind215271/yt2bulk/internal/youtube"
// )

// type WindowFingerprint struct {
// 	Word     string
// 	Count    int
// 	Duration float64
// 	Score    float64
// }

// func RunWindowWordStatsEncoded(video youtube.Video) VideoTestResult {
// 	if video.Transcript == nil {
// 		return VideoTestResult{Name: "window_fingerprint_codec_v6"}
// 	}

// 	rawText := video.Transcript.ToPlainText()
// 	originalSize := len(rawText)

// 	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// 	// -----------------------------
// 	// WINDOW BUILD
// 	// -----------------------------
// 	windowMap := make(map[int]map[string]*WindowFingerprint)

// 	wordGlobalCount := make(map[string]int)

// 	for _, seg := range video.Transcript.Segments {
// 		widx := int(seg.Start / windowSize)

// 		if windowMap[widx] == nil {
// 			windowMap[widx] = map[string]*WindowFingerprint{}
// 		}

// 		clean := reg.ReplaceAllString(strings.ToLower(seg.Text), "")
// 		words := strings.Fields(clean)
// 		if len(words) == 0 {
// 			continue
// 		}

// 		dur := seg.End - seg.Start
// 		perWord := dur / float64(len(words))

// 		for _, w := range words {
// 			if IsHeavyStopWord(w) {
// 				continue
// 			}

// 			wordGlobalCount[w]++

// 			if windowMap[widx][w] == nil {
// 				windowMap[widx][w] = &WindowFingerprint{Word: w}
// 			}

// 			windowMap[widx][w].Count++
// 			windowMap[widx][w].Duration += perWord
// 		}
// 	}

// 	// -----------------------------
// 	// BUILD GLOBAL DICTIONARY
// 	// -----------------------------
// 	type dictItem struct {
// 		word  string
// 		count int
// 	}

// 	var dictList []dictItem
// 	for w, c := range wordGlobalCount {
// 		if c < 2 {
// 			continue
// 		}
// 		dictList = append(dictList, dictItem{w, c})
// 	}

// 	sort.Slice(dictList, func(i, j int) bool {
// 		return dictList[i].count > dictList[j].count
// 	})

// 	wordToID := map[string]int{}
// 	id := 0
// 	for _, d := range dictList {
// 		wordToID[d.word] = id
// 		id++
// 	}

// 	// -----------------------------
// 	// OUTPUT
// 	// -----------------------------
// 	var b strings.Builder

// 	fmt.Fprintf(&b,
// 		"WS=%.0f\nDICT_SIZE=%d\nFORMAT: window_id|high|mid\n\n",
// 		windowSize,
// 		len(wordToID),
// 	)

// 	// -----------------------------
// 	// SORT WINDOWS
// 	// -----------------------------
// 	keys := make([]int, 0, len(windowMap))
// 	for k := range windowMap {
// 		keys = append(keys, k)
// 	}
// 	sort.Ints(keys)

// 	maxWindowsPerWord := 2
// 	wordWindowUse := map[string]int{}

// 	// -----------------------------
// 	// PROCESS WINDOWS
// 	// -----------------------------
// 	for i, idx := range keys {

// 		m := windowMap[idx]

// 		type item struct {
// 			word  string
// 			count int
// 			score float64
// 		}

// 		var list []item
// 		var maxScore float64

// 		for w, v := range m {

// 			if v.Count < 2 {
// 				continue
// 			}

// 			score := math.Log1p(float64(v.Count))*2 +
// 				math.Log1p(v.Duration)

// 			if score > maxScore {
// 				maxScore = score
// 			}

// 			list = append(list, item{
// 				word:  w,
// 				count: v.Count,
// 				score: score,
// 			})
// 		}

// 		if len(list) == 0 {
// 			continue
// 		}

// 		for i := range list {
// 			list[i].score /= maxScore
// 		}

// 		sort.Slice(list, func(i, j int) bool {
// 			return list[i].score > list[j].score
// 		})

// 		var highIDs []string
// 		var midIDs []string

// 		for _, it := range list {
// 			id, ok := wordToID[it.word]
// 			if !ok {
// 				continue
// 			}

// 			// CRITICAL: reduce cross-window repetition
// 			if wordWindowUse[it.word] >= maxWindowsPerWord {
// 				continue
// 			}
// 			wordWindowUse[it.word]++

// 			switch {
// 			case it.count >= 8:
// 				highIDs = append(highIDs, fmt.Sprintf("%d", id))
// 			default:
// 				midIDs = append(midIDs, fmt.Sprintf("%d", id))
// 			}
// 		}

// 		if len(highIDs) == 0 && len(midIDs) == 0 {
// 			continue
// 		}

// 		// compact join (NO commas → smaller)
// 		line := fmt.Sprintf("%d|%s|%s\n",
// 			i,
// 			strings.Join(highIDs, ","),
// 			strings.Join(midIDs, ","),
// 		)

// 		b.WriteString(line)
// 	}

// 	output := b.String()
// 	resultSize := len(output)

// 	reduction := 0.0
// 	if originalSize > 0 {
// 		reduction = float64(originalSize-resultSize) / float64(originalSize) * 100
// 	}

// 	fmt.Printf("\n=== WINDOW FINGERPRINT CODEC V6 ===\n")
// 	fmt.Printf("Video ID     : %s\n", video.ID)
// 	fmt.Printf("Original     : %d\n", originalSize)
// 	fmt.Printf("Output       : %d\n", resultSize)
// 	fmt.Printf("Dict Size    : %d\n", len(wordToID))
// 	fmt.Printf("Reduction    : %.2f%%\n", reduction)
// 	fmt.Printf("===================================\n")

// 	return VideoTestResult{
// 		Name:             "window_fingerprint_codec_v6",
// 		OriginalSize:     originalSize,
// 		ResultSize:       resultSize,
// 		ReductionPercent: reduction,
// 		Output:           output,
// 	}
// }