package parser

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"regexp"
	"strings"
)

const TOKEN_REGEX = `\$\$TOKEN_([A-Z_]*)_\d*\$\$`
const TOKEN_TEMPLATE = "$$TOKEN_%s_%d$$"
const TOKEN_INTERNAL_LINK = "INTERNAL_LINK"
const TOKEN_EXTERNAL_LINK = "EXTERNAL_LINK"
const TOKEN_EXTERNAL_LINK_URL = "EXTERNAL_LINK_URL"
const TOKEN_EXTERNAL_LINK_TEXT = "EXTERNAL_LINK_TEXT"

// Marker do not appear in the token map. A marker does not contain further information, it just marks e.g. the start
// and end of a primitive block of content (like a block of bold text)
const MARKER_REGEX = `\$\$TOKEN_[A-Z_]*_\d*\$\$`
const MARKER_BOLD_OPEN = "$$MARKER_BOLD_OPEN$$"
const MARKER_BOLD_CLOSE = "$$MARKER_BOLD_CLOSE$$"
const MARKER_ITALIC_OPEN = "$$MARKER_ITALIC_OPEN$$"
const MARKER_ITALIC_CLOSE = "$$MARKER_ITALIC_CLOSE$$"

var tokenCounter = 0

func getToken(tokenType string) string {
	token := fmt.Sprintf(TOKEN_TEMPLATE, tokenType, tokenCounter)
	tokenCounter++
	return token
}

// https://www.mediawiki.org/wiki/Markup_spec
func tokenize(content string, tokenMap map[string]string) string {
	tokenizationHappened := false
	for {
		content, tokenizationHappened = parseBoldAndItalic(content, tokenMap)
		if tokenizationHappened {
			continue
		}

		content, tokenizationHappened = parseInternalLinks(content, tokenMap)
		if tokenizationHappened {
			continue
		}

		content, tokenizationHappened = parseExternalLinks(content, tokenMap)
		if tokenizationHappened {
			continue
		}

		break
	}

	return content
}

func parseBoldAndItalic(content string, tokenMap map[string]string) (string, bool) {
	index := strings.Index(content, "''")
	if index != -1 {
		content, _, _, _ = tokenizeBoldAndItalic(content, index, tokenMap, false, false)
		return content, true
	}
	return content, false
}

// tokenizeByRegex applies the regex which must have exactly one group. The tokenized content is returned and a flag
// saying if something changed (i.e. is a tokenization happened).
//func tokenizeByRegex(content string, tokenMap map[string]string, regexString string, tokenType string) (string, bool) {
//	regex := regexp.MustCompile(regexString)
//	matches := regex.FindStringSubmatch(content)
//	if len(matches) != 0 {
//		content = processMatch(content, tokenMap, matches[0], matches[1], tokenType)
//		return content, true
//	}
//	return content, false
//}
//
//func processMatch(content string, tokenMap map[string]string, wholeMatch string, untokenizedMatch string, tokenType string) string {
//	token := getToken(tokenType)
//	tokenizedString := Tokenize(untokenizedMatch, tokenMap)
//	tokenMap[token] = tokenizedString
//	return strings.Replace(content, wholeMatch, token, 1)
//}

func tokenizeBoldAndItalic(content string, index int, tokenMap map[string]string, isBoldOpen bool, isItalicOpen bool) (string, int, bool, bool) {
	for {
		// iIn case of last opened italic marker
		sigolo.Info("Check index %d of %d long content: %s", index, len(content), content[index:index+3])
		if content[index:index+3] == "'''" {
			if !isBoldOpen {
				// -3 +3 to replace the ''' as well
				content = strings.Replace(content, content[index:index+3], MARKER_BOLD_OPEN, 1)
				index = index + len(MARKER_BOLD_OPEN)

				// Check for further nested italic tags
				content, index, isBoldOpen, isItalicOpen = tokenizeBoldAndItalic(content, index, tokenMap, true, isItalicOpen)
			} else {
				// +3 to replace the '''
				content = strings.Replace(content, content[index:index+3], MARKER_BOLD_CLOSE, 1)

				// -3 because of the ''' we replaced above
				return content, index + len(MARKER_BOLD_CLOSE), false, isItalicOpen
			}
		} else if content[index:index+2] == "''" {
			if !isItalicOpen {
				// +2 to replace the ''
				content = strings.Replace(content, content[index:index+2], MARKER_ITALIC_OPEN, 1)
				index = index + len(MARKER_ITALIC_OPEN)

				// Check for further nested italic tags
				content, index, isBoldOpen, isItalicOpen = tokenizeBoldAndItalic(content, index, tokenMap, isBoldOpen, true)
			} else {
				// +2 to replace the ''
				content = strings.Replace(content, content[index:index+2], MARKER_ITALIC_CLOSE, 1)

				// -2 because of the '' we replaced above
				return content, index + len(MARKER_ITALIC_CLOSE), isBoldOpen, false
			}
		} else {
			index++
		}

		if !isBoldOpen && !isItalicOpen {
			break
		}
	}

	return content, index, false, false
}

func parseInternalLinks(content string, tokenMap map[string]string) (string, bool) {
	tokenizationHappened := false
	regex := regexp.MustCompile(`\[\[(.*?)]]`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		token := getToken(TOKEN_INTERNAL_LINK)
		tokenMap[token] = submatch[1]
		content = strings.Replace(content, submatch[0], token, 1)
		tokenizationHappened = true
	}

	return content, tokenizationHappened
}

func parseExternalLinks(content string, tokenMap map[string]string) (string, bool) {
	tokenizationHappened := false
	regex := regexp.MustCompile(`([^\[])\[([^ ]*) ?(.*)\]`)
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		tokenUrl := getToken(TOKEN_EXTERNAL_LINK_URL)
		tokenMap[tokenUrl] = submatch[2]

		if submatch[3] == "" {
			// Use URL as text
			submatch[3] = submatch[2]
		}

		tokenText := getToken(TOKEN_EXTERNAL_LINK_TEXT)
		tokenMap[tokenText] = submatch[3]

		token := getToken(TOKEN_EXTERNAL_LINK)
		tokenMap[token] = tokenUrl + " " + tokenText

		content = strings.Replace(content, submatch[0], submatch[1]+token, 1)
		tokenizationHappened = true
	}

	return content, tokenizationHappened
}
