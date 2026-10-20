package video

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

type WindowFP struct {
	Word  string
	Count int
}

func encodeIDs(ids []uint16) string {
	if len(ids) == 0 {
		return ""
	}

	buf := make([]byte, len(ids)*2)
	for i, id := range ids {
		binary.LittleEndian.PutUint16(buf[i*2:], id)
	}

	return base64.StdEncoding.EncodeToString(buf)
}

func RunWindowFingerprint(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{Name: "window_fp_v9"}
	}

	raw := video.Transcript.ToPlainText()
	originalSize := len(raw)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	windowMap := map[int]map[string]*WindowFP{}
	global := map[string]int{}

	// -----------------------------
	// BUILD
	// -----------------------------
	for _, seg := range video.Transcript.Segments {
		widx := int(seg.Start / windowSize)

		if windowMap[widx] == nil {
			windowMap[widx] = map[string]*WindowFP{}
		}

		clean := reg.ReplaceAllString(strings.ToLower(seg.Text), "")
		words := strings.Fields(clean)

		for _, w := range words {
			if IsHeavyStopWord(w) {
				continue
			}

			global[w]++
			if windowMap[widx][w] == nil {
				windowMap[widx][w] = &WindowFP{Word: w}
			}
			windowMap[widx][w].Count++
		}
	}

	// -----------------------------
	// CODEBOOK
	// -----------------------------
	type dw struct {
		word  string
		count int
	}

	var dict []dw
	for w, c := range global {
		if c >= 2 {
			dict = append(dict, dw{w, c})
		}
	}

	sort.Slice(dict, func(i, j int) bool {
		return dict[i].count > dict[j].count
	})

	wordToID := map[string]uint16{}
	idToWord := map[uint16]string{}

	for i, d := range dict {
		wordToID[d.word] = uint16(i)
		idToWord[uint16(i)] = d.word
	}

	// -----------------------------
	// OUTPUT
	// -----------------------------
	var b strings.Builder

	fmt.Fprintf(&b, "WS=%d\nD=%d\n\n", int(windowSize), len(wordToID))

	// 🔥 CODEBOOK (REQUIRED FOR RECONSTRUCTION)
	b.WriteString("C:")
	first := true
	for i, w := range idToWord {
		if !first {
			b.WriteString(";")
		}
		first = false
		b.WriteString(fmt.Sprintf("%d=%s", i, w))
	}
	b.WriteString("\n\n")

	keys := make([]int, 0, len(windowMap))
	for k := range windowMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// limit repetition across windows
	maxUse := 2
	use := map[string]int{}

	// -----------------------------
	// STREAM
	// -----------------------------
	for wi, k := range keys {
		m := windowMap[k]

		type item struct {
			word  string
			count int
			score float64
		}

		var items []item
		var max float64

		for w, v := range m {
			if v.Count < 2 {
				continue
			}

			s := math.Log1p(float64(v.Count))
			items = append(items, item{w, v.Count, s})
			if s > max {
				max = s
			}
		}

		for i := range items {
			items[i].score /= max
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].score > items[j].score
		})

		var high, mid []uint16

		for _, it := range items {
			id, ok := wordToID[it.word]
			if !ok {
				continue
			}

			if use[it.word] >= maxUse {
				continue
			}
			use[it.word]++

			if it.count >= 8 {
				high = append(high, id)
			} else {
				mid = append(mid, id)
			}
		}

		hEnc := encodeIDs(high)
		mEnc := encodeIDs(mid)

		if hEnc == "" && mEnc == "" {
			continue
		}

		// stream format
		fmt.Fprintf(&b, "%d|%s|%s\n", wi, hEnc, mEnc)
	}

	output := b.String()
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) / float64(originalSize) * 100
	}

	fmt.Printf("\n=== WINDOW FP V9 (CODEBOOK + STREAM) ===\n")
	fmt.Printf("Original : %d\n", originalSize)
	fmt.Printf("Output   : %d\n", resultSize)
	fmt.Printf("Dict     : %d\n", len(wordToID))
	fmt.Printf("Reduction: %.2f%%\n", reduction)
	fmt.Printf("=======================================\n")

	return VideoTestResult{
		Name:             "window_fp_v9",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}
