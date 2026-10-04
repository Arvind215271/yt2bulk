package preflight

type PreflightRequest struct {
	Inputs []string `json:"inputs"`
}

type PreflightResponse struct {
	Valid              bool   `json:"valid"`
	Error              string `json:"error,omitempty"`
	UniqueVideoCount   int    `json:"unique_video_count"`
	TotalResourceCount int    `json:"total_resource_count"`
	EstimatedWork      int    `json:"estimated_work"`
}
