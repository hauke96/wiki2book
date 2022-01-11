package parser

import (
	"github.com/hauke96/sigolo"
	"sort"
)

const IMAGE_REGEX = `\[\[((Datei|File):([^|^\]]*))(\|([^\]]*))?]]`

func Parse(content string, title string, tokenizer Tokenizer) Article {
	content = tokenizer.tokenize(content)

	sigolo.Info("Token map length: %d", len(tokenizer.tokenMap))

	// print some debug information if wanted
	if sigolo.LogLevel >= sigolo.LOG_DEBUG {
		sigolo.Debug(content)

		keys := make([]string, 0, len(tokenizer.tokenMap))
		for k := range tokenizer.tokenMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sigolo.Debug("%s : %s\n", k, tokenizer.tokenMap[k])
		}
	}

	return Article{
		Title:    title,
		TokenMap: tokenizer.tokenMap,
		Images:   images,
		Content:  content,
	}
}
