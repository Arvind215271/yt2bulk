package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/Arvind215271/askito/internal/cache"
	"github.com/Arvind215271/askito/internal/logger"
	"github.com/Arvind215271/askito/internal/youtube"
	"github.com/Arvind215271/askito/internal/youtube/description"
	"github.com/Arvind215271/askito/internal/youtube/export"
	youtubeurl "github.com/Arvind215271/askito/internal/youtube/input"
	"github.com/Arvind215271/askito/internal/youtube/metadata"
	"github.com/Arvind215271/askito/internal/youtube/metadata/ytdlp/python"
	"github.com/Arvind215271/askito/internal/youtube/subtitle"
	"github.com/Arvind215271/askito/internal/youtube/transcript"

	debugvideo "github.com/Arvind215271/askito/debug/testing/video"
)

func DebugInput(
	ctx context.Context,
	log *logger.Logger,
	youtubeSvc *metadata.Service,
	transcriptSvc *transcript.Service,
	exportSvc *export.Service,
) {

	// In debug mode, create a dummy cache manager
	cacheManager := cache.NewManager(cache.Config{
		CacheDir: "./.cache/debug",
		TTLDays:  1,
		MaxFiles: 10,
	}, log)

	pythonPool, err := python.NewSinglePool(1, log, cacheManager)
	if err != nil {
		log.Fatal("failed to create python pool", "error", err)
	}

	inputStr := `
https://youtube.com/playlist?list=PLKnIA16_Rmvbr7zKYQuBfsVkjoLcJgxHH

https://www.youtube.com/watch?v=7HKot-brXFE	
https://www.youtube.com/watch?v=8jLOx1hD3_o 
https://www.youtube.com/watch?v=NWONeJKn6kc
 https://www.youtube.com/watch?v=hDKCxebp88A
https://www.youtube.com/watch?v=Oe421EPjeBE
https://www.youtube.com/watch?v=V_xro1bcAuA
https://www.youtube.com/watch?v=un6ZyFkqFKo
https://www.youtube.com/watch?v=qwAFL1597eM	
https://www.youtube.com/watch?v=LzMnsfqjzkA	
https://www.youtube.com/watch?v=mHxLXzYjQRE	
https://www.youtube.com/watch?v=n1sfrc-RjyM
https://www.youtube.com/watch?v=xwI5OBEnsZU
https://www.youtube.com/watch?v=gmuTjeQUbTM
https://youtu.be/yK1uBHPdp30

random garbage text

`

	results := youtubeurl.ParseMany(
		inputStr,
	)

	if len(results) == 0 {

		log.Warn(
			"no youtube urls found",
		)

		return
	}

	fmt.Printf(
		"\n================ INPUT RESULTS (%d) ================\n",
		len(results),
	)

	for _, result := range results {

		if result.Error != nil {

			log.Error(
				"failed to parse input",
				"error",
				result.Error,
			)

			continue
		}

		item := result.Input

		fmt.Printf(
			"\n====================================================\n",
		)

		fmt.Printf(
			"TYPE : %s\n",
			item.InputType,
		)

		fmt.Printf(
			"ID   : %s\n",
			item.ID,
		)

		fmt.Printf(
			"URL  : %s\n",
			item.NormalizedURL,
		)

		switch item.InputType {

		case youtubeurl.InputTypePlaylist:

			debugPlaylist(
				ctx,
				log,
				youtubeSvc,
				exportSvc,
				item.ID,
			)

		case youtubeurl.InputTypeVideo:

			debugVideo(
				ctx,
				log,
				youtubeSvc,
				transcriptSvc,
				exportSvc,
				item.ID,
				cacheManager,
				pythonPool,
			)

		}
	}
}

func debugPlaylist(
	ctx context.Context,
	log *logger.Logger,
	youtubeSvc *metadata.Service,
	exportSvc *export.Service,
	playlistID string,
) {

	fmt.Printf(
		"\n================ PLAYLIST =================\n",
	)

	playlist, err := youtubeSvc.GetPlaylistMetadata(
		ctx,
		playlistID,
		metadata.ProviderAPI,
	)
	if err != nil {

		log.Error(
			"failed to fetch playlist",
			"playlist_id",
			playlistID,
			"error",
			err,
		)

		return
	}

	items, _, err := youtubeSvc.GetPlaylistItemsWithStats(
		ctx,
		playlistID,
		50,
		metadata.ProviderAPI,
	)
	if err != nil {
		log.Error("failed to fetch playlist items", "playlist_id", playlistID, "error", err)
		return
	}

	for _, item := range items {
		video, _, err := youtubeSvc.GetVideoWithStats(ctx, item.VideoID, metadata.ProviderAPI)
		if err != nil {
			log.Warn("failed to fetch video", "video_id", item.VideoID, "error", err)
			continue
		}

		playlist.Videos = append(playlist.Videos, youtube.PlaylistVideo{
			Video:    video,
			Position: item.Position,
			AddedAt:  item.AddedAt,
		})
	}

	// enrich playlist videos with chapter data
	for i := range playlist.Videos {
		playlist.Videos[i].Video.DescriptionMetadata = description.ProcessDescription(playlist.Videos[i].Video.Description)
	}

	printSize(
		"playlist",
		playlist,
	)

	fmt.Println(
		"Title       :",
		playlist.Title,
	)

	fmt.Println(
		"Channel     :",
		playlist.ChannelTitle,
	)

	fmt.Println(
		"Video Count :",
		len(playlist.Videos),
	)

	if len(playlist.Videos) > 0 {

		first := playlist.Videos[0]

		fmt.Printf(
			"\n-------------- FIRST VIDEO --------------\n",
		)

		fmt.Println(
			"Title      :",
			first.Video.Title,
		)

		fmt.Println(
			"Chapters   :",
			len(first.Video.DescriptionMetadata.Chapters.List),
		)

		fmt.Println(
			"Duration   :",
			first.Video.Duration,
		)

		fmt.Println(
			"Views      :",
			first.Video.ViewCount,
		)

		fmt.Println(
			"Position   :",
			first.Position,
		)

		fmt.Println(
			"Added At   :",
			first.AddedAt,
		)
	}

	exportJSON, err := exportSvc.ExportPlaylist(
		playlist,
		export.PlaylistExportRequest{
			Format: export.FormatJSON,

			VideoFields: nil, // Update for fix
		},
	)
	if err != nil {

		log.Error(
			"failed playlist export",
			"playlist_id",
			playlistID,
			"error",
			err,
		)

		return
	}

	fmt.Printf(
		"\n-------------- EXPORT --------------\n",
	)

	fmt.Printf(
		"Export Size : %.2f KB\n",
		float64(len(exportJSON))/1024,
	)

	filePath := filepath.Join(
		"testdata",
		"youtube",
		"playlists",
		playlistID+".json",
	)

	if err := saveFile(
		filePath,
		exportJSON,
	); err != nil {

		log.Error(
			"failed to save playlist export",
			"path",
			filePath,
			"error",
			err,
		)

	} else {

		fmt.Println(
			"saved:",
			filePath,
		)
	}

	aiText := buildPlaylistAIText(
		playlist,
	)

	aiPath := filepath.Join(
		"testdata",
		"youtube",
		"ai",
		"playlists",
		playlistID+".txt",
	)

	if err := saveFile(
		aiPath,
		[]byte(aiText),
	); err != nil {

		log.Error(
			"failed to save ai video",
			"path",
			aiPath,
			"error",
			err,
		)

	} else {

		fmt.Println(
			"saved:",
			aiPath,
		)
	}
}

func debugVideo(
	ctx context.Context,
	log *logger.Logger,
	youtubeSvc *metadata.Service,
	transcriptSvc *transcript.Service,
	exportSvc *export.Service,
	videoID string,
	cacheManager *cache.Manager,
	pythonPool *python.SinglePool,
) {
	fmt.Printf("\n================ VIDEO =================\n")

	video, err := youtubeSvc.GetVideo(ctx, videoID, metadata.ProviderAPI)
	if err != nil {
		log.Error("failed to fetch video", "video_id", videoID, "error", err)
		return
	}

	// Extract chapters from description
	video.DescriptionMetadata = description.ProcessDescription(video.Description)

	printSize("video", video)

	fmt.Println("Title      :", video.Title)
	fmt.Println("Duration   :", video.Duration)
	fmt.Println("Views      :", video.ViewCount)
	fmt.Println("Channel    :", video.ChannelTitle)

	// Transcript fetch
	subService := subtitle.NewSubtitleService(cacheManager, log, pythonPool)
	result, err := subService.DownloadSubtitle(ctx, subtitle.DownloadRequest{
		VideoID:  videoID,
		Type:     "automatic",
		Language: "en",
		Format:   "json3",
	}, video.SubtitleMetadata)
	if err != nil {
		log.Warn("transcript unavailable", "video_id", videoID, "error", err)
	} else {
		transcriptData, err := transcriptSvc.Parse(result)
		if err != nil {
			log.Error("failed to parse transcript", "video_id", videoID, "error", err)
		} else {
			video.Transcript = transcriptData
			// still normalize for saving
			transcriptPath := filepath.Join(
				"testdata",
				"youtube",
				"transcripts",
				videoID+".txt",
			)

			tmp := transcriptData.GroupByDuration(30)
			transcriptData.Segments = tmp.Segments

			if err := saveFile(
				transcriptPath,
				[]byte(transcriptData.ToTimelineText()),
			); err != nil {
				log.Error("failed to save transcript", "path", transcriptPath, "error", err)
			} else {
				fmt.Println("saved:", transcriptPath)
			}
		}
	}

	exportJSON, err := exportSvc.ExportVideo(
		video,
		export.VideoExportRequest{
			Format: export.FormatJSON,
			Fields: nil, // Update for fix
		},
	)
	if err != nil {
		log.Error("failed video export", "video_id", videoID, "error", err)
		return
	}

	fmt.Printf("\n-------------- EXPORT --------------\n")
	fmt.Printf("Export Size : %.2f KB\n", float64(len(exportJSON))/1024)

	filePath := filepath.Join(
		"testdata",
		"youtube",
		"videos",
		videoID+".json",
	)

	if err := saveFile(filePath, exportJSON); err != nil {
		log.Error("failed to save video export", "path", filePath, "error", err)
	} else {
		fmt.Println("saved:", filePath)
	}

	var transcriptText string
	if video.Transcript != nil {
		transcriptText = video.Transcript.ToTimelineText()
	}

	aiText := buildVideoAIText(video, transcriptText)

	aiPath := filepath.Join(
		"testdata",
		"youtube",
		"ai",
		"videos",
		videoID+".txt",
	)

	if err := saveFile(aiPath, []byte(aiText)); err != nil {
		log.Error("failed to save ai video", "path", aiPath, "error", err)
	} else {
		fmt.Println("saved:", aiPath)
	}

	err = debugvideo.DebugWindowTranscript(video)
	if err != nil {
		log.Error("Window Word Stats Signal failed", "error", err)
	}
}

func printSize(
	name string,
	v any,
) {

	b, _ := json.Marshal(v)

	fmt.Printf(
		"%-20s %10.2f KB\n",
		name,
		float64(len(b))/1024,
	)
}

func saveFile(
	path string,
	data []byte,
) error {

	if err := os.MkdirAll(
		filepath.Dir(path),
		0755,
	); err != nil {
		return err
	}

	return os.WriteFile(
		path,
		data,
		0644,
	)
}

func buildPlaylistAIText(
	playlist youtube.Playlist,
) string {

	var b strings.Builder

	b.WriteString(
		fmt.Sprintf(
			"Playlist Title:\n%s\n\n",
			playlist.Title,
		),
	)

	if playlist.Description != "" {

		b.WriteString(
			"Playlist Description:\n",
		)

		b.WriteString(
			playlist.Description,
		)

		b.WriteString(
			"\n\n",
		)
	}

	b.WriteString(
		fmt.Sprintf(
			"Channel:\n%s\n\n",
			playlist.ChannelTitle,
		),
	)

	for _, pv := range playlist.Videos {

		v := pv.Video

		b.WriteString(
			"----------------------------------------\n",
		)

		b.WriteString(
			fmt.Sprintf(
				"Position: %d\n",
				pv.Position,
			),
		)

		b.WriteString(
			fmt.Sprintf(
				"Title: %s\n\n",
				v.Title,
			),
		)

		if v.DescriptionMetadata.Chapters.Text() != "" {

			b.WriteString(
				"Chapters:\n",
			)

			b.WriteString(
				v.DescriptionMetadata.Chapters.Text(),
			)

			b.WriteString(
				"\n\n",
			)
		}
	}

	return b.String()
}

func buildVideoAIText(
	video youtube.Video,
	transcript string,
) string {

	var b strings.Builder

	b.WriteString(
		fmt.Sprintf(
			"Title:\n%s\n\n",
			video.Title,
		),
	)

	if video.Description != "" {

		b.WriteString(
			"Description:\n",
		)

		b.WriteString(
			video.Description,
		)

		b.WriteString(
			"\n\n",
		)
	}

	if video.DescriptionMetadata.Chapters.Text() != "" {

		b.WriteString(
			"Chapters:\n",
		)

		b.WriteString(
			video.DescriptionMetadata.Chapters.Text(),
		)

		b.WriteString(
			"\n\n",
		)
	}

	if transcript != "" {

		b.WriteString(
			"Transcript:\n",
		)

		b.WriteString(
			transcript,
		)
	}

	return b.String()
}
