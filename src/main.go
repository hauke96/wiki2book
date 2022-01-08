package main

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/generator/epub"
	"github.com/hauke96/wiki2book/src/generator/html"
	"github.com/hauke96/wiki2book/src/parser"
	"github.com/hauke96/wiki2book/src/project"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	generateEbook()
}

func generateTestEbook() {
	imageFolder := "../test/images"
	templateFolder := "../test/templates"

	fileContent, err := ioutil.ReadFile("../test/test.mediawiki")
	sigolo.FatalCheck(err)

	articleName := parser.Parse(string(fileContent), "test", imageFolder, templateFolder)

	err = api.DownloadImages(articleName.Images, imageFolder)
	sigolo.FatalCheck(err)

	_, err = html.Generate(articleName, "../test/", "../example/style.css", imageFolder)
	sigolo.FatalCheck(err)

	sigolo.Info("Start generating EPUB file")
	metadata := project.Metadata{
		Title: "Foobar",
	}
	err = epub.Generate([]string{"../test/test.html"}, "../test/test.epub", "../example/style.css", "../example/wikipedia-astronomie-cover.png", metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}

func generateEbook() {
	projectFile := os.Args[1]

	if "test" == projectFile {
		sigolo.Info("Use test file instead of real project file")
		generateTestEbook()
		os.Exit(0)
	}

	sigolo.Info("Use project file: %s", projectFile)

	directory, _ := filepath.Split(projectFile)
	err := os.Chdir(directory)
	sigolo.FatalCheck(err)

	project, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	var articleFiles []string

	for _, articleName := range project.Articles {
		sigolo.Info("Start processing articleName %s", articleName)

		wikiPageDto, err := api.DownloadPage(project.Domain, articleName)
		sigolo.FatalCheck(err)

		article := parser.Parse(wikiPageDto.Parse.Wikitext.Content, wikiPageDto.Parse.Title, project.ImageFolder, project.TemplateFolder)

		err, outputFile := generateHtml(article, project.Style)
		sigolo.FatalCheck(err)

		articleFiles = append(articleFiles, outputFile)

		sigolo.Info("Succeesfully created HTML for articleName %s", articleName)
	}

	sigolo.Info("Start generating EPUB file")
	err = epub.Generate(articleFiles, project.OutputFile, project.Style, project.Cover, project.Metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}

func generateHtml(wikiPage parser.Article, styleFile string) (error, string) {
	imageFolder := "./images"

	err := api.DownloadImages(wikiPage.Images, imageFolder)
	sigolo.FatalCheck(err)

	outputFile, err := html.Generate(wikiPage, "./", styleFile, imageFolder)
	sigolo.FatalCheck(err)
	return err, outputFile
}
