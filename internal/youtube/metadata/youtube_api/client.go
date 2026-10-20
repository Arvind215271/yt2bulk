package youtubeapi

import (
	"context"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	youtube "github.com/Arvind215271/yt2bulk/internal/youtube"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

type Client struct {
	service *yt.Service
	logger  *logger.Logger
}

func NewClient(
	ctx context.Context,
	apiKey string,
	logger *logger.Logger,
) (*Client, error) {

	service, err := yt.NewService(
		ctx,
		option.WithAPIKey(apiKey),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		service: service,
		logger:  logger,
	}, nil
}

// returns metadata about the playlist, but does not contain any videoID in it
func (c *Client) GetPlaylist(
	ctx context.Context,
	playlistID string,
) (*yt.Playlist, error) {

	c.logger.Debug("starting GetPlaylist", "playlistID", playlistID)

	resp, err := c.service.Playlists.
		List([]string{
			"snippet",
			"status",
			"contentDetails",

		}).
		Id(playlistID).
		Context(ctx).
		Do()

	if err != nil {
		return nil, youtube.Err.Playlist.FetchFailed().Wrap(err)
	}

	if len(resp.Items) == 0 {
		return nil, youtube.Err.Playlist.NotFound()
	}

	item := resp.Items[0]

	c.logger.Debug("finished GetPlaylist", "playlistID", playlistID, "title", item.Snippet.Title)
	c.logger.Info("playlist metadata fetched", "playlistID", playlistID)

	return item, nil
}
// this returns the videoID in the playlist with some data
func (c *Client) GetPlaylistItems(
	ctx context.Context,
	playlistID string,
) ([]*yt.PlaylistItem, error) {

	c.logger.Debug("starting GetPlaylistItems", "playlistID", playlistID)

	var items []*yt.PlaylistItem
	// items represent an array of playlist items.
	// the reason is that YT api returns upto 0 to 50 videoID per call.
	// Thus, to retrieve them all, we need to append those into the system for them to be resuable and such.
	// so we have to use the pagetoken for the upcoming ID if exist.
	var pageToken string

	for {
		c.logger.Debug("fetching page of playlist items", "playlistID", playlistID, "pageToken", pageToken)
		resp, err := c.service.PlaylistItems.
			List([]string{
				"snippet",
				"contentDetails",
			}).
			PlaylistId(playlistID).
			MaxResults(50).
			PageToken(pageToken).
			Context(ctx).
			Do()

		if err != nil {
			return nil, youtube.Err.Playlist.FetchFailed().Wrap(err)
		}
		// add the recieved items to an array.
		items = append(items, resp.Items...)
		c.logger.Debug("received items page", "playlistID", playlistID, "count", len(resp.Items))

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	c.logger.Debug("finished GetPlaylistItems", "playlistID", playlistID, "total", len(items))
	c.logger.Info("playlist items fetched", "playlistID", playlistID, "total", len(items))

	return items, nil
}

// fetches actual video resources from YouTube
func (c *Client) GetVideos(
	ctx context.Context,
	videoIDs []string,
) ([]*yt.Video, error) {

	if len(videoIDs) == 0 {
		return nil, nil
	}

	c.logger.Debug("starting GetVideos", "videoIDsCount", len(videoIDs))

	var videos []*yt.Video

	// youtube allows max 50 ids per request
	for start := 0; start < len(videoIDs); start += 50 {

		end := start + 50
		if end > len(videoIDs) {
			end = len(videoIDs)
		}

		chunk := videoIDs[start:end]
		c.logger.Debug("fetching chunk of videos", "chunkSize", len(chunk))

		resp, err := c.service.Videos.
			List([]string{
				"snippet",
				"contentDetails",
				"statistics",
				"status",
			}).
			Id(chunk...).
			Context(ctx).
			Do()

		if err != nil {
			return nil, youtube.Err.Video.FetchFailed().Wrap(err)
		}

		videos = append(videos, resp.Items...)
	}

	c.logger.Debug("finished GetVideos", "totalFetched", len(videos))
	c.logger.Info("videos metadata fetched", "total", len(videos))

	return videos, nil
}
