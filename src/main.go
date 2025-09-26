package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"time"
	"wiki2book/cache"
	"wiki2book/config"
	"wiki2book/generator"
	"wiki2book/generator/epub"
	"wiki2book/generator/html"
	"wiki2book/http"
	"wiki2book/image"
	"wiki2book/parser"
	"wiki2book/util"
	"wiki2book/wikipedia"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

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

	rootCmd.Version = util.VERSION
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cliOutputFile = initialize(cliLogging, cliConfig, cliConfigFile, cliDiagnosticsProfiling, cliDiagnosticsTrace, cliOutputFile)
		start = time.Now()
	}
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		end := time.Now()
		sigolo.Debugf("Start   : %s", start.Format(util.RFC1123Millis))
		sigolo.Debugf("End     : %s", end.Format(util.RFC1123Millis))
		sigolo.Debugf("Duration: %f seconds", end.Sub(start).Seconds())
	}

	rootCmd.PersistentFlags().BoolP("help", "h", false, "This help message.")
	rootCmd.PersistentFlags().StringVarP(&cliConfigFile, "config", "c", "", "The path to the overall application config. If not specified, default values are used.")
	rootCmd.PersistentFlags().StringVarP(&cliLogging, "logging", "l", "info", "Logging verbosity. Possible values: \"info\", \"debug\", \"trace\".")

	rootCmd.PersistentFlags().StringVarP(&cliOutputFile, cliOutputFileArgKey, "o", "ebook.epub", "The path to the output file.")

	rootCmd.PersistentFlags().BoolVar(&cliDiagnosticsProfiling, "diagnostics-profiling", cliDiagnosticsProfiling, "Enable profiling and write results to ./profiling.prof.")
	rootCmd.PersistentFlags().BoolVar(&cliDiagnosticsTrace, "diagnostics-trace", cliDiagnosticsTrace, "Enable tracing to analyse memory usage and write results to ./trace.out.")

	rootCmd.PersistentFlags().BoolVarP(&cliConfig.ForceRegenerateHtml, "force-regenerate-html", "r", cliConfig.ForceRegenerateHtml, "Forces wiki2book to recreate HTML files even if they exists from a previous run.")
	rootCmd.PersistentFlags().BoolVar(&cliConfig.SvgSizeToViewbox, "svg-size-to-viewbox", cliConfig.SvgSizeToViewbox, "Sets the 'width' and 'height' property of an SimpleSvgAttributes image to its viewbox width and height. This might fix wrong SVG sizes on some eBook-readers.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.OutputType, "output-type", cliConfig.OutputType, "The output file type. Possible values are: 'epub2', 'epub3', 'stats'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.OutputDriver, "output-driver", cliConfig.OutputDriver, "The method to generate the output file. Available driver: 'pandoc', 'internal' (experimental!)")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CacheDir, "cache-dir", cliConfig.CacheDir, "The directory where all cached files will be written to.")
	rootCmd.PersistentFlags().Int64Var(&cliConfig.CacheMaxSize, "cache-max-size", cliConfig.CacheMaxSize, "The maximum size of the file cache in bytes.")
	rootCmd.PersistentFlags().Int64Var(&cliConfig.CacheMaxAge, "cache-max-age", cliConfig.CacheMaxAge, "The maximum age in minutes of files in the cache. All files older than this, will be downloaded/recreated again.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CacheEvictionStrategy, "cache-eviction-strategy", cliConfig.CacheEvictionStrategy, "The strategy by which files are removed from the case when it's full. Can be: 'none', 'lru', 'largest'")
	rootCmd.PersistentFlags().StringVar(&cliConfig.StyleFile, "style-file", cliConfig.StyleFile, "The CSS file that should be used.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CoverImage, "cover-image", cliConfig.CoverImage, "A cover image for the front cover of the eBook.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateSvgToPng, "command-template-svg-to-png", cliConfig.CommandTemplateSvgToPng, "Command template to use for SVG to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateMathSvgToPng, "command-template-math-svg-to-png", cliConfig.CommandTemplateMathSvgToPng, "Command template to use for math SVG to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateImageProcessing, "command-template-image-processing", cliConfig.CommandTemplateImageProcessing, "Command template to use for math SVG to PNG conversion. Disables processing and uses original images when empty. When set, it must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplatePdfToPng, "command-template-pdf-to-png", cliConfig.CommandTemplatePdfToPng, "Command template to use for PDF to PNG conversion. Must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.CommandTemplateWebpToPng, "command-template-webp-to-png", cliConfig.CommandTemplateWebpToPng, "Command template to use for math WebP to PNG conversion. Disables conversion when empty. When set, it must contain the placeholders '{INPUT}' and '{OUTPUT}'.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.PandocExecutable, "pandoc-executable", cliConfig.PandocExecutable, "The executable name or file for pandoc.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.PandocDataDir, "pandoc-data-dir", cliConfig.PandocDataDir, "The data directory for pandoc. This enables you to override pandocs defaults for HTML and therefore EPUB generation.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.FontFiles, "font-files", cliConfig.FontFiles, "A list of font files that should be used. They are references in your style file.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredTemplates, "ignored-templates", cliConfig.IgnoredTemplates, "List of templates that should be ignored and removed from the input wikitext. The list must be in lower case.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.TrailingTemplates, "trailing-templates", cliConfig.TrailingTemplates, "List of templates that will be moved to the end of the document.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredImageParams, "ignored-image-params", cliConfig.IgnoredImageParams, "Parameters of images that should be ignored. The list must be in lower case.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.IgnoredMediaTypes, "ignored-media-types", cliConfig.IgnoredMediaTypes, "List of media types to ignore, i.e. list of file extensions.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaInstance, "wikipedia-instance", cliConfig.WikipediaInstance, "The subdomain of the Wikipedia instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaHost, "wikipedia-host", cliConfig.WikipediaHost, "The domain of the Wikipedia instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaImageHost, "wikipedia-image-host", cliConfig.WikipediaImageHost, "The domain of the Wikipedia image instance.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.WikipediaMathRestApi, "wikipedia-math-rest-api", cliConfig.WikipediaMathRestApi, "The URL to the math API of wikipedia.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.WikipediaImageArticleHosts, "wikipedia-image-article-hosts", cliConfig.WikipediaImageArticleHosts, "Hosts used to search for image article files.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.FilePrefixe, "file-prefixe", cliConfig.FilePrefixe, "A list of prefixes to detect files, e.g. in 'File:picture.jpg' the substring 'File' is the image prefix.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.AllowedLinkPrefixes, "allowed-link-prefixe", cliConfig.AllowedLinkPrefixes, "A list of prefixes that are considered links and are therefore not removed.")
	rootCmd.PersistentFlags().StringArrayVar(&cliConfig.CategoryPrefixes, "category-prefixes", cliConfig.CategoryPrefixes, "A list of category prefixes, which are technically internals links.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.MathConverter, "math-converter", cliConfig.MathConverter, "Converter turning math SVGs into PNGs.")
	rootCmd.PersistentFlags().IntVar(&cliConfig.TocDepth, "toc-depth", cliConfig.TocDepth, "Depth of the table of content. Allowed range is 0 - 6.")
	rootCmd.PersistentFlags().IntVar(&cliConfig.WorkerThreads, "worker-threads", cliConfig.WorkerThreads, "Number of threads to process the articles. Only affects projects but not single articles or the standalone mode. The value must at least be 1.")
	rootCmd.PersistentFlags().StringVar(&cliConfig.UserAgentTemplate, "user-agent-template", cliConfig.UserAgentTemplate, "Template for the user-agent used in HTTP requests.")

	projectCmd := getCommand("project [file]", "Uses a project file to create the eBook.")
	projectCmd.Args = cobra.MatchAll(cobra.ExactArgs(1))
	projectCmd.Run = func(cmd *cobra.Command, args []string) {
		sigolo.Infof("Prepare generating eBook from project")
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

	articleCmd := getCommand("article [name]", "Renders a single article into an eBook.")
	articleCmd.Args = cobra.MatchAll(cobra.ExactArgs(1))
	articleCmd.Run = func(cmd *cobra.Command, args []string) {
		sigolo.Infof("Prepare generating eBook from single article")
		config.MergeIntoCurrentConfig(cliConfig)
		generateArticleEbook(
			args[0],
			cliOutputFile,
		)
	}

	standaloneCmd := getCommand("standalone [file]", "Renders a single mediawiki file into an eBook.")
	standaloneCmd.Args = cobra.MatchAll(cobra.ExactArgs(1))
	standaloneCmd.Run = func(cmd *cobra.Command, args []string) {
		sigolo.Infof("Prepare generating eBook from standalone mediawiki file")
		config.MergeIntoCurrentConfig(cliConfig)
		generateStandaloneEbook(
			args[0],
			cliOutputFile,
		)
	}

	rootCmd.AddCommand(projectCmd, articleCmd, standaloneCmd)

	rootCmd.InitDefaultHelpCmd()
	var helpCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "help" {
			helpCommand = cmd
			break
		}
	}
	if helpCommand == nil {
		sigolo.Fatal("Help command not found")
	}
	helpCommand.Short = "Help about any command."
	helpCommand.Long = "Help provides help for any command in the application. Simply type 'help [command]' for further details on that command."

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

func getCommand(use string, shortDoc string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
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

	proj, err := config.LoadProject(projectFile)
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

	outputFile = ensurePathsAndClearTempDir(outputFile)

	config.Current.AssertFilesAndPathsExists()

	wikipediaService := wikipedia.NewWikipediaService(
		config.Current.WikipediaInstance,
		config.Current.WikipediaHost,
		config.Current.WikipediaImageArticleHosts,
		config.Current.WikipediaImageHost,
		config.Current.WikipediaMathRestApi,
		image.NewImageProcessingService(),
		http.NewDefaultHttpService(),
	)

	tokenizer := parser.NewTokenizer(wikipediaService)
	article, err := tokenizer.Tokenize(string(fileContent), title)
	sigolo.FatalCheck(err)

	err = wikipediaService.DownloadImages(article.Images)
	sigolo.FatalCheck(err)

	// TODO Adjust this when additional non-epub output types are supported.
	htmlFilePath := path.Join(cache.HtmlCacheDirName, article.Title+".html")
	if shouldRecreateHtml(htmlFilePath, config.Current.ForceRegenerateHtml) {
		htmlGenerator := &html.HtmlGenerator{
			TokenMap:         article.TokenMap,
			WikipediaService: wikipediaService,
		}
		htmlFilePath, err = htmlGenerator.Generate(article)
		sigolo.FatalCheck(err)
	}

	sigolo.Infof("Start generating %s file", config.Current.OutputType)
	metadata := config.Metadata{
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

	err = os.RemoveAll(cache.GetTempPath())
	if err != nil {
		sigolo.Warnf("Error cleaning up '%s' directory", cache.GetTempPath())
	}

	absoluteOutputFile, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Infof("Successfully created %s file %s", config.Current.OutputType, absoluteOutputFile)
}

func generateArticleEbook(articleName string, outputFile string) {
	var articles []string
	articles = append(articles, articleName)

	proj := &config.Project{}
	proj.Metadata = config.Metadata{}
	proj.OutputFile = outputFile
	proj.Articles = articles

	config.Current.Print()

	generateBookFromArticles(proj)
}

func generateBookFromArticles(project *config.Project) {
	articles := project.Articles
	metadata := project.Metadata
	outputFile := project.OutputFile

	outputFile = ensurePathsAndClearTempDir(outputFile)

	config.Current.AssertFilesAndPathsExists()

	numberOfArticles := len(articles)
	articleFiles := make([]string, numberOfArticles)

	articleChan := make(chan string, config.Current.WorkerThreads)
	sigolo.Debugf("Use %d worker threads to process the articles", config.Current.WorkerThreads)

	wikipediaService := wikipedia.NewWikipediaService(
		config.Current.WikipediaInstance,
		config.Current.WikipediaHost,
		config.Current.WikipediaImageArticleHosts,
		config.Current.WikipediaImageHost,
		config.Current.WikipediaMathRestApi,
		image.NewImageProcessingService(),
		http.NewDefaultHttpService(),
	)

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

				thisArticleFile := processArticle(articleName, articleNumber+1, numberOfArticles, wikipediaService)
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

	err = os.RemoveAll(cache.GetTempPath())
	if err != nil {
		sigolo.Warnf("Error cleaning up '%s' directory", cache.GetTempPath())
	}

	absoluteOutputFile, err := util.ToAbsolutePath(outputFile)
	sigolo.FatalCheck(err)
	sigolo.Infof("Successfully created %s file %s", config.Current.OutputType, absoluteOutputFile)
}

func processArticle(articleName string, currentArticleNumber int, totalNumberOfArticles int, wikipediaService *wikipedia.DefaultWikipediaService) string {
	sigolo.Infof("Article '%s' (%d/%d): Start processing", articleName, currentArticleNumber, totalNumberOfArticles)

	wikipediaArticleHost := fmt.Sprintf("%s.%s", config.Current.WikipediaInstance, config.Current.WikipediaHost)
	htmlFilePath := filepath.Join(cache.HtmlCacheDirName, articleName+".html")
	if !shouldRecreateHtml(htmlFilePath, config.Current.ForceRegenerateHtml) {
		sigolo.Debugf("Article '%s' (%d/%d): HTML for article does already exist. Skip parsing and HTML generation.", articleName, currentArticleNumber, totalNumberOfArticles)
	} else {

		sigolo.Debugf("Article '%s' (%d/%d): Download article", articleName, currentArticleNumber, totalNumberOfArticles)
		wikiArticleDto, err := wikipediaService.DownloadArticle(wikipediaArticleHost, articleName)
		sigolo.FatalCheck(err)

		sigolo.Debugf("Article '%s' (%d/%d): Tokenize content", articleName, currentArticleNumber, totalNumberOfArticles)
		tokenizer := parser.NewTokenizer(wikipediaService)
		article, err := tokenizer.Tokenize(wikiArticleDto.Parse.Wikitext.Content, wikiArticleDto.Parse.OriginalTitle)
		sigolo.FatalCheck(err)

		sigolo.Debugf("Article '%s' (%d/%d): Download images", articleName, currentArticleNumber, totalNumberOfArticles)
		err = wikipediaService.DownloadImages(article.Images)
		sigolo.FatalCheck(err)

		switch config.Current.OutputType {
		case config.OutputTypeEpub2:
			fallthrough
		case config.OutputTypeEpub3:
			sigolo.Debugf("Article '%s' (%d/%d): Generate HTML", articleName, currentArticleNumber, totalNumberOfArticles)
			htmlGenerator := &html.HtmlGenerator{
				TokenMap:         article.TokenMap,
				WikipediaService: wikipediaService,
			}
			htmlFilePath, err = htmlGenerator.Generate(article)
			sigolo.FatalCheck(err)
		case config.OutputTypeStats:
			sigolo.Debugf("Article '%s' (%d/%d): Generate stats", articleName, currentArticleNumber, totalNumberOfArticles)
			// TODO
		}
	}

	sigolo.Debugf("Article '%s' (%d/%d): Finished processing", articleName, currentArticleNumber, totalNumberOfArticles)

	return htmlFilePath
}

func Generate(outputDriver string, articleFiles []string, outputFile string, outputType string, styleFile string, coverImageFile string, pandocDataDir string, fontFiles []string, tocDepth int, metadata config.Metadata) error {
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

// ensurePathsAndClearTempDir ensures that the output folder for the given outputFile exists and clears up any
// temporary files in the temp files folder that might still exist from previous runs.
func ensurePathsAndClearTempDir(outputFile string) string {
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

	err = os.RemoveAll(cache.GetTempPath())
	sigolo.FatalCheck(errors.Wrapf(err, "Error removing '%s' directory", cache.GetTempPath()))

	sigolo.Debug("Ensure cache directories exist")
	util.EnsureDirectory(cache.GetTempPath())
	util.EnsureDirectory(cache.GetDirPathInCache(cache.ArticleCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.HtmlCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.ImageCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.MathCacheDirName))
	util.EnsureDirectory(cache.GetDirPathInCache(cache.TemplateCacheDirName))

	return outputFile
}
