package parser

import (
	"fmt"
	"regexp"
	"strings"
)

func clean(content string) string {
	content = removeUnwantedTags(content)
	//content = moveCitationsToEnd(content)
	return content
}

func removeUnwantedTags(content string) string {
	regex := regexp.MustCompile("\\[\\[Kategorie:.*?]]\n?")
	content = regex.ReplaceAllString(content, "")

	ignoreTemplates := []string{
		"siehe auch",
		"Exzellent",
		"Normdaten",
		"Hauptartikel",
		"Begriffsklärungshinweis",
		"Weiterleitungshinweis",
		"Dieser Artikel",
		"Commons",
		"Wikiquote",
		"Wiktionary",
		"Wikibooks",
		"Wikisource",
		"Alpha Centauri",
		"Panorama",
		".*(box|Box).*",
	}

	for _, template := range ignoreTemplates {
		regex = regexp.MustCompile(`(?i)(\* )?\{\{` + template + `(.|\n|\r)*?}}\n?`)
		content = regex.ReplaceAllString(content, "")
	}

	return content
}

func moveCitationsToEnd(content string) string {
	counter := 0
	citations := ""

	regex := regexp.MustCompile(`</?references.*?/?>\n?`)
	contentParts := regex.Split(content, -1)

	regex = regexp.MustCompile(`<ref.*?((>((.|\n|\r)*?)</ref>)|/>)`)
	contentParts[0] = regex.ReplaceAllStringFunc(contentParts[0], func(match string) string {
		counter++
		citations += fmt.Sprintf("[%d] %s\n", counter, strings.ReplaceAll(match, "\n", ""))
		return fmt.Sprintf("<sup>%d</sup>", counter)
	})

	if len(contentParts) > 1 {
		citations += "\n"
		contentParts[1] = regex.ReplaceAllStringFunc(contentParts[1], func(match string) string {
			citations += fmt.Sprintf("* %s\n", strings.ReplaceAll(match, "\n", ""))
			return match
		})
	}

	result := contentParts[0] + citations
	if len(contentParts) > 2 {
		result += contentParts[2]
	}

	return result
}
