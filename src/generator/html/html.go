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
</body>
</html>
`

const HREF_TEMPLATE = "<a href=\"%s\">%s</a>"

func Generate(wikiPage parser.Article, outputFolder string, styleFile string) (string, error) {
	content := strings.ReplaceAll(HEADER, "{{STYLE}}", styleFile)
	content += "\n<h1>" + wikiPage.Title + "</h1>"
	content += expand(wikiPage.Content, wikiPage.TokenMap)
	content += FOOTER
	return write(wikiPage.Title, outputFolder, content)
}

func expand(content string, tokenMap map[string]string) string {
	regex := regexp.MustCompile(parser.TOKEN_REGEX)
	submatches := regex.FindAllStringSubmatch(content, -1)

	if len(submatches) == 0 {
		// no token in content
		return content
	}

	for _, submatch := range submatches {
		sigolo.Info("Found token %s", submatch[1])
		switch submatch[1] {
		case parser.TOKEN_EXTERNAL_LINK:
			htmlLink := expandExternalLink(submatch[0], tokenMap)
			content = strings.Replace(content, submatch[0], htmlLink, 1)
		}
	}

	return content
}

func expandExternalLink(tokenString string, tokenMap map[string]string) string {
	splittedToken := strings.Split(tokenMap[tokenString], " ")
	url := tokenMap[splittedToken[0]]
	text := expand(tokenMap[splittedToken[1]], tokenMap)
	return fmt.Sprintf(HREF_TEMPLATE, url, text)
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
