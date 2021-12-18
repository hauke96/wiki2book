package main

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/generator/epub"
	"github.com/hauke96/wiki2book/src/generator/html"
	"github.com/pkg/errors"
	"os"
)

func main() {
	title := "Stern"

	err := createAndUseFolder(title)
	sigolo.FatalCheck(err)

	wikiPageDto, err := downloadPage("de", title)
	sigolo.FatalCheck(err)

	wikiPage := parse(wikiPageDto)

	err = downloadImages(wikiPage.Images, "./images")
	sigolo.FatalCheck(err)

	outputFile, err := html.Generate(wikiPage, "./")
	sigolo.FatalCheck(err)

	err = epub.Generate(outputFile, wikiPage.Title+".epub", "../../style.css", "../../wikipedia-astronomie-cover.png", "Astronomie")
	sigolo.FatalCheck(err)
}

// createAndUseFolder creates a folder with the given name and goes into that folder.
func createAndUseFolder(title string) error {
	err := os.Mkdir(title, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Error creating output directory %s", title))
	}

	err = os.Chdir(title)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error switching into output directory %s", title))
	}

	return nil
}
