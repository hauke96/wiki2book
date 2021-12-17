package main

import "github.com/hauke96/sigolo"

func main() {
	wikiPage, err := downloadPage("Stern")
	sigolo.FatalCheck(err)

	sigolo.Info("content: %s", wikiPage.Parse.Wikitext)
}
