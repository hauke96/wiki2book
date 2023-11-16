package parser

import "strings"

type MathToken struct {
	Token
	Content string
}

func (t *Tokenizer) parseMath(content string) string {
	matches := mathRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		tokenKey := t.getToken(TOKEN_MATH)
		t.setRawToken(tokenKey, MathToken{
			Content: match[1],
		})
		content = strings.Replace(content, match[0], tokenKey, 1)
	}
	return content
}
