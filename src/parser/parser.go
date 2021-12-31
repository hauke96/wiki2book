package parser

import (
	"github.com/hauke96/sigolo"
	"sort"
)

const IMAGE_REGEX = `\[\[((Datei|File):([^|^\]]*))(\|([^\]]*))?]]`

func Parse(content string, title string, imageFolder string, templateFolder string) Article {
	parser := Parser{
		tokenMap:       map[string]string{},
		tokenCounter:   0,
		imageFolder:    imageFolder,
		templateFolder: templateFolder,
	}

	content = parser.tokenize(content)

	sigolo.Info("Token map length: %d", len(parser.tokenMap))

	// print some debug information if wanted
	if sigolo.LogLevel >= sigolo.LOG_DEBUG {
		sigolo.Debug(content)

		keys := make([]string, 0, len(parser.tokenMap))
		for k := range parser.tokenMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sigolo.Debug("%s : %s\n", k, parser.tokenMap[k])
		}
	}

	return Article{
		Title:    title,
		TokenMap: parser.tokenMap,
		Images:   images,
		Content:  content,
	}
}
