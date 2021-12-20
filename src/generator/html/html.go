package html

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/parser"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const HEADER = `
<html>
<head>
<link rel="stylesheet" href="{{STYLE}}">
</head>
<body>
`
const FOOTER = `
</html>
</body>
`

func Generate(wikiPage parser.Article, outputFolder string, styleFile string) (string, error) {
	latexFileContent := strings.ReplaceAll(HEADER, "{{STYLE}}", styleFile)
	latexFileContent += "\n<h1>" + wikiPage.Title + "</h1>"

	//content := wikiPage.Content
	content := escapeSpecialCharacters(wikiPage.Content)
	content = replaceInternalLinks(content)
	content = replaceImages(content)
	content = replaceSections(content)
	content = replaceFormattings(content)
	content = replaceLinks(content)
	//content = replaceUnorderedLists(content)
	content = Parse(content)
	content = removeEmptyLines(content)

	latexFileContent += content
	latexFileContent += FOOTER

	return write(wikiPage.Title, outputFolder, latexFileContent)
}

func Parse(content string) string {
	i := 0
	lines := strings.Split(content, "\n")

	// TODO replace fixed sized areas (like images, tables, everything with start and end marker)

	for i < len(lines) {
		line := lines[i]

		regex := regexp.MustCompile("^\\* ")
		if regex.MatchString(line) {
			originalString, html, newIndex := parseUnorderedList(lines, i)
			content = strings.ReplaceAll(content, originalString, html)
			i = newIndex
		}

		// Image?
		//regex := regexp.MustCompile("^\\[\\[(Datei|File):")
		//if regex.MatchString(line) {
		//	TODO parse link
		//continue
		//}

		i++
	}

	return content
}

func parseUnorderedList(lines []string, i int) (completeListString string, html string, newIndex int) {
	newIndex = i
	completeListString = ""

	regex := regexp.MustCompile(`^([*:\[|! ].*)?$`)
	for {
		if i == len(lines) || !regex.MatchString(lines[i]) {
			break
		}
		completeListString += lines[i] + "\n"
		i++
	}

	html = parseUnorderedListString(completeListString)
	newIndex = i

	return
}

func parseUnorderedListString(completeListString string) string {
	listItemRegex := regexp.MustCompile(`(^|\n)\* `)
	if !listItemRegex.MatchString(completeListString) {
		return completeListString
	}

	html := "<ul>\n"

	// Ignore first item as it's always empty
	listItems := listItemRegex.Split(completeListString, -1)[1:]

	for _, listItem := range listItems {
		listItemLines := strings.Split(listItem, "\n")

		html += "<li>\n"
		html += listItemLines[0] + "\n"

		if len(listItemLines) > 1 {
			// Turn sub-lists starting with "** foo" into "* foo" for recursive parsing
			listItem = strings.Join(listItemLines[1:], "\n")
			listItem = strings.ReplaceAll(listItem, "* ", " ")
			html += parseUnorderedListString(listItem)
		}

		html += "</li>\n"
	}

	html += "</ul>\n"
	return html
}

func replaceTables(content string) string {
	regex := regexp.MustCompile(":?\\{\\|.*\\n(.|\\n|\\r)*?\\|}")
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		rawTableString := submatch[0]
		htmlTable := parseTable(rawTableString)
		content = strings.ReplaceAll(content, rawTableString, htmlTable)
	}

	return content
}

func parseTable(content string) string {
	rowRegex := regexp.MustCompile("\\|-(.|\\n|\\r)*?\\|-")

	headerMatch := rowRegex.FindString(content)
	htmlRow := parseRow(headerMatch, "!", "th")
	content = strings.ReplaceAll(content, headerMatch, htmlRow)

	for {
		if !rowRegex.MatchString(content) {
			break
		}

		match := rowRegex.FindString(content)
		htmlRow = parseRow(match, "|", "td")
		content = strings.ReplaceAll(content, match, htmlRow)
	}

	content = regexp.MustCompile(":\\{\\|.*").ReplaceAllString(content, "<table><tbody>")
	content = regexp.MustCompile("\\|\\-\\n\\|}").ReplaceAllString(content, "</table></tbody>")

	return content
}

func parseRow(content string, splitChar string, htmlColumnElement string) string {
	content = strings.ReplaceAll(content, "\n"+splitChar, splitChar+splitChar)
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, "|-", "")
	content = strings.ReplaceAll(content, splitChar+splitChar, "\n"+splitChar)
	content = strings.Trim(content, "\n")
	content = strings.Trim(content, "|")

	result := "<tr>\n"

	for _, column := range strings.Split(content, "\n") {
		// There might be styling before the actual data which is separated with |
		entries := strings.Split(column, splitChar)

		result += "<" + htmlColumnElement + ">"
		result += entries[len(entries)-1]
		result += "</" + htmlColumnElement + ">\n"
	}

	// Append |- for further parsing of following rows
	result += "</tr>\n|-"

	return result
}

func write(title string, outputFolder string, content string) (string, error) {
	// Create the output folder
	err := os.Mkdir(outputFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", outputFolder))
	}

	// Create output file
	outputFilepath := filepath.Join(outputFolder, title+".html")
	sigolo.Info("Write to %s", outputFilepath)
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to create LaTeX output file %s", outputFilepath))
	}
	defer outputFile.Close()

	// Write data to file
	_, err = outputFile.WriteString(content)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable write LaTeX data to file %s", outputFilepath))
	}

	return outputFilepath, nil
}

func escapeSpecialCharacters(content string) string {
	//content = strings.ReplaceAll(content, "&", "\\&")
	//content = strings.ReplaceAll(content, "_", "\\_")
	//content = strings.ReplaceAll(content, "#", "\\#")
	//content = strings.ReplaceAll(content, "{", "\\{")
	//content = strings.ReplaceAll(content, "}", "\\}")
	//content = strings.ReplaceAll(content, "$", "\\$")
	//content = strings.ReplaceAll(content, "%", "\\%")
	content = strings.ReplaceAll(content, " < ", " &lt; ")
	content = strings.ReplaceAll(content, " > ", " &gt; ")
	return content
}

func replaceSections(content string) string {
	content = strings.ReplaceAll(content, "\n== ", "\n<h2>")
	content = strings.ReplaceAll(content, " ==\n", "</h2>\n")

	content = strings.ReplaceAll(content, "\n=== ", "\n<h3>")
	content = strings.ReplaceAll(content, " ===\n", "</h3>\n")

	content = strings.ReplaceAll(content, "\n==== ", "\n<h4>")
	content = strings.ReplaceAll(content, " ====\n", "</h4>\n")

	content = strings.ReplaceAll(content, "\n===== ", "\n<h5>")
	content = strings.ReplaceAll(content, " =====\n", "</h5>\n")

	content = strings.ReplaceAll(content, "\n====== ", "\n<h6>")
	content = strings.ReplaceAll(content, " ======\n", "</h6>\n")

	return content
}

func replaceFormattings(content string) string {
	regex := regexp.MustCompile("'''(.+?)'''")
	content = regex.ReplaceAllString(content, "<b>$1</b>")

	regex = regexp.MustCompile("''(.+?)''")
	content = regex.ReplaceAllString(content, "<i>$1</i>")

	return content
}

func replaceLinks(content string) string {
	// Format of Links: [https://... Text]
	// To not be confused with internal [[Article]] links, the [^\[] parts say that we do NOT want brackets before and
	// after the potential match
	regex := regexp.MustCompile("([^\\[])\\[(http[^\\[]*?) ([^\\]]*?)]([^\\]])")
	content = regex.ReplaceAllString(content, "$1<a href=\"$2\">$3</a>$4")

	regex = regexp.MustCompile("([^\\[])\\[(http[^\\[]*?)]([^\\]])")
	// Also match normal URLs like [https://...]
	content = regex.ReplaceAllString(content, "$1<a href=\"$2\"></a>$3")

	return content
}

func replaceInternalLinks(content string) string {
	fileRegex := regexp.MustCompile("^\\[\\[(Datei|File):")

	regex := regexp.MustCompile("\\[\\[([^|]*?)]]")
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		if !fileRegex.MatchString(submatch[0]) {
			content = strings.ReplaceAll(content, submatch[0], submatch[1])
		}
	}

	regex = regexp.MustCompile("\\[\\[[^\\[]*?\\|([^|]*?)]]")
	submatches = regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		if !fileRegex.MatchString(submatch[0]) {
			content = strings.ReplaceAll(content, submatch[0], submatch[1])
		}
	}

	return content
}

func replaceImages(content string) string {
	template := "<br><div class=\"figure\"><img alt=\"{{ALT}}\" src=\"./images/{{FILE}}\"><div class=\"caption\">{{CAPTION}}</div></div>"

	regex := regexp.MustCompile("\\[\\[(Datei|File):((.|\\n|\\r)*?)\\|.*?\\|([^\\|]*?)]]")
	submatches := regex.FindAllStringSubmatch(content, -1)

	for _, submatch := range submatches {
		includeCommand := ""

		if !strings.HasSuffix(submatch[2], ".svg") && !strings.HasSuffix(submatch[2], ".gif") {
			filename := strings.ReplaceAll(submatch[2], "\\_", "_")

			includeCommand = strings.ReplaceAll(template, "{{FILE}}", filename)
			includeCommand = strings.ReplaceAll(includeCommand, "{{ALT}}", filename)
			includeCommand = strings.ReplaceAll(includeCommand, "{{CAPTION}}", submatch[4])
		}

		content = strings.ReplaceAll(content, submatch[0], includeCommand)
	}

	return content
}

func replaceUnorderedLists(content string) string {
	for {
		// Match from beginning of list to beginning of a new line which is either a new paragraph or a new text line
		ulRegex := regexp.MustCompile("\\n(\\* (.|\\n|\\r)*?)\\n([a-zA-Z0-9]|<h|<p)")
		if !ulRegex.MatchString(content) {
			break
		}

		submatches := ulRegex.FindAllStringSubmatch(content, -1)
		for _, list := range submatches {
			// list = list with possible sub-lists

			result := evaluateUnorderedListString("\\*", list[1])
			if result == "" {
				break
			}

			result += "\n" + list[3]

			content = strings.ReplaceAll(content, list[0], result)
		}
	}

	return content
}

func evaluateUnorderedListString(listPrefix string, listString string) string {
	result := ""
	listString = listString

	bulletRegex := regexp.MustCompile("(^|\\n)" + listPrefix + " ")
	if !bulletRegex.MatchString(listString) {
		return listString
	}

	listItems := bulletRegex.Split(listString, -1)
	if len(listItems) == 0 {
		return listString
	}

	for i, listItem := range listItems {
		if i == 0 {
			// First item is either empty (for new starting list) or consists of the parent item in case we're in a
			// sub-list here. I that case, we want to just add the parent items text and then add our new <ul> sub-list.
			if listItem != "" {
				result += "\n" + listItem + "\n"
			}
			result += "<ul>\n"
			continue
		}

		result += "<li>\n"
		result += evaluateUnorderedListString(listPrefix+"\\*", listItem)
		result += "</li>\n"
	}

	result += "</ul>\n"
	return result
}

func removeEmptyLines(content string) string {
	regex := regexp.MustCompile("\\n\\n\\n")
	for {
		if !regex.MatchString(content) {
			break
		}
		content = regex.ReplaceAllString(content, "\n\n")
	}

	return content
}
