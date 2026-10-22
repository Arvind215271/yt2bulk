package subtitle

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/export"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
	"github.com/Arvind215271/yt2bulk/internal/youtube/preflight"
	"github.com/Arvind215271/yt2bulk/internal/youtube/resource"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	resourceService  *resource.Service
	subtitleService  *subtitle.SubtitleService
	exportService    *export.Service
	preflightService preflight.Service
}

func NewHandler(
	resourceService *resource.Service,
	subtitleService *subtitle.SubtitleService,
	exportService *export.Service,
	preflightService preflight.Service,
) *Handler {
	return &Handler{
		resourceService:  resourceService,
		subtitleService:  subtitleService,
		exportService:    exportService,
		preflightService: preflightService,
	}
}

func (h *Handler) GetSubtitleOptions(c *echo.Context) error {
	var req SubtitleOptionsRequest
	if err := c.Bind(&req); err != nil {
		return Err.BadRequest("invalid request body").Wrap(err)
	}

	if len(req.Inputs) == 0 {
		return Err.BadRequest("inputs required")
	}

	parsedInputsForPreflight := make([]*youtubeurl.YouTubeInput, len(req.Inputs))
	for i, in := range req.Inputs {
		parsed, err := youtubeurl.Parse(in)
		if err == nil && parsed != nil {
			parsedInputsForPreflight[i] = parsed
		} else {
			parsedInputsForPreflight[i] = &youtubeurl.YouTubeInput{InputType: youtubeurl.InputTypeVideo, ID: in}
		}
	}
	preflightReq := preflight.Request{Inputs: parsedInputsForPreflight}
	ctx := c.Request().Context()
	res, err := h.preflightService.Check(ctx, preflightReq)
	if err != nil {
		return Err.BadRequest(err.Error()).Wrap(err)
	}
	if !res.Valid {
		return Err.BadRequest(res.Error.Error()).Wrap(res.Error)
	}

	fieldPlanner, err := fields.NewPlanner([]string{"id", "subtitle_metadata"})
	if err != nil {
		return err
	}

	inputs := make([]youtubeurl.YouTubeInput, len(req.Inputs))
	for i, in := range req.Inputs {
		parsed, err := youtubeurl.Parse(in)
		if err == nil && parsed != nil {
			inputs[i] = *parsed
		} else {
			inputs[i] = youtubeurl.YouTubeInput{InputType: youtubeurl.InputTypeVideo, ID: in}
		}
	}

	executionPlan := planner.Build(inputs, fieldPlanner)

	pipelineReq := &pipeline.Request{
		FieldPlanner:  fieldPlanner,
		ExecutionPlan: executionPlan,
	}

	resources := h.resourceService.ProcessResources(ctx, inputs, pipelineReq)

	data, err := h.exportService.ExportBatchResource(resources, export.FormatJSON, fieldPlanner, nil)
	if err != nil {
		return err
	}

	contentType := "application/json"

	return c.Blob(
		http.StatusOK,
		contentType,
		data,
	)
}

func (h *Handler) DownloadSubtitle(c *echo.Context) error {
	var req SubtitleDownloadRequest
	if err := c.Bind(&req); err != nil {
		return Err.BadRequest("invalid request body").Wrap(err)
	}

	if len(req.Inputs) == 0 {
		return Err.BadRequest("inputs required")
	}

	parsedInputsForPreflight := make([]*youtubeurl.YouTubeInput, len(req.Inputs))
	for i, in := range req.Inputs {
		parsed, err := youtubeurl.Parse(in)
		if err == nil && parsed != nil {
			parsedInputsForPreflight[i] = parsed
		} else {
			parsedInputsForPreflight[i] = &youtubeurl.YouTubeInput{InputType: youtubeurl.InputTypeVideo, ID: in}
		}
	}
	preflightReq := preflight.Request{Inputs: parsedInputsForPreflight}
	ctx := c.Request().Context()
	_, err := h.preflightService.Check(ctx, preflightReq)
	if err != nil {
		return Err.BadRequest(err.Error()).Wrap(err)
	}

	parsedInputs := make([]youtubeurl.YouTubeInput, len(req.Inputs))
	hasPlaylist := false
	hasVideo := false
	playlistCount := 0

	for i, in := range req.Inputs {
		parsed, err := youtubeurl.Parse(in)
		if err == nil && parsed != nil {
			parsedInputs[i] = *parsed
		} else {
			parsedInputs[i] = youtubeurl.YouTubeInput{InputType: youtubeurl.InputTypeVideo, ID: in}
		}

		if parsedInputs[i].InputType == youtubeurl.InputTypePlaylist {
			hasPlaylist = true
			playlistCount++
		} else {
			hasVideo = true
		}
	}

	if playlistCount > 1 {
		return Err.MultiplePlaylistsNotAllowed()
	}
	if hasPlaylist && hasVideo {
		return Err.MixedResourcesNotAllowed()
	}

	fieldPlanner, err := fields.NewPlanner([]string{"id", "title", "subtitle_metadata"})
	if err != nil {
		return err
	}

	preferences, err := BuildPreferences(req.Preferences)
	if err != nil {
		return Err.InvalidPreference(err)
	}

	format := req.Format
	if format == "" {
		format = "vtt"
	}

	inputs := parsedInputs
	executionPlan := planner.Build(inputs, fieldPlanner)

	pipelineReq := &pipeline.Request{
		FieldPlanner:  fieldPlanner,
		ExecutionPlan: executionPlan,
		Subtitle: &subtitle.DownloadRequest{
			Format: format,
		},
		Preferences: preferences,
		Format:      format,
	}

	resources := h.resourceService.ProcessResources(ctx, inputs, pipelineReq)

	type subFile struct {
		name    string
		content []byte
	}

	var files []subFile

	for _, res := range resources {
		if res.Type == youtube.ResourceTypeVideo && res.Video != nil {
			v := res.Video
			downloadReq, err := h.subtitleService.ResolveDownloadRequest(v.SubtitleMetadata, preferences, format)
			if err == nil {
				downloadReq.VideoID = v.ID
				subRes, err := h.subtitleService.DownloadSubtitle(ctx, downloadReq, v.SubtitleMetadata)
				if err == nil {
					filename := fmt.Sprintf("%s.%s.%s", v.ID, subRes.Language, subRes.Format)
					files = append(files, subFile{name: filename, content: subRes.Content})
				}
			}
		} else if res.Type == youtube.ResourceTypePlaylist && res.Playlist != nil {
			p := res.Playlist
			for idx, v := range p.Videos {
				downloadReq, err := h.subtitleService.ResolveDownloadRequest(v.SubtitleMetadata, preferences, format)
				if err == nil {
					downloadReq.VideoID = v.ID
					subRes, err := h.subtitleService.DownloadSubtitle(ctx, downloadReq, v.SubtitleMetadata)
					if err == nil {
						titleSlug := v.ID
						if v.Title != "" {
							titleSlug = sanitizeFilename(v.Title)
						}
						filename := fmt.Sprintf("%03d-%s.%s.%s", idx+1, titleSlug, subRes.Language, subRes.Format)
						files = append(files, subFile{name: filename, content: subRes.Content})
					}
				}
			}
		}
	}

	if len(files) == 0 {
		return Err.NoSubtitlesFound()
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	for _, f := range files {
		w, err := zipWriter.Create(f.name)
		if err != nil {
			return Err.ZipCreationFailed(err)
		}
		_, err = w.Write(f.content)
		if err != nil {
			return Err.ZipCreationFailed(err)
		}
	}
	err = zipWriter.Close()
	if err != nil {
		return Err.ZipCreationFailed(err)
	}

	c.Response().Header().Set("Content-Type", "application/zip")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=subtitles.zip")
	return c.Blob(http.StatusOK, "application/zip", buf.Bytes())
}

func sanitizeFilename(s string) string {
	buf := bytes.NewBuffer(nil)
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			buf.WriteRune(r)
		} else if r == ' ' {
			buf.WriteRune('-')
		}
	}
	if buf.Len() == 0 {
		return "video"
	}
	return buf.String()
}
