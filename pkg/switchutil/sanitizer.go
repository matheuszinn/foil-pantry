package switchutil

import (
	"regexp"
	"strings"
)

var bracketRegex = regexp.MustCompile(`\[[^\]]+\]`)
var spaceRegex = regexp.MustCompile(`\s+`)

// CleanTitle removes bracketed tags (e.g. [Update], [DLC], [USA], [0100...]) from a title,
// preserving the file extension (.torrent), and standardizes whitespace.
func CleanTitle(title string) string {
	// Check if there is a .torrent extension and temporarily strip it to simplify cleaning
	hasTorrentSuffix := strings.HasSuffix(strings.ToLower(title), ".torrent")
	base := title
	if hasTorrentSuffix {
		base = title[:len(title)-len(".torrent")]
	}

	// Remove all bracketed tags
	cleaned := bracketRegex.ReplaceAllString(base, " ")

	// Replace multiple whitespace characters with a single space
	cleaned = spaceRegex.ReplaceAllString(cleaned, " ")

	// Trim leading/trailing whitespace
	cleaned = strings.TrimSpace(cleaned)

	if hasTorrentSuffix {
		cleaned = cleaned + ".torrent"
	}

	return cleaned
}
