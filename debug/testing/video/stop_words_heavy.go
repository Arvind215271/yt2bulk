// stopwords.go
package video

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"regexp"

	youtube "github.com/Arvind215271/yt2bulk/internal/youtube"
)

var heavyStopWords map[string]struct{}

func LoadHeavyStopWords(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	heavyStopWords = make(map[string]struct{})

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())

		if word == "" {
			continue
		}

		if strings.HasPrefix(word, "#") {
			continue
		}

		heavyStopWords[strings.ToLower(word)] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Printf(
		"Loaded %d heavy stop words\n",
		len(heavyStopWords),
	)

	return nil
}

func IsHeavyStopWord(word string) bool {
	_, exists := heavyStopWords[strings.ToLower(word)]
	return exists
}



func RunHeavyStopWordReductionTest(video youtube.Video) VideoTestResult {
	if video.Transcript == nil {
		return VideoTestResult{
			Name: "heavy_stop_word_reduction",
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
		if IsHeavyStopWord(word) {
			continue
		}

		remaining = append(remaining, word)
	}

	output := strings.Join(remaining, " ")
	resultSize := len(output)

	reduction := 0.0
	if originalSize > 0 {
		reduction = float64(originalSize-resultSize) /
			float64(originalSize) * 100
	}

	fmt.Printf("\n=== HEAVY STOP WORD REDUCTION ===\n")
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
	fmt.Printf("=================================\n")

	return VideoTestResult{
		Name:             "heavy_stop_word_reduction",
		OriginalSize:     originalSize,
		ResultSize:       resultSize,
		ReductionPercent: reduction,
		Output:           output,
	}
}