package main

import (
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/generator/html"
	"github.com/hauke96/wiki2book/src/parser"
	"io/ioutil"
	"os"
)

func main() {
	imageFolder := "./images"

	fileContent, err := ioutil.ReadFile("./test.mediawiki")
	sigolo.FatalCheck(err)

	articleName := parser.Parse(string(fileContent), "test", imageFolder)

	err = api.DownloadImages(articleName.Images, imageFolder)
	sigolo.FatalCheck(err)

	html.Generate(articleName, ".", "../example/style.css")
	os.Exit(0)

	//projectFile := os.Args[1]
	//
	//directory, _ := filepath.Split(projectFile)
	//err := os.Chdir(directory)
	//sigolo.FatalCheck(err)
	//
	//project, err := project.LoadProject(projectFile)
	//sigolo.FatalCheck(err)
	//
	//var articleFiles []string
	//
	//for _, articleName := range project.Articles {
	//	sigolo.Info("Start processing articleName %s", articleName)
	//
	//	wikiPageDto, err := api.DownloadPage(project.Domain, articleName)
	//	sigolo.FatalCheck(err)
	//
	//	article := parser.Parse(wikiPageDto.Parse.Wikitext.Content, wikiPageDto.Parse.Title)
	//
	//	err, outputFile := generateHtml(article, project.Style)
	//	sigolo.FatalCheck(err)
	//
	//	articleFiles = append(articleFiles, outputFile)
	//
	//	sigolo.Info("Succeesfully created HTML for articleName %s", articleName)
	//}
	//
	//sigolo.Info("Start generating EPUB file")
	//err = epub.Generate(articleFiles, project.Output, project.Style, project.Cover, project.Metadata)
	//sigolo.FatalCheck(err)
	//sigolo.Info("Successfully created EPUB file")
}

func generateHtml(wikiPage parser.Article, styleFile string) (error, string) {
	err := api.DownloadImages(wikiPage.Images, "./images")
	sigolo.FatalCheck(err)

	outputFile, err := html.Generate(wikiPage, "./", styleFile)
	sigolo.FatalCheck(err)
	return err, outputFile
}
