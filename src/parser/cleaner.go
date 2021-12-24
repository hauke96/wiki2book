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
		"Alpha Centauri",
		"Begriffsklärungshinweis",
		"Commons",
		"Dieser Artikel",
		"Exzellent",
		"Hauptartikel",
		"Lesenswert",
		"Navigationsleiste",
		"Normdaten",
		"Panorama",
		"siehe auch",
		"Weiterleitungshinweis",
		"Wikibooks",
		"Wikiquote",
		"Wikisource",
		"Wiktionary",
		"Toter Link",
	}

	for _, template := range ignoreTemplates {
		regex = regexp.MustCompile(`(?i)(\* )?\{\{` + template + `(.|\n|\r)*?}}\n?`)
		content = regex.ReplaceAllString(content, "")
	}

	return content
}
