package main

import (
	"github.com/hauke96/sigolo"
	"regexp"
	"strings"
)

type WikiPage struct {
	Title   string
	Content string
	Images  []WikiImage
}

type WikiImage struct {
	Filename string
	Caption string
}

func parse(wikiPageDto *WikiPageDto) WikiPage {
	images := getListOfImages(wikiPageDto.Parse.Wikitext.Content)
	return WikiPage{
		Images: images,
	}
}

// TODO remove unwanted stuff: templates, references, etc.

// TODO expand templates and parse the HTML: \{\{[a-zA-Z0-9äöüÄÖÜ\\|\s,.-_\(\)=\[\]\{\}]*\}\} -> https://www.mediawiki.org/wiki/API:Expandtemplates#GET_request

func getListOfImages(content string) []WikiImage {
	var result []WikiImage

	regex := regexp.MustCompile("\\[\\[((Datei|File):.*)\\]\\]")
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		splittedMatch := strings.Split(submatch[1], "|")

		filename := splittedMatch[0]
		caption := splittedMatch[len(splittedMatch)-1]

		result = append(result, WikiImage{
			Filename: filename,
			Caption: caption,
		})

		sigolo.Info("Found image: %s", filename)
	}

	sigolo.Info("Found %d images", len(submatches))
	return result
}
