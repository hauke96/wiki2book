package main

import (
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/generator/epub"
	"github.com/hauke96/wiki2book/src/generator/html"
	"github.com/hauke96/wiki2book/src/parser"
	"github.com/hauke96/wiki2book/src/project"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

func main() {
	//fileContent, err := ioutil.ReadFile("./test.mediawiki")
	//sigolo.FatalCheck(err)
	//
	//article := parser.Parse(string(fileContent), "test")
	//
	//err = api.DownloadImages(article.Images, "./images")
	//sigolo.FatalCheck(err)
	//
	//html.Generate(article, ".", "../example/style.css")
	//os.Exit(0)

	projectFile := os.Args[1]

	directory, _ := filepath.Split(projectFile)
	err := os.Chdir(directory)
	sigolo.FatalCheck(err)

	project, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	var articleFiles []string

	for _, article := range project.Articles {
		sigolo.Info("Start processing article %s", article)

		err, outputFile := generateHtml(article, project.Domain, project.Style)
		sigolo.FatalCheck(err)

		articleFiles = append(articleFiles, outputFile)

		sigolo.Info("Succeesfully created HTML for article %s", article)
	}

	sigolo.Info("Start generating EPUB file")
	err = epub.Generate(articleFiles, project.Output, project.Style, project.Cover, project.Metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}

func generateHtml(article string, language string, styleFile string) (error, string) {
	wikiPageDto, err := api.DownloadPage(language, article)
	sigolo.FatalCheck(err)

	wikiPage := parser.Parse(wikiPageDto.Parse.Wikitext.Content, wikiPageDto.Parse.Title)

	err = api.DownloadImages(wikiPage.Images, "./images")
	sigolo.FatalCheck(err)

	outputFile, err := html.Generate(wikiPage, "./", styleFile)
	sigolo.FatalCheck(err)
	return err, outputFile
}

// createAndUseFolder creates a folder with the given name and goes into that folder.
func createAndUseFolder(title string) error {
	err := os.Mkdir(title, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, fmt.Sprintf("Error creating output directory %s", title))
	}

	err = os.Chdir(title)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error switching into output directory %s", title))
	}

	return nil
}
