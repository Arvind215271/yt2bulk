package preflight

// PreflightStats contains accounting for resources.
type PreflightStats struct {
	Type             string
	ItemCount        int
	MetadataCached   int
	MetadataUpstream int
	SubtitleCached   int
	SubtitleUpstream int
}

// Add adds the values from another PreflightStats to this one.
func (s *PreflightStats) Add(other PreflightStats) {
	s.ItemCount += other.ItemCount
	s.MetadataCached += other.MetadataCached
	s.MetadataUpstream += other.MetadataUpstream
	s.SubtitleCached += other.SubtitleCached
	s.SubtitleUpstream += other.SubtitleUpstream
}

// PlaylistPreflightStats contains accounting for a playlist.
type PlaylistPreflightStats struct {
	Type           string
	TotalPlaylists int
	TotalVideos    int
	Items          PreflightStats
}

// Add adds the values from another PlaylistPreflightStats to this one.
func (s *PlaylistPreflightStats) Add(other PlaylistPreflightStats) {
	s.TotalPlaylists += other.TotalPlaylists
	s.TotalVideos += other.TotalVideos
	s.Items.Add(other.Items)
}

// ChannelPreflightStats contains accounting for a channel.
type ChannelPreflightStats struct {
	Type          string
	TotalChannels int
	TotalVideos   int
	Items         PreflightStats
}

// Add adds the values from another ChannelPreflightStats to this one.
func (s *ChannelPreflightStats) Add(other ChannelPreflightStats) {
	s.TotalChannels += other.TotalChannels
	s.TotalVideos += other.TotalVideos
	s.Items.Add(other.Items)
}

// ResourcePreflight contains aggregated accounting for all resources.
type ResourcePreflight struct {
	DirectVideos PreflightStats
	Playlists    PlaylistPreflightStats
	Channels     ChannelPreflightStats
	Combined     PreflightStats
}

// Add adds the values from another ResourcePreflight to this one.
func (s *ResourcePreflight) Add(other ResourcePreflight) {
	s.DirectVideos.Add(other.DirectVideos)
	s.Playlists.Add(other.Playlists)
	s.Channels.Add(other.Channels)
	s.Combined.Add(other.Combined)
}
