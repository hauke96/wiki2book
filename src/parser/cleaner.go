package parser

import (
	"regexp"
)

func clean(content string) string {
	content = removeUnwantedTags(content)
	// TODO remove div tags?
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
		"Linkbox",
		"Graph:Chart",
		"Manueller Rahmen",
	}

	for _, template := range ignoreTemplates {
		regex = regexp.MustCompile(`(?i)(\* )?\{\{` + template + `(.|\n|\r)*?}}\n?`)
		content = regex.ReplaceAllString(content, "")
	}

	return content
}
