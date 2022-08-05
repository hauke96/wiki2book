package generator

import "github.com/hauke96/wiki2book/src/parser"

type Generator interface {
	Generate(wikiArticle parser.Article, outputFolder string, styleFile string, imgFolder string, mathFolder string, articleFolder string) (string, error)
}
