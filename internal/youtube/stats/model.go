package stats

type MetadataStats struct {
	Attempts                 int
	Successes                int
	Failures                 int
	CacheHits                int
	CacheMisses              int
	UpstreamFetches          int
	EstimatedUpstreamFetches int
	YTDLPRequests            int
	APIRequests              int
	Fallbacks                int
}

func (s *MetadataStats) Add(other MetadataStats) {
	s.Attempts += other.Attempts
	s.Successes += other.Successes
	s.Failures += other.Failures
	s.CacheHits += other.CacheHits
	s.CacheMisses += other.CacheMisses
	s.UpstreamFetches += other.UpstreamFetches
	s.EstimatedUpstreamFetches += other.EstimatedUpstreamFetches
	s.YTDLPRequests += other.YTDLPRequests
	s.APIRequests += other.APIRequests
	s.Fallbacks += other.Fallbacks
}

type SubtitleStats struct {
	Attempts                 int
	Successes                int
	Failures                 int
	CacheHits                int
	CacheMisses              int
	UpstreamFetches          int
	PotentialSubtitleFetches int
	Fallbacks                int
	ManualRequests           int
	AutomaticRequests        int
}

func (s *SubtitleStats) Add(other SubtitleStats) {
	s.Attempts += other.Attempts
	s.Successes += other.Successes
	s.Failures += other.Failures
	s.CacheHits += other.CacheHits
	s.CacheMisses += other.CacheMisses
	s.UpstreamFetches += other.UpstreamFetches
	s.PotentialSubtitleFetches += other.PotentialSubtitleFetches
	s.Fallbacks += other.Fallbacks
	s.ManualRequests += other.ManualRequests
	s.AutomaticRequests += other.AutomaticRequests
}

type PlaylistStats struct {
	Attempts        int
	Successes       int
	Failures        int
	ItemsDiscovered int
	Fallbacks       int
	Metadata        MetadataStats
}

func (s *PlaylistStats) Add(other PlaylistStats) {
	s.Attempts += other.Attempts
	s.Successes += other.Successes
	s.Failures += other.Failures
	s.ItemsDiscovered += other.ItemsDiscovered
	s.Fallbacks += other.Fallbacks
	s.Metadata.Add(other.Metadata)
}

type PipelineStats struct {
	Metadata MetadataStats
	Subtitle SubtitleStats
}

func (s *PipelineStats) Add(other PipelineStats) {
	s.Metadata.Add(other.Metadata)
	s.Subtitle.Add(other.Subtitle)
}

type ResourceStats struct {
	VideosProcessed int
	VideoFailures   int
	Metadata        MetadataStats
	Subtitle        SubtitleStats
}

func (s *ResourceStats) Add(other ResourceStats) {
	s.VideosProcessed += other.VideosProcessed
	s.VideoFailures += other.VideoFailures
	s.Metadata.Add(other.Metadata)
	s.Subtitle.Add(other.Subtitle)
}

type ProcessStats struct {
	ResourcesRequested int
	ResourcesSucceeded int
	ResourceFailures   int
	VideosProcessed    int
	Metadata           MetadataStats
	Subtitle           SubtitleStats
}

func (s *ProcessStats) Add(other ProcessStats) {
	s.ResourcesRequested += other.ResourcesRequested
	s.ResourcesSucceeded += other.ResourcesSucceeded
	s.ResourceFailures += other.ResourceFailures
	s.VideosProcessed += other.VideosProcessed
	s.Metadata.Add(other.Metadata)
	s.Subtitle.Add(other.Subtitle)
}
