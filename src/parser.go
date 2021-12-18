package main

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/wiki"
	"regexp"
	"strings"
)

func parse(wikiPageDto *WikiPageDto) wiki.Article {
	content := moveCitationsToEnd(wikiPageDto.Parse.Wikitext.Content)
	content = removeUnwantedTags(content)
	content = evaluateTemplates(content)
	content, images := processImages(content)
	return wiki.Article{
		Title:   wikiPageDto.Parse.Title,
		Images:  images,
		Content: content,
	}
}

func removeUnwantedTags(content string) string {
	regex := regexp.MustCompile("<references.*?\\/>\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\[\\[Kategorie:.*?]]\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Gesprochener Artikel(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Exzellent(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Normdaten(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Hauptartikel(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile("\\{\\{Begriffsklärungshinweis(.|\\n|\\r)*?}}\n?")
	content = regex.ReplaceAllString(content, "")

	return content
}

func moveCitationsToEnd(content string) string {
	counter := 0
	citations := ""

	regex := regexp.MustCompile("<ref.*?>(.*?)</ref>")
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		counter++
		if counter > 1 {
			citations += "<br>"
		}
		citations += fmt.Sprintf("\n[%d] %s", counter, match)
		return fmt.Sprintf("[%d]", counter)
	})

	return content + citations
}

func evaluateTemplates(content string) string {
	regex := regexp.MustCompile("\\{\\{(.*?)}}")
	content = regex.ReplaceAllStringFunc(content, func(match string) string {
		evaluatedTemplate, err := evaluateTemplate(match)
		if err != nil {
			sigolo.Stack(err)
			return ""
		}
		return evaluatedTemplate
	})
	return content
}

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

		sigolo.Debug("Found image: %s", filename)
	}

	sigolo.Info("Found and embedded %d images", len(submatches))
	return content, result
}