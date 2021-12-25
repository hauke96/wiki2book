package parser

import (
	"github.com/hauke96/sigolo"
	"math"
	"regexp"
	"strings"
)

var images = []string{}

// escapeImages returns the list of all images and also escapes the image names in the content
func escapeImages(content string) string {
	var result []string

	// Remove videos and gifs
	regex := regexp.MustCompile(`\[\[((Datei|File):.*?\.(webm|gif|ogv|mp3|mp4|ogg|wav)).*(]]|\|)`)
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile(IMAGE_REGEX)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		filePrefix := submatch[2]
		filename := submatch[3]
		filename = strings.ReplaceAll(filename, " ", "_")
		filename = strings.ReplaceAll(filename, "%20", "_")
		filename = filePrefix + ":" + strings.ToUpper(string(filename[0])) + filename[1:]

		content = strings.ReplaceAll(content, submatch[1], filename)

		result = append(result, filename)

		sigolo.Debug("Found image: %s", filename)
	}

	images = append(images, result...)

	firstPartOfContent := content[:int(math.Min(float64(len(content)), 30))]
	if len(content) > 30 {
		firstPartOfContent += "..."
	}

	sigolo.Info("Found and embedded %d images in content %s", len(submatches), firstPartOfContent)
	return content
}
