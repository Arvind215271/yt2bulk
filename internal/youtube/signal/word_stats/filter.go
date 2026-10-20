package wordstats

import (
	stopwords "github.com/Arvind215271/yt2bulk/internal/youtube/signal/stopwords"

)

type FilterConfig struct {
	MinFreq int

	Depth float64 // 0.0 - 1.0 percentage

	MaxChars int
}


// FilterWords applies deterministic multi-constraint filtering
func FilterWords(words []*WordStats, cfg FilterConfig) []*WordStats {
	if len(words) == 0 {
		return words
	}

	totalChars := 0
	for _, w := range words {
		totalChars += len(w.Word) + 1
	}

	charBudget := cfg.MaxChars
	if charBudget <= 0 {
		charBudget = totalChars
	}

	depthBudget := float64(len(words))
	if cfg.Depth > 0 && cfg.Depth <= 1 {
		depthBudget = float64(len(words)) * cfg.Depth
	}

	out := make([]*WordStats, 0, len(words))

	currChars := 0
	currDepth := 0

	for _, w := range words {

		// freq gate
		if cfg.MinFreq > 0 && w.Count < cfg.MinFreq {
			continue
		}

		// depth gate
		if float64(currDepth) >= depthBudget {
			break
		}

		// char gate
		nextSize := len(w.Word) + 1
		if currChars+nextSize > charBudget {
			break
		}

		out = append(out, w)
		currChars += nextSize
		currDepth++
	}

	return out
}



// FilterHeavyStopWords removes words that are marked as heavy stopwords
func FilterHeavyStopWords(words []*WordStats) []*WordStats {
	if len(words) == 0 {
		return words
	}

	out := make([]*WordStats, 0, len(words))

	for _, w := range words {
		if stopwords.IsHeavy(w.Word) {
			continue
		}
		out = append(out, w)
	}

	return out
}