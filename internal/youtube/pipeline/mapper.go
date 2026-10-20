package pipeline

import (
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/description"
	wordstats "github.com/Arvind215271/yt2bulk/internal/youtube/signal/word_stats"
	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

// func MapMetadata(video *youtube.Video, meta *youtube.Video) {
// 	video.Title = meta.Title
// 	video.ChannelTitle = meta.ChannelTitle
// 	video.Duration = meta.Duration
// 	video.PublishedAt = meta.PublishedAt
// }

func MapDescription(video *youtube.Video, desc *description.Metadata) {
	video.DescriptionChapters = desc.Chapters.Text()
	video.DescriptionLinks = desc.Links
	video.DescriptionEmails = desc.Emails
	video.DescriptionCleaned = desc.Cleaned
}

func MapTranscript(video *youtube.Video, trans *transcript.Transcript) {
	video.Transcript = trans
	video.TranscriptText = trans.ToTimelineText()
}

func MapSignal(video *youtube.Video, sig *wordstats.Result) {
	video.TranscriptSignal = wordstats.Export(*sig, wordstats.ExportCompact)
}
