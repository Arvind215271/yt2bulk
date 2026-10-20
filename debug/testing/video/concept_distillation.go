

package video

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

type conceptAgg struct {
	Concept   string
	Score     float64
	Freq      int
	Words     map[string]int
}

func RunConceptDistillation(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{
			Name: "concept_distillation_test",
		}
	}

	rawText := video.Transcript.ToPlainText()
	originalSize := len(rawText)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	// -----------------------------
	// STEP 1: build windowed word signals
	// -----------------------------
	// windowSize := 45.0

	type window struct {
		start float64
		end   float64
		words []string
	}

	var windows []window

	for _, seg := range video.Transcript.Segments {
		clean := reg.ReplaceAllString(strings.ToLower(seg.Text), "")
		words := strings.Fields(clean)

		windows = append(windows, window{
			start: seg.Start,
			end:   seg.End,
			words: words,
		})
	}

	// -----------------------------
	// STEP 2: global frequency map
	// -----------------------------
	globalFreq := map[string]int{}
	totalWords := 0

	for _, w := range windows {
		for _, word := range w.words {
			if IsHeavyStopWord(word) {
				continue
			}
			globalFreq[word]++
			totalWords++
		}
	}

	// -----------------------------
	// STEP 3: concept assignment (heuristic clustering)
	// -----------------------------
	concepts := map[string]*conceptAgg{}

	// pick top N global words as "concept seeds"
	type kv struct {
		word string
		freq int
	}

	var top []kv
	for w, f := range globalFreq {
		top = append(top, kv{w, f})
	}

	sort.Slice(top, func(i, j int) bool {
		return top[i].freq > top[j].freq
	})

	// if len(top) > 20 {
	// 	top = top[:20]
	// }

	// initialize concepts
	for _, t := range top {
		concepts[t.word] = &conceptAgg{
			Concept: t.word,
			Words:   map[string]int{},
		}
	}

	// assign words to nearest concept (simple overlap heuristic)
	for _, w := range windows {
		localFreq := map[string]int{}

		for _, word := range w.words {
			if IsHeavyStopWord(word){
				continue
			}
			localFreq[word]++
		}

		for word, freq := range localFreq {
			// find best concept match
			var bestConcept string
			bestScore := 0

			for c := range concepts {
				score := 0
				if c == word {
					score += 5
				}
				if strings.Contains(word, c) || strings.Contains(c, word) {
					score += 2
				}

				if score > bestScore {
					bestScore = score
					bestConcept = c
				}
			}

			if bestConcept == "" {
				continue
			}

			concepts[bestConcept].Freq += freq
			concepts[bestConcept].Words[word] += freq
		}
	}

	// -----------------------------
	// STEP 4: scoring concepts
	// -----------------------------
	var result []conceptAgg

	for _, c := range concepts {
		// hybrid importance score
		score := float64(c.Freq) * 1.0

		c.Score = score
		result = append(result, *c)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	// -----------------------------
	// STEP 5: OUTPUT FORMAT
	// -----------------------------
	var b strings.Builder

	fmt.Fprintf(&b,
		"# CONCEPT DISTILLATION (HIGH → LOW IMPORTANCE)\n"+
			"# FORMAT: concept,score,freq,topWords(space separated)\n\n",
	)

	for _, c := range result {
		// collect top words
		type wf struct {
			w string
			f int
		}

		var wfList []wf
		for w, f := range c.Words {
			wfList = append(wfList, wf{w, f})
		}

		sort.Slice(wfList, func(i, j int) bool {
			return wfList[i].f > wfList[j].f
		})

		topWords := []string{}
		limit := len(wfList)
		// if len(wfList) < limit {
		// 	limit = len(wfList)
		// }

		for i := 0; i < limit; i++ {
			topWords = append(topWords, wfList[i].w)
		}

		fmt.Fprintf(
			&b,
			"%s,%.0f,%d,%s\n",
			c.Concept,
			c.Score,
			c.Freq,
			strings.Join(topWords, " "),
		)
	}

	output := b.String()
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) / float64(originalSize) * 100
	}

	// -----------------------------
	// DEBUG PRINT (same style as yours)
	// -----------------------------
	fmt.Printf("\n=== CONCEPT DISTILLATION TEST ===\n")
	fmt.Printf("Video ID       : %s\n", video.ID)
	fmt.Printf("Original Chars : %d\n", originalSize)
	fmt.Printf("Output Chars   : %d\n", resultSize)
	fmt.Printf("Concepts       : %d\n", len(result))
	fmt.Printf("Reduction      : %.2f%%\n", reduction)
	fmt.Printf("=================================\n")

	return VideoTestResult{
		Name:             "concept_distillation_test",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}