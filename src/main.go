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
	"sync"
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

const VERSION = "v0.3.0"
const RFC1123Millis = "Mon, 02 Jan 2006 15:04:05.999 MST"

type Cli struct {
	Version VersionFlag `help:"Print version information and quit" name:"version" short:"v"`

	Config  string `help:"The path to the overall application config. If not specified, default values are used." type:"existingfile" short:"c" placeholder:"<file>"`
	Logging string `help:"Logging verbosity. Possible values: \"info\" (default), \"debug\", \"trace\"." short:"l" default:"info"`

	DiagnosticsProfiling bool `help:"Enable profiling and write results to ./profiling.prof."`
	DiagnosticsTrace     bool `help:"Enable tracing to analyse memory usage and write results to ./trace.out."`

	OutputFile string `help:"The path to the output file." short:"o" placeholder:"<file>"`

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
		config.MergeIntoCurrentConfig(&cli.Configuration)
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
		config.MergeIntoCurrentConfig(&cli.Configuration)
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

func generateProjectEbook(projectFile string, outputFile string) {
	var err error

	sigolo.Infof("Use project file: %s", projectFile)

	directory, projectFile := filepath.Split(projectFile)
	if directory != "" {
		sigolo.Debugf("Go into folder %s", directory)
		err = os.Chdir(directory)
		sigolo.FatalCheck(err)
	}

	proj, err := project.LoadProject(projectFile)
	sigolo.FatalCheck(err)

	if outputFile != "" {
		sigolo.Tracef("Project has no output file set, so I'll use %s", outputFile)
		proj.OutputFile = outputFile
	}

	sigolo.Debug("Turn output file path into absolute path")
	proj.OutputFile, err = util.ToAbsolutePath(proj.OutputFile)
	sigolo.FatalCheck(err)

	config.MergeIntoCurrentConfig(&proj.Configuration)
	config.MergeIntoCurrentConfig(&cli.Configuration)

	config.Current.Print()
	proj.Print()

	generateBookFromArticles(proj)
}

func generateStandaloneEbook(inputFile string, outputFile string) {
	config.Current.Print()

	fileContent, err := os.ReadFile(inputFile)
	sigolo.FatalCheck(err)

	_, inputFileName := path.Split(inputFile)
	title := strings.Split(inputFileName, ".")[0]

	outputFile, imageCache, mathCache, templateCache, articleCache, htmlOutputFolder, relativeStyleFile := ensurePathsAndGoIntoCacheFolder(outputFile)

	config.Current.AssertFilesAndPathsExists()

	tokenizer := parser.NewTokenizer(imageCache, templateCache)
	article, err := tokenizer.Tokenize(string(fileContent), title)
	sigolo.FatalCheck(err)

	err = api.DownloadImages(article.Images, imageCache, articleCache, config.Current.SvgSizeToViewbox, config.Current.ImagesToGrayscale, config.Current.ConvertPDFsToImages, config.Current.ConvertSvgToPng)
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

	err = Generate(
		config.Current.OutputDriver,
		[]string{htmlFilePath},
		outputFile,
		config.Current.OutputType,
		config.Current.StyleFile,
		config.Current.CoverImage,
		config.Current.PandocDataDir,
		config.Current.FontFiles,
		*config.Current.TocDepth,
		metadata,
	)
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

	config.Current.Print()

	generateBookFromArticles(proj)
}

func generateBookFromArticles(project *project.Project) {
	articles := project.Articles
	metadata := project.Metadata
	outputFile := project.OutputFile

	outputFile, imageCache, mathCache, templateCache, articleCache, htmlOutputFolder, relativeStyleFile := ensurePathsAndGoIntoCacheFolder(outputFile)

	config.Current.AssertFilesAndPathsExists()

	numberOfArticles := len(articles)
	articleFiles := make([]string, numberOfArticles)

	articleChan := make(chan string, *config.Current.WorkerThreads)
	sigolo.Debugf("Use %d worker threads to process the articles", *config.Current.WorkerThreads)

	// Create a wait-group that is zero when all threads are done
	threadPoolWaitGroup := &sync.WaitGroup{}
	threadPoolWaitGroup.Add(*config.Current.WorkerThreads)

	// Start threads which pick an article from the channel to work on
	for i := 0; i < *config.Current.WorkerThreads; i++ {
		sigolo.Debugf("Start worker thread %d", i)

		go func(threadNumber int) {
			for articleName := range articleChan {
				sigolo.Debugf("Processing article %s on thread %d", articleName, threadNumber)

				articleNumber := 0
				for ; articleNumber < len(articles); articleNumber++ {
					if articleName == articles[articleNumber] {
						break
					}
				}

				thisArticleFile := processArticle(articleName, articleNumber, numberOfArticles, htmlOutputFolder, articleCache, imageCache, templateCache, mathCache, relativeStyleFile)
				articleFiles[articleNumber] = thisArticleFile
			}

			// This thread will close, so mark it as done in the wait-group
			sigolo.Debugf("Mark thread %d as done", threadNumber)
			threadPoolWaitGroup.Done()
		}(i)
	}

	for _, articleName := range articles {
		sigolo.Debugf("Pass article %s to worker threads", articleName)
		articleChan <- articleName
	}
	close(articleChan)

	// Wait for threads to be done with processing
	sigolo.Debugf("Wait for threads to finish processing articles ...")
	threadPoolWaitGroup.Wait()
	sigolo.Debugf("Worker threads are done processing articles")

	sigolo.Infof("Start generating %s file", config.Current.OutputType)
	err := Generate(
		config.Current.OutputDriver,
		articleFiles,
		outputFile,
		config.Current.OutputType,
		config.Current.StyleFile,
		config.Current.CoverImage,
		config.Current.PandocDataDir,
		config.Current.FontFiles,
		*config.Current.TocDepth,
		metadata,
	)
	sigolo.FatalCheck(err)

	err = os.RemoveAll(util.TempDirName)
	if err != nil {
		sigolo.Warnf("Error cleaning up '%s' directory", util.TempDirName)
	}

	absoluteOutputFile, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Infof("Successfully created %s file %s", config.Current.OutputType, absoluteOutputFile)
}

func processArticle(articleName string, currentArticleNumber int, totalNumberOfArticles int, htmlOutputFolder string, articleCache string, imageCache string, templateCache string, mathCache string, relativeStyleFile string) string {
	sigolo.Infof("Article '%s' (%d/%d): Start processing", articleName, currentArticleNumber, totalNumberOfArticles)

	htmlFilePath := filepath.Join(htmlOutputFolder, articleName+".html")
	if !shouldRecreateHtml(htmlFilePath, config.Current.ForceRegenerateHtml) {
		sigolo.Infof("Article '%s' (%d/%d): HTML for article does already exist. Skip parsing and HTML generation.", articleName, currentArticleNumber, totalNumberOfArticles)
	} else {
		sigolo.Infof("Article '%s' (%d/%d): Download article", articleName, currentArticleNumber, totalNumberOfArticles)
		wikiArticleDto, err := api.DownloadArticle(config.Current.WikipediaInstance, config.Current.WikipediaHost, articleName, articleCache)
		sigolo.FatalCheck(err)

		sigolo.Infof("Article '%s' (%d/%d): Tokenize content", articleName, currentArticleNumber, totalNumberOfArticles)
		tokenizer := parser.NewTokenizer(imageCache, templateCache)
		article, err := tokenizer.Tokenize(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.OriginalTitle)
		sigolo.FatalCheck(err)

		sigolo.Infof("Article '%s' (%d/%d): Download images", articleName, currentArticleNumber, totalNumberOfArticles)
		err = api.DownloadImages(article.Images, imageCache, articleCache, config.Current.SvgSizeToViewbox, config.Current.ImagesToGrayscale, config.Current.ConvertPDFsToImages, config.Current.ConvertSvgToPng)
		sigolo.FatalCheck(err)

		// TODO Adjust this when additional non-epub output types are supported.
		sigolo.Infof("Article '%s' (%d/%d): Generate HTML", articleName, currentArticleNumber, totalNumberOfArticles)
		htmlGenerator := &html.HtmlGenerator{
			ImageCacheFolder:   imageCache,
			MathCacheFolder:    mathCache,
			ArticleCacheFolder: articleCache,
			TokenMap:           article.TokenMap,
		}
		htmlFilePath, err = htmlGenerator.Generate(article, htmlOutputFolder, relativeStyleFile)
		sigolo.FatalCheck(err)
	}

	sigolo.Infof("Article '%s' (%d/%d): Finished processing", articleName, currentArticleNumber, totalNumberOfArticles)

	return htmlFilePath
}

func Generate(outputDriver string, articleFiles []string, outputFile string, outputType string, styleFile string, coverImageFile string, pandocDataDir string, fontFiles []string, tocDepth int, metadata project.Metadata) error {
	var err error

	switch outputDriver {
	case generator.OutputDriverPandoc:
		err = epub.Generate(articleFiles, outputFile, outputType, styleFile, coverImageFile, pandocDataDir, fontFiles, tocDepth, metadata)
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

	err = os.RemoveAll(util.TempDirName)
	sigolo.FatalCheck(errors.Wrapf(err, "Error removing '%s' directory", util.TempDirName))
	util.EnsureDirectory(util.TempDirName)

	// Make all relevant paths relative again. This ensures that the locations within the HTML files are independent
	// of the systems' directory structure.
	relativeStyleFile, err := util.ToRelativePath(config.Current.StyleFile)
	sigolo.FatalCheck(err)

	return outputFile, imageCache, mathCache, templateCache, articleCache, htmlOutputFolder, relativeStyleFile
}
