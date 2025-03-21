package parser

import (
	"fmt"
	"strings"
	"wiki2book/config"
	"wiki2book/util"
)

type LinkType int

const (
	LINK_TYPE_INTERNAL LinkType = iota
	LINK_TYPE_EXTERNAL
)

type InternalLinkToken struct {
	Token
	ArticleName string
	LinkText    string
}

type ExternalLinkToken struct {
	Token
	URL      string
	LinkText string
}

func (t *Tokenizer) parseInternalLinks(content string) string {
	return t.parseLink(content, "[[", "]]", "|", LINK_TYPE_INTERNAL, false, true)
}

func (t *Tokenizer) parseExternalLinks(content string) string {
	return t.parseLink(content, "[", "]", " ", LINK_TYPE_EXTERNAL, true, false)
}

// parseLink takes the given bracket type and tries to find the link content in between them and replaces it with a
// token. The parameter delimiterRequired specifies if the link must definitely have two parts (URL/Article and a
// display text) or if the delimiter is optional. The parameter removeSectionReference specifies whether everything
// behind the first "#" should be ignored or not. Link to e.g. categories start with a ":" and will be treated as normal
// links, which means, in case of a missing delimiter, the text of the part behind the last ":" is used as link text.
func (t *Tokenizer) parseLink(content string, openingBrackets string, closingBrackets string, linkDelimiter string, linkType LinkType, delimiterRequired bool, removeSectionReference bool) string {
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
		if strings.Contains(splitItem, ":") {
			filePrefix := strings.ToLower(strings.SplitN(splitItem, ":", 2)[0])
			if util.Contains(config.Current.FilePrefixe, filePrefix) {
				resultSegments = append(resultSegments, openingBrackets)
				resultSegments = append(resultSegments, splitItem)
				continue
			}
		}

		segments := strings.Split(splitItem, closingBrackets)

		possibleLinkWikitext := segments[0]

		wikitextElements := strings.Split(possibleLinkWikitext, linkDelimiter)
		linkTarget := wikitextElements[0]
		if removeSectionReference {
			linkTarget = strings.SplitN(linkTarget, "#", 2)[0]
		}
		linkTarget = strings.TrimSpace(linkTarget)

		var linkText string
		if len(wikitextElements) == 1 {
			if delimiterRequired {
				// We need at least one delimiter in this link but found none -> Abort parsing this link.
				resultSegments = append(resultSegments, openingBrackets)
				resultSegments = append(resultSegments, splitItem)
				continue
			}

			if linkTarget[0] == ':' {
				// A link starting with ":" indicates e.g. a link to a category. Because we have no delimiter here, the
				// link text is the part behind the last ":".
				linkSegments := strings.SplitN(linkTarget, ":", -1)
				linkText = linkSegments[len(linkSegments)-1]
			} else {
				// A normal link but without the delimiter. Therefore, the link text is the whole link content, e.g. the
				// name of the article.
				linkText = linkTarget
			}
		} else {
			// We might have a normal link or a link to a category (or other category-like object) here. However, that
			// is irrelevant, because we have a delimiter and use the content behind the delimiter for the link text.
			linkText = strings.Join(wikitextElements[1:], linkDelimiter)
		}

		var token Token
		var linkTokenKey string
		if linkType == LINK_TYPE_INTERNAL {
			token = InternalLinkToken{
				ArticleName: linkTarget,
				LinkText:    t.tokenizeContent(t, linkText),
			}
			linkTokenKey = TOKEN_INTERNAL_LINK
		} else if linkType == LINK_TYPE_EXTERNAL {
			// See https://www.mediawiki.org/wiki/API:Extlinks property "elprotocol" for all possible values that are
			// supported by MediaWiki. However, these are the most likely ones:
			if util.HasAnyPrefix(linkTarget, "http://", "https://") {
				token = ExternalLinkToken{
					URL:      linkTarget,
					LinkText: t.tokenizeContent(t, linkText),
				}
				linkTokenKey = TOKEN_EXTERNAL_LINK
			} else if strings.Contains(linkTarget, "://") {
				token = StringToken{
					String: t.tokenizeContent(t, linkText),
				}
				linkTokenKey = TOKEN_STRING
			}
			// No supported or unsupported protocol -> leave the text as is
		}

		if token != nil {
			tokenKey := t.getToken(linkTokenKey)
			t.setRawToken(tokenKey, token)

			resultSegments = append(resultSegments, tokenKey)
		} else {
			// No actual link found, just some words in brackets:
			resultSegments = append(resultSegments, fmt.Sprintf("%s%s%s", openingBrackets, possibleLinkWikitext, closingBrackets))
		}

		if len(segments) > 1 {
			// Add all uninteresting segments behind the link
			contentBehindLink := strings.Join(segments[1:], closingBrackets)
			resultSegments = append(resultSegments, contentBehindLink)
		}
	}

	return strings.Join(resultSegments, "")
}
