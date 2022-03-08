package parser

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/util"
	"regexp"
	"strings"
)

var images = []string{}

// escapeImages escapes the image names in the content and returns the updated content.
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

		// Replace spaces with underscore because wikimedia doesn't know spaces in file names:
		filename = strings.ReplaceAll(filename, " ", "_")
		filename = strings.ReplaceAll(filename, "%20", "_")

		// Make first letter upper case as wikimedia want's it this way:
		filename = strings.ToUpper(string(filename[0])) + filename[1:]

		// Attach prefix like "File:" to reconstruct the wikitext entry:
		filename = filePrefix + ":" + filename

		content = strings.ReplaceAll(content, submatch[1], filename)
		result = append(result, filename)

		sigolo.Debug("Found image: %s", filename)
	}

	images = append(images, result...)

	firstPartOfContent := util.TruncString(content)

	sigolo.Info("Found and embedded %d images in content %s", len(submatches), firstPartOfContent)
	return content
}
