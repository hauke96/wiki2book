package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/parser"
	"wiki2book/util"

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
	NumberOfExternalLinks  int            `json:"numberOfExternalLinks"`
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

func (g *StatsGenerator) expandSimpleString(content string) string {
	g.stats.NumberOfCharacters += len([]rune(content))
	return content
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
	return expand(g, token.Content)
}

func (g *StatsGenerator) expandInlineImage(token parser.InlineImageToken) (string, error) {
	return "", nil
}

func (g *StatsGenerator) expandImage(token parser.ImageToken) (string, error) {
	return expand(g, token.Caption.Content)
}

func (g *StatsGenerator) expandInternalLink(token parser.InternalLinkToken) (string, error) {
	expandedContent, err := expand(g, token.LinkText)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfInternalLinks++
	g.stats.InternalLinks[token.ArticleName]++

	return expandedContent, nil
}

func (g *StatsGenerator) expandExternalLink(token parser.ExternalLinkToken) (string, error) {
	expandedContent, err := expand(g, token.LinkText)
	if err != nil {
		return "", err
	}

	g.stats.NumberOfExternalLinks++

	return expandedContent, nil
}

func (g *StatsGenerator) expandTable(token parser.TableToken) (string, error) {
	result := ""
	for _, rowToken := range token.Rows {
		expandedRow, err := expand(g, rowToken)
		if err != nil {
			return "", err
		}
		result += expandedRow
	}

	expandedCaption, err := expand(g, token.Caption)
	if err != nil {
		return "", err
	}
	result += expandedCaption

	return result, nil
}

func (g *StatsGenerator) expandTableRow(token parser.TableRowToken) (string, error) {
	result := ""
	for _, colToken := range token.Columns {
		expandedItem, err := expand(g, colToken)
		if err != nil {
			return "", err
		}
		result += expandedItem
	}
	return result, nil
}

func (g *StatsGenerator) expandTableColumn(token parser.TableColToken) (string, error) {
	return expand(g, token.Content)
}

func (g *StatsGenerator) expandTableCaption(token parser.TableCaptionToken) (string, error) {
	return expand(g, token.Content)
}

func (g *StatsGenerator) expandUnorderedList(token parser.UnorderedListToken) (string, error) {
	return g.expandListItems(token.Items)
}

func (g *StatsGenerator) expandOrderedList(token parser.OrderedListToken) (string, error) {
	return g.expandListItems(token.Items)
}

func (g *StatsGenerator) expandDescriptionList(token parser.DescriptionListToken) (string, error) {
	return g.expandListItems(token.Items)
}

func (g *StatsGenerator) expandListItems(items []parser.ListItemToken) (string, error) {
	result := ""
	for _, item := range items {
		expandedItem, err := expand(g, item)
		if err != nil {
			return "", err
		}
		result += expandedItem
	}
	return result, nil
}

func (g *StatsGenerator) expandListItem(token parser.ListItemToken) (string, error) {
	return expand(g, token.Content)
}

func (g *StatsGenerator) expandRefDefinition(token parser.RefDefinitionToken) (string, error) {
	return expand(g, token.Content)
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

	sigolo.Debugf("Generate stats to '%s' for articles %v", outputFilePath, statFiles)

	combinedStats := &articleStats{
		InternalLinks:          map[string]int{},
		UncoveredInternalLinks: map[string]int{},
	}
	articles := map[string]*articleStats{}

	for _, statFile := range statFiles {
		var fileContent []byte

		sigolo.Debugf("Read and process stats file '%s'", statFile)

		fileContent, err = util.CurrentFilesystem.ReadFile(statFile)
		if err != nil {
			return errors.Wrapf(err, "Error reading stats file '%s'", statFile)
		}

		stats := &articleStats{}
		err = json.Unmarshal(fileContent, stats)
		if err != nil {
			return errors.Wrapf(err, "Error parsing JSON from stats file '%s'", statFile)
		}

		articles[stats.ArticleName] = stats
		combinedStats.NumberOfInternalLinks += stats.NumberOfInternalLinks
		combinedStats.NumberOfExternalLinks += stats.NumberOfExternalLinks
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

	var outputConent []byte

	if config.Current.OutputType == config.OutputTypeStatsJson {
		outputConent, err = generateJsonStatsContent(combinedStats)
	} else if config.Current.OutputType == config.OutputTypeStatsTxt {
		outputConent, err = generateTxtStatsContent(articles, combinedStats)
	} else {
		return errors.Errorf("Invalid output type '%s' for stats", config.Current.OutputType)
	}

	// Just to make it easier to simply open the read the file in CLI because usually files have a newline at the end.
	outputConent = append(outputConent, byte('\n'))

	var outputFile util.FileLike
	outputFile, err = util.CurrentFilesystem.Create(outputFilePath)
	if err != nil {
		return errors.Wrapf(err, "Error opening output stats file '%s'", outputFilePath)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, bytes.NewReader(outputConent))
	if err != nil {
		return errors.Wrapf(err, "Error writing combined stats to output file '%s'", outputFilePath)
	}

	return nil
}

func generateJsonStatsContent(combinedStats *articleStats) ([]byte, error) {
	outputConent, err := json.MarshalIndent(combinedStats, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "Error creating JSON for the combined stats")
	}
	return outputConent, nil
}

func generateTxtStatsContent(articles map[string]*articleStats, combinedStats *articleStats) ([]byte, error) {
	result := "Statistics generated with wiki2book:\n\n"

	result += "General:\n"
	result += fmt.Sprintf("  Number of characters    : %d\n", combinedStats.NumberOfCharacters)
	result += fmt.Sprintf("  Number of internal links: %d\n", combinedStats.NumberOfInternalLinks)
	result += "\n"

	// Article names
	result += "Article:\n"
	for _, article := range articles {
		result += "  " + article.ArticleName + "\n"
	}
	result += "\n"

	// Most uncovered links
	result += "Top 20 linked articles that are not part of the book (total: " + strconv.Itoa(len(combinedStats.UncoveredInternalLinks)) + "):\n"

	type uncoveredLinkEntry struct {
		articleName string
		occurrences int
	}
	var sortedUncoveredLinks []uncoveredLinkEntry
	for k, v := range combinedStats.UncoveredInternalLinks {
		sortedUncoveredLinks = append(sortedUncoveredLinks, uncoveredLinkEntry{k, v})
	}
	sort.Slice(sortedUncoveredLinks, func(i, j int) bool {
		return sortedUncoveredLinks[i].occurrences > sortedUncoveredLinks[j].occurrences
	})

	for i := 0; i < int(math.Min(float64(len(sortedUncoveredLinks)), 20)); i++ {
		result += "  " + sortedUncoveredLinks[i].articleName + " (" + strconv.Itoa(sortedUncoveredLinks[i].occurrences) + ")\n"
	}
	//result += "\n"

	return []byte(result), nil
}
