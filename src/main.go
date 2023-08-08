package main

import (
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"
	"wiki2book/api"
	"wiki2book/config"
	"wiki2book/generator/epub"
	"wiki2book/generator/html"
	"wiki2book/parser"
	"wiki2book/project"
	"wiki2book/util"
)

var cli struct {
	Debug               bool   `help:"Enable debug mode." short:"d"`
	Profiling           bool   `help:"Enable profiling and write results to ./profiling.prof."`
	ForceRegenerateHtml bool   `help:"Forces wiki2book to recreate HTML files even if they exists from a previous run." short:"r"`
	SvgSizeToViewbox    bool   `help:"Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers."`
	Config              string `help:"The path to the overall application config. If not specified, default values are used." type:"existingfile" short:"c" placeholder:"<file>"`
	Standalone          struct {
		File          string `help:"A mediawiki file tha should be rendered to an eBook." arg:""`
		OutputFile    string `help:"The path to the EPUB-file." short:"o" default:"ebook.epub" placeholder:"<file>"`
		CacheDir      string `help:"The directory where all cached files will be written to." default:".wiki2book" placeholder:"<dir>"`
		StyleFile     string `help:"The CSS file that should be used." short:"s" placeholder:"<file>"`
		CoverImage    string `help:"A cover image for the front cover of the eBook." short:"i" placeholder:"<file>"`
		PandocDataDir string `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p" placeholder:"<dir>"`
	} `cmd:"" help:"Renders a single mediawiki file into an eBook."`
	Project struct {
		ProjectFile string `help:"A project JSON-file tha should be used to create an eBook." type:"existingfile:" arg:"" placeholder:"<file>"`
	} `cmd:"" help:"Uses a project file to create the eBook."`
	Article struct {
		ArticleName   string `help:"The name of the article to render." arg:""`
		OutputFile    string `help:"The path to the EPUB-file." short:"o" default:"ebook.epub" placeholder:"<file>"`
		CacheDir      string `help:"The directory where all cached files will be written to." default:".wiki2book" placeholder:"<dir>"`
		StyleFile     string `help:"The CSS file that should be used." short:"s" placeholder:"<file>"`
		CoverImage    string `help:"A cover image for the front cover of the eBook." short:"i" placeholder:"<file>"`
		PandocDataDir string `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p" placeholder:"<dir>"`
	} `cmd:"" help:"Renders a single article into an eBook."`
}

func main() {
	ctx := kong.Parse(&cli)

	if cli.Debug {
		sigolo.LogLevel = sigolo.LOG_DEBUG
	}

	if cli.Config != "" {
		err := config.LoadConfig(cli.Config)
		sigolo.FatalCheck(err)
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
		generateStandaloneEbook(
			cli.Standalone.File,
			cli.Standalone.OutputFile,
			cli.Standalone.CacheDir,
			cli.Standalone.StyleFile,
			cli.Standalone.CoverImage,
			cli.Standalone.PandocDataDir,
			cli.ForceRegenerateHtml,
			cli.SvgSizeToViewbox,
		)
	case "project <project-file>":
		generateProjectEbook(cli.Project.ProjectFile, cli.ForceRegenerateHtml, cli.SvgSizeToViewbox)
	case "article <article-name>":
		generateArticleEbook(
			cli.Article.ArticleName,
			cli.Article.OutputFile,
			cli.Article.CacheDir,
			cli.Article.StyleFile,
			cli.Article.CoverImage,
			cli.Article.PandocDataDir,
			cli.ForceRegenerateHtml,
			cli.SvgSizeToViewbox,
		)
	default:
		sigolo.Fatal("Unknown command: %v\n%#v", ctx.Command(), ctx)
	}

	end := time.Now()
	sigolo.Debug("Start   : %s", start.Format(time.RFC1123))
	sigolo.Debug("End     : %s", end.Format(time.RFC1123))
	sigolo.Debug("Duration: %f seconds", end.Sub(start).Seconds())
}

func generateProjectEbook(projectFile string, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	var err error

	sigolo.Info("Use project file: %s", projectFile)

	directory, projectFile := filepath.Split(projectFile)
	sigolo.Debug("Go into folder %s", directory)
	err = os.Chdir(directory)
	sigolo.FatalCheck(err)

	project, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	articles := project.Articles
	cacheDir := project.CacheDir
	styleFile := project.Style
	coverFile := project.Cover
	metadata := project.Metadata
	outputFile := project.OutputFile
	pandocDataDir := project.PandocDataDir

	generateEpubFromArticles(articles, cacheDir, styleFile, outputFile, coverFile, pandocDataDir, metadata, forceHtmlRecreate, svgSizeToViewbox)
}

func generateStandaloneEbook(inputFile string, outputFile string, cacheDir string, styleFile string, coverImageFile string, pandocDataDir string, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	var err error

	imageCache := "images"
	mathCache := "math"
	templateCache := "templates"
	articleCache := "articles"
	htmlOutputFolder := "./"

	_, inputFileName := path.Split(inputFile)
	title := strings.Split(inputFileName, ".")[0]

	fileContent, err := os.ReadFile(inputFile)
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
	paths, err := util.ToAbsolute(styleFile, outputFile, coverImageFile, pandocDataDir)
	sigolo.FatalCheck(err)
	styleFile = paths[0]
	outputFile = paths[1]
	coverImageFile = paths[2]
	pandocDataDir = paths[3]

	// Create cache dir and go into it
	err = os.MkdirAll(cacheDir, os.ModePerm)
	sigolo.FatalCheck(err)

	err = os.Chdir(cacheDir)
	sigolo.FatalCheck(err)

	// Make all relevant paths relative again. This ensures that the locations within the HTML files are independent
	// of the systems' directory structure.
	paths, err = util.ToRelative(styleFile, outputFile, coverImageFile)
	sigolo.FatalCheck(err)
	styleFile = paths[0]
	outputFile = paths[1]
	coverImageFile = paths[2]

	tokenizer := parser.NewTokenizer(imageCache, templateCache)
	article := tokenizer.Tokenize(string(fileContent), title)

	err = api.DownloadImages(article.Images, imageCache, articleCache, svgSizeToViewbox)
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

	absoluteOutputFile, err := util.MakePathAbsolute(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file %s", absoluteOutputFile)
}

func generateArticleEbook(articleName string, outputFile string, cacheDir string, styleFile string, coverImageFile string, pandocDataDir string, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	//var err error
	// Enable this to create a profiling file. Then use the command "go tool pprof wiki2book ./profiling.prof" and enter "web" to open a diagram in your browser.
	//f, err := os.Create("profiling.prof")
	//sigolo.FatalCheck(err)
	//
	//err = pprof.StartCPUProfile(f)
	//sigolo.FatalCheck(err)
	//defer pprof.StopCPUProfile()

	var articles []string
	articles = append(articles, articleName)

	generateEpubFromArticles(articles,
		cacheDir,
		styleFile,
		outputFile,
		coverImageFile,
		pandocDataDir,
		project.Metadata{},
		forceHtmlRecreate,
		svgSizeToViewbox,
	)
}

func generateEpubFromArticles(articles []string, cacheDir string, styleFile string, outputFile string, coverImageFile string, pandocDataDir string, metadata project.Metadata, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	var articleFiles []string
	var err error

	// Make all relevant paths absolute
	paths, err := util.ToAbsolute(styleFile, outputFile, coverImageFile, pandocDataDir)
	sigolo.FatalCheck(err)
	styleFile = paths[0]
	outputFile = paths[1]
	coverImageFile = paths[2]
	pandocDataDir = paths[3]

	// Create cache dir and go into it
	err = os.MkdirAll(cacheDir, os.ModePerm)
	sigolo.FatalCheck(err)

	err = os.Chdir(cacheDir)
	sigolo.FatalCheck(err)

	// Make all relevant paths relative again. This ensures that the locations within the HTML files are independent
	// of the systems' directory structure.
	paths, err = util.ToRelative(styleFile, outputFile, coverImageFile)
	sigolo.FatalCheck(err)
	styleFile = paths[0]
	outputFile = paths[1]
	coverImageFile = paths[2]

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
			sigolo.Info("Download article %s", articleName)
			wikiArticleDto, err := api.DownloadArticle(config.Current.WikipediaInstance, articleName, articleCache)
			sigolo.FatalCheck(err)

			sigolo.Info("Tokenize article %s", articleName)
			tokenizer := parser.NewTokenizer(imageCache, templateCache)
			article := tokenizer.Tokenize(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.OriginalTitle)

			sigolo.Info("Download images from article %s", articleName)
			err = api.DownloadImages(article.Images, imageCache, articleCache, svgSizeToViewbox)
			sigolo.FatalCheck(err)

			sigolo.Info("Generate HTML for article %s", articleName)
			htmlGenerator := &html.HtmlGenerator{}
			htmlFileName, err = htmlGenerator.Generate(article, htmlOutputFolder, styleFile, imageCache, mathCache, articleCache)
			sigolo.FatalCheck(err)

		}

		sigolo.Info("Finished processing article %s", articleName)
		articleFiles = append(articleFiles, htmlFileName)
	}

	sigolo.Info("Start generating EPUB file")
	err = epub.Generate(articleFiles, outputFile, styleFile, coverImageFile, pandocDataDir, metadata)
	sigolo.FatalCheck(err)

	absoluteOutputFile, err := util.MakePathAbsolute(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file %s", absoluteOutputFile)
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
