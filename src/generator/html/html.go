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

	//content := wikiPage.Content
	content := escapeSpecialCharacters(wikiPage.Content)
	content = replaceImages(content)
	content = replaceInternalLinks(content)
	content = replaceSections(content)
	content = replaceFormattings(content)
	content = replaceLinks(content)
	content = replaceUnorderedList(content)
	content = removeEmptyLines(content)

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
	regex := regexp.MustCompile("([^\\[])\\[(http[^\\[]+) ([^\\]]+)]([^\\]])")
	content = regex.ReplaceAllString(content, "$1<a href=\"$2\">$3</a>$4")

	regex = regexp.MustCompile("([^\\[])\\[(http[^\\[]+)]([^\\]])")
	// Also match normal URLs like [https://...]
	content = regex.ReplaceAllString(content, "$1<a href=\"$2\"></a>$3")

	return content
}

func replaceInternalLinks(content string) string {
	regex := regexp.MustCompile("\\[\\[([^|]*?)]]")
	content = regex.ReplaceAllString(content, "$1")

	regex = regexp.MustCompile("\\[\\[[^\\[]*?\\|(.*?)]]")
	content = regex.ReplaceAllString(content, "$1")

	return content
}

func replaceImages(content string) string {
	start := "<br><img src=\"./images/"
	end := "\"><br>"

	regex := regexp.MustCompile("\\[\\[(Datei|File):(.*?)\\|(.|\\n|\\r)*?]]")
	submatches := regex.FindAllStringSubmatch(content, -1)
	for _, submatch := range submatches {
		includeCommand := start + strings.ReplaceAll(submatch[2], "\\_", "_") + end
		if strings.HasSuffix(submatch[2], ".svg") || strings.HasSuffix(submatch[2], ".gif") {
			includeCommand = ""
		}
		content = strings.ReplaceAll(content, submatch[0], includeCommand)
	}

	return content
}

func replaceUnorderedList(content string) string {
	ulRegex := regexp.MustCompile("((\\n\\*+ .*)+)")
	liRegex := regexp.MustCompile("\\n\\* (.*)")
	nestedLiRegex := regexp.MustCompile("\\n\\*(\\*+ .*)")

	for {
		if !ulRegex.MatchString(content) {
			break
		}

		content = ulRegex.ReplaceAllString(content, "\n<ul>$1\n</ul>")
		content = liRegex.ReplaceAllString(content, "\n<li>$1</li>")
		content = nestedLiRegex.ReplaceAllString(content, "\n$1") // Remove a * from nested items so that they will be replaced in one of the following iterations
	}

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
