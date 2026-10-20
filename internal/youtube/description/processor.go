package description

import (
	// "encoding/json"
	"regexp"
	"strings"

	//"github.com/Arvind215271/yt2bulk/internal/youtube"
)

type Metadata struct {
	Chapters Chapters `json:"chapters"`
	Links    []string `json:"links"`
	Emails   []string `json:"emails"`
	Cleaned  string   `json:"cleaned"`
}

var (
	linkRegex  = regexp.MustCompile(`https?://[^\s]+`)
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
)

func ProcessDescription(description string) Metadata {
	chapters := ExtractChapters(description)

	return Metadata{
		Chapters: chapters,
		Links:    extractLinks(description),
		Emails:   extractEmails(description),
		Cleaned:  cleanDescription(description),
	}
}


func extractLinks(description string) []string {
	return linkRegex.FindAllString(description, -1)
}

func extractEmails(description string) []string {
	return emailRegex.FindAllString(description, -1)
}

func cleanDescription(description string) string {
	// Remove chapters
	cleaned := description
	chapters := ExtractChapters(description)
	cleaned = strings.ReplaceAll(cleaned, chapters.Text(), "")

	// Remove links
	for _, link := range extractLinks(description) {
		cleaned = strings.ReplaceAll(cleaned, link, "")
	}

	// Remove emails
	for _, email := range extractEmails(description) {
		cleaned = strings.ReplaceAll(cleaned, email, "")
	}

	return strings.TrimSpace(cleaned)
}
