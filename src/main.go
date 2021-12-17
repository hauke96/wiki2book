package main

import "github.com/hauke96/sigolo"

func main() {
	wikiPageDto, err := downloadPage("de", "Stern")
	sigolo.FatalCheck(err)

	wikiPage := parse(wikiPageDto)

	err = downloadImages(wikiPage.Images, "./images")
	sigolo.FatalCheck(err)
}
