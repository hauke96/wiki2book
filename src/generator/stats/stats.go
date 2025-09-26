package stats

import (
	"strings"
	"wiki2book/cache"
	"wiki2book/parser"
)

type StatsGenerator struct {
	TokenMap map[string]parser.Token
}

func (g *StatsGenerator) GenerateForArticle(wikiArticle *parser.Article) (string, error) {
	filename := wikiArticle.Title + ".txt"
	stringReader := strings.NewReader("TODO")
	return cache.CacheToFile(cache.StatsCacheDirName, filename, stringReader)
}
