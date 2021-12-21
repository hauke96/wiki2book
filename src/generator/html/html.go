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

const HEADER = `<html>
<head>
<meta charset="utf-8">
<link rel="stylesheet" href="{{STYLE}}">
</head>
<body>
`
const FOOTER = `</body>
</html>
`

const HREF_TEMPLATE = "<a href=\"%s\">%s</a>"
const TABLE_TEMPLATE = `<table>
%s
</table>`
const TABLE_TEMPLATE_HEAD = `<th>
%s
</th>
`
const TABLE_TEMPLATE_ROW = `<tr>
%s
</tr>
`
const TABLE_TEMPLATE_COL = `<td>
%s
</td>
`
const TEMPLATE_UL = `<ul>
%s
</ul>
`
const TEMPLATE_LI = `<li>
%s
</li>
`

func Generate(wikiPage parser.Article, outputFolder string, styleFile string) (string, error) {
	content := strings.ReplaceAll(HEADER, "{{STYLE}}", styleFile)
	content += "\n<h1>" + wikiPage.Title + "</h1>"
	content += expand(wikiPage.Content, wikiPage.TokenMap)
	content += FOOTER
	return write(wikiPage.Title, outputFolder, content)
}

func expand(content string, tokenMap map[string]string) string {
	content = expandMarker(content)

	regex := regexp.MustCompile(parser.TOKEN_REGEX)
	submatches := regex.FindAllStringSubmatch(content, -1)

	if len(submatches) == 0 {
		// no token in content
		return content
	}

	for _, submatch := range submatches {
		sigolo.Info("Found token %s", submatch[1])

		html := submatch[0]

		switch submatch[1] {
		case parser.TOKEN_EXTERNAL_LINK:
			html = expandExternalLink(submatch[0], tokenMap)
		case parser.TOKEN_INTERNAL_LINK:
			html = expandInternalLint(submatch[0], tokenMap)
		case parser.TOKEN_TABLE:
			html = expandTable(submatch[0], tokenMap)
		case parser.TOKEN_TABLE_HEAD:
			html = expandTableHead(submatch[0], tokenMap)
		case parser.TOKEN_TABLE_ROW:
			html = expandTableRow(submatch[0], tokenMap)
		case parser.TOKEN_TABLE_COL:
			html = expandTableColumn(submatch[0], tokenMap)
		case parser.TOKEN_UNORDERED_LIST:
			html = expandUnorderedList(submatch[0], tokenMap)
		case parser.TOKEN_LIST_ITEM:
			html = expandListItem(submatch[0], tokenMap)
		}

		content = strings.Replace(content, submatch[0], html, 1)
	}

	return content
}

func expandMarker(content string) string {
	content = strings.ReplaceAll(content, parser.MARKER_BOLD_OPEN, "<b>")
	content = strings.ReplaceAll(content, parser.MARKER_BOLD_CLOSE, "</b>")
	content = strings.ReplaceAll(content, parser.MARKER_ITALIC_OPEN, "<i>")
	content = strings.ReplaceAll(content, parser.MARKER_ITALIC_CLOSE, "</i>")
	return content
}

func expandInternalLint(tokenString string, tokenMap map[string]string) string {
	splittedToken := strings.Split(tokenMap[tokenString], " ")
	text := expand(tokenMap[splittedToken[1]], tokenMap)
	// Yeah, let's not add an link to the article in an eBook. Maybe make it configurable some day...
	return text
}

func expandExternalLink(tokenString string, tokenMap map[string]string) string {
	splittedToken := strings.Split(tokenMap[tokenString], " ")
	url := tokenMap[splittedToken[0]]
	text := expand(tokenMap[splittedToken[1]], tokenMap)
	return fmt.Sprintf(HREF_TEMPLATE, url, text)
}

func expandTable(tokenString string, tokenMap map[string]string) string {
	tokenContent := tokenMap[tokenString]
	return fmt.Sprintf(TABLE_TEMPLATE, expand(tokenContent, tokenMap))
}

func expandTableHead(tokenString string, tokenMap map[string]string) string {
	tokenContent := tokenMap[tokenString]
	return fmt.Sprintf(TABLE_TEMPLATE_HEAD, expand(tokenContent, tokenMap))
}

func expandTableRow(tokenString string, tokenMap map[string]string) string {
	tokenContent := tokenMap[tokenString]
	return fmt.Sprintf(TABLE_TEMPLATE_ROW, expand(tokenContent, tokenMap))
}

func expandTableColumn(tokenString string, tokenMap map[string]string) string {
	tokenContent := tokenMap[tokenString]
	return fmt.Sprintf(TABLE_TEMPLATE_COL, expand(tokenContent, tokenMap))
}

func expandUnorderedList(tokenString string, tokenMap map[string]string) string {
	tokenContent := tokenMap[tokenString]
	return fmt.Sprintf(TEMPLATE_UL, expand(tokenContent, tokenMap))
}

func expandListItem(tokenString string, tokenMap map[string]string) string {
	tokenContent := tokenMap[tokenString]
	return fmt.Sprintf(TEMPLATE_LI, expand(tokenContent, tokenMap))
}

func escapeSpecialCharacters(content string) string {
	content = strings.ReplaceAll(content, " < ", " &lt; ")
	content = strings.ReplaceAll(content, " > ", " &gt; ")
	return content
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
