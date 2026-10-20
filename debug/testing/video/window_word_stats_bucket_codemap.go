package video

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"strconv"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

type WindowConceptCode struct {
	Word     string
	Count    int
	Duration float64
	Score    float64
}

type encodedWindow struct {
	G1 []string
	G2 []string
	G3 []string
}

type item struct {
	word  string
	count int
	score float64
}

func RunWindowCompressedConceptCodeMapTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{
			Name: "window_concept_codemap_v1",
		}
	}

	rawText := video.Transcript.ToPlainText()
	originalSize := len(rawText)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	// Build all windows and collect the emitted concepts.
	windows, usedWords, globalFreq := buildWindowData(video, reg)

	// Build the codemap and convert every window into numeric ids.
	codeMap, encoded := buildCodeMapAndEncode(
		windows,
		usedWords,
		globalFreq,
	)

	output := writeOutput(
		codeMap,
		encoded,
	)

	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) /
			float64(originalSize) * 100
	}

	fmt.Printf("\n=== WINDOW CONCEPT CODEMAP v1 ===\n")
	fmt.Printf("Video ID   : %s\n", video.ID)
	fmt.Printf("Windows    : %d\n", len(encoded))
	fmt.Printf("Original   : %d\n", originalSize)
	fmt.Printf("Reduced    : %d\n", resultSize)
	fmt.Printf("Reduction  : %.2f%%\n", reduction)
	fmt.Printf("=================================\n")

	return VideoTestResult{
		Name:             "window_concept_codemap_v1",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}

func buildBudgetedWindow(
	list []item, 
	totalWindows int,
) encodedWindow {

	targetTotalChars := 30000

	budgetPerWindow := targetTotalChars / totalWindows
	if budgetPerWindow < 80 {
		budgetPerWindow = 80
	}

	headerSize := len("#9999\n")
	remaining := budgetPerWindow - headerSize

	var (
		g1 []string
		g2 []string
		g3 []string
	)

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

	writeBudget := func(words []string) []string {

		if remaining <= 0 {
			return nil
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

		remaining -= chars + 1 // newline

		return out
	}

	core := writeBudget(g1)

	midBudget := remaining * 7 / 10
	old := remaining
	remaining = midBudget

	mid := writeBudget(g2)

	remaining = old - (midBudget - remaining)

	low := writeBudget(g3)

	return encodedWindow{
		G1: core,
		G2: mid,
		G3: low,
	}
}




func buildWindowData(
	video youtube.Video,
	reg *regexp.Regexp,
) (
	[]encodedWindow,
	map[string]bool,
	map[string]int,
) {

	windowMap := make(map[int]map[string]*WindowConceptCode)
	globalFreq := make(map[string]int)

	//------------------------------------------
	// Build window statistics
	//------------------------------------------
	for _, seg := range video.Transcript.Segments {

		widx := int(seg.Start / windowSize)

		if windowMap[widx] == nil {
			windowMap[widx] = map[string]*WindowConceptCode{}
		}

		clean := reg.ReplaceAllString(
			strings.ToLower(seg.Text),
			"",
		)

		words := strings.Fields(clean)

		if len(words) == 0 {
			continue
		}

		dur := seg.End - seg.Start
		perWord := dur / float64(len(words))

		for _, w := range words {

			if IsHeavyStopWord(w) {
				continue
			}

			globalFreq[w]++

			if windowMap[widx][w] == nil {
				windowMap[widx][w] = &WindowConceptCode{
					Word: w,
				}
			}

			windowMap[widx][w].Count++
			windowMap[widx][w].Duration += perWord
		}
	}

	//------------------------------------------
	// Sort windows
	//------------------------------------------
	keys := make([]int, 0, len(windowMap))

	for k := range windowMap {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	usedWords := make(map[string]bool)

	var windows []encodedWindow

	//------------------------------------------
	// Build grouped windows
	//------------------------------------------
	for _, idx := range keys {

		

		var (
			list     []item
			maxScore float64
		)

		for _, w := range windowMap[idx] {

			score :=
				math.Log1p(float64(w.Count))*2 +
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

		window := buildBudgetedWindow(
			list,
			len(keys),
		)

		for _, w := range window.G1 {
			usedWords[w] = true
		}
		for _, w := range window.G2 {
			usedWords[w] = true
		}
		for _, w := range window.G3 {
			usedWords[w] = true
		}

		windows = append(windows, window)
	}

	return windows, usedWords, globalFreq
}


func buildCodeMapAndEncode(
	windows []encodedWindow,
	usedWords map[string]bool,
	globalFreq map[string]int,
) (
	map[int]string,
	[]encodedWindow,
) {

	type codeWord struct {
		Word  string
		Count int
	}

	//------------------------------------------
	// Collect only emitted words
	//------------------------------------------
	var words []codeWord

	for word := range usedWords {
		words = append(words, codeWord{
			Word:  word,
			Count: globalFreq[word],
		})
	}

	//------------------------------------------
	// Sort:
	// 1. Global frequency (desc)
	// 2. Alphabetically (asc)
	//------------------------------------------
	sort.Slice(words, func(i, j int) bool {

		if words[i].Count == words[j].Count {
			return words[i].Word < words[j].Word
		}

		return words[i].Count > words[j].Count
	})

	//------------------------------------------
	// Build codemap
	//------------------------------------------
	wordToCode := make(map[string]string)
	codeMap := make(map[int]string)

	for i, w := range words {

		code := strconv.Itoa(i + 1)

		wordToCode[w.Word] = code
		codeMap[i+1] = w.Word
	}

	//------------------------------------------
	// Encode every window
	//------------------------------------------
	encoded := make([]encodedWindow, 0, len(windows))

	for _, win := range windows {

		var out encodedWindow

		for _, w := range win.G1 {
			if code, ok := wordToCode[w]; ok {
				out.G1 = append(out.G1, code)
			}
		}

		for _, w := range win.G2 {
			if code, ok := wordToCode[w]; ok {
				out.G2 = append(out.G2, code)
			}
		}

		for _, w := range win.G3 {
			if code, ok := wordToCode[w]; ok {
				out.G3 = append(out.G3, code)
			}
		}

		encoded = append(encoded, out)
	}

	return codeMap, encoded
}



func writeOutput(
	codeMap map[int]string,
	windows []encodedWindow,
) string {

	var b strings.Builder

	fmt.Fprintf(
		&b,
		"WINDOW_SIZE=%d seconds\n\n"+
			"FORMAT:\n"+
			"CODEMAP              (numeric id -> concept)\n"+
			"WINDOWS              (concept timeline)\n"+
			"#<window_id>         (window number)\n"+
			"line1                (core concepts, freq>=8)\n"+
			"line2                (mid concepts, freq=4-7)\n"+
			"line3                (low concepts, freq=2-3)\n"+
			"VALUES               (comma-separated CODEMAP ids)\n\n"+
			"CODEMAP:\n",
		int(windowSize),
	)

	//------------------------------------------
	// CODEMAP (packed horizontally)
	//------------------------------------------
	ids := make([]int, 0, len(codeMap))
	for id := range codeMap {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	const maxLine = 110

	lineLen := 0

	for _, id := range ids {

		entry := fmt.Sprintf("%d=%s", id, codeMap[id])

		// first item on line
		if lineLen == 0 {
			b.WriteString(entry)
			lineLen = len(entry)
			continue
		}

		// wrap if needed
		if lineLen+2+len(entry) > maxLine {
			b.WriteByte('\n')
			b.WriteString(entry)
			lineLen = len(entry)
			continue
		}

		b.WriteString(", ")
		b.WriteString(entry)
		lineLen += 2 + len(entry)
	}

	b.WriteString("\n\nWINDOWS:\n\n")

	//------------------------------------------
	// WINDOWS
	//------------------------------------------
	for i, win := range windows {

		fmt.Fprintf(&b, "#%d\n", i+1)

		b.WriteString(strings.Join(win.G1, ","))
		b.WriteByte('\n')

		b.WriteString(strings.Join(win.G2, ","))
		b.WriteByte('\n')

		b.WriteString(strings.Join(win.G3, ","))
		b.WriteByte('\n')
	}

	return b.String()
}