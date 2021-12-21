package parser

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"regexp"
	"sort"
	"strings"
)

const IMAGE_REGEX = `\[\[((Datei|File):([^|]*))((.|\n)*?\|([^|]*?))?]]`

func Parse(content string, title string) Article {
	content = clean(content)
	content = evaluateTemplates(content)
	content, images := processImages(content)

	tokenMap := map[string]string{}
	content = tokenize(content, tokenMap)
	sigolo.Debug(content)

	keys := make([]string, 0, len(tokenMap))
	for k := range tokenMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sigolo.Debug("%s : %s\n", k, tokenMap[k])
	}

	return Article{
		Title:    title,
		TokenMap: tokenMap,
		Images:   images,
		Content:  content,
	}
}

func evaluateTemplates(content string) string {
	regex := regexp.MustCompile("\\{\\{(.*?)}}")
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		evaluatedTemplate, err := api.EvaluateTemplate(match)
		if err != nil {
			sigolo.Stack(err)
			return ""
		}
		return evaluatedTemplate
	})
	return content
}

// processImages returns the list of all images and also escapes the image names in the content
func processImages(content string) (string, []string) {
	var result []string

	// Remove videos and gifs
	regex := regexp.MustCompile(`\[\[((Datei|File):.*?\.(webm|gif|ogv|mp3|mp4)).*(]]|\|)`)
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile(IMAGE_REGEX)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		filename := strings.ReplaceAll(submatch[3], " ", "_")
		filename = submatch[2] + ":" + strings.ToUpper(string(filename[0])) + filename[1:]

		content = strings.ReplaceAll(content, submatch[1], filename)

		result = append(result, filename)

		sigolo.Debug("Found image: %s", filename)
	}

	sigolo.Info("Found and embedded %d images", len(submatches))
	return content, result
}
