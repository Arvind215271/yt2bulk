package export

import (
	youtube "github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
)

// BuildPlaylist converts a playlist domain model into common ExportData.
func BuildPlaylist(
	playlist youtube.Playlist,
	planner *fields.Planner,
) (ExportData, error) {

	result, err := exportStruct(playlist, planner)
	if err != nil {
		return nil, youtube.Err.Export.MarshalFailed().Wrap(err)
	}

	if len(playlist.Videos) > 0 {
		videos := make([]any, 0, len(playlist.Videos))
		for _, v := range playlist.Videos {
			videoData, err := BuildVideo(v.Video, planner)
			if err != nil {
				return nil, err
			}

			videoData["position"] = v.Position
			videoData["added_at"] = v.AddedAt
			videoData["id"] = v.Video.ID

			videos = append(videos, videoData)
		}
		result["videos"] = videos
	}

	return result, nil
}

// BuildVideo converts a video domain model into common ExportData.
func BuildVideo(
	video youtube.Video,
	planner *fields.Planner,
) (ExportData, error) {

	data, err := exportStruct(video, planner)
	if err != nil {
		return nil, youtube.Err.Export.MarshalFailed().Wrap(err)
	}

	return data, nil
}

// BuildBatchVideo converts multiple video domain models into common ExportData.
func BuildBatchVideo(
	videos []youtube.Video,
	planner *fields.Planner,
) (ExportData, error) {
	exportedVideos := make([]any, 0, len(videos))

	for _, v := range videos {
		videoData, err := exportStruct(v, planner)
		if err != nil {
			return nil, youtube.Err.Export.MarshalFailed().Wrap(err)
		}
		exportedVideos = append(exportedVideos, videoData)
	}

	return ExportData{
		"videos": exportedVideos,
	}, nil
}

// BuildResource converts a single youtube.Resource container into common ExportData.
func BuildResource(
	resource youtube.Resource,
	planner *fields.Planner,
) (ExportData, error) {
	var data ExportData
	var err error

	switch resource.Type {
	case youtube.ResourceTypePlaylist:
		if resource.Playlist != nil {
			data, err = BuildPlaylist(*resource.Playlist, planner)
		}
	case youtube.ResourceTypeVideo:
		fallthrough
	default:
		if resource.Video != nil {
			data, err = BuildVideo(*resource.Video, planner)
		}
	}

	if err != nil {
		return nil, err
	}

	if data == nil {
		data = make(ExportData)
	}

	data["id"] = resource.ID
	data["type"] = resource.Type

	return data, nil
}

// BuildBatchResource converts multiple youtube.Resource containers into common ExportData.
func BuildBatchResource(
	resources []youtube.Resource,
	planner *fields.Planner,
) (ExportData, error) {
	exportedResources := make([]any, 0, len(resources))

	for _, res := range resources {
		resData, err := BuildResource(res, planner)
		if err != nil {
			return nil, err
		}
		exportedResources = append(exportedResources, resData)
	}

	return ExportData{
		"resources": exportedResources,
	}, nil
}
