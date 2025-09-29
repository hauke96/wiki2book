package generator

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
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
	ArticleName            string         `json:"articleName"`
	NumberOfCharacters     int            `json:"numberOfCharacters"`
	NumberOfInternalLinks  int            `json:"numberOfInternalLinks"`
	InternalLinks          map[string]int `json:"internalLinks"`
	UncoveredInternalLinks map[string]int `json:"uncoveredInternalLinks"`
}

func NewStatsGenerator(tokenMap map[string]parser.Token) *StatsGenerator {
	return &StatsGenerator{
		tokenMap: tokenMap,
		stats:    &articleStats{},
	}
}

func (g *StatsGenerator) Generate(wikiArticle *parser.Article) (string, error) {
	filename := wikiArticle.Title + ".json"

	g.stats.ArticleName = wikiArticle.Title
	g.stats.InternalLinks = map[string]int{}
	g.stats.UncoveredInternalLinks = map[string]int{}

	_, err := expand(g, wikiArticle.Content)
	sigolo.FatalCheck(err)

	statsBytes, err := json.MarshalIndent(g.stats, "", "  ")
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
	g.stats.InternalLinks[token.ArticleName]++

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

func GenerateCombinedStats(statFiles []string, outputFilePath string) error {
	var err error

	combinedStats := &articleStats{
		InternalLinks:          map[string]int{},
		UncoveredInternalLinks: map[string]int{},
	}
	articles := map[string]interface{}{}

	for _, statFile := range statFiles {
		var fileContent []byte

		sigolo.Debugf("Read and process stats file '%s'", statFile)

		fileContent, err = os.ReadFile(statFile)
		if err != nil {
			return errors.Wrapf(err, "Error reading stats file '%s'", statFile)
		}

		stats := &articleStats{}
		err = json.Unmarshal(fileContent, stats)
		if err != nil {
			return errors.Wrapf(err, "Error parsing JSON from stats file '%s'", statFile)
		}

		articles[stats.ArticleName] = nil
		combinedStats.NumberOfInternalLinks += stats.NumberOfInternalLinks
		combinedStats.NumberOfCharacters += stats.NumberOfCharacters

		for articleName, count := range stats.InternalLinks {
			combinedStats.InternalLinks[articleName] += count
		}
	}

	for articleName, count := range combinedStats.InternalLinks {
		if _, ok := articles[articleName]; !ok {
			combinedStats.UncoveredInternalLinks[articleName] += count
		}
	}

	sigolo.Debugf("Write combined stats to output file '%s'", outputFilePath)

	var outputJson []byte
	var outputFile *os.File
	outputJson, err = json.MarshalIndent(combinedStats, "", "  ")
	if err != nil {
		return errors.Wrap(err, "Error creating JSON for the combined stats")
	}

	// Just to make it easier to simply open the read the file in CLI because usually files have a newline at the end.
	outputJson = append(outputJson, byte('\n'))

	outputFile, err = os.Create(outputFilePath)
	if err != nil {
		return errors.Wrapf(err, "Error opening output stats file '%s'", outputFilePath)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, bytes.NewReader(outputJson))
	if err != nil {
		return errors.Wrapf(err, "Error writing combined stats to output file '%s'", outputFilePath)
	}

	return nil
}
