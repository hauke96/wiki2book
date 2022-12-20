package main

import (
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/generator/epub"
	"github.com/hauke96/wiki2book/src/generator/html"
	"github.com/hauke96/wiki2book/src/parser"
	"github.com/hauke96/wiki2book/src/project"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var cli struct {
	Debug      bool `help:"Enable debug mode." short:"d"`
	Standalone struct {
		File          string `help:"A mediawiki file tha should be rendered to an eBook." arg:""`
		OutputDir     string `help:"The directory where all the files should be put into." short:"o"`
		CacheDir      string `help:"The directory where all cached files will be written to." default:".wiki2book"` // TODO add this to the other commands as well
		StyleFile     string `help:"The CSS file that should be used." short:"s"`
		CoverImage    string `help:"A cover image for the front cover of the eBook." short:"c"`
		PandocDataDir string `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p"`
	} `cmd:"" help:"Renders a single mediawiki file into an eBook."`
	Project struct {
		ProjectFile string `help:"A project JSON-file tha should be used to create an eBook." type:"existingfile:" arg:""`
	} `cmd:"" help:"Uses a project file to create the eBook."`
	Article struct {
		ArticleName   string `help:"The name of the article to render." arg:""`
		OutputFile    string `help:"The path to the EPUB-file." short:"o" default:"ebook.epub"`
		StyleFile     string `help:"The CSS file that should be used." short:"s"`
		CoverImage    string `help:"A cover image for the front cover of the eBook." short:"c"`
		PandocDataDir string `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p"`
	} `cmd:"" help:"Renders a specific article into an eBook."`
}

func main() {
	ctx := kong.Parse(&cli)

	if cli.Debug {
		sigolo.LogLevel = sigolo.LOG_DEBUG
	}

	start := time.Now()

	switch ctx.Command() {
	case "standalone <file>":
		assertFileExists(cli.Standalone.StyleFile)
		assertFileExists(cli.Standalone.CoverImage)
		generateStandaloneEbook(cli.Standalone.File, cli.Standalone.OutputDir, cli.Standalone.CacheDir, cli.Standalone.StyleFile, cli.Standalone.CoverImage, cli.Standalone.PandocDataDir)
	case "project <project-file>":
		generateProjectEbook(cli.Project.ProjectFile)
	case "article <article-name>":
		generateArticleEbook(cli.Article.ArticleName, cli.Article.OutputFile, cli.Article.StyleFile, cli.Article.CoverImage, cli.Article.PandocDataDir)
	default:
		sigolo.Fatal("Unknown command: %v\n%#v", ctx.Command(), ctx)
	}

	end := time.Now()
	sigolo.Debug("Start   : %s", start.Format(time.RFC1123))
	sigolo.Debug("End     : %s", end.Format(time.RFC1123))
	sigolo.Debug("Duration: %f seconds", end.Sub(start).Seconds())
}

func generateProjectEbook(projectFile string) {
	var err error
	// Enable this to create a profiling file. Then use the command "go tool pprof src ./profiling.prof" and enter "web" to open a diagram in your browser.
	//f, err := os.Create("profiling.prof")
	//sigolo.FatalCheck(err)
	//
	//err = pprof.StartCPUProfile(f)
	//sigolo.FatalCheck(err)
	//defer pprof.StopCPUProfile()

	sigolo.Info("Use project file: %s", projectFile)

	directory, projectFile := filepath.Split(projectFile)
	sigolo.Debug("Go into folder %s", directory)
	err = os.Chdir(directory)
	sigolo.FatalCheck(err)

	project, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	articles := project.Articles
	wikipediaDomain := project.Domain
	articleCache := project.Caches.Articles
	imageCache := project.Caches.Images
	templateCache := project.Caches.Templates
	mathCache := project.Caches.Math
	styleFile := project.Style
	coverFile := project.Cover
	metadata := project.Metadata
	outputFile := project.OutputFile
	pandocDataDir := project.PandocDataDir

	generateEpubFromArticles(articles, wikipediaDomain, articleCache, imageCache, templateCache, styleFile, mathCache, outputFile, coverFile, pandocDataDir, metadata)
}

func generateStandaloneEbook(inputFile string, outputFolder string, cacheFolder string, styleFile string, coverImageFile string, pandocDataDir string) {
	var err error

	imageFolder := path.Join(cacheFolder, "images")
	mathFolder := path.Join(cacheFolder, "math")
	templateFolder := path.Join(cacheFolder, "templates")
	articleFolder := path.Join(cacheFolder, "articles")

	_, inputFileName := path.Split(inputFile)
	title := strings.Split(inputFileName, ".")[0]

	fileContent, err := ioutil.ReadFile(inputFile)
	sigolo.FatalCheck(err)

	err = os.MkdirAll(cacheFolder, os.ModePerm)
	sigolo.FatalCheck(err)

	tokenizer := parser.NewTokenizer(imageFolder, templateFolder)
	article := parser.Parse(string(fileContent), title, &tokenizer)

	err = api.DownloadImages(article.Images, imageFolder, articleFolder)
	sigolo.FatalCheck(err)

	htmlGenerator := &html.HtmlGenerator{}
	_, err = htmlGenerator.Generate(article, outputFolder, styleFile, imageFolder, mathFolder, articleFolder)
	sigolo.FatalCheck(err)

	sigolo.Info("Start generating EPUB file")
	metadata := project.Metadata{
		Title: title,
	}

	htmlFile := path.Join(outputFolder, title+".html")
	epubFile := path.Join(outputFolder, title+".epub")

	err = epub.Generate([]string{htmlFile}, epubFile, styleFile, coverImageFile, pandocDataDir, metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}

func generateArticleEbook(articleName string, outputFile string, styleFile string, coverImageFile string, pandocDataDir string) {
	var err error

	// Make ".wiki2book" a parameter
	cacheFolder := ".wiki2book"
	imageFolder := path.Join(cacheFolder, "images")
	mathFolder := path.Join(cacheFolder, "math")
	templateFolder := path.Join(cacheFolder, "templates")
	articleFolder := path.Join(cacheFolder, "articles")

	err = os.MkdirAll(cacheFolder, os.ModePerm)
	sigolo.FatalCheck(err)

	var articles []string
	articles = append(articles, articleName)

	// TODO Parameterize all of these fixed strings
	generateEpubFromArticles(articles,
		"de",
		articleFolder,
		imageFolder,
		templateFolder,
		styleFile,
		mathFolder,
		outputFile,
		coverImageFile,
		pandocDataDir,
		project.Metadata{})
}

func generateEpubFromArticles(articles []string, wikipediaDomain string, articleCache string, imageCache string, templateCache string, styleFile string, mathCache string, outputFile string, coverFile string, pandocDataDir string, metadata project.Metadata) {
	var articleFiles []string

	for _, articleName := range articles {
		sigolo.Info("Start processing articleName %s", articleName)

		wikiArticleDto, err := api.DownloadArticle(wikipediaDomain, articleName, articleCache)
		sigolo.FatalCheck(err)

		tokenizer := parser.NewTokenizer(imageCache, templateCache)
		article := parser.Parse(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.Title, &tokenizer)

		err = api.DownloadImages(article.Images, imageCache, articleCache)
		sigolo.FatalCheck(err)

		htmlGenerator := &html.HtmlGenerator{}
		outputFile, err := htmlGenerator.Generate(article, "./", styleFile, imageCache, mathCache, articleCache)
		sigolo.FatalCheck(err)

		articleFiles = append(articleFiles, outputFile)

		sigolo.Info("Succeesfully created HTML for articleName %s", articleName)
	}

	sigolo.Info("Start generating EPUB file")
	err := epub.Generate(articleFiles, outputFile, styleFile, coverFile, pandocDataDir, metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}

func assertFileExists(path string) {
	if _, err := os.Stat(path); strings.TrimSpace(path) != "" && err != nil {
		sigolo.Fatal("File path '%s' does not exist", path)
	}
}
