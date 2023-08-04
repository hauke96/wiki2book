package generator

import "wiki2book/parser"

type Generator interface {
	Generate(wikiArticle parser.Article, outputFolder string, styleFile string, imgFolder string, mathFolder string, articleFolder string) (string, error)
}
