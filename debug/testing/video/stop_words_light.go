package video
import (
	"fmt"
	"strings"
	"regexp"

	youtube "github.com/Arvind215271/yt2bulk/internal/youtube"
)

var stopWords = map[string]bool{
	// Articles & common words
	"the": true,
	"and": true,
	"a":   true,
	"of":  true,
	"to":  true,
	"is":  true,
	"in":  true,

	"that": true,
	"it":   true,
	"for":  true,
	"on":   true,
	"with": true,
	"as":   true,
	"at":   true,

	"by":   true,
	"an":   true,
	"be":   true,
	"this": true,
	"from": true,
	"or":   true,
	"are":  true,

	// Pronouns
	"you":  true,
	"we":   true,
		"i":    true,
	"me":   true,
	"my":   true,
	"he":   true,
	"she":  true,

	"they": true,
	"them": true,
	"your": true,
	"our":  true,
	"us":   true,

	// Common verbs
	"do":     true,
	"does":   true,
	"did":    true,
	"have":   true,
	"has":    true,
	"had":    true,
	"can":    true,
	"could":  true,
	"will":   true,
	"would":  true,
	"should": true,
	"was":    true,
	"were":   true,
	"been":   true,
	"going":  true,
	"go":     true,
	"get":    true,
	"got":    true,

	// Spoken filler words
	"yeah":      true,
	"guys":      true,
	"video":     true,
	"think":     true,
	"know":      true,
	"talk":      true,
	"about":     true,
	"right":     true,
	"basically": true,
	"cool":      true,
	"like":      true,
	"just":      true,
	"what":      true,
	"then":      true,
	"here":      true,
	"there":     true,
	"so":        true,
	"okay":      true,
	"now":       true,
	"actually":  true,
	"something": true,
	"someone":   true,
}

func RunLightStopWordReductionTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{
			Name: "light_stop_word_reduction",
		}
	}

	rawText := video.Transcript.ToPlainText()

	originalSize := len(rawText)

	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	cleanText := reg.ReplaceAllString(
		strings.ToLower(rawText),
		"",
	)

	words := strings.Fields(cleanText)

	var remaining []string

	for _, word := range words {
		if stopWords[word] {
			continue
		}

		remaining = append(
			remaining,
			strings.Title(word),
		)
	}

	output := strings.Join(remaining, " ")
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) /
			float64(originalSize) * 100
	}

	fmt.Printf("\n===LIGHT STOP WORD REDUCTION ===\n")
	fmt.Printf("Video ID     : %s\n", video.ID)
	fmt.Printf("Title        : %s\n", video.Title)
	fmt.Printf("Original     : %d chars (%.2f KB)\n",
		originalSize,
		float64(originalSize)/1024,
	)
	fmt.Printf("Reduced      : %d chars (%.2f KB)\n",
		resultSize,
		float64(resultSize)/1024,
	)
	fmt.Printf("Reduction    : %.2f%%\n", reduction)
	fmt.Printf("===========================\n")

	return VideoTestResult{
		Name:             "light_stop_word_reduction",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}

