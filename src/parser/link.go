package parser

import (
	"strings"
)

func (t *Tokenizer) parseInternalLinks(content string) string {
	return t.parseLink(content, "[[", "]]", "|", TOKEN_INTERNAL_LINK_ARTICLE, TOKEN_INTERNAL_LINK_TEXT, TOKEN_INTERNAL_LINK, false, true)
}

func (t *Tokenizer) parseExternalLinks(content string) string {
	return t.parseLink(content, "[", "]", " ", TOKEN_EXTERNAL_LINK_URL, TOKEN_EXTERNAL_LINK_TEXT, TOKEN_EXTERNAL_LINK, true, false)
}

// parseLink takes the given bracket type and tries to find the link content in between them and replaces it with a
// token. The parameter delimiterRequired specified if the link must definitely have two parts (URL/Article and a
// display text). The parameter removeSectionReference specifies whether or not everything behind the first "#" should
// be ignored or not.
func (t *Tokenizer) parseLink(content string, openingBrackets string, closingBrackets string, linkDelimiter string, targetTokenString string, linkTextTokenString string, linkTokenString string, delimiterRequired bool, removeSectionReference bool) string {
	splitContent := strings.Split(content, openingBrackets)
	var resultSegments []string

	// The following steps are performed for INTERNAL links (same steps for EXTERNAL links but of course with different
	// brackets and delimiters):
	//   1. Split by  [[  since it's the start of an internal link.
	//   2. Split each element (except the first one, see below) at  ]]  since it's the end of an internal link. The
	//      first slice element is now the link content between the brackets and the rest is just text after the link.
	//   3. Split the link content by  |  since it's the delimiter for target page and link text.
	//   4. Create the token for the link target, link content and overall internal link.
	//   5. Continue with step 2 until all elements have been processed.

	for i, splitItem := range splitContent {
		if i == 0 {
			// The first string is never the start of a link. It's either an empty string (in case the content directly
			// starts with a link) or it's the text before the first link.
			resultSegments = append(resultSegments, splitItem)
			continue
		}

		// Ignore all kind of files, they are parsed elsewhere
		if imagePrefixRegex.MatchString(splitItem) {
			resultSegments = append(resultSegments, openingBrackets)
			resultSegments = append(resultSegments, splitItem)
			continue
		}

		segments := strings.Split(splitItem, closingBrackets)

		possibleLinkWikitext := segments[0]

		wikitextElements := strings.Split(possibleLinkWikitext, linkDelimiter)
		linkTarget := wikitextElements[0]
		if removeSectionReference {
			linkTarget = strings.SplitN(linkTarget, "#", 2)[0]
		}

		var linkText string
		if len(wikitextElements) == 1 {
			if delimiterRequired {
				// We need at least one delimiter in this link but found none -> Abort parsing this link.
				resultSegments = append(resultSegments, openingBrackets)
				resultSegments = append(resultSegments, splitItem)
				continue
			}
			linkText = linkTarget
		} else {
			linkText = strings.Join(wikitextElements[1:], linkDelimiter)
		}

		tokenTarget := t.getToken(targetTokenString)
		t.setRawToken(tokenTarget, linkTarget)

		tokenLinkText := t.getToken(linkTextTokenString)
		t.setRawToken(tokenLinkText, linkText)

		token := t.getToken(linkTokenString)
		t.setRawToken(token, tokenTarget+" "+tokenLinkText)

		resultSegments = append(resultSegments, token)

		if len(segments) > 1 {
			// Add all uninteresting segments behind the link
			resultSegments = append(resultSegments, strings.Join(segments[1:], closingBrackets))
		}
	}

	return strings.Join(resultSegments, "")
}
