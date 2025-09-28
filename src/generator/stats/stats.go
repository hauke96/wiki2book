package stats

import (
	"encoding/json"
	"strings"
	"wiki2book/cache"
	"wiki2book/parser"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

type StatsGenerator struct {
}

type articleStats struct {
	NumberOfCharacters    int `json:"numberOfCharacters"`
	NumberOfInternalLinks int `json:"numberOfInternalLinks"`
}

func (g *StatsGenerator) GenerateForArticle(wikiArticle *parser.Article) (string, error) {
	filename := wikiArticle.Title + ".json"

	// TODO fill articleStats for article
	stats := &articleStats{
		NumberOfCharacters:    g.determineNumberOfCharacters(wikiArticle),
		NumberOfInternalLinks: g.determineNumberOfLinks(wikiArticle),
	}

	statsBytes, err := json.Marshal(stats)
	sigolo.FatalCheck(errors.Wrapf(err, "Error creating stats for article '%s'", wikiArticle.Title))

	stringReader := strings.NewReader(string(statsBytes))
	return cache.CacheToFile(cache.StatsCacheDirName, filename, stringReader)
}

func (g *StatsGenerator) determineNumberOfCharacters(article *parser.Article) int {
	counter := 0

	for _, token := range article.TokenMap {
		switch concreteToken := token.(type) {
		case parser.InternalLinkToken:
			counter += len([]rune(concreteToken.LinkText))
		}
	}

	return counter
}

func (g *StatsGenerator) determineNumberOfLinks(article *parser.Article) int {
	counter := 0

	for _, token := range article.TokenMap {
		switch token.(type) {
		case parser.InternalLinkToken:
			counter++
		}
	}

	return counter
}
