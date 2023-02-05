package parser

import (
	"fmt"
	"strings"
)

func (t *Tokenizer) parseHeadings(content string) string {
	lines := strings.Split(content, "\n")

	// Start with large headings to only match them and then go down in size to match smaller ones.
	for headingDepth := 7; headingDepth > 0; headingDepth-- {
		headingMediawikiMarker := strings.Repeat("=", headingDepth)

		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])

			if strings.HasPrefix(line, headingMediawikiMarker) && strings.HasSuffix(line, headingMediawikiMarker) {
				headingText := strings.ReplaceAll(line, headingMediawikiMarker, "")
				headingText = strings.TrimSpace(headingText)

				token := t.getToken(fmt.Sprintf(TOKEN_HEADING_TEMPLATE, headingDepth))
				t.setToken(token, headingText)
				lines[i] = token
			}
		}
	}

	return strings.Join(lines, "\n")
}
