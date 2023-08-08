package parser

import (
	"strings"
	"wiki2book/config"
	"wiki2book/util"
)

const semiHeadingDepth = 10

func clean(content string) string {
	content = removeComments(content)
	content = removeUnwantedInternalLinks(content)
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

			linkContent := content[i+2 : endIndex]
			allPrefixes := strings.SplitN(linkContent, ":", -1)

			if len(allPrefixes) > 0 && util.Contains(config.Current.FilePrefixe, strings.ToLower(allPrefixes[0])) {
				// We found an image. This extra treatment exists because images might contain colons and that would
				// disturb the rest of the parsing below. Therefore, images get this fast exit.
				continue
			}

			// Go through all prefixes and see if any one is forbidden. Use the "-1" to skip the last element, which is
			// the actual link/article and not a prefix anymore.
			for j := 0; j < len(allPrefixes)-1; j++ {
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

func removeUnwantedTemplates(content string) string {
	// All lower case. Makes things easier below.
	ignoreTemplates := util.AllToLower(config.Current.IgnoredTemplates)

	for {
		originalContent := content

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

				if util.Contains(ignoreTemplates, templateName) {
					// Replace the template with an empty string, since it should be ignored.
					content = strings.Replace(content, templateText, "", 1)

					// Continue from the original template position (only +1 and not +2 to skip the "}}" token, because
					// of the i++ of the loop)
					i = closedTemplateIndex + 1
				}
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
	matches := headingRegex.FindAllStringSubmatch(line, -1)
	if len(matches) >= 1 && len(matches[0]) == 3 {
		lenHeadingPrefix := len(matches[0][1])
		lenHeadingSuffix := len(matches[0][2])
		if lenHeadingPrefix > 0 && lenHeadingSuffix > 0 && lenHeadingPrefix == lenHeadingSuffix {
			return len(matches[0][1])
		}
	}

	lineIsSemiHeading := semiHeadingRegex.MatchString(line)
	if lineIsSemiHeading {
		// This is a semi heading: Just bold text in this line -> Interpret this as most insignificant heading
		return semiHeadingDepth
	}

	return 0
}
