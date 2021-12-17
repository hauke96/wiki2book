package generator

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/wiki"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

const header = `
% !TeX program = xelatex

\documentclass{book}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage[ngerman]{babel}

\begin{document}
	\tableofcontents
`
const footer = `
\end{document}
`

func Generate(wikiPage wiki.Article, outputFolder string) error {
	latexFileContent := header

	content := escapeSpecialCharacters(wikiPage.Content)
	content = replaceMathMode(content)
	content = replaceSections(content)

	latexFileContent += content
	latexFileContent += footer

	return write(wikiPage.Title, outputFolder, latexFileContent)
}

func write(title string, outputFolder string, content string) error {
	// Create the output folder
	err := os.Mkdir(outputFolder, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Unable to create output folder %s", outputFolder))
	}

	// Create output file
	outputFilepath := filepath.Join(outputFolder, title+".tex")
	sigolo.Info("Content to write %s", content)
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
	content = strings.ReplaceAll(content, "&", "\\&")
	content = strings.ReplaceAll(content, "_", "\\_")
	content = strings.ReplaceAll(content, "#", "\\#")
	content = strings.ReplaceAll(content, "{", "\\{")
	content = strings.ReplaceAll(content, "}", "\\}")
	content = strings.ReplaceAll(content, "$", "\\$")
	content = strings.ReplaceAll(content, "%", "\\%")
	return content
}

func replaceSections(content string) string {
	content = strings.ReplaceAll(content, "\n== ", "\n\\section{")
	content = strings.ReplaceAll(content, " ==\n", "}\n")
	return content
}

func replaceMathMode(content string) string {
	content = strings.ReplaceAll(content, "<math>", "$")
	content = strings.ReplaceAll(content, "</math>", "$")
	return content
}