package parser

import (
	"regexp"
	"strings"
)

func clean(content string) string {
	content = removeUnwantedCategories(content)
	content = removeUnwantedTemplates(content)
	content = removeUnwantedHtml(content)
	content = removeEmptySections(content)
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
	regex := regexp.MustCompile(`</?(div|span)[^>]*>`)
	return regex.ReplaceAllString(content, "")
}

func removeEmptySections(content string) string {
	lines := strings.Split(content, "\n")
	var resultLines []string

	for i := 0; i < len(lines); i++ {
		line := getTrimmedLine(lines, i)

		// Is heading? -> Check if section is empty
		if isHeading(line) {
			sectionStartIndex := i
			i++
			if i >= len(lines) {
				break
			}
			line = getTrimmedLine(lines, i)
			sectionIsEmpty := true

			for i < len(lines) && !isHeading(line) && sectionIsEmpty {
				sectionIsEmpty = sectionIsEmpty && len(line) == 0
				i++
				if i >= len(lines) {
					break
				}
				line = getTrimmedLine(lines, i)
			}

			// If the section was not empty, go back to the first line of the section. This causes the loop to go over
			// the lines again and this will e.g. add them to the result list.
			if !sectionIsEmpty {
				resultLines = append(resultLines, lines[sectionStartIndex])
				i = sectionStartIndex
				continue
			}

			// When the exit condition of the above loop was "!isHeading", then we need to go one step back to process
			// that heading during the next run of the outer loop.
			if isHeading(line) {
				i--
			}
		} else {
			resultLines = append(resultLines, lines[i])
		}
	}

	return strings.Join(resultLines, "\n")
}

func getTrimmedLine(lines []string, i int) string {
	return strings.TrimSpace(lines[i])
}

func isHeading(line string) bool {
	return strings.HasPrefix(line, "=") && strings.HasSuffix(line, "=")
}
