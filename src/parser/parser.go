package parser

import (
	"github.com/hauke96/sigolo"
	"sort"
)

const IMAGE_REGEX = `\[\[((Datei|File):([^|^\]]*))(\|([^\]]*))?]]`

func Parse(content string, title string, tokenizer ITokenizer) Article {
	content = tokenizer.tokenize(content)

	sigolo.Info("Token map length: %d", len(tokenizer.getTokenMap()))

	// print some debug information if wanted
	if sigolo.LogLevel >= sigolo.LOG_DEBUG {
		sigolo.Debug(content)

		keys := make([]string, 0, len(tokenizer.getTokenMap()))
		for k := range tokenizer.getTokenMap() {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			sigolo.Debug("%s : %s\n", k, tokenizer.getTokenMap()[k])
		}
	}

	return Article{
		Title:    title,
		TokenMap: tokenizer.getTokenMap(),
		Images:   images,
		Content:  content,
	}
}
