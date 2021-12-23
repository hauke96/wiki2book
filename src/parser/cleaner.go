package parser

import (
	"regexp"
)

func clean(content string) string {
	content = removeUnwantedTags(content)
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
	}

	for _, template := range ignoreTemplates {
		regex = regexp.MustCompile(`(?i)(\* )?\{\{` + template + `(.|\n|\r)*?}}\n?`)
		content = regex.ReplaceAllString(content, "")
	}

	return content
}
