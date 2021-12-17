package html

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/wiki"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const HEADER = `
<html>
<body>
`
const FOOTER = `
</html>
</body>
`
const FILE_PLACEHOLDER = "__FILE__"

func Generate(wikiPage wiki.Article, outputFolder string) error {
	latexFileContent := HEADER
	latexFileContent += "\n<h1>" + wikiPage.Title + "</h1>"

	content := escapeSpecialCharacters(wikiPage.Content)
	//content = replaceMathMode(content)
	content = prepareImages(content)
	content = replaceInternalLinks(content)
	content = replaceImages(content)
	content = replaceSections(content)

	latexFileContent += content
	latexFileContent += FOOTER

	return write(wikiPage.Title, outputFolder, latexFileContent)
}

func write(title string, outputFolder string, content string) error {
	// Create the output folder
	err := os.Mkdir(outputFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", outputFolder))
	}

	// Create output file
	outputFilepath := filepath.Join(outputFolder, title+".html")
	sigolo.Info("Write to %s", outputFilepath)
	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create LaTeX output file %s", outputFilepath))
	}
	defer outputFile.Close()

	// Write data to file
	_, err = outputFile.WriteString(content)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable write LaTeX data to file %s", outputFilepath))
	}

	return nil
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

	content = strings.ReplaceAll(content, "\n=== ", "\n<h4>")
	content = strings.ReplaceAll(content, " ===\n", "</h4>\n")

	return content
}

// Protect the [[File:... tags against replacement by replaceInternalLinks.
func prepareImages(content string) string {
	regex := regexp.MustCompile("\\[\\[(Datei|File):")
	content = regex.ReplaceAllString(content, FILE_PLACEHOLDER)

	return content
}

func replaceImages(content string) string {
	start := "<br><img src=\"./images/"
	end := "\"><br>"

	regex := regexp.MustCompile(FILE_PLACEHOLDER + "(.*?)\\|(.|\\n|\\r)*?]]")
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		includeCommand := start + strings.ReplaceAll(submatch[1], "\\_", "_") + end
		if strings.HasSuffix(submatch[1], ".svg") || strings.HasSuffix(submatch[1], ".gif") {
			includeCommand = ""
		}
		content = strings.ReplaceAll(content, submatch[0], includeCommand)
	}

	return content
}

func replaceInternalLinks(content string) string {
	regex := regexp.MustCompile("\\[\\[([^|]*?)]]")
	content = regex.ReplaceAllString(content, "$1")

	regex = regexp.MustCompile("\\[\\[.*?\\|(.*?)]]")
	content = regex.ReplaceAllString(content, "$1")

	return content
}

func replaceMathMode(content string) string {
	content = strings.ReplaceAll(content, "<math>", "$")
	content = strings.ReplaceAll(content, "</math>", "$")
	return content
}
