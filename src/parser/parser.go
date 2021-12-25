package parser

import (
	"github.com/hauke96/sigolo"
	"sort"
)

const TEMPLATE_FOLDER = "./templates/"

const IMAGE_REGEX = `\[\[((Datei|File):([^|^\]]*))(\|([^\]]*))?]]`

func Parse(content string, title string) Article {
	tokenMap := map[string]string{}

	content = tokenize(content, tokenMap)

	sigolo.Info("Token map length: %d", len(tokenMap))

	// print some debug information if wanted
	if sigolo.LogLevel >= sigolo.LOG_DEBUG {
		sigolo.Debug(content)

		keys := make([]string, 0, len(tokenMap))
		for k := range tokenMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sigolo.Debug("%s : %s\n", k, tokenMap[k])
		}
	}

	return Article{
		Title:    title,
		TokenMap: tokenMap,
		Images:   images,
		Content:  content,
	}
}
