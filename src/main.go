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
	config.Configuration

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

	var err error
	cli.OutputFile, err = util.ToAbsolutePath(cli.OutputFile)
	sigolo.FatalCheck(err)

	start := time.Now()

	switch ctx.Command() {
	case "standalone <file>":
		mergeConfigIntoMainConfig(&cli.Configuration)
		generateStandaloneEbook(
			cli.Standalone.File,
			cli.OutputFile,
		)
	case "project <project-file>":
		generateProjectEbook(
			cli.Project.ProjectFile,
			cli.OutputFile,
		)
	case "article <article-name>":
		mergeConfigIntoMainConfig(&cli.Configuration)
		generateArticleEbook(
			cli.Article.ArticleName,
			cli.OutputFile,
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

func mergeConfigIntoMainConfig(c *config.Configuration) {
	if c.ForceRegenerateHtml {
		sigolo.Tracef("Override outputType from project file with %s", c.OutputType)
		config.Current.ForceRegenerateHtml = c.ForceRegenerateHtml
	}
	if c.SvgSizeToViewbox {
		sigolo.Tracef("Override svgSizeToViewbox from project file with %v", c.SvgSizeToViewbox)
		config.Current.SvgSizeToViewbox = c.SvgSizeToViewbox
	}
	if c.OutputType != "" {
		sigolo.Tracef("Override outputType from project file with %s", c.OutputType)
		config.Current.OutputType = c.OutputType
	}
	if c.OutputDriver != "" {
		sigolo.Tracef("Override OutputDriver from project file with %s", c.OutputDriver)
		config.Current.OutputDriver = c.OutputDriver
	}
	if c.CacheDir != "" {
		absolutePath, err := util.ToAbsolutePath(c.CacheDir)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override CacheDir from project file with %s", absolutePath)
		config.Current.CacheDir = absolutePath
	}
	if c.StyleFile != "" {
		absolutePath, err := util.ToAbsolutePath(c.StyleFile)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override StyleFile from project file with %s", absolutePath)
		config.Current.StyleFile = absolutePath
	}
	if c.CoverImage != "" {
		absolutePath, err := util.ToAbsolutePath(c.CoverImage)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override CoverImage from project file with %s", absolutePath)
		config.Current.CoverImage = absolutePath
	}
	if c.RsvgConvertExecutable != "" {
		absolutePath, err := util.ToAbsolutePath(c.RsvgConvertExecutable)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override RsvgConvertExecutable from project file with %s", c.RsvgConvertExecutable)
		config.Current.RsvgConvertExecutable = absolutePath
	}
	if c.RsvgMathStylesheet != "" {
		absolutePath, err := util.ToAbsolutePath(c.RsvgMathStylesheet)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override RsvgMathStylesheet from project file with %s", c.RsvgMathStylesheet)
		config.Current.RsvgMathStylesheet = absolutePath
	}
	if c.ImageMagickExecutable != "" {
		absolutePath, err := util.ToAbsolutePath(c.ImageMagickExecutable)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override ImageMagickExecutable from project file with %s", c.ImageMagickExecutable)
		config.Current.ImageMagickExecutable = absolutePath
	}
	if c.PandocExecutable != "" {
		absolutePath, err := util.ToAbsolutePath(c.PandocExecutable)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override PandocExecutable from project file with %s", c.PandocExecutable)
		config.Current.PandocExecutable = absolutePath
	}
	if c.PandocDataDir != "" {
		absolutePath, err := util.ToAbsolutePath(c.PandocDataDir)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override PandocDataDir from project file with %s", absolutePath)
		config.Current.PandocDataDir = absolutePath
	}
	if c.FontFiles != nil {
		absolutePaths, err := util.ToAbsolutePaths(c.FontFiles...)
		sigolo.FatalCheck(err)
		sigolo.Tracef("Override FontFiles from project file with %v", c.SvgSizeToViewbox)
		config.Current.FontFiles = absolutePaths
	}
	if c.ImagesToGrayscale {
		sigolo.Tracef("Override ImagesToGrayscale from project file with %v", c.ImagesToGrayscale)
		config.Current.ImagesToGrayscale = c.ImagesToGrayscale
	}
	if c.ConvertPDFsToImages {
		sigolo.Tracef("Override ConvertPDFsToImages from project file with %v", c.ConvertPDFsToImages)
		config.Current.ConvertPDFsToImages = c.ConvertPDFsToImages
	}
	if c.IgnoredTemplates != nil {
		sigolo.Tracef("Override IgnoredTemplates from project file with %v", c.IgnoredTemplates)
		config.Current.IgnoredTemplates = c.IgnoredTemplates
	}
	if c.TrailingTemplates != nil {
		sigolo.Tracef("Override TrailingTemplates from project file with %v", c.TrailingTemplates)
		config.Current.TrailingTemplates = c.TrailingTemplates
	}
	if c.IgnoredImageParams != nil {
		sigolo.Tracef("Override IgnoredImageParams from project file with %v", c.IgnoredImageParams)
		config.Current.IgnoredImageParams = c.IgnoredImageParams
	}
	if c.IgnoredMediaTypes != nil {
		sigolo.Tracef("Override IgnoredMediaTypes from project file with %v", c.IgnoredMediaTypes)
		config.Current.IgnoredMediaTypes = c.IgnoredMediaTypes
	}
	if c.WikipediaInstance != "" {
		sigolo.Tracef("Override WikipediaInstance from project file with %s", c.WikipediaInstance)
		config.Current.WikipediaInstance = c.WikipediaInstance
	}
	if c.WikipediaHost != "" {
		sigolo.Tracef("Override WikipediaHost from project file with %s", c.WikipediaHost)
		config.Current.WikipediaHost = c.WikipediaHost
	}
	if c.WikipediaImageHost != "" {
		sigolo.Tracef("Override WikipediaImageHost from project file with %s", c.WikipediaImageHost)
		config.Current.WikipediaImageHost = c.WikipediaImageHost
	}
	if c.WikipediaMathRestApi != "" {
		sigolo.Tracef("Override WikipediaMathRestApi from project file with %s", c.WikipediaMathRestApi)
		config.Current.WikipediaMathRestApi = c.WikipediaMathRestApi
	}
	if c.WikipediaImageArticleInstances != nil {
		sigolo.Tracef("Override WikipediaImageArticleInstances from project file with %v", c.WikipediaImageArticleInstances)
		config.Current.WikipediaImageArticleInstances = c.WikipediaImageArticleInstances
	}
	if c.FilePrefixe != nil {
		sigolo.Tracef("Override FilePrefixe from project file with %v", c.FilePrefixe)
		config.Current.FilePrefixe = c.FilePrefixe
	}
	if c.AllowedLinkPrefixes != nil {
		sigolo.Tracef("Override AllowedLinkPrefixes from project file with %v", c.AllowedLinkPrefixes)
		config.Current.AllowedLinkPrefixes = c.AllowedLinkPrefixes
	}
	if c.CategoryPrefixes != nil {
		sigolo.Tracef("Override CategoryPrefixes from project file with %v", c.CategoryPrefixes)
		config.Current.CategoryPrefixes = c.CategoryPrefixes
	}
	if c.MathConverter != "" {
		sigolo.Tracef("Override MathConverter from project file with %s", c.MathConverter)
		config.Current.MathConverter = c.MathConverter
	}

	config.Current.MakePathsAbsoluteToWorkingDir()

	config.Current.AssertValidity()
}

func generateProjectEbook(projectFile string, outputFile string) {
	var err error

	sigolo.Infof("Use project file: %s", projectFile)

	sigolo.Debug("Turn paths from CLI arguments into absolute paths before going into the project file directory")
	if outputFile != "" {
		outputFile, err = util.ToAbsolutePath(outputFile)
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

	mergeConfigIntoMainConfig(&proj.Configuration)
	mergeConfigIntoMainConfig(&cli.Configuration)

	generateBookFromArticles(proj)
}

func generateStandaloneEbook(inputFile string, outputFile string) {
	fileContent, err := os.ReadFile(inputFile)
	sigolo.FatalCheck(err)

	_, inputFileName := path.Split(inputFile)
	title := strings.Split(inputFileName, ".")[0]

	outputFile, imageCache, mathCache, templateCache, articleCache, htmlOutputFolder, relativeStyleFile := ensurePathsAndGoIntoCacheFolder(outputFile)

	config.Current.AssertFilesAndPathsExists()

	tokenizer := parser.NewTokenizer(imageCache, templateCache)
	article, err := tokenizer.Tokenize(string(fileContent), title)
	sigolo.FatalCheck(err)

	err = api.DownloadImages(article.Images, imageCache, articleCache, config.Current.SvgSizeToViewbox, config.Current.ImagesToGrayscale, config.Current.ConvertPDFsToImages)
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

func generateArticleEbook(articleName string, outputFile string) {
	var articles []string
	articles = append(articles, articleName)

	proj := &project.Project{}
	proj.Metadata = project.Metadata{}
	proj.OutputFile = outputFile
	proj.Articles = articles

	generateBookFromArticles(proj)
}

func generateBookFromArticles(project *project.Project) {
	var articleFiles []string

	articles := project.Articles
	metadata := project.Metadata
	outputFile := project.OutputFile

	outputFile, imageCache, mathCache, templateCache, articleCache, htmlOutputFolder, relativeStyleFile := ensurePathsAndGoIntoCacheFolder(outputFile)

	config.Current.AssertFilesAndPathsExists()

	var images []string

	numberOfArticles := len(articles)
	for i, articleName := range articles {
		sigolo.Infof("Article '%s' (%d/%d): Start processing", articleName, i, numberOfArticles)

		htmlFilePath := filepath.Join(htmlOutputFolder, articleName+".html")
		if !shouldRecreateHtml(htmlFilePath, config.Current.ForceRegenerateHtml) {
			sigolo.Infof("Article '%s' (%d/%d): HTML for article does already exist. Skip parsing and HTML generation.", articleName, i, numberOfArticles)
		} else {
			sigolo.Infof("Article '%s' (%d/%d): Download article", articleName, i, numberOfArticles)
			wikiArticleDto, err := api.DownloadArticle(config.Current.WikipediaInstance, config.Current.WikipediaHost, articleName, articleCache)
			sigolo.FatalCheck(err)

			sigolo.Infof("Article '%s' (%d/%d): Tokenize content", articleName, i, numberOfArticles)
			tokenizer := parser.NewTokenizer(imageCache, templateCache)
			article, err := tokenizer.Tokenize(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.OriginalTitle)
			sigolo.FatalCheck(err)
			images = append(images, article.Images...)

			sigolo.Infof("Article '%s' (%d/%d): Download images", articleName, i, numberOfArticles)
			err = api.DownloadImages(article.Images, imageCache, articleCache, config.Current.SvgSizeToViewbox, config.Current.ImagesToGrayscale, config.Current.ConvertPDFsToImages)
			sigolo.FatalCheck(err)

			// TODO Adjust this when additional non-epub output types are supported.
			sigolo.Infof("Article '%s' (%d/%d): Generate HTML", articleName, i, numberOfArticles)
			htmlGenerator := &html.HtmlGenerator{
				ImageCacheFolder:   imageCache,
				MathCacheFolder:    mathCache,
				ArticleCacheFolder: articleCache,
				TokenMap:           article.TokenMap,
			}
			htmlFilePath, err = htmlGenerator.Generate(article, htmlOutputFolder, relativeStyleFile)
			sigolo.FatalCheck(err)
		}

		sigolo.Infof("Article '%s' (%d/%d): Finished processing", articleName, i, numberOfArticles)
		articleFiles = append(articleFiles, htmlFilePath)
	}

	images = util.RemoveDuplicates(images)

	sigolo.Infof("Start generating %s file", config.Current.OutputType)
	err := Generate(config.Current.OutputDriver, articleFiles, outputFile, config.Current.OutputType, config.Current.StyleFile, config.Current.CoverImage, config.Current.PandocDataDir, config.Current.FontFiles, metadata)
	sigolo.FatalCheck(err)

	err = os.RemoveAll(util.TempDirName)
	if err != nil {
		sigolo.Warnf("Error cleaning up '%s' directory", util.TempDirName)
	}

	absoluteOutputFile, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Infof("Successfully created %s file %s", config.Current.OutputType, absoluteOutputFile)
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

func ensurePathsAndGoIntoCacheFolder(outputFile string) (string, string, string, string, string, string, string) {
	imageCache := "images"
	mathCache := "math"
	templateCache := "templates"
	articleCache := "articles"
	htmlOutputFolder := "html"

	var file *os.File
	if _, err := os.Stat(outputFile); err != nil {
		// Output file does not exist
		sigolo.Debugf("Output file %s does not exists, I'll create it", outputFile)

		err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
		sigolo.FatalCheck(errors.Wrapf(err, "Error creating output file %s", outputFile))

		file, err = os.Create(outputFile)
		sigolo.FatalCheck(err)
	} else {
		file, err = os.Open(outputFile)
		sigolo.FatalCheck(err)
	}

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
	util.EnsureDirectory(config.Current.CacheDir)
	err = os.Chdir(config.Current.CacheDir)
	sigolo.FatalCheck(err)

	util.EnsureDirectory(util.TempDirName)

	// Make all relevant paths relative again. This ensures that the locations within the HTML files are independent
	// of the systems' directory structure.
	relativeStyleFile, err := util.ToRelativePath(config.Current.StyleFile)
	sigolo.FatalCheck(err)

	return outputFile, imageCache, mathCache, templateCache, articleCache, htmlOutputFolder, relativeStyleFile
}
