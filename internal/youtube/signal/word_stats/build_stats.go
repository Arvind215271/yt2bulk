package wordstats

import (
	"regexp"
	"strings"
	"math"

	transcript "github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

var cleanRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// RAW Word Stats
func BuildWordStats(t *transcript.Transcript) map[string]*WordStats {
	stats := make(map[string]*WordStats)

	if t == nil {
		return stats
	}

	for _, seg := range t.Segments {
		clean := cleanRegex.ReplaceAllString(strings.ToLower(seg.Text), "")
		words := strings.Fields(clean)

		duration := seg.End - seg.Start

		for _, w := range words {
			if w == "" {
				continue
			}

			if _, ok := stats[w]; !ok {
				stats[w] = &WordStats{Word: w}
			}

			stats[w].Count++
			stats[w].Duration += duration
		}
	}

	return stats
}





// ADD SCORE
func ScoreWords(stats map[string]*WordStats) []*WordStats {
	var list []*WordStats

	for _, w := range stats {
		w.Score = math.Log1p(float64(w.Count))*2 +
			math.Log1p(w.Duration)

		list = append(list, w)
	}

	return list
}



type Window struct {
	Words map[string]*WordStats

	OriginalWords int
	OriginalChars int
}


func BuildWindowStats(t *transcript.Transcript, windowSize float64) map[int]*Window {

	windowMap := make(map[int]*Window)

	if t == nil {
		return windowMap
	}

	for _, seg := range t.Segments {

		wid := int(seg.Start / windowSize)

		if windowMap[wid] == nil {
			windowMap[wid] = &Window{
				Words: make(map[string]*WordStats),
			}
		}

		w := windowMap[wid]

		// ORIGINAL SOURCE STATS (correct now)
		w.OriginalWords += len(strings.Fields(seg.Text))
		w.OriginalChars += len([]rune(seg.Text))

		clean := cleanRegex.ReplaceAllString(strings.ToLower(seg.Text), "")
		words := strings.Fields(clean)

		if len(words) == 0 {
			continue
		}

		duration := seg.End - seg.Start
		perWord := duration / float64(len(words))

		for _, token := range words {

			if token == "" {
				continue
			}

			if w.Words[token] == nil {
				w.Words[token] = &WordStats{Word: token}
			}

			ws := w.Words[token]
			ws.Count++
			ws.Duration += perWord
		}
	}

	return windowMap
}