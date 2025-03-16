package parser

import (
	"strings"
	"wiki2book/config"
	"wiki2book/util"
)

const semiHeadingDepth = 10

func clean(content string) (string, error) {
	var err error

	content = removeComments(content)
	content = removeUnwantedInternalLinks(content)
	content = handleUnwantedAndTrailingTemplates(content)
	// Disabled for now. This caused problems, because sometimes the wikitext does contain useful HTML.
	//content = removeUnwantedHtml(content)
	content = removeUnwantedWikitext(content)
	content = removeEmptyListEntries(content)
	content = removeEmptySections(content)

	content, err = hackGermanRailwayTemplates(content, 0)
	if err != nil {
		return "", err
	}

	return content, nil
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
			nonCommentSegment := segments[0]

			// Remove potentially trailing newline:
			if len(nonCommentSegment) > 0 && nonCommentSegment[len(nonCommentSegment)-1] == '\n' {
				// If this segments ends with a newline, the comment following it started with a newline. We remove the
				// newline, because otherwise, there will be additional blank lines between former comments. Example:
				// "foo\n<!--comment-->\nbar"  would turn into  "foo\n\nbar"  instead of  "foo\nbar"
				nonCommentSegment = nonCommentSegment[:len(nonCommentSegment)-1]
			}

			resultSegments = append(resultSegments, nonCommentSegment)
		}
	}

	return strings.Join(resultSegments, "")
}

// removeUnwantedInternalLinks removes all kind of unwanted links. This method leaves all allowed internal links
// unchanged (links with a certain prefix). Category prefixes and all other not explicitly allowed prefixes are
// considered unwanted and each such link will be removed.
func removeUnwantedInternalLinks(content string) string {
	// Go through all characters with a 2-char sliding window, hence the "-2".
	for i := 0; i < len(content)-2; i++ {
		cursor := content[i : i+2]

		if cursor == "[[" {
			endIndex := findCorrespondingCloseToken(content, i+2, "[", "]")

			totalLinkContent := content[i+2 : endIndex]
			linkSegments := strings.SplitN(totalLinkContent, ":", -1)
			allPrefixes := linkSegments[:len(linkSegments)-1]

			// It's possible to link categories with a leading colon, e.g. [[:SomeCategory:SomeLinkText]]. They should
			// be treated as normal internal links, i.e. the "SomeLinkText" in this example should stay.
			isNormalLinkToCategory := content[i+2] == ':'

			if len(allPrefixes) > 0 && util.Contains(config.Current.FilePrefixe, strings.ToLower(allPrefixes[0])) {
				// We found an image. This extra treatment exists because images might contain colons and that would
				// disturb the rest of the parsing below. Therefore, images get this fast exit.
				continue
			}

			if isNormalLinkToCategory {
				// We found a category (or other category-like object), but it starts with a ":" and should therefore
				// be treated as normal link. This treatment happens in the parsing of normal links, which means there's
				// nothing to clean here.
				continue
			}

			// Go through all prefixes and see if any one is forbidden.
			for j := 0; j < len(allPrefixes); j++ {
				linkPrefix := strings.ToLower(allPrefixes[j])

				isForbiddenPrefix := !util.Contains(config.Current.FilePrefixe, linkPrefix) && !util.Contains(config.Current.AllowedLinkPrefixes, linkPrefix)
				isCategory := util.Contains(config.Current.CategoryPrefixes, linkPrefix)

				if linkPrefix != "" && (isForbiddenPrefix || isCategory) {
					content = content[0:i] + content[endIndex+2:]

					// Compensate "i++" from loop to not skip a character and continue with the outer loop to find the
					// next link.
					i--
					break
				}
			}
		}
	}

	return content
}

func handleUnwantedAndTrailingTemplates(content string) string {
	// All lower case. Makes things easier below.
	ignoreTemplates := util.AllToLower(config.Current.IgnoredTemplates)
	trailingTemplates := util.AllToLower(config.Current.TrailingTemplates)
	var foundTrailingTemplates []string

	for i := 0; i < len(content)-1; i++ {
		cursor := content[i : i+2]

		if cursor == "{{" {
			// Get the index on which the template is closed
			closedTemplateIndex := findCorrespondingCloseToken(content, i+2, "{{", "}}")
			if closedTemplateIndex == -1 {
				// no closing tag found -> move on in the normal text
				continue
			}

			templateText := content[i : closedTemplateIndex+2]
			templateNameMatches := templateNameRegex.FindStringSubmatch(templateText)
			if templateNameMatches == nil {
				// No match found
				continue
			}

			templateName := strings.ToLower(templateNameMatches[1])
			templateName = strings.TrimSpace(templateName)

			if util.Contains(ignoreTemplates, templateName) || util.Contains(trailingTemplates, templateName) {
				// Replace the template with an empty string, since it should be ignored.
				content = strings.Replace(content, templateText, "", 1)

				// Collect templates that should be moved to the bottom of the article
				if util.Contains(trailingTemplates, templateName) {
					foundTrailingTemplates = append(foundTrailingTemplates, templateName)
				}

				// Because the template was removed, we have to start from the original location again in the next
				// loop iteration. Therefore, we have to compensate the +1 of the loop counter.
				i--
			}
		}
	}

	// Re-add trailing templates to the bottom of the article
	for _, foundTrailingTemplate := range foundTrailingTemplates {
		content += "\n{{" + foundTrailingTemplate + "}}"
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
	var resultLines []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if len(line) == 0 || !isListBeginning(rune(line[0])) {
			// Line has no list beginning whatsoever -> not a list line -> don't ignore it
			resultLines = append(resultLines, line)
			continue
		}

		numberOfListPrefixChars := 0
		for _, c := range line {
			if isListBeginning(c) {
				numberOfListPrefixChars++
			}
		}

		if numberOfListPrefixChars != len(line) {
			// Number of counted list prefix chars differs from line length -> some other chars in line -> line not empty
			resultLines = append(resultLines, line)
		}
	}

	return strings.Join(resultLines, "\n")
}

func isListBeginning(c rune) bool {
	return c == '*' || c == '#' || c == ';' || c == ':'
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
				// This line is considered to be a heading. But: If this is only a semi-heading we leave it. It's
				// possible that the whole content is e.g. just a table cell or link text and therefore not a multi-row
				// article.
				if currentHeadingDepth == semiHeadingDepth {
					resultLines = append(resultLines, line)
				}
				break
			}

			// Go through lines of this "current" section until the end of data has been reached OR the current line is
			// a heading.
			var sectionIsEmpty bool
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
	if !strings.HasPrefix(line, "=") || !strings.HasSuffix(line, "=") {
		// No real heading. But maybe a semi-heading (line with only bold text)?
		if strings.HasPrefix(line, "'''") && strings.HasSuffix(line, "'''") {
			return semiHeadingDepth
		}
		// Not even semi-heading -> No heading at all
		return 0
	}

	headingDepthCounter := 0
	lineRunes := []rune(line)
	lineLength := len(lineRunes)
	for i, c := range lineRunes {
		if c != '=' && lineRunes[lineLength-1-i] != '=' {
			// Valid heading end (prefix and suffix have same length)
			break
		} else if c != lineRunes[lineLength-1-i] {
			// Invalid heading form (prefix and suffix have unequal length)
			return 0
		}
		headingDepthCounter++
	}

	return headingDepthCounter
	//
	//matches := headingRegex.FindAllStringSubmatch(line, -1)
	//if len(matches) >= 1 && len(matches[0]) == 3 {
	//	lenHeadingPrefix := len(matches[0][1])
	//	lenHeadingSuffix := len(matches[0][2])
	//	if lenHeadingPrefix > 0 && lenHeadingSuffix > 0 && lenHeadingPrefix == lenHeadingSuffix {
	//		return len(matches[0][1])
	//	}
	//}
	//
	//lineIsSemiHeading := semiHeadingRegex.MatchString(line)
	//if lineIsSemiHeading {
	//	// This is a semi heading: Just bold text in this line -> Interpret this as most insignificant heading
	//	return semiHeadingDepth
	//}
	//
	//return 0
}
