package generator

import (
	"strings"
	"wiki2book/parser"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

func expandToken(expansionHandler ExpansionHandler, token parser.Token) (string, error) {
	var err error = nil
	var html = ""

	switch t := token.(type) {
	case parser.HeadingToken:
		html, err = expansionHandler.expandHeadings(t)
	case parser.InlineImageToken:
		html, err = expansionHandler.expandInlineImage(t)
	case parser.ImageToken:
		html, err = expansionHandler.expandImage(t)
	case parser.ExternalLinkToken:
		html, err = expansionHandler.expandExternalLink(t)
	case parser.InternalLinkToken:
		html, err = expansionHandler.expandInternalLink(t)
	case parser.UnorderedListToken:
		html, err = expansionHandler.expandUnorderedList(t)
	case parser.OrderedListToken:
		html, err = expansionHandler.expandOrderedList(t)
	case parser.DescriptionListToken:
		html, err = expansionHandler.expandDescriptionList(t)
	case parser.ListItemToken:
		html, err = expansionHandler.expandListItem(t)
	case parser.TableToken:
		html, err = expansionHandler.expandTable(t)
	case parser.TableRowToken:
		html, err = expansionHandler.expandTableRow(t)
	case parser.TableColToken:
		html, err = expansionHandler.expandTableColumn(t)
	case parser.TableCaptionToken:
		html, err = expansionHandler.expandTableCaption(t)
	case parser.MathToken:
		html, err = expansionHandler.expandMath(t)
	case parser.RefDefinitionToken:
		html, err = expansionHandler.expandRefDefinition(t)
	case parser.RefUsageToken:
		html = expansionHandler.expandRefUsage(t)
	case parser.NowikiToken:
		html = expansionHandler.expandNowiki(t)
	}

	if err != nil {
		return "", err
	}

	return html, nil
}

func expand(expansionHandler ExpansionHandler, content interface{}) (string, error) {
	switch content.(type) {
	case string:
		return expandString(expansionHandler, content.(string))
	case parser.Token:
		return expandToken(expansionHandler, content.(parser.Token))
	}

	return "", errors.Errorf("Unsupported type to expand: %T", content)
}

func expandString(g ExpansionHandler, content string) (string, error) {
	content = g.expandMarker(content)

	matches := tokenRegex.FindAllString(content, -1)

	if len(matches) == 0 {
		// no token in content
		return content, nil
	}

	for _, tokenKey := range matches {
		tokenContent, hasTokenKey := g.getToken(tokenKey)
		if !hasTokenKey {
			return "", errors.Errorf("Token key %s not found in token map", tokenKey)
		}
		sigolo.Tracef("Found token %s -> %#v", tokenKey, tokenContent)

		html, err := expand(g, tokenContent)
		if err != nil {
			return "", err
		}

		content = strings.Replace(content, tokenKey, html, 1)
	}

	return content, nil
}

type ExpansionHandler interface {
	getToken(string) (parser.Token, bool)
	expandMarker(content string) string
	expandHeadings(token parser.HeadingToken) (string, error)
	expandInlineImage(token parser.InlineImageToken) (string, error)
	expandImage(token parser.ImageToken) (string, error)
	expandInternalLink(token parser.InternalLinkToken) (string, error)
	expandExternalLink(token parser.ExternalLinkToken) (string, error)
	expandTable(token parser.TableToken) (string, error)
	expandTableRow(token parser.TableRowToken) (string, error)
	expandTableColumn(token parser.TableColToken) (string, error)
	expandTableCaption(token parser.TableCaptionToken) (string, error)
	expandUnorderedList(token parser.UnorderedListToken) (string, error)
	expandOrderedList(token parser.OrderedListToken) (string, error)
	expandDescriptionList(token parser.DescriptionListToken) (string, error)
	expandListItems(items []parser.ListItemToken) (string, error)
	expandListItem(token parser.ListItemToken) (string, error)
	expandRefDefinition(token parser.RefDefinitionToken) (string, error)
	expandRefUsage(token parser.RefUsageToken) string
	expandMath(token parser.MathToken) (string, error)
	expandNowiki(token parser.NowikiToken) string
}
