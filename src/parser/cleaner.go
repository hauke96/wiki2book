package parser

import (
	"github.com/hauke96/wiki2book/src/util"
	"regexp"
	"strings"
)

var headingRegex = regexp.MustCompile(`^(=*)[^=]+(=*)$`)
var semiHeadingRegex = regexp.MustCompile(`^'''.+'''$`)
var categoryRegex = regexp.MustCompile(`\[\[(Kategorie|Category):[^]]*?]]\n?`)
var templateNameRegex = regexp.MustCompile(`{{\s*([^\n|}]+)`)
var unwantedHtmlRegex = regexp.MustCompile(`</?(div|span)[^>]*>`)

// Use multi-line matches (?m) to also match on \n
// TODO Edge case: Empty list item at the end of the content (with no trailing newline)
var emptyListItemRegex = regexp.MustCompile(`(?m)^(\s*[*#:;]+\s*\n)`)

func clean(content string) string {
	content = removeComments(content)
	content = removeUnwantedCategories(content)
	content = removeUnwantedTemplates(content)
	content = removeUnwantedHtml(content)
	content = removeUnwantedWikitext(content)
	content = removeEmptyListEntries(content)
	content = removeEmptySections(content)
	return content
}

func removeComments(content string) string {
	// The following steps are performed:
	//   1. Split by the end token "-->" of comments
	//   2. For each element in that slice, split by start token "<!--" of comments
	//   3. Only append the non-comment parts of the splits to the result segments

	splitContent := strings.Split(content, "-->")
	var resultSegments []string

	for i, splitItem := range splitContent {
		if i == len(splitContent)-1 {
			// The last string is never the end of a comment. It's either an empty string (in case the content directly
			// ends with a comment) or it's the text after the last comment.
			resultSegments = append(resultSegments, splitItem)
			continue
		}

		segments := strings.Split(splitItem, "<!--")
		if len(segments) == 1 {
			resultSegments = append(resultSegments, segments...)
			resultSegments = append(resultSegments, "-->")
		} else {
			resultSegments = append(resultSegments, segments[0])
		}
	}

	return strings.Join(resultSegments, "")
}

func removeUnwantedCategories(content string) string {
	return categoryRegex.ReplaceAllString(content, "")
}

func removeUnwantedTemplates(content string) string {
	// All lower case. Makes things easier below.
	ignoreTemplates := []string{
		"alpha centauri",
		"begriffsklärung",
		"begriffsklärungshinweis",
		"belege fehlen",
		"belege",
		"commons",
		"commonscat",
		"dieser artikel",
		"exzellent",
		"gesprochener",
		"gesprochene version",
		"graph:chart",
		"hauptartikel",
		"klade",
		"lesenswert",
		"linkbox",
		"lückenhaft",
		"manueller rahmen",
		"navigationsleiste",
		"navigationsleiste sonnensystem",
		"naviblock",
		"normdaten",
		"panorama",
		"portal",
		"quellen",
		"redundanztext",
		"siehe auch",
		"staatslastig",
		"toc",
		"toter link",
		"überarbeiten",
		"veraltet",
		"weiterleitungshinweis",
		"wikibooks",
		"wikiquote",
		"wikisource",
		"wiktionary",
	}

	lastOpeningTemplateIndex := -1

	for {
		originalContent := content

		for i := 0; i < len(content)-1; i++ {
			cursor := content[i : i+2]

			if cursor == "{{" {
				lastOpeningTemplateIndex = i
			} else if lastOpeningTemplateIndex != -1 && cursor == "}}" {
				templateText := content[lastOpeningTemplateIndex : i+2]

				matches := templateNameRegex.FindStringSubmatch(templateText)
				if matches == nil {
					// No match found
					lastOpeningTemplateIndex = -1
					continue
				}

				templateName := strings.ToLower(matches[1])
				templateName = strings.TrimSpace(templateName)

				if util.Contains(ignoreTemplates, templateName) {
					// Replace the template with an empty string, since it should be ignored.
					content = strings.Replace(content, templateText, "", 1)

					// Continue from the original template position (-1 because of the i++ of the loop)
					i = lastOpeningTemplateIndex - 1
				}

				lastOpeningTemplateIndex = -1
			}
		}

		if content == originalContent {
			break
		}
	}

	return content
}

func removeUnwantedHtml(content string) string {
	return unwantedHtmlRegex.ReplaceAllString(content, "")
}

func removeUnwantedWikitext(content string) string {
	return strings.ReplaceAll(content, "__NOTOC__", "")
}

func removeEmptyListEntries(content string) string {
	emptyListItemMatches := emptyListItemRegex.FindAllStringSubmatch(content, -1)
	for _, match := range emptyListItemMatches {
		content = strings.Replace(content, match[1], "", 1)
	}

	return content
}

func removeEmptySections(content string) string {
	lines := strings.Split(content, "\n")
	var resultLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Is heading? -> Check if section is empty
		if currentHeadingDepth := headingDepth(line); currentHeadingDepth > 0 {
			sectionStartIndex := i

			// Get next line after "current heading"
			i++
			if i >= len(lines) {
				break
			}

			// Go through lines of this "current" section until the end of data has been reached OR the current line is
			// a heading.
			sectionIsEmpty := true
			i, sectionIsEmpty = walkSection(i, lines, currentHeadingDepth)

			// If the section was not empty, go back to the first line of the section. This causes the loop to go over
			// the lines again and this will especially add them to the result list.
			if !sectionIsEmpty {
				resultLines = append(resultLines, lines[sectionStartIndex])
				i = sectionStartIndex
				continue
			}

			// If the section was in deed empty, then we are probably sitting on a new heading (or the end of lines) and
			// want to parse this next section. To do this, we go one step back to compensate the incrementation of the
			// outer for loop and to process that heading during the next run of the outer loop.
			if i < len(lines) && headingDepth(lines[i]) > 0 {
				i--
			}
		} else {
			resultLines = append(resultLines, lines[i])
		}
	}

	return strings.Join(resultLines, "\n")
}

// walkSection goes through the lines from index i till the end or the next heading. For i the condition i < len(lines)
// has to hold.
func walkSection(i int, lines []string, previousHeadingDepth int) (int, bool) {
	sectionIsEmpty := true

	for i < len(lines) {
		line := getTrimmedLine(lines, i)

		// Is heading? -> We're done in this loop and return
		if headingDepth := headingDepth(line); headingDepth > 0 {
			// But if this next heading is a sub-heading of the current one, then ...
			if headingDepth > previousHeadingDepth {
				// ... interpret this section as non-empty to keep it, as it structures the document in a helpful way.
				sectionIsEmpty = false
			}
			break
		}

		// Not a heading ->
		sectionIsEmpty = sectionIsEmpty && len(line) == 0
		i++
		if i >= len(lines) {
			break
		}
	}

	return i, sectionIsEmpty
}

func getTrimmedLine(lines []string, i int) string {
	return strings.TrimSpace(lines[i])
}

// headingDepth returns the number of "=" characters. When 0 is returned, the line is not a heading.
func headingDepth(line string) int {
	matches := headingRegex.FindAllStringSubmatch(line, -1)
	if len(matches) >= 1 && len(matches[0]) == 3 {
		lenHeadingPrefix := len(matches[0][1])
		lenHeadingSuffix := len(matches[0][2])
		if lenHeadingPrefix > 0 && lenHeadingSuffix > 0 && lenHeadingPrefix == lenHeadingSuffix {
			return len(matches[0][1])
		}
	}

	matcheSemiHeading := semiHeadingRegex.MatchString(line)
	if matcheSemiHeading {
		// This is a semi heading: Just bold text in this line -> Interpret this as most insignificant heading
		return 10
	}

	return 0
}
