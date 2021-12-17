package main

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/generator"
)

func main() {
	wikiPageDto, err := downloadPage("de", "Stern")
	sigolo.FatalCheck(err)

	wikiPage := parse(wikiPageDto)

	err = downloadImages(wikiPage.Images, "./images")
	sigolo.FatalCheck(err)

	err = generator.Generate(wikiPage, "./" + wikiPage.Title)
	sigolo.FatalCheck(err)
}
