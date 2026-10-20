package debug

// import (

// 	"regexp"
// 	"strings"

// 	youtube "github.com/Arvind215271/yt2bulk/internal/youtube"
// 	transcript "github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
// )


// // PlaylistSummary is the highly condensed payload optimized for LLMs
// type PlaylistSummary struct {
// 	PlaylistID   string         `json:"playlist_id"`
// 	Title        string         `json:"title"`
// 	ChannelTitle string         `json:"channel_title"`
// 	TotalVideos  int            `json:"total_videos"`
// 	Curriculum   []VideoSummary `json:"curriculum"`
// }

// type VideoSummary struct {
// 	Position int      `json:"position"`
// 	VideoID  string   `json:"video_id"`
// 	Title    string   `json:"video_title"`
// 	Topics   []string `json:"topics"` // Extracted from Chapters OR Transcript
// }





// // GenerateLLMContext distills your huge Playlist struct into a tiny, context-dense summary
// func GenerateLLMContext(pl youtube.Playlist) PlaylistSummary {
// 	summary := PlaylistSummary{
// 		PlaylistID:   pl.ID,
// 		Title:        pl.Title,
// 		ChannelTitle: pl.ChannelTitle,
// 		TotalVideos:  len(pl.Videos),
// 		Curriculum:   make([]VideoSummary, 0, len(pl.Videos)),
// 	}

// 	titleCleanRegex := regexp.MustCompile(`(?i)\bDay \d+\b|[|\-]`)

// 	for _, pv := range pl.Videos {
// 		// 1. Clean up the main video title
// 		cleanTitle := titleCleanRegex.Split(pv.Title, 2)[0]
// 		cleanTitle = strings.TrimSpace(cleanTitle)

// 		vSummary := VideoSummary{
// 			Position: pv.Position,
// 			VideoID:  pv.ID,
// 			Title:    cleanTitle,
// 			Topics:   []string{},
// 		}

// 		// 2. Strategy A: If human-defined chapters exist, use them!
// 		if len(pv.Video.Chapters.Text()) > 0 {
// 			seenChapters := make(map[string]bool)
// 			for _, ch := range pv.Video.Chapters.List {
// 				cleanCh := sanitizeTopicName(ch.Title)
// 				if cleanCh != "" && !seenChapters[cleanCh] {
// 					seenChapters[cleanCh] = true
// 					vSummary.Topics = append(vSummary.Topics, ch.Title) // Keeps original formatting or use cleanCh
// 				}
// 			}
// 		} else if pv.Transcript != nil && len(pv.Transcript.Segments) > 0 {
// 			// 3. Strategy B: Fallback to transcript processing (Every 3 minutes / 180 seconds)
// 			vSummary.Topics = ExtractTopicsFromTranscript(*pv.Transcript, 180)
// 		}

// 		summary.Curriculum = append(summary.Curriculum, vSummary)
// 	}

// 	return summary
// }

// // Helper to filter out meta-words from chapters (e.g., "Intro", "Outro")
// func sanitizeTopicName(title string) string {
// 	low := strings.ToLower(strings.TrimSpace(title))
// 	fillers := map[string]bool{
// 		"intro": true, "outro": true, "conclusion": true, "summary": true, "recap": true,
// 	}
// 	if fillers[low] || len(low) < 2 {
// 		return ""
// 	}
// 	return low
// }



// var stopWords = map[string]bool{
// 	"the": true, "and": true, "a": true, "of": true, "to": true, "is": true, "in": true,
// 	"that": true, "it": true, "you": true, "we": true, "this": true, "for": true, "on": true,
// 	"machine": true, "learning": true, "data": true, "course": true, "video": true,
// }

// func ExtractTopicsFromTranscript(t transcript.Transcript, windowSeconds float64) []string {
// 	if len(t.Segments) == 0 {
// 		return nil
// 	}

// 	var topics []string
// 	var currentWindowWords []string
// 	windowStart := t.Segments[0].Start
// 	reg := regexp.MustCompile(`[^\w\s]`)

// 	for _, seg := range t.Segments {
// 		if seg.Start-windowStart >= windowSeconds {
// 			if phrase := extractTopPhrase(currentWindowWords); phrase != "" {
// 				topics = append(topics, phrase)
// 			}
// 			currentWindowWords = nil
// 			windowStart = seg.Start
// 		}
// 		cleanText := reg.ReplaceAllString(strings.ToLower(seg.Text), "")
// 		currentWindowWords = append(currentWindowWords, strings.Fields(cleanText)...)
// 	}

// 	if len(currentWindowWords) > 0 {
// 		if phrase := extractTopPhrase(currentWindowWords); phrase != "" {
// 			topics = append(topics, phrase)
// 		}
// 	}

// 	return deduplicate(topics)
// }

// func extractTopPhrase(words []string) string {
// 	if len(words) < 2 {
// 		return ""
// 	}
// 	counts := make(map[string]int)
// 	for i := 0; i < len(words)-1; i++ {
// 		w1, w2 := words[i], words[i+1]
// 		if stopWords[w1] || stopWords[w2] {
// 			continue
// 		}
// 		bigram := w1 + " " + w2
// 		counts[bigram]++

// 		if i < len(words)-2 {
// 			w3 := words[i+2]
// 			if !stopWords[w3] {
// 				counts[bigram+" "+w3] += 2 
// 			}
// 		}
// 	}

// 	best := ""
// 	max := 0
// 	for p, c := range counts {
// 		if c > max {
// 			max = c
// 			best = p
// 		}
// 	}
// 	return strings.Title(best)
// }

// func deduplicate(input []string) []string {
// 	var res []string
// 	seen := make(map[string]bool)
// 	for _, val := range input {
// 		low := strings.ToLower(val)
// 		if !seen[low] {
// 			seen[low] = true
// 			res = append(res, val)
// 		}
// 	}
// 	return res
// }


