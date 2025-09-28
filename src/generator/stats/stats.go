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
	numberOfCharacters int
	numberOfLinks      int
}

func (g *StatsGenerator) GenerateForArticle(wikiArticle *parser.Article) (string, error) {
	filename := wikiArticle.Title + ".json"

	// TODO fill articleStats for article
	stats := &articleStats{
		numberOfCharacters: 123,
		numberOfLinks:      234,
	}

	statsBytes, err := json.Marshal(stats)
	sigolo.FatalCheck(errors.Wrapf(err, "Error creating stats for article '%s'", wikiArticle.Title))

	stringReader := strings.NewReader(string(statsBytes))
	return cache.CacheToFile(cache.StatsCacheDirName, filename, stringReader)
}
