package parser

import "strings"

func (t *Tokenizer) parseMath(content string) string {
	matches := mathRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		token := t.getToken(TOKEN_MATH)
		t.setRawToken(token, match[1])
		content = strings.Replace(content, match[0], token, 1)
	}
	return content
}
