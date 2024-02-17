package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
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

const VERSION = "v0.1.0"
const RFC1123Millis = "Mon, 02 Jan 2006 15:04:05.999 MST"

var cli struct {
	Logging              string      `help:"Logging verbosity. Possible values: debug, trace" short:"l"`
	DiagnosticsProfiling bool        `help:"Enable profiling and write results to ./profiling.prof."`
	DiagnosticsTrace     bool        `help:"Enable tracing to analyse memory usage and write results to ./trace.out."`
	ForceRegenerateHtml  bool        `help:"Forces wiki2book to recreate HTML files even if they exists from a previous run." short:"r"`
	SvgSizeToViewbox     bool        `help:"Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers."`
	Config               string      `help:"The path to the overall application config. If not specified, default values are used." type:"existingfile" short:"c" placeholder:"<file>"`
	Version              VersionFlag `help:"Print version information and quit" name:"version" short:"v"`
	Standalone           struct {
		File          string   `help:"A mediawiki file tha should be rendered to an eBook." arg:""`
		OutputFile    string   `help:"The path to the EPUB-file." short:"o" default:"ebook.epub" placeholder:"<file>"`
		OutputType    string   `help:"The EPUB type. Possible values are epub2 and epub3, see pandoc '-t' parameter." short:"t" default:"epub2" placeholder:"<type>"`
		CacheDir      string   `help:"The directory where all cached files will be written to." default:".wiki2book" placeholder:"<dir>"`
		StyleFile     string   `help:"The CSS file that should be used." short:"s" placeholder:"<file>"`
		CoverImage    string   `help:"A cover image for the front cover of the eBook." short:"i" placeholder:"<file>"`
		PandocDataDir string   `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p" placeholder:"<dir>"`
		FontFiles     []string `help:"A list of font files that should be used. They are references in your style file." short:"f" placeholder:"<file> ..."`
	} `cmd:"" help:"Renders a single mediawiki file into an eBook."`
	Project struct {
		ProjectFile string `help:"A project JSON-file tha should be used to create an eBook." type:"existingfile:" arg:"" placeholder:"<file>"`
	} `cmd:"" help:"Uses a project file to create the eBook."`
	Article struct {
		ArticleName   string   `help:"The name of the article to render." arg:""`
		OutputFile    string   `help:"The path to the EPUB-file." short:"o" default:"ebook.epub" placeholder:"<file>"`
		OutputType    string   `help:"The EPUB type. Possible values are epub2 and epub3, see pandoc '-t' parameter." short:"t" default:"epub2" placeholder:"<type>"`
		CacheDir      string   `help:"The directory where all cached files will be written to." default:".wiki2book" placeholder:"<dir>"`
		StyleFile     string   `help:"The CSS file that should be used." short:"s" placeholder:"<file>"`
		CoverImage    string   `help:"A cover image for the front cover of the eBook." short:"i" placeholder:"<file>"`
		PandocDataDir string   `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p" placeholder:"<dir>"`
		FontFiles     []string `help:"A list of font files that should be used. They are references in your style file." short:"f" placeholder:"<file>"`
	} `cmd:"" help:"Renders a single article into an eBook."`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Name("wiki2book"),
		kong.Description("A CLI tool to turn one or multiple Wikipedia articles into a good-looking eBook."),
		kong.Vars{
			"version": VERSION,
		},
	)

	if strings.ToLower(cli.Logging) == "debug" {
		sigolo.LogLevel = sigolo.LOG_DEBUG
	} else if strings.ToLower(cli.Logging) == "trace" {
		sigolo.LogLevel = sigolo.LOG_TRACE
	}

	sigolo.Trace("CLI config:\n%+v", cli)

	if cli.Config != "" {
		err := config.LoadConfig(cli.Config)
		sigolo.FatalCheck(err)
	}

	if cli.DiagnosticsProfiling {
		f, err := os.Create("profiling.prof")
		sigolo.FatalCheck(err)

		err = pprof.StartCPUProfile(f)
		sigolo.FatalCheck(err)
		defer pprof.StopCPUProfile()
	}

	if cli.DiagnosticsTrace {
		f, err := os.Create("trace.out")
		sigolo.FatalCheck(err)

		err = trace.Start(f)
		sigolo.FatalCheck(err)
		defer trace.Stop()
	}

	start := time.Now()

	switch ctx.Command() {
	case "standalone <file>":
		util.AssertFileExists(cli.Standalone.StyleFile)
		util.AssertFileExists(cli.Standalone.CoverImage)
		generateStandaloneEbook(
			cli.Standalone.File,
			cli.Standalone.OutputFile,
			cli.Standalone.OutputType,
			cli.Standalone.CacheDir,
			cli.Standalone.StyleFile,
			cli.Standalone.CoverImage,
			cli.Standalone.PandocDataDir,
			cli.Standalone.FontFiles,
			cli.ForceRegenerateHtml,
			cli.SvgSizeToViewbox,
		)
	case "project <project-file>":
		generateProjectEbook(cli.Project.ProjectFile, cli.ForceRegenerateHtml, cli.SvgSizeToViewbox)
	case "article <article-name>":
		generateArticleEbook(
			cli.Article.ArticleName,
			cli.Article.OutputFile,
			cli.Article.OutputType,
			cli.Article.CacheDir,
			cli.Article.StyleFile,
			cli.Article.CoverImage,
			cli.Article.PandocDataDir,
			cli.Article.FontFiles,
			cli.ForceRegenerateHtml,
			cli.SvgSizeToViewbox,
		)
	default:
		if sigolo.LogLevel > sigolo.LOG_DEBUG {
			sigolo.Trace("CLI config:\n%+v", cli)
		}
		sigolo.Fatal("Unknown command: %v", ctx.Command())
	}

	end := time.Now()
	sigolo.Debug("Start   : %s", start.Format(RFC1123Millis))
	sigolo.Debug("End     : %s", end.Format(RFC1123Millis))
	sigolo.Debug("Duration: %f seconds", end.Sub(start).Seconds())
}

func generateProjectEbook(projectFile string, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	var err error

	sigolo.Info("Use project file: %s", projectFile)

	directory, projectFile := filepath.Split(projectFile)
	sigolo.Debug("Go into folder %s", directory)
	err = os.Chdir(directory)
	sigolo.FatalCheck(err)

	proj, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	generateEpubFromArticles(proj, forceHtmlRecreate, svgSizeToViewbox)
}

func generateStandaloneEbook(inputFile string, outputFile string, outputType string, cacheDir string, styleFile string, coverImageFile string, pandocDataDir string, fontFiles []string, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	var err error

	imageCache := "images"
	mathCache := "math"
	templateCache := "templates"
	articleCache := "articles"
	htmlOutputFolder := "html"

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
	article, err := tokenizer.Tokenize(string(fileContent), title)
	sigolo.FatalCheck(err)

	err = api.DownloadImages(article.Images, imageCache, articleCache, svgSizeToViewbox)
	sigolo.FatalCheck(err)

	htmlFilePath := path.Join(htmlOutputFolder, article.Title+".html")
	if shouldRecreateHtml(htmlFilePath, forceHtmlRecreate) {
		htmlGenerator := &html.HtmlGenerator{
			ImageCacheFolder:   imageCache,
			MathCacheFolder:    mathCache,
			ArticleCacheFolder: articleCache,
			TokenMap:           article.TokenMap,
		}
		htmlFilePath, err = htmlGenerator.Generate(article, htmlOutputFolder, styleFile)
		sigolo.FatalCheck(err)
	}

	sigolo.Info("Start generating EPUB file")
	metadata := project.Metadata{
		Title: title,
	}

	err = epub.Generate([]string{htmlFilePath}, outputFile, outputType, styleFile, coverImageFile, pandocDataDir, fontFiles, metadata)
	sigolo.FatalCheck(err)

	absoluteOutputFile, err := util.MakePathAbsolute(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file %s", absoluteOutputFile)
}

func generateArticleEbook(articleName string, outputFile string, outputType string, cacheDir string, styleFile string, coverImageFile string, pandocDataDir string, fontFiles []string, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	var articles []string
	articles = append(articles, articleName)

	proj := project.NewWithDefaults()
	proj.Metadata = project.Metadata{}
	proj.OutputFile = outputFile
	proj.OutputType = outputType
	proj.CacheDir = cacheDir
	proj.Cover = coverImageFile
	proj.Style = styleFile
	proj.PandocDataDir = pandocDataDir
	proj.Articles = articles
	proj.FontFiles = fontFiles

	generateEpubFromArticles(
		proj,
		forceHtmlRecreate,
		svgSizeToViewbox,
	)
}

func generateEpubFromArticles(project *project.Project, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	var articleFiles []string
	var err error

	articles := project.Articles
	cacheDir := project.CacheDir
	styleFile := project.Style
	coverImageFile := project.Cover
	metadata := project.Metadata
	outputFile := project.OutputFile
	//outputType := project.OutputType
	pandocDataDir := project.PandocDataDir
	fontFiles := project.FontFiles

	imageCache := "images"
	mathCache := "math"
	templateCache := "templates"
	articleCache := "articles"
	htmlOutputFolder := "html"

	// Make all relevant paths absolute
	paths, err := util.ToAbsolute(styleFile, outputFile, coverImageFile, pandocDataDir)
	sigolo.FatalCheck(err)
	styleFile = paths[0]
	outputFile = paths[1]
	coverImageFile = paths[2]
	pandocDataDir = paths[3]

	// Create cache dir and go into it
	sigolo.Debug("Ensure cache folder '%s'", cacheDir)
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

	var images []string

	for _, articleName := range articles {
		sigolo.Info("Article '%s': Start processing", articleName)

		htmlFilePath := filepath.Join(htmlOutputFolder, articleName+".html")
		if !shouldRecreateHtml(htmlFilePath, forceHtmlRecreate) {
			sigolo.Info("Article '%s': HTML for article does already exist. Skip parsing and HTML generation.", articleName)
		} else {
			sigolo.Info("Article '%s': Download article", articleName)
			wikiArticleDto, err := api.DownloadArticle(config.Current.WikipediaInstance, articleName, articleCache)
			sigolo.FatalCheck(err)

			sigolo.Info("Article '%s': Tokenize content", articleName)
			tokenizer := parser.NewTokenizer(imageCache, templateCache)
			article, err := tokenizer.Tokenize(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.OriginalTitle)
			sigolo.FatalCheck(err)
			images = append(images, article.Images...)

			sigolo.Info("Article '%s': Download images", articleName)
			err = api.DownloadImages(article.Images, imageCache, articleCache, svgSizeToViewbox)
			sigolo.FatalCheck(err)

			sigolo.Info("Article '%s': Generate HTML", articleName)
			htmlGenerator := &html.HtmlGenerator{
				ImageCacheFolder:   imageCache,
				MathCacheFolder:    mathCache,
				ArticleCacheFolder: articleCache,
				TokenMap:           article.TokenMap,
			}
			htmlFilePath, err = htmlGenerator.Generate(article, htmlOutputFolder, "../css/style.css")
			sigolo.FatalCheck(err)
		}

		sigolo.Info("Article '%s': Finished processing", articleName)
		articleFiles = append(articleFiles, htmlFilePath)
	}

	images = util.RemoveDuplicates(images)

	sigolo.Info("Start generating EPUB file")
	// TODO Add switch in CLI/config to decide between pandoc and internal library
	//err = epub.Generate(articleFiles, outputFile, outputType, styleFile, coverImageFile, pandocDataDir, fontFiles, metadata)
	err = epub.GenerateWithGoLibrary(articleFiles, outputFile, coverImageFile, styleFile, fontFiles, metadata)
	sigolo.FatalCheck(err)

	absoluteOutputFile, err := util.MakePathAbsolute(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Info("Successfully created EPUB file %s", absoluteOutputFile)
}

func shouldRecreateHtml(htmlFilePath string, forceHtmlRecreate bool) bool {
	if forceHtmlRecreate {
		return true
	}

	// Check if HTML file already exists. If so, no recreate is wanted.
	_, err := os.Stat(htmlFilePath)
	htmlFileExists := err == nil

	return !htmlFileExists
}
