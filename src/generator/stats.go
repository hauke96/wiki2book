package generator

import (
	"encoding/json"
	"strings"
	"wiki2book/cache"
	"wiki2book/parser"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

type StatsGenerator struct {
	tokenMap map[string]parser.Token
	stats    *articleStats
}

type articleStats struct {
	NumberOfCharacters    int `json:"numberOfCharacters"`
	NumberOfInternalLinks int `json:"numberOfInternalLinks"`
}

func NewStatsGenerator(tokenMap map[string]parser.Token) *StatsGenerator {
	return &StatsGenerator{
		tokenMap: tokenMap,
		stats:    &articleStats{},
	}
}

func (g *StatsGenerator) Generate(wikiArticle *parser.Article) (string, error) {
	filename := wikiArticle.Title + ".json"

	_, err := expand(g, wikiArticle.Content)
	sigolo.FatalCheck(err)

	statsBytes, err := json.Marshal(g.stats)
	sigolo.FatalCheck(errors.Wrapf(err, "Error creating stats for article '%s'", wikiArticle.Title))

	stringReader := strings.NewReader(string(statsBytes))
	return cache.CacheToFile(cache.StatsCacheDirName, filename, stringReader)
}

func (g *StatsGenerator) getToken(tokenKey string) (parser.Token, bool) {
	token, hasToken := g.tokenMap[tokenKey]
	return token, hasToken
}

func (g *StatsGenerator) expandMarker(content string) string {
	content = strings.ReplaceAll(content, parser.MARKER_BOLD_OPEN, "")
	content = strings.ReplaceAll(content, parser.MARKER_BOLD_CLOSE, "")
	content = strings.ReplaceAll(content, parser.MARKER_ITALIC_OPEN, "")
	content = strings.ReplaceAll(content, parser.MARKER_ITALIC_CLOSE, "")
	content = strings.ReplaceAll(content, parser.MARKER_PARAGRAPH, "")
	return content
}

func (g *StatsGenerator) expandHeadings(token parser.HeadingToken) (string, error) {
	expandedContent, err := expand(g, token.Content)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfCharacters += len([]rune(expandedContent))
	return "", nil
}

func (g *StatsGenerator) expandInlineImage(token parser.InlineImageToken) (string, error) {
	return "", nil
}

func (g *StatsGenerator) expandImage(token parser.ImageToken) (string, error) {
	return "", nil
}

func (g *StatsGenerator) expandInternalLink(token parser.InternalLinkToken) (string, error) {
	expandedContent, err := expand(g, token.LinkText)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfCharacters += len([]rune(expandedContent))
	g.stats.NumberOfInternalLinks++

	return "", nil
}

func (g *StatsGenerator) expandExternalLink(token parser.ExternalLinkToken) (string, error) {
	expandedContent, err := expand(g, token.LinkText)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfCharacters += len([]rune(expandedContent))
	return "", nil
}

func (g *StatsGenerator) expandTable(token parser.TableToken) (string, error) {
	for _, rowToken := range token.Rows {
		_, err := expand(g, rowToken)
		if err != nil {
			return "", err
		}
	}

	_, err := expand(g, token.Caption)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (g *StatsGenerator) expandTableRow(token parser.TableRowToken) (string, error) {
	for _, colToken := range token.Columns {
		_, err := expand(g, colToken)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func (g *StatsGenerator) expandTableColumn(token parser.TableColToken) (string, error) {
	expandedContent, err := expand(g, token.Content)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfCharacters += len([]rune(expandedContent))
	return "", nil
}

func (g *StatsGenerator) expandTableCaption(token parser.TableCaptionToken) (string, error) {
	expandedContent, err := expand(g, token.Content)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfCharacters += len([]rune(expandedContent))
	return "", nil
}

func (g *StatsGenerator) expandUnorderedList(token parser.UnorderedListToken) (string, error) {
	_, err := g.expandListItems(token.Items)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (g *StatsGenerator) expandOrderedList(token parser.OrderedListToken) (string, error) {
	_, err := g.expandListItems(token.Items)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (g *StatsGenerator) expandDescriptionList(token parser.DescriptionListToken) (string, error) {
	_, err := g.expandListItems(token.Items)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (g *StatsGenerator) expandListItems(items []parser.ListItemToken) (string, error) {
	for _, item := range items {
		_, err := g.expandListItem(item)
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

func (g *StatsGenerator) expandListItem(token parser.ListItemToken) (string, error) {
	expandedContent, err := expand(g, token.Content)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfCharacters += len([]rune(expandedContent))
	return "", nil
}

func (g *StatsGenerator) expandRefDefinition(token parser.RefDefinitionToken) (string, error) {
	expandedContent, err := expand(g, token.Content)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfCharacters += len([]rune(expandedContent))
	return "", nil
}

func (g *StatsGenerator) expandRefUsage(token parser.RefUsageToken) string {
	return ""
}

func (g *StatsGenerator) expandMath(token parser.MathToken) (string, error) {
	return "", nil
}

func (g *StatsGenerator) expandNowiki(token parser.NowikiToken) string {
	return ""
}
