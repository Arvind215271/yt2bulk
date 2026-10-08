package preflight

type PreflightRequest struct {
	Inputs []string `json:"inputs"`
}

type PreflightStatsResponse struct {
	Type             string `json:"type"`
	ItemCount        int    `json:"item_count"`
	MetadataCached   int    `json:"metadata_cached"`
	MetadataUpstream int    `json:"metadata_upstream"`
	SubtitleCached   int    `json:"subtitle_cached"`
	SubtitleUpstream int    `json:"subtitle_upstream"`
}

type PlaylistPreflightStatsResponse struct {
	Type           string                 `json:"type"`
	TotalPlaylists int                    `json:"total_playlists"`
	TotalVideos    int                    `json:"total_videos"`
	Items          PreflightStatsResponse `json:"items"`
}

type PreflightResponse struct {
	Valid              bool                           `json:"valid"`
	Error              string                         `json:"error,omitempty"`
	UniqueVideoCount   int                            `json:"unique_video_count"`
	TotalResourceCount int                            `json:"total_resource_count"`
	EstimatedWork      int                            `json:"estimated_work"`
	DirectVideos       PreflightStatsResponse         `json:"direct_videos"`
	Playlists          PlaylistPreflightStatsResponse `json:"playlists"`
	Combined           PreflightStatsResponse         `json:"combined"`
}
