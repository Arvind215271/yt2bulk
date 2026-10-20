package wordstats

import transcript "github.com/Arvind215271/yt2bulk/internal/youtube/transcript"



type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) AnalyzeVideoWordStats(
	t *transcript.Transcript,
	cfg AnalysisConfig,
) Result {

	raw := BuildWordStats(t)

	return RunPipeline(raw, cfg)
}



func RunPipeline(
	raw map[string]*WordStats,
	cfg AnalysisConfig,
) Result {

	scored := ScoreWords(raw)

	if cfg.UseHeavyStopWords {
		scored = FilterHeavyStopWords(scored)
	}

	scored = FilterWords(scored, FilterConfig{
		MinFreq:  cfg.MinFreq,
		Depth:    cfg.Depth,
		MaxChars: cfg.MaxChars,
	})

	buckets := CreateBuckets(scored, BucketConfig{
		Count: cfg.BucketCount,
	})

	outputChars := 0
	for _, w := range scored {
		outputChars += len(w.Word)
	}

	return Result{
		Buckets: buckets,

		OriginalWords: len(raw),
		KeptWords:     len(scored),

		OriginalChars: 0, // window/global decides this externally
		OutputChars:   outputChars,
	}
}



func (s *Service) AnalyzeVideoWindowed(
	t *transcript.Transcript,
	cfg AnalysisConfig,
) []Result {

	windows := BuildWindowStats(t, cfg.WindowSize)

	out := make([]Result, 0, len(windows))

	for _, w := range windows {

		res := RunPipeline(w.Words, cfg)

		// attach correct source stats
		res.OriginalWords = w.OriginalWords
		res.OriginalChars = w.OriginalChars

		out = append(out, res)
	}

	return out
}