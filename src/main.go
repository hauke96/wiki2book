package main

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/generator/epub"
	"github.com/hauke96/wiki2book/src/generator/html"
	"os"
)

func main() {
	wikiPageDto, err := downloadPage("de", "Stern")
	sigolo.FatalCheck(err)

	wikiPage := parse(wikiPageDto)

	err = os.Mkdir(wikiPage.Title, os.ModePerm)
	if err != nil {
		sigolo.Fatal("Error creating output directory %s", wikiPage.Title)
	}
	err = os.Chdir(wikiPage.Title)
	if err != nil {
		sigolo.Fatal("Error switching into output directory %s", wikiPage.Title)
	}

	err = downloadImages(wikiPage.Images, "./images")
	sigolo.FatalCheck(err)

	outputFile, err := html.Generate(wikiPage, "./")
	sigolo.FatalCheck(err)

	err = epub.Generate(outputFile, wikiPage.Title+".epub", "../../style.css")
	if err != nil {
		sigolo.Stack(err)
	}
	sigolo.FatalCheck(err)
}
