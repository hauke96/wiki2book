package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"path/filepath"
	"strconv"
	"strings"
	"wiki2book/util"
)

var images []string

var imageIgnoreParameters = []string{
	"alt",
	"alternativtext",
	"baseline",
	"border",
	"bottom",
	"center",
	"class",
	"framed",
	"frameless",
	"gerahmt",
	"hochkant",
	"lang",
	"left",
	"link",
	"links",
	"middle",
	"none",
	"ohne",
	"page",
	"rahmenlos",
	"rand",
	"rechts",
	"right",
	"seite",
	"sprache",
	"sub",
	"super",
	"text-bottom",
	"text-top",
	"top",
	"upright",
	"verweis",
	"zentriert",
}
var imageNonInlineParameters = []string{
	"mini",
	"thumb",
}

// escapeImages escapes the image names in the image specification and returns the updated spec. The spec is expected to
// be the complete spec, not just the image name, so everything between "[[" and "]]". If the media type if not
// supported, an empty string is returned.
func escapeImages(content string) string {
	segments := strings.Split(content, "|")

	fileSegments := strings.SplitN(segments[0], ":", 2)
	var mediaType string
	var filename string
	if len(fileSegments) == 1 {
		mediaType = "File"
		filename = strings.TrimSpace(fileSegments[0])
	} else {
		mediaType = fileSegments[0]
		filename = strings.TrimSpace(fileSegments[1])
	}

	// Check if this media type is unwanted
	fileExtension := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if util.Contains(unwantedMediaTypes, fileExtension) {
		return ""
	}

	sigolo.Debug("Found image: %s", filename)

	// Replace spaces with underscore because wikimedia doesn't know spaces in file names:
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "%20", "_")

	// Make first letter upper case as wikimedia wants it this way:
	filenameRunes := []rune(filename)
	filename = strings.ToUpper(string(filenameRunes[0])) + string(filenameRunes[1:])

	filenameWithMediaType := mediaType + ":" + filename
	images = append(images, filenameWithMediaType)
	segments[0] = filenameWithMediaType

	return strings.Join(segments, "|")
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
		}

		if withinGallery {
			// We're within a gallery -> turn each line into a correct wikitext image with "[[File:...]]"

			if trimmedLine == "" {
				continue
			}

			lineSegments := strings.Split(trimmedLine, "|")

			if !hasNonInlineParameterRegex.MatchString(trimmedLine) {
				// Line has no non-inline parameter -> Add one to make it a non-inline image in further image parsing/escaping
				// The last parameter is the caption, so the non-inline parameter is added right behind the file name
				newLineSegments := make([]string, len(lineSegments)+1)
				for j, v := range lineSegments {
					if j == 0 {
						// Add filename and "mini"
						newLineSegments[0] = lineSegments[0]
						newLineSegments[1] = "mini"
					} else {
						// Add parameter
						newLineSegments[j+1] = v
					}
				}
				trimmedLine = strings.Join(newLineSegments, "|")
			}

			line = escapeImages(trimmedLine)
			if line != "" {
				line = fmt.Sprintf("[[%s]]", line)
			}
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
	startIndex := imageStartRegex.FindStringIndex(content)
	for startIndex != nil {
		// Use the end-index of the match, since it points to the ":" of the "[[File:" match
		endIndex := findCorrespondingCloseToken(content, startIndex[1], "[", "]")

		// +1 to jump over the ":" after "File". Hence, the resulting string starts at the first character of the image name.
		imageContent := content[startIndex[1]:endIndex]
		imageContent = t.tokenizeContent(t, imageContent)
		imageContent = escapeImages(imageContent)

		if imageContent == "" {
			content = content[0:startIndex[0]] + content[endIndex+2:]
		} else {
			options := strings.Split(imageContent, "|")

			filename := strings.SplitN(options[0], ":", 2)[1]
			imageFilepath := filepath.Join(t.imageFolder, filename)
			filenameToken := t.getToken(TOKEN_IMAGE_FILENAME)
			t.setRawToken(filenameToken, imageFilepath)

			tokenKey := TOKEN_IMAGE_INLINE
			imageSizeToken := ""
			captionToken := ""

			// Do some cleanup: Remove definitely uninteresting options.
			var filteredOptions []string
			for _, option := range options {
				if !util.ElementHasPrefix(option, imageIgnoreParameters) {
					filteredOptions = append(filteredOptions, option)
				}
			}

			for i, option := range filteredOptions {
				if util.ElementHasPrefix(option, imageNonInlineParameters) {
					tokenKey = TOKEN_IMAGE
				} else if strings.HasSuffix(option, "px") {
					option = strings.TrimSuffix(option, "px")
					sizes := strings.Split(option, "x")

					// Valid formats:
					//   {width}px
					//   x{height}px
					//   {width}x{height}px
					xSize := strings.TrimSpace(sizes[0])
					ySize := ""

					if len(sizes) == 2 {
						ySize = strings.TrimSpace(sizes[1])
					} else if len(sizes) > 2 {
						sigolo.Error("Invalid size specification %spx of image %s", option, filename)
					}

					xSizeInt, _ := strconv.Atoi(xSize)
					ySizeInt, _ := strconv.Atoi(ySize)
					// Too large images should not be considered inline. The exact values are just guesses and may change over time.
					if ySizeInt >= 50 || xSizeInt >= 100 {
						tokenKey = TOKEN_IMAGE
					}

					imageSizeString := fmt.Sprintf("%sx%s", xSize, ySize)
					imageSizeToken = t.getToken(TOKEN_IMAGE_SIZE)
					t.setRawToken(imageSizeToken, imageSizeString)
				} else if i == len(filteredOptions)-1 && tokenKey == TOKEN_IMAGE {
					// Last remaining option is the caption. We ignore captions on inline images.
					captionToken = t.getToken(TOKEN_IMAGE_CAPTION)
					t.setToken(captionToken, option)
				}
			}

			token := t.getToken(tokenKey)
			resultTokenContent := filenameToken
			if captionToken != "" {
				resultTokenContent += " " + captionToken
			}

			if imageSizeToken != "" {
				resultTokenContent += " " + imageSizeToken
			}
			t.setRawToken(token, resultTokenContent)

			content = content[0:startIndex[0]] + token + content[endIndex+2:]
		}

		// Find next image
		startIndex = imageStartRegex.FindStringIndex(content)
	}

	return content
}
