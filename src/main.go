package main

import (
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	"github.com/hauke96/wiki2book/src/api"
	"github.com/hauke96/wiki2book/src/generator/epub"
	"github.com/hauke96/wiki2book/src/generator/html"
	"github.com/hauke96/wiki2book/src/parser"
	"github.com/hauke96/wiki2book/src/project"
	"github.com/hauke96/wiki2book/src/util"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"
)

var cli struct {
	Debug      bool `help:"Enable debug mode." short:"d"`
	Profiling  bool `help:"Enable profiling and write results to ./profiling.prof."`
	Standalone struct {
		File                string `help:"A mediawiki file tha should be rendered to an eBook." arg:""`
		OutputFile          string `help:"The path to the EPUB-file." short:"o" default:"ebook.epub"`
		CacheDir            string `help:"The directory where all cached files will be written to." default:".wiki2book"`
		StyleFile           string `help:"The CSS file that should be used." short:"s"`
		CoverImage          string `help:"A cover image for the front cover of the eBook." short:"c"`
		PandocDataDir       string `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p"`
		ForceRegenerateHtml bool   `help:"Forces wiki2book to recreate HTML files even if they exists from a previous run." short:"r"`
	} `cmd:"" help:"Renders a single mediawiki file into an eBook."`
	Project struct {
		ProjectFile         string `help:"A project JSON-file tha should be used to create an eBook." type:"existingfile:" arg:""`
		ForceRegenerateHtml bool   `help:"Forces wiki2book to recreate HTML files even if they exists from a previous run." short:"r"`
	} `cmd:"" help:"Uses a project file to create the eBook."`
	Article struct {
		ArticleName         string `help:"The name of the article to render." arg:""`
		OutputFile          string `help:"The path to the EPUB-file." short:"o" default:"ebook.epub"`
		CacheDir            string `help:"The directory where all cached files will be written to." default:".wiki2book"`
		StyleFile           string `help:"The CSS file that should be used." short:"s"`
		CoverImage          string `help:"A cover image for the front cover of the eBook." short:"c"`
		PandocDataDir       string `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p"`
		WikipediaInstance   string `help:"The Wikipedia-server that should be used. For example 'en' for en.wikipedia.org or 'de' for de.wikipedia.org." short:"i" default:"de"`
		ForceRegenerateHtml bool   `help:"Forces wiki2book to recreate HTML files even if they exists from a previous run." short:"r"`
	} `cmd:"" help:"Renders a single article into an eBook."`
}

func main() {
	ctx := kong.Parse(&cli)

	if cli.Debug {
		sigolo.LogLevel = sigolo.LOG_DEBUG
	}

	if cli.Profiling {
		f, err := os.Create("profiling.prof")
		sigolo.FatalCheck(err)

		err = pprof.StartCPUProfile(f)
		sigolo.FatalCheck(err)
		defer pprof.StopCPUProfile()
	}

	start := time.Now()

	switch ctx.Command() {
	case "standalone <file>":
		util.AssertFileExists(cli.Standalone.StyleFile)
		util.AssertFileExists(cli.Standalone.CoverImage)
		generateStandaloneEbook(cli.Standalone.File, cli.Standalone.OutputFile, cli.Standalone.CacheDir, cli.Standalone.StyleFile, cli.Standalone.CoverImage, cli.Standalone.PandocDataDir, cli.Standalone.ForceRegenerateHtml)
	case "project <project-file>":
		generateProjectEbook(cli.Project.ProjectFile, cli.Project.ForceRegenerateHtml)
	case "article <article-name>":
		generateArticleEbook(cli.Article.ArticleName, cli.Article.OutputFile, cli.Article.CacheDir, cli.Article.StyleFile, cli.Article.CoverImage, cli.Article.PandocDataDir, cli.Article.WikipediaInstance, cli.Article.ForceRegenerateHtml)
	default:
		sigolo.Fatal("Unknown command: %v\n%#v", ctx.Command(), ctx)
	}

	end := time.Now()
	sigolo.Debug("Start   : %s", start.Format(time.RFC1123))
	sigolo.Debug("End     : %s", end.Format(time.RFC1123))
	sigolo.Debug("Duration: %f seconds", end.Sub(start).Seconds())
}

func generateProjectEbook(projectFile string, forceHtmlRecreate bool) {
	var err error

	sigolo.Info("Use project file: %s", projectFile)

	directory, projectFile := filepath.Split(projectFile)
	sigolo.Debug("Go into folder %s", directory)
	err = os.Chdir(directory)
	sigolo.FatalCheck(err)

	project, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	articles := project.Articles
	wikipediaDomain := project.Domain
	cacheDir := project.CacheDir
	styleFile := project.Style
	coverFile := project.Cover
	metadata := project.Metadata
	outputFile := project.OutputFile
	pandocDataDir := project.PandocDataDir

	generateEpubFromArticles(articles, wikipediaDomain, cacheDir, styleFile, outputFile, coverFile, pandocDataDir, metadata, forceHtmlRecreate)
}

func generateStandaloneEbook(inputFile string, outputFile string, cacheDir string, styleFile string, coverImageFile string, pandocDataDir string, forceHtmlRecreate bool) {
	var err error

	imageCache := "images"
	mathCache := "math"
	templateCache := "templates"
	articleCache := "articles"
	htmlOutputFolder := "./"

	_, inputFileName := path.Split(inputFile)
	title := strings.Split(inputFileName, ".")[0]

	fileContent, err := ioutil.ReadFile(inputFile)
	sigolo.FatalCheck(err)

	file, err := os.Open(outputFile)
	sigolo.FatalCheck(err)

	// Assign default EPUB file name if given path is a directory
	fileInfo, err := file.Stat()
	sigolo.FatalCheck(err)

	if fileInfo.IsDir() {
		outputFile = path.Join(outputFile, "standalone.epub")
	}

	// Make all relevant paths absolute
	styleFile, outputFile, coverImageFile, pandocDataDir, err = toAbsolute(styleFile, outputFile, coverImageFile, pandocDataDir)

	// Create cache dir and go into it
	err = os.MkdirAll(cacheDir, os.ModePerm)
	sigolo.FatalCheck(err)

	err = os.Chdir(cacheDir)
	sigolo.FatalCheck(err)

	// Make all relevant paths relative again. This ensures that the locations within the HTML files are independent
	// of the systems' directory structure.
	styleFile, outputFile, coverImageFile, err = toRelative(styleFile, outputFile, coverImageFile)

	tokenizer := parser.NewTokenizer(imageCache, templateCache)
	article := tokenizer.Tokenize(string(fileContent), title)

	err = api.DownloadImages(article.Images, imageCache, articleCache)
	sigolo.FatalCheck(err)

	htmlFileName := article.Title + ".html"
	htmlFile := path.Join(htmlOutputFolder, htmlFileName)
	if shouldRecreateHtml(htmlOutputFolder, htmlFileName, forceHtmlRecreate) {
		htmlGenerator := &html.HtmlGenerator{}
		htmlFile, err = htmlGenerator.Generate(article, htmlOutputFolder, styleFile, imageCache, mathCache, articleCache)
		sigolo.FatalCheck(err)
	}

	sigolo.Info("Start generating EPUB file")
	metadata := project.Metadata{
		Title: title,
	}

	err = epub.Generate([]string{htmlFile}, outputFile, styleFile, coverImageFile, pandocDataDir, metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}

func generateArticleEbook(articleName string, outputFile string, cacheDir string, styleFile string, coverImageFile string, pandocDataDir string, instance string, forceHtmlRecreate bool) {
	//var err error
	// Enable this to create a profiling file. Then use the command "go tool pprof src ./profiling.prof" and enter "web" to open a diagram in your browser.
	//f, err := os.Create("profiling.prof")
	//sigolo.FatalCheck(err)
	//
	//err = pprof.StartCPUProfile(f)
	//sigolo.FatalCheck(err)
	//defer pprof.StopCPUProfile()

	var articles []string
	articles = append(articles, articleName)

	generateEpubFromArticles(articles,
		instance,
		cacheDir,
		styleFile,
		outputFile,
		coverImageFile,
		pandocDataDir,
		project.Metadata{},
		forceHtmlRecreate)
}

func generateEpubFromArticles(articles []string, wikipediaDomain string, cacheDir string, styleFile string, outputFile string, coverFile string, pandocDataDir string, metadata project.Metadata, forceHtmlRecreate bool) {
	var articleFiles []string
	var err error

	// Make all relevant paths absolute
	styleFile, outputFile, coverFile, pandocDataDir, err = toAbsolute(styleFile, outputFile, coverFile, pandocDataDir)

	// Create cache dir and go into it
	err = os.MkdirAll(cacheDir, os.ModePerm)
	sigolo.FatalCheck(err)

	err = os.Chdir(cacheDir)
	sigolo.FatalCheck(err)

	// Make all relevant paths relative again. This ensures that the locations within the HTML files are independent
	// of the systems' directory structure.
	styleFile, outputFile, coverFile, err = toRelative(styleFile, outputFile, coverFile)

	imageCache := "images"
	mathCache := "math"
	templateCache := "templates"
	articleCache := "articles"

	htmlOutputFolder := "./"
	for _, articleName := range articles {
		sigolo.Info("Start processing article %s", articleName)

		htmlFileName := articleName + ".html"
		if !shouldRecreateHtml(htmlOutputFolder, htmlFileName, forceHtmlRecreate) {
			sigolo.Info("HTML for article %s does already exist. Skip parsing and HTML generation.", articleName)
		} else {
			wikiArticleDto, err := api.DownloadArticle(wikipediaDomain, articleName, articleCache)
			sigolo.FatalCheck(err)

			tokenizer := parser.NewTokenizer(imageCache, templateCache)
			article := tokenizer.Tokenize(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.Title)

			err = api.DownloadImages(article.Images, imageCache, articleCache)
			sigolo.FatalCheck(err)

			htmlGenerator := &html.HtmlGenerator{}
			htmlFileName, err = htmlGenerator.Generate(article, htmlOutputFolder, styleFile, imageCache, mathCache, articleCache)
			sigolo.FatalCheck(err)

			sigolo.Info("Succeesfully created HTML for article %s", articleName)
		}

		articleFiles = append(articleFiles, htmlFileName)
	}

	sigolo.Info("Start generating EPUB file")
	err = epub.Generate(articleFiles, outputFile, styleFile, coverFile, pandocDataDir, metadata)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file")
}

func shouldRecreateHtml(htmlOutputFolder string, htmlFileName string, forceHtmlRecreate bool) bool {
	if forceHtmlRecreate {
		return true
	}

	// Check if HTML file already exists. If so, no recreate is wanted.
	htmlFilePath := filepath.Join(htmlOutputFolder, htmlFileName)
	_, err := os.Stat(htmlFilePath)
	htmlFileExists := err == nil

	return !htmlFileExists
}

func toRelative(styleFile string, outputFile string, coverFile string) (string, string, string, error) {
	var err error

	styleFile, err = util.MakePathRelative(styleFile)
	sigolo.FatalCheck(err)

	outputFile, err = util.MakePathRelative(outputFile)
	sigolo.FatalCheck(err)

	coverFile, err = util.MakePathRelative(coverFile)
	sigolo.FatalCheck(err)

	return styleFile, outputFile, coverFile, err
}

func toAbsolute(styleFile string, outputFile string, coverFile string, pandocDir string) (string, string, string, string, error) {
	var err error

	styleFile, err = util.MakePathAbsolute(styleFile, err)
	sigolo.FatalCheck(err)

	outputFile, err = util.MakePathAbsolute(outputFile, err)
	sigolo.FatalCheck(err)

	coverFile, err = util.MakePathAbsolute(coverFile, err)
	sigolo.FatalCheck(err)

	pandocDir, err = util.MakePathAbsolute(pandocDir, err)
	sigolo.FatalCheck(err)

	return styleFile, outputFile, coverFile, pandocDir, err
}
