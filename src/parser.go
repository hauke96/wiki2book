package main

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/wiki"
	"regexp"
	"strings"
)

func parse(wikiPageDto *WikiPageDto) wiki.Article {
	content, images := processImages(wikiPageDto.Parse.Wikitext.Content)
	return wiki.Article{
		Title:   wikiPageDto.Parse.Title,
		Images:  images,
		Content: content,
	}
}

// TODO remove unwanted stuff: templates, references, etc.

// TODO expand templates and parse the HTML: \{\{[a-zA-Z0-9äöüÄÖÜ\\|\s,.-_\(\)=\[\]\{\}]*\}\} -> https://www.mediawiki.org/wiki/API:Expandtemplates#GET_request

// processImages returns the list of all images and also escapes the image names in the content
func processImages(content string) (string, []wiki.Image) {
	var result []wiki.Image

	regex := regexp.MustCompile("\\[\\[((Datei|File):.*?)(]]|\\|)")
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		splittedMatch := strings.Split(submatch[1], "|")

		filename := strings.ReplaceAll(splittedMatch[0], " ", "_")
		caption := splittedMatch[len(splittedMatch)-1]

		content = strings.ReplaceAll(content, splittedMatch[0], filename)

		result = append(result, wiki.Image{
			Filename: filename,
			Caption:  caption,
		})

		sigolo.Info("Found image: %s", filename)
	}

	sigolo.Info("Found %d images", len(submatches))
	return content, result
}
