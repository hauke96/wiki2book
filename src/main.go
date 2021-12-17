package main

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/generator/html"
)

func main() {
	wikiPageDto, err := downloadPage("de", "Stern")
	sigolo.FatalCheck(err)

	wikiPage := parse(wikiPageDto)

	err = downloadImages(wikiPage.Images, "./"+wikiPage.Title+"/images")
	sigolo.FatalCheck(err)

	err = html.Generate(wikiPage, "./"+wikiPage.Title)
	sigolo.FatalCheck(err)
}
