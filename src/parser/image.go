package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/util"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var images []string

// Remove videos and gifs
var nonImageRegex = regexp.MustCompile(`\[\[((` + FILE_PREFIXES + `):.*?\.(webm|gif|ogv|mp3|mp4|ogg|wav)).*(]]|\|)`)
var imagePrefixRegex = regexp.MustCompile("^(" + FILE_PREFIXES + ")")
var imageRegex = regexp.MustCompile(IMAGE_REGEX_PATTERN)

// escapeImages escapes the image names in the content and returns the updated content.
func escapeImages(content string) string {
	var result []string

	content = nonImageRegex.ReplaceAllString(content, "")

	submatches := imageRegex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		filePrefix := submatch[2]
		filename := submatch[3]
		filename = strings.TrimSpace(filename)

		// Replace spaces with underscore because wikimedia doesn't know spaces in file names:
		filename = strings.ReplaceAll(filename, " ", "_")
		filename = strings.ReplaceAll(filename, "%20", "_")

		// Make first letter upper case as wikimedia wants it this way:
		filenameRunes := []rune(filename)
		filename = strings.ToUpper(string(filenameRunes[0])) + string(filenameRunes[1:])

		// Attach prefix like "File:" to reconstruct the wikitext entry:
		filename = filePrefix + ":" + filename

		content = strings.ReplaceAll(content, submatch[1], filename)
		result = append(result, filename)

		sigolo.Debug("Found image: %s", filename)
	}

	images = append(images, result...)

	firstPartOfContent := util.TruncString(content)

	sigolo.Debug("Found and embedded %d images in content %s", len(submatches), firstPartOfContent)
	return content
}

func (t *Tokenizer) parseGalleries(content string) string {
	lines := strings.Split(content, "\n")
	withinGallery := false
	var resultLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Gallery ends -> Simply remove line and end "withinGallery" mode
		if strings.HasPrefix(trimmedLine, "</gallery>") {
			withinGallery = false

			if trimmedLine == "</gallery>" {
				// This line just contains the tag -> ignore it and proceed with parsing
				continue
			}

			// If the line contains more than the closing tag -> Keep it and proceed with the processing
			line = strings.ReplaceAll(line, "</gallery>", "")
		} else if galleryStartRegex.MatchString(trimmedLine) {
			withinGallery = true

			// Gallery starts -> Remove tag and see if the line also contains the first image
			trimmedLine = galleryStartRegex.ReplaceAllString(trimmedLine, "")
			if trimmedLine != "" {
				// This line contained more than just the start tag -> handle line again
				lines[i] = trimmedLine
				i--
			}

			continue
		} else if withinGallery {
			// We're within a gallery -> turn each line into a correct wikitext image with "[[File:...]]"

			if trimmedLine == "" {
				continue
			}

			if !hasNonInlineParameterRegex.MatchString(trimmedLine) {
				// Line has no non-inline parameter -> Add one to make it a non-inline image in further image parsing/escaping
				lineSegments := strings.Split(trimmedLine, "|")
				// The last parameter is the caption, so the non-inline parameter is added right behind the file name
				lineSegments[0] += "|mini"
				trimmedLine = strings.Join(lineSegments, "|")
			}

			if !imagePrefixRegex.MatchString(trimmedLine) {
				// Files with and without "File:" prefixes are allowed. This line has no such prefix -> add valid prefix
				trimmedLine = "File:" + trimmedLine
			}

			trimmedLine = fmt.Sprintf("[[%s]]", trimmedLine)
			line = escapeImages(trimmedLine)
		}

		// Normal line or line has been processed -> anyway, add it to the result list
		resultLines = append(resultLines, line)
	}

	content = strings.Join(resultLines, "\n")
	return content
}

func (t *Tokenizer) parseImageMaps(content string) string {
	lines := strings.Split(content, "\n")

	withinImageMap := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Delete uninteresting lines (end of map or all the polygon-map-stuff in between)
		if withinImageMap || line == "</imagemap>" {
			// delete this line i
			lines = append(lines[:i], lines[i+1:]...)
			i--

			// Imagemap ends -> end "withinImageMap" mode
			if line == "</imagemap>" {
				withinImageMap = false
			}

			continue
		}

		// Image map starts -> Parse the image
		if imagemapStartRegex.MatchString(line) {
			line = strings.TrimSpace(imagemapStartRegex.ReplaceAllString(line, ""))
			if line == "" {
				// if empty -> delete this line i, the next line contains the image
				lines = append(lines[:i], lines[i+1:]...)
				line = lines[i]
			}

			// "line" contains definitely the image of the imagemap
			lines[i] = escapeImages(fmt.Sprintf("[[%s]]", line))

			withinImageMap = true
			continue
		}
	}

	content = strings.Join(lines, "\n")
	return content
}

func (t *Tokenizer) parseImages(content string) string {
	submatches := imageRegex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		filename := filepath.Join(t.imageFolder, submatch[3])
		filenameToken := t.getToken(TOKEN_IMAGE_FILENAME)
		t.setRawToken(filenameToken, filename)

		tokenString := TOKEN_IMAGE_INLINE
		imageSizeToken := ""
		captionToken := ""
		if len(submatch) >= 4 {
			options := strings.Split(submatch[5], "|")

			for i, option := range options {
				if util.ElementHasPrefix(option, imageNonInlineParameters) {
					tokenString = TOKEN_IMAGE
				} else if util.ElementHasPrefix(option, imageIgnoreParameters) {
					continue
				} else if strings.HasSuffix(option, "px") && tokenString != TOKEN_IMAGE {
					option = strings.TrimSuffix(option, "px")
					sizes := strings.Split(option, "x")

					xSize := sizes[0]
					ySize := ""
					if len(sizes) == 2 {
						ySize = sizes[1]
					}

					xSizeInt, _ := strconv.Atoi(xSize)
					ySizeInt, _ := strconv.Atoi(ySize)
					// Too large images should not be considered inline. The exact values are just guesses and may change over time.
					if ySizeInt > 50 && xSizeInt > 100 {
						tokenString = TOKEN_IMAGE
					}

					imageSizeString := fmt.Sprintf("%sx%s", xSize, ySize)
					imageSizeToken = t.getToken(TOKEN_IMAGE_SIZE)
					t.setRawToken(imageSizeToken, imageSizeString)
				} else if i == len(options)-1 && tokenString == TOKEN_IMAGE {
					// last remaining option is caption as long as we do NOT have an inline image
					captionToken = t.getToken(TOKEN_IMAGE_CAPTION)
					t.setToken(captionToken, option)
				}
			}
		}

		token := t.getToken(tokenString)
		t.setRawToken(token, filenameToken+" "+captionToken+" "+imageSizeToken)

		content = strings.Replace(content, submatch[0], token, 1)
	}

	return content
}
