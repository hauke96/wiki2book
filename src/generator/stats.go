package generator

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
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
	ArticleName                 string         `json:"article-name"`
	NumberOfCharacters          int            `json:"number-of-characters"`
	NumberOfWords               int            `json:"number-of-words"`
	NumberOfInternalLinks       int            `json:"number-of-internal-links"`
	NumberOfExternalLinks       int            `json:"number-of-external-links"`
	NumberOfImages              int            `json:"number-of-images"`
	NumberOfMath                int            `json:"number-of-math"`
	NumberOfRefDefinitions      int            `json:"number-of-ref-definitions"`
	NumberOfRefUsages           int            `json:"number-of-ref-usages"`
	InternalLinks               map[string]int `json:"internal-links"`
	UncoveredInternalLinks      map[string]int `json:"uncovered-internal-links"`
	Top10UncoveredInternalLinks map[string]int `json:"top-10-uncovered-internal-links"`
	EstimatedReadingTimeMinutes int            `json:"estimated-reading-time-minutes"`
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

	readingTimeInMinutes := g.getEstimatedReadingTimeInMinutes()
	g.stats.EstimatedReadingTimeMinutes = readingTimeInMinutes

	statsBytes, err := json.MarshalIndent(g.stats, "", "  ")
	sigolo.FatalCheck(errors.Wrapf(err, "Error creating stats for article '%s'", wikiArticle.Title))

	stringReader := strings.NewReader(string(statsBytes))
	return cache.CacheToFile(cache.StatsCacheDirName, filename, stringReader)
}

func (g *StatsGenerator) getEstimatedReadingTimeInMinutes() int {
	// Average word/min based on 17 major languages according to a 2012 study by Trauzettel-Klosinski et al.
	averageWordsPerMinute := 183
	switch config.Current.WikipediaInstance {
	case "en":
		averageWordsPerMinute = 228
	case "de":
		averageWordsPerMinute = 179
	case "fr":
		averageWordsPerMinute = 195
	case "nl":
		averageWordsPerMinute = 202
	case "fi":
		averageWordsPerMinute = 161
	case "he":
		averageWordsPerMinute = 187
	case "it":
		averageWordsPerMinute = 188
	case "jp":
		averageWordsPerMinute = 193
	case "pl":
		averageWordsPerMinute = 166
	case "pt":
		averageWordsPerMinute = 181
	case "ru":
		averageWordsPerMinute = 184
	case "sl":
		averageWordsPerMinute = 180
	case "es":
		averageWordsPerMinute = 218
	case "se":
		averageWordsPerMinute = 199
	case "tr":
		averageWordsPerMinute = 166
	}
	return g.stats.NumberOfWords / averageWordsPerMinute
}

func (g *StatsGenerator) getToken(tokenKey string) (parser.Token, bool) {
	token, hasToken := g.tokenMap[tokenKey]
	return token, hasToken
}

func (g *StatsGenerator) expandSimpleString(content string) string {
	g.stats.NumberOfCharacters += len([]rune(content))

	inWord := false
	for _, r := range content {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if !inWord {
				// New word starts
				g.stats.NumberOfWords++
			}
			inWord = true
		} else {
			inWord = false
		}
	}

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
	g.stats.NumberOfImages++
	return "", nil
}

func (g *StatsGenerator) expandImage(token parser.ImageToken) (string, error) {
	g.stats.NumberOfImages++
	return expand(g, token.Caption.Content)
}

func (g *StatsGenerator) expandInternalLink(token parser.InternalLinkToken) (string, error) {
	g.stats.NumberOfInternalLinks++
	g.stats.InternalLinks[token.ArticleName]++
	return expand(g, token.LinkText)
}

func (g *StatsGenerator) expandExternalLink(token parser.ExternalLinkToken) (string, error) {
	g.stats.NumberOfExternalLinks++
	return expand(g, token.LinkText)
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
	g.stats.NumberOfRefDefinitions++
	return expand(g, token.Content)
}

func (g *StatsGenerator) expandRefUsage(token parser.RefUsageToken) string {
	g.stats.NumberOfRefUsages++
	return ""
}

func (g *StatsGenerator) expandMath(token parser.MathToken) (string, error) {
	g.stats.NumberOfMath++
	return "", nil
}

func (g *StatsGenerator) expandNowiki(token parser.NowikiToken) string {
	return ""
}

func GenerateCombinedStats(statFiles []string, outputFilePath string) error {
	var err error

	sigolo.Debugf("Generate stats to '%s' for articles %v", outputFilePath, statFiles)

	combinedStats := &articleStats{
		InternalLinks:               map[string]int{},
		UncoveredInternalLinks:      map[string]int{},
		Top10UncoveredInternalLinks: map[string]int{},
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
		combinedStats.NumberOfCharacters += stats.NumberOfCharacters
		combinedStats.NumberOfWords += stats.NumberOfWords
		combinedStats.NumberOfInternalLinks += stats.NumberOfInternalLinks
		combinedStats.NumberOfExternalLinks += stats.NumberOfExternalLinks
		combinedStats.NumberOfImages += stats.NumberOfImages
		combinedStats.NumberOfMath += stats.NumberOfMath
		combinedStats.NumberOfRefDefinitions += stats.NumberOfRefDefinitions
		combinedStats.NumberOfRefUsages += stats.NumberOfRefUsages
		combinedStats.EstimatedReadingTimeMinutes += stats.EstimatedReadingTimeMinutes

		for articleName, count := range stats.InternalLinks {
			combinedStats.InternalLinks[articleName] += count
		}
	}

	var uncoveredArticles []string
	for articleName, count := range combinedStats.InternalLinks {
		if _, ok := articles[articleName]; !ok {
			combinedStats.UncoveredInternalLinks[articleName] += count
			uncoveredArticles = append(uncoveredArticles, articleName)
		}
	}

	slices.SortFunc(uncoveredArticles, func(a, b string) int {
		return cmp.Compare(combinedStats.UncoveredInternalLinks[b], combinedStats.UncoveredInternalLinks[a])
	})
	for i, uncoveredArticle := range uncoveredArticles {
		if i >= 10 {
			break
		}
		combinedStats.Top10UncoveredInternalLinks[uncoveredArticle] = combinedStats.UncoveredInternalLinks[uncoveredArticle]
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
