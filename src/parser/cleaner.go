package parser

import (
	"regexp"
)

func clean(content string) string {
	content = removeUnwantedCategories(content)
	content = removeUnwantedTemplates(content)
	content = removeUnwantedHtml(content)
	return content
}

func removeUnwantedCategories(content string) string {
	regex := regexp.MustCompile(`\[\[(Kategorie|Category):[^]]*?]]\n?`)
	return regex.ReplaceAllString(content, "")
}

func removeUnwantedTemplates(content string) string {
	ignoreTemplates := []string{
		"Alpha Centauri",
		"Begriffsklärungshinweis",
		"Commons",
		"Dieser Artikel",
		"Exzellent",
		"Gesprochener",
		"Graph:Chart",
		"Hauptartikel",
		"Lesenswert",
		"Linkbox",
		"Manueller Rahmen",
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
		regex := regexp.MustCompile(`(?i)(\* )?\{\{` + template + `[^}]*?}}\n?`)
		content = regex.ReplaceAllString(content, "")
	}

	return content
}

func removeUnwantedHtml(content string) string {
	regex := regexp.MustCompile(`<div[^>]*>`)
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile(`</div>`)
	content = regex.ReplaceAllString(content, "")

	return content
}
