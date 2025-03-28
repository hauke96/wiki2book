package main

import (
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

const VERSION = "v0.4.0"
const RFC1123Millis = "Mon, 02 Jan 2006 15:04:05.999 MST"

var cliConfig = config.NewDefaultConfig()

func main() {
	rootCmd := initCli()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initCli() *cobra.Command {
	var cliConfigFile = ""
	var cliLogging = ""
	var cliOutputFileArgKey = "output"
	var cliOutputFile = ""
	var cliDiagnosticsProfiling = false
	var cliDiagnosticsTrace = false
	var start time.Time

	rootCmd := &cobra.Command{
		Use:   "wiki2book",
		Short: "A CLI tool to turn one or multiple Wikipedia articles into a good-looking eBook.",
		Long:  "A CLI tool to turn one or multiple Wikipedia articles into a good-looking eBook.",
	}

	rootCmd.Version = VERSION
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cliOutputFile = initialize(cliLogging, cliConfig, cliConfigFile, cliDiagnosticsProfiling, cliDiagnosticsTrace, cliOutputFile)
		start = time.Now()
	}
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		end := time.Now()
		sigolo.Debugf("Start   : %s", start.Format(RFC1123Millis))
		sigolo.Debugf("End     : %s", end.Format(RFC1123Millis))
		sigolo.Debugf("Duration: %f seconds", end.Sub(start).Seconds())
	}

	rootCmd.PersistentFlags().StringVarP(&cliConfigFile, "config", "c", "", "The path to the overall application config. If not specified, default values are used.")
	rootCmd.PersistentFlags().StringVarP(&cliLogging, "logging", "l", "info", "Logging verbosity. Possible values: \"info\" (default), \"debug\", \"trace\".")

	rootCmd.PersistentFlags().StringVarP(&cliOutputFile, cliOutputFileArgKey, "o", "ebook.epub", "The path to the output file.")

	rootCmd.PersistentFlags().BoolVar(&cliDiagnosticsProfiling, "diagnostics-profiling", cliDiagnosticsProfiling, "Enable profiling and write results to ./profiling.prof.")
	rootCmd.PersistentFlags().BoolVar(&cliDiagnosticsTrace, "diagnostics-trace", cliDiagnosticsTrace, "Enable tracing to analyse memory usage and write results to ./trace.out.")

	rootCmd.PersistentFlags().BoolVarP(&cliConfig.ForceRegenerateHtml, "force-regenerate-html", "r", cliConfig.ForceRegenerateHtml, "Forces wiki2book to recreate HTML files even if they exists from a previous run.")
	rootCmd.PersistentFlags().BoolVar(&cliConfig.SvgSizeToViewbox, "svg-size-to-viewbox", cliConfig.SvgSizeToViewbox, "Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.OutputType, "output-type", cliConfig.OutputType, "The output file type. Possible values are: 'epub2' (default), 'epub3'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.OutputDriver, "output-driver", cliConfig.OutputDriver, "The method to generate the output file. Available driver: 'pandoc' (default), 'internal' (experimental!)")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CacheDir, "cache-dir", cliConfig.CacheDir, "The directory where all cached files will be written to.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.StyleFile, "style-file", cliConfig.StyleFile, "The CSS file that should be used.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CoverImage, "cover-image", cliConfig.CoverImage, "A cover image for the front cover of the eBook.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateSvgToPng, "command-template-svg-to-png", cliConfig.CommandTemplateSvgToPng, "Command template to use for SVG to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateMathSvgToPng, "command-template-math-svg-to-png", cliConfig.CommandTemplateMathSvgToPng, "Command template to use for math SVG to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateImageProcessing, "command-template-image-processing", cliConfig.CommandTemplateImageProcessing, "Command template to use for math SVG to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplatePdfToPng, "command-template-pdf-to-png", cliConfig.CommandTemplatePdfToPng, "The executable name or file for ImageMagick.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.PandocExecutable, "pandoc-executable", cliConfig.PandocExecutable, "The executable name or file for pandoc.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.PandocDataDir, "pandoc-data-dir", cliConfig.PandocDataDir, "The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.FontFiles, "font-files", cliConfig.FontFiles, "A list of font files that should be used. They are references in your style file.")
	rootCmd.PersistentFlags().BoolVar(&cliConfig.ConvertPdfToPng, "convert-pdf-to-png", cliConfig.ConvertPdfToPng, "Set to true in order to convert referenced PDFs into images.")
	rootCmd.PersistentFlags().BoolVar(&cliConfig.ConvertSvgToPng, "convert-svg-to-png", cliConfig.ConvertSvgToPng, "Set to true in order to convert referenced SVGs into raster images.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredTemplates, "ignored-templates", cliConfig.IgnoredTemplates, "List of templates that should be ignored and removed from the input wikitext. The list must be in lower case.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.TrailingTemplates, "trailing-templates", cliConfig.TrailingTemplates, "List of templates that will be moved to the end of the document.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredImageParams, "ignored-image-params", cliConfig.IgnoredImageParams, "Parameters of images that should be ignored. The list must be in lower case.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredMediaTypes, "ignored-media-types", cliConfig.IgnoredMediaTypes, "List of media types to ignore, i.e. list of file extensions.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaInstance, "wikipedia-instance", cliConfig.WikipediaInstance, "The subdomain of the Wikipedia instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaHost, "wikipedia-host", cliConfig.WikipediaHost, "The domain of the Wikipedia instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaImageHost, "wikipedia-image-host", cliConfig.WikipediaImageHost, "The domain of the Wikipedia image instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaMathRestApi, "wikipedia-math-rest-api", cliConfig.WikipediaMathRestApi, "The URL to the math API of wikipedia.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.WikipediaImageArticleInstances, "wikipedia-image-article-instances", cliConfig.WikipediaImageArticleInstances, "Wikipedia instances (subdomains) of the wikipedia image host where images should be searched for.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.FilePrefixe, "file-prefixe", cliConfig.FilePrefixe, "A list of prefixes to detect files, e.g. in 'File:picture.jpg' the substring 'File' is the image prefix.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.AllowedLinkPrefixes, "allowed-link-prefixe", cliConfig.AllowedLinkPrefixes, "A list of prefixes that are considered links and are therefore not removed.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.CategoryPrefixes, "category-prefixes", cliConfig.CategoryPrefixes, "A list of category prefixes, which are technically internals links.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.MathConverter, "math-converter", cliConfig.MathConverter, "Converter turning math SVGs into PNGs.")
	rootCmd.PersistentFlags().IntVar(&cliConfig.TocDepth, "toc-depth", cliConfig.TocDepth, "Depth of the table of content. Allowed range is 0 - 6.")
	rootCmd.PersistentFlags().IntVar(&cliConfig.WorkerThreads, "worker-threads", cliConfig.WorkerThreads, "Number of threads to process the articles. Only affects projects but not single articles or the standalone mode. The value must at least be 1.")

	projectCmd := getCommand("project", "Uses a project file to create the eBook.")
	projectCmd.Run = func(cmd *cobra.Command, args []string) {
		if !rootCmd.PersistentFlags().Changed(cliOutputFileArgKey) {
			// In case the output file was not specified, we don't want to use this file path but see if the project
			// file contains the output file. This is handled in the function called here.
			cliOutputFile = ""
		}

		generateProjectEbook(
			args[0],
			cliOutputFile,
		)
	}

	articleCmd := getCommand("article", "Renders a single article into an eBook.")
	articleCmd.Run = func(cmd *cobra.Command, args []string) {
		config.MergeIntoCurrentConfig(cliConfig)
		generateArticleEbook(
			args[0],
			cliOutputFile,
		)
	}

	standaloneCmd := getCommand("standalone", "Renders a single mediawiki file into an eBook.")
	standaloneCmd.Run = func(cmd *cobra.Command, args []string) {
		config.MergeIntoCurrentConfig(cliConfig)
		generateStandaloneEbook(
			args[0],
			cliOutputFile,
		)
	}

	rootCmd.AddCommand(projectCmd, articleCmd, standaloneCmd)

	return rootCmd
}

func initialize(cliLogging string, cliConfig *config.Configuration, cliConfigFile string, cliDiagnosticsProfiling bool, cliDiagnosticsTrace bool, cliOutputFile string) string {
	if strings.ToLower(cliLogging) == "debug" {
		sigolo.SetDefaultLogLevel(sigolo.LOG_DEBUG)
	} else if strings.ToLower(cliLogging) == "trace" {
		sigolo.SetDefaultLogLevel(sigolo.LOG_TRACE)
	} else if strings.ToLower(cliLogging) == "info" {
		sigolo.SetDefaultLogLevel(sigolo.LOG_INFO)
		sigolo.SetDefaultFormatFunctionAll(sigolo.LogPlain)
	} else {
		sigolo.SetDefaultFormatFunctionAll(sigolo.LogPlain)
		sigolo.Fatalf("Unknown logging level '%s'", cliLogging)
	}

	sigolo.Tracef("CLI config:\n%+v", cliConfig)

	if cliConfigFile != "" {
		err := config.LoadConfig(cliConfigFile)
		sigolo.FatalCheck(err)
	}

	if cliDiagnosticsProfiling {
		f, err := os.Create("profiling.prof")
		sigolo.FatalCheck(err)

		err = pprof.StartCPUProfile(f)
		sigolo.FatalCheck(err)
		defer pprof.StopCPUProfile()
	}

	if cliDiagnosticsTrace {
		f, err := os.Create("trace.out")
		sigolo.FatalCheck(err)

		err = trace.Start(f)
		sigolo.FatalCheck(err)
		defer trace.Stop()
	}

	cliOutputFile, err := util.ToAbsolutePath(cliOutputFile)
	sigolo.FatalCheck(err)

	return cliOutputFile
}

func getCommand(name string, shortDoc string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: shortDoc,
		Long:  shortDoc,
		Args:  cobra.ExactArgs(1),
	}

	return cmd
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
	config.MergeIntoCurrentConfig(cliConfig)

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

	err = api.DownloadImages(article.Images, imageCache, articleCache, config.Current.SvgSizeToViewbox, config.Current.ConvertPdfToPng, config.Current.ConvertSvgToPng)
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
		config.Current.TocDepth,
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

	articleChan := make(chan string, config.Current.WorkerThreads)
	sigolo.Debugf("Use %d worker threads to process the articles", config.Current.WorkerThreads)

	// Create a wait-group that is zero when all threads are done
	threadPoolWaitGroup := &sync.WaitGroup{}
	threadPoolWaitGroup.Add(config.Current.WorkerThreads)

	// Start threads which pick an article from the channel to work on
	for i := 0; i < config.Current.WorkerThreads; i++ {
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
		config.Current.TocDepth,
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
		err = api.DownloadImages(article.Images, imageCache, articleCache, config.Current.SvgSizeToViewbox, config.Current.ConvertPdfToPng, config.Current.ConvertSvgToPng)
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
		outputFile = path.Join(outputFile, "ebook.epub")
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
