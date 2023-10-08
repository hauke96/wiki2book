package parser

import (
	"strings"
)

type HeadingToken struct {
	Token
	Content string // TODO replace by "[]*Token" when ready
	Depth   int
}

func (t *Tokenizer) parseHeadings(content string) string {
	lines := strings.Split(content, "\n")

	// Start with large headings to only match them and then go down in size to match smaller ones.
	for depth := 7; depth > 0; depth-- {
		headingMediawikiMarker := strings.Repeat("=", depth)

		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])

			if strings.HasPrefix(line, headingMediawikiMarker) && strings.HasSuffix(line, headingMediawikiMarker) {
				headingText := strings.ReplaceAll(line, headingMediawikiMarker, "")
				headingText = strings.TrimSpace(headingText)

				token := t.getToken(TOKEN_HEADING)
				t.setRawToken(token, &HeadingToken{
					Content: t.tokenizeContent(t, headingText),
					Depth:   depth,
				})
				lines[i] = token
			}
		}
	}

	return strings.Join(lines, "\n")
}
