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
	mathFolder := "../test/math"
	templateFolder := "../test/templates"

	fileContent, err := ioutil.ReadFile("../test/test.mediawiki")
	sigolo.FatalCheck(err)

	tokenizer := parser.NewTokenizer(imageFolder, templateFolder)
	articleName := parser.Parse(string(fileContent), "test", &tokenizer)

	err = api.DownloadImages(articleName.Images, imageFolder)
	sigolo.FatalCheck(err)

	_, err = html.Generate(articleName, "../test/", "../example/style.css", imageFolder, mathFolder)
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

		wikiPageDto, err := api.DownloadPage(project.Domain, articleName, project.Caches.Articles)
		sigolo.FatalCheck(err)

		tokenizer := parser.NewTokenizer(project.Caches.Images, project.Caches.Templates)
		article := parser.Parse(wikiPageDto.Parse.Wikitext.Content, wikiPageDto.Parse.Title, &tokenizer)

		err, outputFile := generateHtml(article, project.Style, project.Caches.Images, project.Caches.Math)
		sigolo.FatalCheck(err)

		articleFiles = append(articleFiles, outputFile)

		sigolo.Info("Succeesfully created HTML for articleName %s", articleName)
	}

	sigolo.Info("Start generating EPUB file")
	err = epub.Generate(articleFiles, project.OutputFile, project.Style, project.Cover, project.Metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}

func generateHtml(wikiPage parser.Article, styleFile string, imageCacheFolder string, mathCacheFolder string) (error, string) {
	err := api.DownloadImages(wikiPage.Images, imageCacheFolder)
	sigolo.FatalCheck(err)

	outputFile, err := html.Generate(wikiPage, "./", styleFile, imageCacheFolder, mathCacheFolder)
	sigolo.FatalCheck(err)
	return err, outputFile
}
