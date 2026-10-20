package video

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

type TranscriptWindow struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
	Chars int     `json:"chars"`
}

type TranscriptTopicSignal struct {
	Topic    string  `json:"topic"`
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
	Strength float64 `json:"strength"`
	Sample   string  `json:"sample"`
	Chars    int     `json:"chars"`
}

// -------------------------
// CLEANING (LIGHT ONLY)
// -------------------------

var noiseRegex = regexp.MustCompile(`(?i)\b(um|uh|like|you know|basically|so)\b`)
var spaceRegex = regexp.MustCompile(`\s+`)

func cleanText(text string) string {
	text = strings.ToLower(text)
	text = noiseRegex.ReplaceAllString(text, "")
	text = spaceRegex.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// -------------------------
// WINDOWING (CHAR-BASED BUILD)
// -------------------------

func BuildTranscriptWindows(t transcript.Transcript, windowSize float64) []TranscriptWindow {
	if len(t.Segments) == 0 {
		return nil
	}

	var windows []TranscriptWindow

	var current TranscriptWindow
	started := false

	for _, s := range t.Segments {
		if !started {
			current.Start = s.Start
			started = true
		}

		clean := cleanText(s.Text)

		current.Text += " " + clean
		current.End = s.End

		if current.End-current.Start >= windowSize {
			current.Text = strings.TrimSpace(current.Text)
			current.Chars = len(current.Text)
			windows = append(windows, current)

			current = TranscriptWindow{}
			started = false
		}
	}

	if strings.TrimSpace(current.Text) != "" {
		current.Text = strings.TrimSpace(current.Text)
		current.Chars = len(current.Text)
		windows = append(windows, current)
	}

	return windows
}

// -------------------------
// TOPIC SIGNAL EXTRACTION
// -------------------------

func ExtractTopicSignals(windows []TranscriptWindow) []TranscriptTopicSignal {
	var signals []TranscriptTopicSignal

	for _, window := range windows {

		words := strings.Fields(window.Text)

		freq := map[string]int{}

		for _, w := range words {
			if len(w) < 4 {
				continue
			}
			freq[w]++
		}

		bestWord := ""
		bestCount := 0

		for w, c := range freq {
			if c > bestCount {
				bestCount = c
				bestWord = w
			}
		}

		if bestWord == "" {
			continue
		}

		signals = append(signals, TranscriptTopicSignal{
			Topic:    bestWord,
			Start:    window.Start,
			End:      window.End,
			Strength: float64(bestCount),
			Sample:   window.Text,
			Chars:    window.Chars,
		})
	}

	return signals
}

// -------------------------
// TEST RUNNER 1
// -------------------------

func RunTranscriptWindowTest(t transcript.Transcript, videoID string) VideoTestResult {
	raw := t.ToPlainText()
	originalChars := len(raw)

	windows := BuildTranscriptWindows(t, 45)
	outputChars := 0

	for _, w := range windows {
		outputChars += w.Chars
	}

	reduction := calcCharReduction(originalChars, outputChars)

	output := marshal(windows)

	fmt.Printf("\n=== TRANSCRIPT WINDOW STATS ===\n")
	fmt.Printf("Video ID       : %s\n", videoID)
	fmt.Printf("Window Size (s): %.0f\n", 45.0)
	fmt.Printf("Original Chars : %d\n", originalChars)
	fmt.Printf("Output Chars   : %d\n", outputChars)
	fmt.Printf("Windows        : %d\n", len(windows))
	fmt.Printf("Reduction      : %.2f%%\n", reduction)
	fmt.Printf("=========================================\n")

	return VideoTestResult{
		Name:             "transcript_window_stats",
		OriginalSize:     originalChars,
		ResultSize:       outputChars,
		ReductionPercent: reduction,
		Output:           output,
	}
}

// -------------------------
// TEST RUNNER 2
// -------------------------

func RunTranscriptTopicSignalTestV2(t transcript.Transcript, videoID string) VideoTestResult {
	windows := BuildTranscriptWindows(t, 45)
	signals := ExtractTopicSignals(windows)

	raw := t.ToPlainText()
	originalChars := len(raw)

	outputChars := 0
	for _, s := range signals {
		outputChars += len(s.Sample)
	}

	reduction := calcCharReduction(originalChars, outputChars)

	output := marshal(signals)

	fmt.Printf("\n=== TRANSCRIPT TOPIC SIGNAL STATS ===\n")
	fmt.Printf("Video ID       : %s\n", videoID)
	fmt.Printf("Original Chars : %d\n", originalChars)
	fmt.Printf("Output Chars   : %d\n", outputChars)
	fmt.Printf("Windows        : %d\n", len(windows))
	fmt.Printf("Signals        : %d\n", len(signals))
	fmt.Printf("Reduction      : %.2f%%\n", reduction)
	fmt.Printf("=========================================\n")

	return VideoTestResult{
		Name:             "transcript_topic_signal_stats",
		OriginalSize:     originalChars,
		ResultSize:       outputChars,
		ReductionPercent: reduction,
		Output:           output,
	}
}

// -------------------------
// HELPERS
// -------------------------

func calcCharReduction(orig, out int) float64 {
	if orig == 0 {
		return 0
	}
	return (1.0 - float64(out)/float64(orig)) * 100.0
}

func marshal(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}