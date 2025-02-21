package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"time"
	"wiki2book/api"
	"wiki2book/config"
	"wiki2book/generator"
	"wiki2book/generator/epub"
	"wiki2book/generator/html"
	"wiki2book/parser"
	"wiki2book/project"
	"wiki2book/util"
)

const VERSION = "v0.2.0"
const RFC1123Millis = "Mon, 02 Jan 2006 15:04:05.999 MST"

type Cli struct {
	Version VersionFlag `help:"Print version information and quit" name:"version" short:"v"`

	Config  string `help:"The path to the overall application config. If not specified, default values are used." type:"existingfile" short:"c" placeholder:"<file>"`
	Logging string `help:"Logging verbosity. Possible values: \"info\" (default), \"debug\", \"trace\"." short:"l" default:"info"`

	DiagnosticsProfiling bool `help:"Enable profiling and write results to ./profiling.prof."`
	DiagnosticsTrace     bool `help:"Enable tracing to analyse memory usage and write results to ./trace.out."`

	OutputFile string `help:"The path to the output file." short:"o" default:"ebook.epub" placeholder:"<file>"`

	// Can be set via and override config file:
	ForceRegenerateHtml            bool     `help:"Forces wiki2book to recreate HTML files even if they exists from a previous run." short:"r"`
	SvgSizeToViewbox               bool     `help:"Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers."`
	OutputType                     string   `help:"The output file type. Possible values are: \"epub2\" (default), \"epub3\"." short:"t" placeholder:"<type>"`
	OutputDriver                   string   `help:"The method to generate the output file. Available driver: \"pandoc\" (default), \"internal\" (experimental!)" short:"d" placeholder:"<driver>"`
	CacheDir                       string   `help:"The directory where all cached files will be written to." placeholder:"<dir>"`
	StyleFile                      string   `help:"The CSS file that should be used." short:"s" placeholder:"<file>"`
	CoverImage                     string   `help:"A cover image for the front cover of the eBook." short:"i" placeholder:"<file>"`
	PandocDataDir                  string   `help:"The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation." short:"p" placeholder:"<dir>"`
	FontFiles                      []string `help:"A list of font files that should be used. They are references in your style file." short:"f" placeholder:"<file>"`
	ImagesToGrayscale              bool     `help:"Set to true in order to convert raster images to grayscale." short:"g"`
	IgnoredTemplates               []string `help:"List of templates that should be ignored and removed from the input wikitext. The list must be in lower case."`
	TrailingTemplates              []string `help:"List of templates that will be moved to the end of the document."`
	IgnoredImageParams             []string `help:"Parameters of images that should be ignored. The list must be in lower case."`
	IgnoredMediaTypes              []string `help:"List of media types to ignore, i.e. list of file extensions."`
	WikipediaInstance              string   `help:"The subdomain of the Wikipedia instance."`
	WikipediaHost                  string   `help:"The domain of the Wikipedia instance."`
	WikipediaImageHost             string   `help:"The domain of the Wikipedia image instance."`
	WikipediaMathRestApi           string   `help:"The URL to the math API of wikipedia."`
	WikipediaImageArticleInstances []string `help:"Wikipedia instances (subdomains) of the wikipedia image host where images should be searched for."`
	FilePrefixe                    []string `help:"A list of prefixes to detect files, e.g. in \"File:picture.jpg\" the substring \"File\" is the image prefix."`
	AllowedLinkPrefixes            []string `help:"A list of prefixes that are considered links and are therefore not removed."`
	CategoryPrefixes               []string `help:"A list of category prefixes, which are technically internals links."`
	MathConverter                  string   `help:"Converter turning math SVGs into PNGs."`
	RsvgMathStylesheet             string   `help:"Stylesheet for rsvg-convert when using the rsvg converter for math SVGs." placeholder:"<file>"`

	Standalone struct {
		File string `help:"A mediawiki file tha should be rendered to an eBook." arg:""`
	} `cmd:"" help:"Renders a single mediawiki file into an eBook."`
	Project struct {
		ProjectFile string `help:"A project JSON-file tha should be used to create an eBook." type:"existingfile:" arg:"" placeholder:"<file>"`
	} `cmd:"" help:"Uses a project file to create the eBook."`
	Article struct {
		ArticleName string `help:"The name of the article to render." arg:""`
	} `cmd:"" help:"Renders a single article into an eBook."`
}

var cli Cli

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
		sigolo.SetDefaultLogLevel(sigolo.LOG_DEBUG)
	} else if strings.ToLower(cli.Logging) == "trace" {
		sigolo.SetDefaultLogLevel(sigolo.LOG_TRACE)
	} else if strings.ToLower(cli.Logging) == "info" {
		sigolo.SetDefaultLogLevel(sigolo.LOG_INFO)
		sigolo.SetDefaultFormatFunctionAll(sigolo.LogPlain)
	} else {
		sigolo.SetDefaultFormatFunctionAll(sigolo.LogPlain)
		sigolo.Fatalf("Unknown logging level '%s'", cli.Logging)
	}

	sigolo.Tracef("CLI config:\n%+v", cli)

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
		mergeCliParametersIntoConfig(&cli)
		generateStandaloneEbook(
			cli.Standalone.File,
			cli.OutputFile,
		)
	case "project <project-file>":
		// TODO First merge project into config and then cli args
		mergeCliParametersIntoConfig(&cli)
		generateProjectEbook(
			cli.Project.ProjectFile,
			cli.OutputFile,
			cli.OutputType,
			cli.OutputDriver,
			cli.CacheDir,
			cli.StyleFile,
			cli.CoverImage,
			cli.PandocDataDir,
			cli.FontFiles,
			cli.ImagesToGrayscale,
			cli.ForceRegenerateHtml,
			cli.SvgSizeToViewbox,
			cli.MathConverter,
			cli.RsvgMathStylesheet,
		)
	case "article <article-name>":
		mergeCliParametersIntoConfig(&cli)
		generateArticleEbook(
			cli.Article.ArticleName,
			cli.OutputFile,
			cli.OutputType,
			cli.OutputDriver,
			cli.CacheDir,
			cli.StyleFile,
			cli.CoverImage,
			cli.PandocDataDir,
			cli.FontFiles,
			cli.ImagesToGrayscale,
			cli.ForceRegenerateHtml,
			cli.SvgSizeToViewbox,
			cli.MathConverter,
			cli.RsvgMathStylesheet,
		)
	default:
		if sigolo.GetCurrentLogLevel() > sigolo.LOG_DEBUG {
			sigolo.Tracef("CLI config:\n%+v", cli)
		}
		sigolo.Fatalf("Unknown command: %v", ctx.Command())
	}

	end := time.Now()
	sigolo.Debugf("Start   : %s", start.Format(RFC1123Millis))
	sigolo.Debugf("End     : %s", end.Format(RFC1123Millis))
	sigolo.Debugf("Duration: %f seconds", end.Sub(start).Seconds())
}

func mergeCliParametersIntoConfig(cli *Cli) {
	if cli.ForceRegenerateHtml {
		config.Current.ForceRegenerateHtml = cli.ForceRegenerateHtml
	}
	if cli.SvgSizeToViewbox {
		config.Current.SvgSizeToViewbox = cli.SvgSizeToViewbox
	}
	if cli.OutputType != "" {
		config.Current.OutputType = cli.OutputType
	}
	if cli.OutputDriver != "" {
		config.Current.OutputDriver = cli.OutputDriver
	}
	if cli.CacheDir != "" {
		absolutePath, err := util.ToAbsolutePath(cli.CacheDir)
		sigolo.FatalCheck(err)
		config.Current.CacheDir = absolutePath
	}
	if cli.StyleFile != "" {
		absolutePath, err := util.ToAbsolutePath(cli.StyleFile)
		sigolo.FatalCheck(err)
		config.Current.StyleFile = absolutePath
	}
	if cli.CoverImage != "" {
		absolutePath, err := util.ToAbsolutePath(cli.CoverImage)
		sigolo.FatalCheck(err)
		config.Current.CoverImage = absolutePath
	}
	if cli.PandocDataDir != "" {
		absolutePath, err := util.ToAbsolutePath(cli.PandocDataDir)
		sigolo.FatalCheck(err)
		config.Current.PandocDataDir = absolutePath
	}
	if cli.FontFiles != nil {
		absolutePaths, err := util.ToAbsolutePaths(cli.FontFiles...)
		sigolo.FatalCheck(err)
		config.Current.FontFiles = absolutePaths
	}
	if cli.ImagesToGrayscale {
		config.Current.ImagesToGrayscale = cli.ImagesToGrayscale
	}
	if cli.IgnoredTemplates != nil {
		config.Current.IgnoredTemplates = cli.IgnoredTemplates
	}
	if cli.TrailingTemplates != nil {
		config.Current.TrailingTemplates = cli.TrailingTemplates
	}
	if cli.IgnoredImageParams != nil {
		config.Current.IgnoredImageParams = cli.IgnoredImageParams
	}
	if cli.IgnoredMediaTypes != nil {
		config.Current.IgnoredMediaTypes = cli.IgnoredMediaTypes
	}
	if cli.WikipediaInstance != "" {
		config.Current.WikipediaInstance = cli.WikipediaInstance
	}
	if cli.WikipediaHost != "" {
		config.Current.WikipediaHost = cli.WikipediaHost
	}
	if cli.WikipediaImageHost != "" {
		config.Current.WikipediaImageHost = cli.WikipediaImageHost
	}
	if cli.WikipediaMathRestApi != "" {
		config.Current.WikipediaMathRestApi = cli.WikipediaMathRestApi
	}
	if cli.WikipediaImageArticleInstances != nil {
		config.Current.WikipediaImageArticleInstances = cli.WikipediaImageArticleInstances
	}
	if cli.FilePrefixe != nil {
		config.Current.FilePrefixe = cli.FilePrefixe
	}
	if cli.AllowedLinkPrefixes != nil {
		config.Current.AllowedLinkPrefixes = cli.AllowedLinkPrefixes
	}
	if cli.CategoryPrefixes != nil {
		config.Current.CategoryPrefixes = cli.CategoryPrefixes
	}
	if cli.MathConverter != "" {
		config.Current.MathConverter = cli.MathConverter
	}
	if cli.RsvgMathStylesheet != "" {
		absolutePath, err := util.ToAbsolutePath(cli.RsvgMathStylesheet)
		sigolo.FatalCheck(err)
		config.Current.RsvgMathStylesheet = absolutePath
	}

	config.Current.AssertValidity()
}

func generateProjectEbook(projectFile string, outputFile string, outputType string, outputDriver string, cacheDir string, styleFile string, coverImageFile string, pandocDataDir string, fontFiles []string, imagesToGrayscale bool, forceHtmlRecreate bool, svgSizeToViewbox bool, mathconverter string, rsvgMathStylesheet string) {
	var err error

	sigolo.Infof("Use project file: %s", projectFile)

	sigolo.Debug("Turn paths from CLI arguments into absolute paths before going into the project file directory")
	if outputFile != "" {
		outputFile, err = util.ToAbsolutePath(outputFile)
		sigolo.FatalCheck(err)
	}
	if cacheDir != "" {
		cacheDir, err = util.ToAbsolutePath(cacheDir)
		sigolo.FatalCheck(err)
	}
	if styleFile != "" {
		styleFile, err = util.ToAbsolutePath(styleFile)
		sigolo.FatalCheck(err)
	}
	if coverImageFile != "" {
		coverImageFile, err = util.ToAbsolutePath(coverImageFile)
		sigolo.FatalCheck(err)
	}
	if pandocDataDir != "" {
		pandocDataDir, err = util.ToAbsolutePath(pandocDataDir)
		sigolo.FatalCheck(err)
	}
	if fontFiles != nil && len(fontFiles) > 0 {
		fontFiles, err = util.ToAbsolutePaths(fontFiles...)
		sigolo.FatalCheck(err)
	}

	directory, projectFile := filepath.Split(projectFile)
	if directory != "" {
		sigolo.Debugf("Go into folder %s", directory)
		err = os.Chdir(directory)
		sigolo.FatalCheck(err)
	}

	proj, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	if outputFile != "" {
		sigolo.Tracef("Override outputFile from project file with %s", outputFile)
		proj.OutputFile = outputFile
	}
	if outputType != "" {
		sigolo.Tracef("Override outputType from project file with %s", outputType)
		proj.OutputType = outputType
	}
	if outputDriver != "" {
		sigolo.Tracef("Override outputDriver from project file with %s", outputDriver)
		proj.OutputDriver = outputDriver
	}
	if cacheDir != "" {
		sigolo.Tracef("Override cacheDir from project file with %s", cacheDir)
		proj.CacheDir = cacheDir
	}
	if styleFile != "" {
		sigolo.Tracef("Override styleFile from project file with %s", styleFile)
		proj.StyleFile = styleFile
	}
	if coverImageFile != "" {
		sigolo.Tracef("Override coverImageFile from project file with %s", coverImageFile)
		proj.CoverImage = coverImageFile
	}
	if pandocDataDir != "" {
		sigolo.Tracef("Override pandocDataDir from project file with %s", pandocDataDir)
		proj.PandocDataDir = pandocDataDir
	}
	if fontFiles != nil && len(fontFiles) > 0 {
		sigolo.Tracef("Override fontFiles from project file with %v", fontFiles)
		proj.FontFiles = fontFiles
	}
	if imagesToGrayscale {
		sigolo.Tracef("Override imagesToGrayscale from project file with %v", imagesToGrayscale)
		proj.ImagesToGrayscale = imagesToGrayscale
	}
	if mathconverter != "" {
		sigolo.Tracef("Override mathConverter from project file with %s", mathconverter)
		proj.MathConverter = mathconverter
	}
	if rsvgMathStylesheet != "" {
		sigolo.Tracef("Override rsvgMathStylesheet from project file with %s", rsvgMathStylesheet)
		proj.RsvgMathStylesheet = rsvgMathStylesheet
	}

	generateBookFromArticles(proj, forceHtmlRecreate, svgSizeToViewbox)
}

func generateStandaloneEbook(inputFile string, outputFile string) {
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

	// Assign default output file name if given path is a directory
	fileInfo, err := file.Stat()
	sigolo.FatalCheck(err)

	if fileInfo.IsDir() {
		// TODO Adjust this when additional non-epub output types are supported.
		outputFile = path.Join(outputFile, "standalone.epub")
	}

	// Make all relevant paths absolute
	absolutePath, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	outputFile = absolutePath

	// Create cache dir and go into it
	ensureDirectory(config.Current.CacheDir)
	err = os.Chdir(config.Current.CacheDir)
	sigolo.FatalCheck(err)

	ensureDirectory(util.TempDirName)

	// Make all relevant paths relative again. This ensures that the locations within the HTML files are independent
	// of the systems' directory structure.
	relativeStyleFile, err := util.ToRelativePath(config.Current.StyleFile)
	sigolo.FatalCheck(err)

	tokenizer := parser.NewTokenizer(imageCache, templateCache)
	article, err := tokenizer.Tokenize(string(fileContent), title)
	sigolo.FatalCheck(err)

	err = api.DownloadImages(article.Images, imageCache, articleCache, config.Current.SvgSizeToViewbox, config.Current.ImagesToGrayscale)
	sigolo.FatalCheck(err)

	// TODO Adjust this when additional non-epub output types are supported.
	htmlFilePath := path.Join(htmlOutputFolder, article.Title+".html")
	if shouldRecreateHtml(htmlFilePath, config.Current.ForceRegenerateHtml) {
		htmlGenerator := &html.HtmlGenerator{
			ImageCacheFolder:   imageCache,
			MathCacheFolder:    mathCache,
			ArticleCacheFolder: articleCache,
			TokenMap:           article.TokenMap,
		}
		htmlFilePath, err = htmlGenerator.Generate(article, htmlOutputFolder, relativeStyleFile)
		sigolo.FatalCheck(err)
	}

	sigolo.Infof("Start generating %s file", config.Current.OutputType)
	metadata := project.Metadata{
		Title: title,
	}

	err = Generate(config.Current.OutputDriver, []string{htmlFilePath}, outputFile, config.Current.OutputType, config.Current.StyleFile, config.Current.CoverImage, config.Current.PandocDataDir, config.Current.FontFiles, metadata)
	sigolo.FatalCheck(err)

	err = os.RemoveAll(util.TempDirName)
	if err != nil {
		sigolo.Warnf("Error cleaning up '%s' directory", util.TempDirName)
	}

	absoluteOutputFile, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Infof("Successfully created %s file %s", config.Current.OutputType, absoluteOutputFile)
}

func generateArticleEbook(articleName string, outputFile string, outputType string, outputDriver string, cacheDir string, styleFile string, coverImageFile string, pandocDataDir string, fontFiles []string, imagesToGrayscale bool, forceHtmlRecreate bool, svgSizeToViewbox bool, mathconverter string, rsvgMathStylesheet string) {
	var articles []string
	articles = append(articles, articleName)

	proj := project.NewWithDefaults()
	proj.Metadata = project.Metadata{}
	proj.OutputFile = outputFile
	proj.OutputType = outputType
	proj.OutputDriver = outputDriver
	proj.CacheDir = cacheDir
	proj.CoverImage = coverImageFile
	proj.StyleFile = styleFile
	proj.PandocDataDir = pandocDataDir
	proj.Articles = articles
	proj.FontFiles = fontFiles
	proj.ImagesToGrayscale = imagesToGrayscale
	proj.MathConverter = mathconverter
	proj.RsvgMathStylesheet = rsvgMathStylesheet

	generateBookFromArticles(
		proj,
		forceHtmlRecreate,
		svgSizeToViewbox,
	)
}

func generateBookFromArticles(project *project.Project, forceHtmlRecreate bool, svgSizeToViewbox bool) {
	var articleFiles []string
	var err error

	articles := project.Articles
	cacheDir := project.CacheDir
	styleFile := project.StyleFile
	coverImageFile := project.CoverImage
	metadata := project.Metadata
	outputFile := project.OutputFile
	outputType := project.OutputType
	outputDriver := project.OutputDriver
	pandocDataDir := project.PandocDataDir
	fontFiles := project.FontFiles
	imagesToGrayscale := project.ImagesToGrayscale

	imageCache := "images"
	mathCache := "math"
	templateCache := "templates"
	articleCache := "articles"
	htmlOutputFolder := "html"

	util.AssertPathExists(styleFile)
	util.AssertPathExists(coverImageFile)

	err = generator.VerifyOutputAndDriver(outputType, outputDriver)
	sigolo.FatalCheck(err)

	// Make all relevant paths absolute
	paths, err := util.ToAbsolutePaths(styleFile, outputFile, coverImageFile, pandocDataDir)
	sigolo.FatalCheck(err)
	styleFile = paths[0]
	outputFile = paths[1]
	coverImageFile = paths[2]
	pandocDataDir = paths[3]

	if project.RsvgMathStylesheet != "" {
		sigolo.Tracef("Override rsvgMathStylesheet from config file with %s", project.RsvgMathStylesheet)
		config.Current.RsvgMathStylesheet = project.RsvgMathStylesheet
	}
	config.Current.RsvgMathStylesheet, err = util.ToAbsolutePath(config.Current.RsvgMathStylesheet)
	sigolo.FatalCheck(err)
	util.AssertPathExists(config.Current.RsvgMathStylesheet)

	// Create cache dir and go into it
	ensureDirectory(cacheDir)
	err = os.Chdir(cacheDir)
	sigolo.FatalCheck(err)

	ensureDirectory(util.TempDirName)

	// Make all relevant paths relative again. This ensures that the locations within the HTML files are independent
	// of the systems' directory structure.
	paths, err = util.ToRelativePaths(styleFile, outputFile, coverImageFile)
	sigolo.FatalCheck(err)
	styleFile = paths[0]
	outputFile = paths[1]
	coverImageFile = paths[2]

	var images []string

	for _, articleName := range articles {
		sigolo.Infof("Article '%s': Start processing", articleName)

		htmlFilePath := filepath.Join(htmlOutputFolder, articleName+".html")
		if !shouldRecreateHtml(htmlFilePath, forceHtmlRecreate) {
			sigolo.Infof("Article '%s': HTML for article does already exist. Skip parsing and HTML generation.", articleName)
		} else {
			sigolo.Infof("Article '%s': Download article", articleName)
			wikiArticleDto, err := api.DownloadArticle(config.Current.WikipediaInstance, config.Current.WikipediaHost, articleName, articleCache)
			sigolo.FatalCheck(err)

			sigolo.Infof("Article '%s': Tokenize content", articleName)
			tokenizer := parser.NewTokenizer(imageCache, templateCache)
			article, err := tokenizer.Tokenize(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.OriginalTitle)
			sigolo.FatalCheck(err)
			images = append(images, article.Images...)

			sigolo.Infof("Article '%s': Download images", articleName)
			err = api.DownloadImages(article.Images, imageCache, articleCache, svgSizeToViewbox, imagesToGrayscale)
			sigolo.FatalCheck(err)

			// TODO Adjust this when additional non-epub output types are supported.
			sigolo.Infof("Article '%s': Generate HTML", articleName)
			htmlGenerator := &html.HtmlGenerator{
				ImageCacheFolder:   imageCache,
				MathCacheFolder:    mathCache,
				ArticleCacheFolder: articleCache,
				TokenMap:           article.TokenMap,
			}
			htmlFilePath, err = htmlGenerator.Generate(article, htmlOutputFolder, "../css/style.css")
			sigolo.FatalCheck(err)
		}

		sigolo.Infof("Article '%s': Finished processing", articleName)
		articleFiles = append(articleFiles, htmlFilePath)
	}

	images = util.RemoveDuplicates(images)

	sigolo.Infof("Start generating %s file", outputType)
	err = Generate(outputDriver, articleFiles, outputFile, outputType, styleFile, coverImageFile, pandocDataDir, fontFiles, metadata)
	sigolo.FatalCheck(err)

	err = os.RemoveAll(util.TempDirName)
	if err != nil {
		sigolo.Warnf("Error cleaning up '%s' directory", util.TempDirName)
	}

	absoluteOutputFile, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Infof("Successfully created %s file %s", outputType, absoluteOutputFile)
}

func Generate(outputDriver string, articleFiles []string, outputFile string, outputType string, styleFile string, coverImageFile string, pandocDataDir string, fontFiles []string, metadata project.Metadata) error {
	var err error

	switch outputDriver {
	case generator.OutputDriverPandoc:
		err = epub.Generate(articleFiles, outputFile, outputType, styleFile, coverImageFile, pandocDataDir, fontFiles, metadata)
	case generator.OutputDriverInternal:
		err = epub.GenerateWithGoLibrary(articleFiles, outputFile, coverImageFile, styleFile, fontFiles, metadata)
	default:
		err = errors.Errorf("No implementation found for output driver %s", outputDriver)
	}

	return err
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

func ensureDirectory(path string) {
	sigolo.Debugf("Ensure an empty '%s' directory exists", path)
	err := os.RemoveAll(path)
	sigolo.FatalCheck(errors.Wrapf(err, "Error removing '%s' directory", path))
	err = os.MkdirAll(path, os.ModePerm)
	sigolo.FatalCheck(errors.Wrapf(err, "Error creating '%s' directory", path))
}
