// // ./internal/youtube/debug_summary.go
package debug

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"math"
// 	"path/filepath"
// 	"strings"
// 	"regexp"

// 	"github.com/Arvind215271/yt2bulk/internal/logger"
// 	"github.com/Arvind215271/yt2bulk/internal/youtube"
// 	"github.com/Arvind215271/yt2bulk/internal/youtube/chapter"
// 	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
// )

// // DebugLLMContext fetches your playlist, runs your context reduction strategy,
// // and exports both a dense JSON and an ultra-compressed plain text file.
// func DebugLLMContext(
// 	ctx context.Context,
// 	log *logger.Logger,
// 	youtubeSvc *youtube.Service,
// 	transcriptSvc *transcript.Service,
// 	playlistID string,
// ) {
// 	fmt.Printf("\n================ COMPRESSION TEST RUN ================\n")

// 	// 1. Fetch the raw playlist data using your existing service architecture
// 	playlist, err := youtubeSvc.GetPlaylist(ctx, playlistID)
// 	if err != nil {
// 		log.Error("failed to fetch playlist for summary optimization test", "playlist_id", playlistID, "error", err)
// 		return
// 	}

// 	// Calculate baseline size before enrichment/compression metrics
// 	printSize("Raw Playlist Structural Metadata", playlist)

// 	// 2. Enrich the structural entities loop exactly like your layout
// 	for i := range playlist.Videos {
// 		// Populate Chapters
// 		playlist.Videos[i].Video.Chapters = chapter.ExtractChapters(playlist.Videos[i].Video.Description)

// 		// Fallback to Transcripts if Chapters do not exist
// 		if len(playlist.Videos[i].Video.Chapters.Text()) == 0 && playlist.Videos[i].Video.CaptionAvailable {
// 			transcriptData, err := transcriptSvc.Get(ctx, playlist.Videos[i].Video.ID)
// 			if err == nil {
// 				playlist.Videos[i].Video.Transcript = transcriptData
// 			} else {
// 				log.Warn("transcript download skipped during fallback pipeline extraction", "video_id", playlist.Videos[i].Video.ID, "error", err)
// 			}
// 		}
// 	}

// 	// 3. Process data using your newly designed algorithm
// 	summaryPayload := GenerateLLMContext(playlist)

// 	// 4. Output Results and Size Comparisons
// 	fmt.Printf("\n-------------- VERIFICATION METRICS --------------\n")
// 	fmt.Println("Processed Title  :", summaryPayload.Title)
// 	fmt.Println("Channel Identity :", summaryPayload.ChannelTitle)
// 	fmt.Println("Total Curriculum :", summaryPayload.TotalVideos, "Videos Map Items")

// 	// --- A. Dense Compact JSON Generation ---
// 	compressedJSON, err := json.Marshal(summaryPayload)
// 	if err != nil {
// 		log.Error("failed to serialize dense optimized output payload matching target design schemas", "error", err)
// 		return
// 	}
// 	fmt.Printf("Dense JSON Payload Structural Size : %.2f KB (Length: %d chars)\n", float64(len(compressedJSON))/1024, len(compressedJSON))

// 	// --- B. Plain Text Curriculum Generation ---
// 	rawTextPayload := BuildRawCurriculumText(summaryPayload)
// 	fmt.Printf("Raw Text Payload Structural Size   : %.2f KB (Length: %d chars)\n", float64(len(rawTextPayload))/1024, len(rawTextPayload))

// 	// 5. Save Output Artifacts
// 	baseDir := filepath.Join("testdata", "youtube", "optimized_summaries")

// 	// Save the Dense JSON version
// 	jsonPath := filepath.Join(baseDir, playlistID+"_llm_curriculum.json")
// 	if err := saveFile(jsonPath, compressedJSON); err != nil {
// 		log.Error("failed storing generated dense optimization summary context json artifact", "path", jsonPath, "error", err)
// 	} else {
// 		fmt.Println("Successfully Exported JSON Profile:", jsonPath)
// 	}

// 	// Save the Ultra-Compressed Raw Text version
// 	textPath := filepath.Join(baseDir, playlistID+"_llm_curriculum.txt")
// 	if err := saveFile(textPath, []byte(rawTextPayload)); err != nil {
// 		log.Error("failed storing generated dense optimization raw text artifact", "path", textPath, "error", err)
// 	} else {
// 		fmt.Println("Successfully Exported Text Profile:", textPath)
// 	}
// }



// // BuildRawCurriculumText evaluates data purely based on text relationships, formatting, 
// // and structural similarity—making it universally valid for any topic or playlist domain.
// func BuildRawCurriculumText(summary PlaylistSummary) string {
// 	var b strings.Builder
// 	b.WriteString(fmt.Sprintf("Playlist: %s\nChannel: %s\n\n", summary.Title, summary.ChannelTitle))

// 	// Global universal timeline structural noise to filter out
// 	universalFluff := map[string]bool{
// 		"introduction": true, "intro": true, "outro": true, "conclusion": true,
// 		"summary": true, "recap": true, "agenda": true, "lesson": true, "part": true,
// 	}

// 	for _, v := range summary.Curriculum {
// 		var activeTopics []string
// 		seen := make(map[string]bool)

// 		// Clean up video branding/suffixes before running comparisons
// 		cleanVideoTitle := v.Title
// 		if idx := strings.IndexAny(cleanVideoTitle, "|-"); idx != -1 {
// 			cleanVideoTitle = strings.TrimSpace(cleanVideoTitle[:idx])
// 		}
// 		lowVideoTitle := strings.ToLower(cleanVideoTitle)

// 		for _, rawTopic := range v.Topics {
// 			// Isolate words to filter layout punctuation safely
// 			words := strings.Fields(strings.ToLower(rawTopic))
// 			var cleanWords []string

// 			for _, w := range words {
// 				// Clean off structural brackets/punctuation from the word boundaries
// 				w = strings.Trim(w, " -+:,[]()️/.\\•*")
				
// 				// Drop purely numeric markers, single characters, or generic timeline words
// 				if w == "" || len(w) <= 1 || universalFluff[w] {
// 					continue
// 				}
// 				cleanWords = append(cleanWords, w)
// 			}

// 			// Reconstruct the sanitized topic phrase
// 			if len(cleanWords) == 0 {
// 				continue
// 			}
// 			lowTopic := strings.Join(cleanWords, " ")

// 			// 1. Two-Way Containment & Similarity Test
// 			// Drops the topic if it heavily duplicates the main video title (or vice-versa)
// 			if strings.Contains(lowVideoTitle, lowTopic) || strings.Contains(lowTopic, lowVideoTitle) {
// 				continue
// 			}
// 			if computeIntersectionScore(lowTopic, lowVideoTitle) > 0.65 {
// 				continue
// 			}

// 			if !seen[lowTopic] {
// 				seen[lowTopic] = true
// 				// Capitalize first letter of words for clean presentation
// 				activeTopics = append(activeTopics, strings.Title(lowTopic))
// 			}
// 		}

// 		// Write output to string builder inline
// 		if len(activeTopics) > 0 {
// 			b.WriteString(fmt.Sprintf("#%d %s (%s)\n", v.Position, cleanVideoTitle, strings.Join(activeTopics, ", ")))
// 		} else {
// 			b.WriteString(fmt.Sprintf("#%d %s\n", v.Position, cleanVideoTitle))
// 		}
// 	}
// 	return b.String()
// }

// // computeIntersectionScore checks how heavily two topic tokens overlap structurally
// func computeIntersectionScore(s1, s2 string) float64 {
// 	w1 := strings.Fields(s1)
// 	w2 := strings.Fields(s2)
// 	if len(w1) == 0 || len(w2) == 0 {
// 		return 0.0
// 	}

// 	matches := 0
// 	targetMap := make(map[string]bool)
// 	for _, word := range w2 {
// 		targetMap[word] = true
// 	}

// 	for _, word := range w1 {
// 		if targetMap[word] {
// 			matches++
// 		}
// 	}

// 	minSize := math.Min(float64(len(w1)), float64(len(w2)))
// 	return float64(matches) / minSize
// }


// // The exact stop words we are testing against
// // Universal, domain-agnostic stop words common to all spoken English videos
// var stopWordsTestList = map[string]bool{
// 	// Structural & Articles
// 	"the": true, "and": true, "a": true, "of": true, "to": true, "is": true, "in": true,
// 	"that": true, "it": true, "for": true, "on": true, "with": true, "as": true, "at": true,
// 	"by": true, "an": true, "be": true, "this": true, "from": true, "or": true, "are": true,
	
// 	// Pronouns
// 	"you": true, "we": true, "i": true, "me": true, "my": true, "he": true, "she": true,
// 	"they": true, "them": true, "your": true, "our": true, "us": true, "itS": true,

// 	// Verbs & Auxiliaries
// 	"do": true, "does": true, "did": true, "have": true, "has": true, "had": true,
// 	"can": true, "could": true, "will": true, "would": true, "should": true, "was": true,
// 	"were": true, "been": true, "going": true, "go": true, "get": true, "got": true,

// 	// Conversational / Video Filler Words
// 	"yeah": true, "guys": true, "video": true, "think": true, "know": true, "talk": true,
// 	"about": true, "right": true, "basically": true, "cool": true, "like": true, "just": true,
// 	"what": true, "then": true, "here": true, "there": true, "so": true, "okay": true,
// 	"now": true, "actually": true, "something": true, "someone": true,
// }


// // TestVideoTranscriptReduction focuses entirely on testing a single video's transcript size reduction
// func TestVideoTranscriptReduction(pv youtube.Video) []string {
// 	if pv.Transcript == nil || len(pv.Transcript.Segments) == 0 {
// 		fmt.Printf("\n[!] Video #%s has no raw transcript data available to test.\n", pv.ID)
// 		return nil
// 	}

// 	// 1. Combine all raw subtitle segments into one continuous baseline text string
// 	var rawSB strings.Builder
// 	for _, seg := range pv.Transcript.Segments {
// 		rawSB.WriteString(seg.Text + " ")
// 	}
// 	originalRawText := rawSB.String()
// 	originalSize := len(originalRawText)

// 	// 2. Clear punctuation symbols to allow clean word extraction matches
// 	reg := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
// 	cleanText := reg.ReplaceAllString(strings.ToLower(originalRawText), "")
// 	allWords := strings.Fields(cleanText)

// 	// 3. Loop through and extract words ONLY if they do not exist in the stopWords map
// 	var remainingWords []string
// 	for _, word := range allWords {
// 		if stopWordsTestList[word] {
// 			continue
// 		}
// 		// Capitalize first letter of words for clean presentation layout consistency
// 		remainingWords = append(remainingWords, strings.Title(word))
// 	}

// 	// 4. Rebuild the text string after stop words filtration processing to calculate final size
// 	textWithNoStopWords := strings.Join(remainingWords, " ")
// 	newSize := len(textWithNoStopWords)

// 	// 5. Compute metrics and log results to the console line boundaries
// 	reductionPercentage := float64(originalSize-newSize) / float64(originalSize) * 100

// 	fmt.Printf("\n=== [VIDEO TEST] TRANSCRIPT COMPRESSION FOR VIDEO ID: %s ===\n", pv.ID)
// 	fmt.Printf("Video Title            : %s\n", pv.Title)
// 	fmt.Printf("Original Raw Text Size : %d characters (%.2f KB)\n", originalSize, float64(originalSize)/1024)
// 	fmt.Printf("After Stop Words Strip : %d characters (%.2f KB)\n", newSize, float64(newSize)/1024)
// 	fmt.Printf("Space Saved / Reduced  : %.2f%%\n", reductionPercentage)
// 	fmt.Printf("=========================================================\n")

// 	// Return the resulting filtered token slice so it can be stored directly into your schema structures
// 	return remainingWords
// }

