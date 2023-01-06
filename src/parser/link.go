package parser

import (
	"github.com/hauke96/wiki2book/src/util"
	"strings"
)

func (t *Tokenizer) parseInternalLinks(content string) string {
	submatches := internalLinkRegex.FindAllStringSubmatch(content, -1)

	for _, submatch := range submatches {
		// Ignore all kind of files, they are parsed elsewhere
		if filePrefixRegex.MatchString(submatch[0]) {
			continue
		}

		articleName := submatch[1]
		linkText := submatch[3]

		tokenArticle := t.getToken(TOKEN_INTERNAL_LINK_ARTICLE)
		t.setRawToken(tokenArticle, articleName)

		if linkText != "" {
			// Use article as text
			linkText = t.tokenizeInline(linkText)
		} else {
			linkText = articleName
		}

		tokenText := t.getToken(TOKEN_INTERNAL_LINK_TEXT)
		t.setRawToken(tokenText, linkText)

		token := t.getToken(TOKEN_INTERNAL_LINK)
		t.setRawToken(token, tokenArticle+" "+tokenText)

		content = strings.Replace(content, submatch[0], token, 1)
	}

	return content
}

func (t *Tokenizer) parseExternalLinks(content string) string {
	submatches := externalLinkRegex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		tokenUrl := t.getToken(TOKEN_EXTERNAL_LINK_URL)
		t.setRawToken(tokenUrl, submatch[2])

		linkText := submatch[2]
		if len(submatch) >= 5 {
			linkText = submatch[4]
		}

		tokenText := t.getToken(TOKEN_EXTERNAL_LINK_TEXT)
		t.setToken(tokenText, linkText)

		token := t.getToken(TOKEN_EXTERNAL_LINK)
		t.setRawToken(token, tokenUrl+" "+tokenText)

		// Remove last characters as it's the first character after the closing  ]  of the file tag.
		totalMatch := submatch[0]
		if totalMatch[len(totalMatch)-1] != ']' {
			totalMatch = util.RemoveLastChar(totalMatch)
		}
		content = strings.Replace(content, totalMatch, submatch[1]+token, 1)
	}

	return content
}
