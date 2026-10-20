package signal

import (
	transcript "github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
	wordstats "github.com/Arvind215271/yt2bulk/internal/youtube/signal/word_stats"
)

type SignalService struct {
	wordStatsService *wordstats.Service
}

func NewSignalService() *SignalService {
	return &SignalService{
		wordStatsService: wordstats.NewService(),
	}
}

func (s *SignalService) AnalyzeWordStats(t *transcript.Transcript, cfg wordstats.AnalysisConfig) wordstats.Result {
	return s.wordStatsService.AnalyzeVideoWordStats(t, cfg)
}

func (s *SignalService) AnalyzeWindowedStats(t *transcript.Transcript, cfg wordstats.AnalysisConfig) []wordstats.Result {
	return s.wordStatsService.AnalyzeVideoWindowed(t, cfg)
}
