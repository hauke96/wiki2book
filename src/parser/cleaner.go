package parser

import (
	"regexp"
	"strings"
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
	// All lower case. Makes things easier below.
	ignoreTemplates := []string{
		"alpha centauri",
		"begriffsklärungshinweis",
		"commons",
		"dieser artikel",
		"exzellent",
		"gesprochener",
		"graph:chart",
		"hauptartikel",
		"lesenswert",
		"linkbox",
		"manueller rahmen",
		"navigationsleiste",
		"normdaten",
		"panorama",
		"siehe auch",
		"weiterleitungshinweis",
		"wikibooks",
		"wikiquote",
		"wikisource",
		"wiktionary",
		"toter link",
	}

	// Find all templates that actually appear in the text
	lowerCaseContent := strings.ToLower(content)
	var ignoreRegexes []*regexp.Regexp
	for _, template := range ignoreTemplates {
		if strings.Contains(lowerCaseContent, template) {
			ignoreRegexes = append(ignoreRegexes, regexp.MustCompile(`(?i)(\* )?\{\{`+template+`[^}]*?}}\n?`))
		}
	}

	var matches []string
	for _, regex := range ignoreRegexes {
		matches = append(matches, regex.FindAllString(content, -1)...)
	}

	for _, match := range matches {
		content = strings.ReplaceAll(content, match, "")
	}

	return content
}

func removeUnwantedHtml(content string) string {
	regex := regexp.MustCompile(`<(div|span)[^>]*>`)
	content = regex.ReplaceAllString(content, "")

	regex = regexp.MustCompile(`</(div|span)>`)
	content = regex.ReplaceAllString(content, "")

	return content
}
