// ./internal/youtube/debug/video.go
package video

import (
	"fmt"

	// "unicode/utf8"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
	wordstats "github.com/Arvind215271/yt2bulk/internal/youtube/signal/word_stats"
)

type VideoTestResult struct {
	Name string `json:"name"`

	OriginalSize int `json:"original_size"`
	ResultSize   int `json:"result_size"`

	ReductionPercent float64 `json:"reduction_percent"`

	Output string `json:"output"`
}

type VideoTestSuite struct {
	VideoID string `json:"video_id"`
	Title   string `json:"title"`

	Results []VideoTestResult `json:"results"`
}

// DebugVideoTranscript executes all transcript-related experiments
// against a single video and exports the resulting artifacts.
// func DebugVideoTranscript(video youtube.Video) error {
// 	if video.Transcript == nil {
// 		return fmt.Errorf("video transcript not available")
// 	}

// 	if err := LoadHeavyStopWords(
// 		"debug/video/heavy_stopwords.txt",
// 	); err != nil {
// 		return err
// 	}

// 	results := []VideoTestResult{
// 		// RunConceptDistillation(video ),

// 		// RunLightStopWordReductionTest(video),
// 		// RunHeavyStopWordReductionTest(video),

// 		// RunUniqueWordReductionTest(video),
// 		// RunHeavyStopWordUniqueWordTest(video),

// 		// RunWordStatsTest(video),
// 		// RunHeavyWordStatsTest(video),
// 		// RunHeavyWordStatsBucketFormatTest(video),

// 		RunWindowWordStatsTest(video),
// 		// RunWindowMergeWordStatsTest(video),
// 		RunWindowCompressedConceptStatsTest(video),
		
// 		// RunWindowWordStatsEncoded(video),
// 		// RunWindowFingerprint(video),
// 		// RunWindowCompressedConceptCodeMapTest(video),

// 		// NEW correct layer
// 		// RunTranscriptWindowTest(*video.Transcript, video.ID),
// 		// RunTranscriptTopicSignalTestV2(*video.Transcript, video.ID),
// 		// RunConceptDistillationTest(video),
// 	}

// 	suite := VideoTestSuite{
// 		VideoID: video.ID,
// 		Title:   video.Title,
// 		Results: results,
// 	}

// 	return ExportVideoTestResults(suite)
// }


type VideoTestCase struct {
	Name     string
	Config   wordstats.AnalysisConfig
}

func mapResult(name string, res wordstats.Result) VideoTestResult {
	return VideoTestResult{
		Name:             name,
		OriginalSize:     res.OriginalChars,
		ResultSize:       res.OutputChars,
		ReductionPercent: res.CompressionRatio,
		Output:           wordstats.Export(res, wordstats.ExportVerbose),
	}
}

func runVideoTests(
	svc *wordstats.Service,
	video *youtube.Video,
	tests []VideoTestCase,
) []VideoTestResult {

	results := make([]VideoTestResult, 0, len(tests))

	for _, t := range tests {
		res := svc.AnalyzeVideoWordStats(video.Transcript, t.Config)
		results = append(results, mapResult(t.Name, res))
	}

	return results
}

func DebugVideoTranscript(video youtube.Video) error {
	if video.Transcript == nil {
		return fmt.Errorf("video transcript not available")
	}

	

	// fmt.Println("Transcript Signal Testing")

	svc := wordstats.NewService()

	tests := []VideoTestCase{

	// BASELINE
	{
		Name: "baseline",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
		},
	},

	// -------------------------
	// DEPTH EXPERIMENTS
	// -------------------------

	{
		Name: "depth_1.0",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			Depth:             1.0,
		},
	},
	{
		Name: "depth_0.8",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			Depth:             0.8,
		},
	},
	{
		Name: "depth_0.5",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			Depth:             0.5,
		},
	},
	{
		Name: "depth_0.3",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			Depth:             0.3,
		},
	},

	// -------------------------
	// FREQUENCY EXPERIMENTS
	// -------------------------

	{
		Name: "freq_1",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			MinFreq:           1,
		},
	},
	{
		Name: "freq_2",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			MinFreq:           2,
		},
	},
	{
		Name: "freq_5",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			MinFreq:           5,
		},
	},
	{
		Name: "freq_10",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			MinFreq:           10,
		},
	},

	// -------------------------
	// CHAR LIMIT EXPERIMENTS
	// -------------------------

	{
		Name: "max_5k",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			MaxChars:          5000,
		},
	},
	{
		Name: "max_10k",
		Config: wordstats.AnalysisConfig{
			UseHeavyStopWords: true,
			MaxChars:          10000,
		},
	},
}

	results := runVideoTests(svc, &video, tests)


	for _, r := range results {
		PrintVideoTestResult(r)
	}

	suite := VideoTestSuite{
		VideoID: video.ID,
		Title:   video.Title,
		Results: results,
	}

	return ExportVideoTestResults(suite)
}

func mustResult(result wordstats.Result) VideoTestResult {

	str := wordstats.Export(result, wordstats.ExportVerbose)

	return VideoTestResult{
		Name: "",

		OriginalSize: result.OriginalChars,
		ResultSize:   result.OutputChars,

		ReductionPercent: result.CompressionRatio,

		Output: str,
	}
}

func PrintVideoTestResult(r VideoTestResult) {

	fmt.Println("===================================")
	fmt.Println("[VIDEO TEST RESULT]")
	fmt.Println("Name:", r.Name)

	fmt.Println("Original Size:", r.OriginalSize)
	fmt.Println("Result Size:", r.ResultSize)

	fmt.Println("Reduction %:", r.ReductionPercent)
	filteredChar:= len([]rune(r.Output))
	fmt.Println("Output Chars:", filteredChar)
	fmt.Println("Output Words:", len(strings.Fields(r.Output)))
	fmt.Println("Output Compression:", float64(r.OriginalSize - filteredChar)/float64(r.OriginalSize) * 100)

	fmt.Println("-----------------------------------")
	fmt.Println("===================================")
}


type WindowTestCase struct {
	Name string

	Config wordstats.AnalysisConfig
	WindowSize float64
}


func runWindowTests(
	svc *wordstats.Service,
	video *youtube.Video,
	tests []WindowTestCase,
) []VideoTestResult {

	results := make([]VideoTestResult, 0, len(tests))

	for _, t := range tests {

		res := svc.AnalyzeVideoWindowed(
			video.Transcript,
			t.Config,
			
		)

		// export window result (NEW exporter you built earlier)
		output := wordstats.ExportWindowVerbose(res, t.Config)

		originalSize := len(video.Transcript.ToPlainText())
		resultSize := len(output)

		results = append(results, VideoTestResult{
			Name:             t.Name,
			OriginalSize:     originalSize,
			ResultSize:       resultSize,
			ReductionPercent: float64(originalSize-resultSize) / float64(originalSize) * 100,
			Output:           output,
		})
	}

	return results
}



func getWindowTestCases() []WindowTestCase {

	return []WindowTestCase{

		// -------------------------
		// DEFAULT BASELINE
		// -------------------------
		{
			Name: "ws_300_default",
			Config: wordstats.AnalysisConfig{
				WindowSize: 300,
				BucketCount: 3,
				UseHeavyStopWords: true,
				MinFreq: 2,
				Depth: 0.8,
			},
		},

		// -------------------------
		// SMALLER WINDOWS (more detail)
		// -------------------------
		{
			Name: "ws_200_dense_3b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 200,
				BucketCount: 3,
				UseHeavyStopWords: true,
				MinFreq: 1,
				Depth: 1.0,
			},
		},
		{
			Name: "ws_200_dense_2b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 200,
				BucketCount: 2,
				UseHeavyStopWords: true,
				MinFreq: 1,
				Depth: 1.0,
			},
		},

		// -------------------------
		// MID RANGE WINDOWS (stable)
		// -------------------------
		{
			Name: "ws_300_balanced_3b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 300,
				BucketCount: 3,
				UseHeavyStopWords: true,
				MinFreq: 2,
				Depth: 0.8,
			},
		},
		{
			Name: "ws_300_balanced_2b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 300,
				BucketCount: 2,
				UseHeavyStopWords: true,
				MinFreq: 2,
				Depth: 0.8,
			},
		},

		// -------------------------
		// LARGER WINDOWS (compression heavy)
		// -------------------------
		{
			Name: "ws_450_compact_3b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 450,
				BucketCount: 3,
				UseHeavyStopWords: true,
				MinFreq: 3,
				Depth: 0.6,
			},
		},
		{
			Name: "ws_450_compact_2b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 450,
				BucketCount: 2,
				UseHeavyStopWords: true,
				MinFreq: 3,
				Depth: 0.6,
			},
		},

		// -------------------------
		// MAX WINDOW RANGE (heavy compression)
		// -------------------------
		{
			Name: "ws_600_ultra_3b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 600,
				BucketCount: 3,
				UseHeavyStopWords: true,
				MinFreq: 4,
				Depth: 0.5,
			},
		},
		{
			Name: "ws_600_ultra_2b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 600,
				BucketCount: 2,
				UseHeavyStopWords: true,
				MinFreq: 5,
				Depth: 0.4,
			},
		},

		// -------------------------
		// EDGE COMPRESSION STRESS TEST
		// -------------------------
		{
			Name: "ws_300_extreme_compress_2b",
			Config: wordstats.AnalysisConfig{
				WindowSize: 300,
				BucketCount: 2,
				UseHeavyStopWords: true,
				MinFreq: 8,
				Depth: 0.3,
			},
		},
	}
}


func DebugWindowTranscript(video youtube.Video) error {

	if video.Transcript == nil {
		return fmt.Errorf("video transcript not available")
	}

	fmt.Println("Window Signal Testing")

	svc := wordstats.NewService()

	tests := getWindowTestCases()

	results := runWindowTests(svc, &video, tests)

	for _, r := range results {
		PrintVideoTestResult(r)
	}

	suite := VideoTestSuite{
		VideoID: video.ID,
		Title:   video.Title,
		Results: results,
	}

	return ExportVideoTestResults(suite)
}