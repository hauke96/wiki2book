package main

import "github.com/hauke96/sigolo"

func main() {
	wikiPageDto, err := downloadPage("de", "Stern")
	sigolo.FatalCheck(err)

	wikiPage := parse(wikiPageDto)

	sigolo.Info("content: %s", wikiPage.Content)
	sigolo.Info("title: %s", wikiPage.Title)
	sigolo.Info("images: %s", wikiPage.Images)
}
